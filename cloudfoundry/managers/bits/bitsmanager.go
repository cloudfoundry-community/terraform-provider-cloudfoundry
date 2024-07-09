package bits

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudfoundry/go-cfclient/v3/client"
	"github.com/cloudfoundry/go-cfclient/v3/resource"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
)

// Manage upload bits like app and buildpack in full stream
type BitsManager struct {
	client *client.Client
}

type job struct {
	Entity struct {
		GUID string `json:"guid"`
	} `json:"entity"`
}

type ZipFile struct {
	r        io.ReadCloser
	baseName string
	filesize int64
}

// NewBitsManager -
func NewBitsManager(client *client.Client) *BitsManager {
	return &BitsManager{
		client: client,
	}
}

// CopyApp - Copy one app to another by using only api
func (m BitsManager) CopyApp(origAppGuid string, newAppGuid string) error {
	// Construct the request URL and data
	path := fmt.Sprintf("/v2/apps/%s/copy_bits", newAppGuid)
	data := strings.NewReader(fmt.Sprintf(`{"source_app_guid":"%s"}`, origAppGuid))

	// Create a new HTTP request
	req, err := http.NewRequest("POST", m.client.ApiURL(path), data)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the HTTP request
	resp, err := m.client.ExecuteRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var j job
	err = json.NewDecoder(resp.Body).Decode(&j)
	if err != nil {
		return err

	}
	// poll job

	opts := client.NewPollingOptions()
	jobGUID := ccv2.Job{GUID: j.Entity.GUID}
	jobGUID_str := fmt.Sprintf("%+v", jobGUID)
	m.client.Jobs.PollComplete(context.Background(), jobGUID_str, opts)
	return nil
}

// UploadBuildpack - Upload buildpack in full stream by setting an uri path
// uri path can be:
// - file:///path/to/my/buildpack.zip
// - http(s)://awesome.buildpack.com/my-buildpack.zip
func (m BitsManager) UploadBuildpack(buildpackGUID string, bpPath string) error {
	zipFile, err := m.RetrieveZip(bpPath)
	if err != nil {
		return err
	}
	zipFileReader := zipFile.r
	baseName := zipFile.baseName
	fileSize := zipFile.filesize
	defer zipFileReader.Close()
	apiURL := fmt.Sprintf("/v2/buildpacks/%s/bits", buildpackGUID)

	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)
	go func() {
		var err error
		defer w.Close()

		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="buildpack"; filename="%s"`, baseName))
		h.Set("Content-Type", "application/zip")
		h.Set("Content-Length", fmt.Sprintf("%d", fileSize))
		h.Set("Content-Transfer-Encoding", "binary")

		part, err := mpw.CreatePart(h)
		if err != nil {
			mpw.Close()
			panic(err)
		}
		if _, err = io.Copy(part, zipFileReader); err != nil {
			mpw.Close()
			panic(err)
		}
		mpw.Close()
	}()

	req, err := http.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return err
	}
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", mpw.Boundary())
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = m.predictPartBuildpack(baseName, fileSize, mpw.Boundary())
	req.Body = r

	resp, err := m.client.ExecuteRequest(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return nil
}

// GetAppEnvironmentVariables - Get app environment variables
func (m BitsManager) GetAppEnvironmentVariables(appGUID string) (map[string]interface{}, error) {
	apiURL := fmt.Sprintf("/v3/apps/%s/environment_variables", appGUID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	var responseBody = struct {
		Var map[string]interface{} `json:"var"`
	}{}
	resp, err := m.client.ExecuteRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&responseBody)
	if err != nil {
		return nil, err
	}
	return responseBody.Var, nil
}

// SetAppEnvironmentVariables - Update application's environment variable
// DOES NOT remove empty string variables
func (m BitsManager) SetAppEnvironmentVariables(appGUID string, env map[string]interface{}) error {
	apiURL := fmt.Sprintf("/v3/apps/%s/environment_variables", appGUID)

	req, err := http.NewRequest("PATCH", apiURL, nil)
	if err != nil {
		return err
	}
	var requestBody = struct {
		Var map[string]interface{} `json:"var"`
	}{}
	requestBody.Var = env
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(requestBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Body = ioutil.NopCloser(body)

	resp, err := m.client.ExecuteRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// UpdateAppEnvVars updates environment variables for an application identified by appGUID
func (m *BitsManager) UpdateAppEnvVars(appGUID string, env map[string]interface{}) (map[string]interface{}, error) {
	// Creating the environment variables request body
	envVars := make(map[string]string)
	for key, value := range env {
		if strValue, ok := value.(string); ok {
			envVars[key] = strValue
		} else {
			return nil, fmt.Errorf("environment variable values must be strings")
		}
	}

	requestBody := &resource.EnvVarGroupUpdate{
		Var: envVars,
	}

	// Context for the request
	ctx := context.Background()

	// Sending the request to update the environment variables
	req, err := m.client.EnvVarGroups.Update(ctx, appGUID, requestBody)
	if err != nil {
		return nil, err
	}

	// Convert map[string]string to map[string]interface{}
	updatedEnvVars := make(map[string]interface{})
	for key, value := range req.Var {
		updatedEnvVars[key] = value
	}
	// Return the updated environment variables
	return updatedEnvVars, err
}

// UploadApp - Upload a zip file containing app code to cloud foundry in full stream
func (m BitsManager) UploadApp(appGUID string, path string) error {
	zipFile, err := m.RetrieveZip(path)
	if err != nil {
		return err
	}
	zipFileReader := zipFile.r
	fileSize := zipFile.filesize
	defer zipFileReader.Close()
	apiURL := fmt.Sprintf("/v2/apps/%s/bits", appGUID)
	r, w := io.Pipe()
	mpw := multipart.NewWriter(w)
	go func() {
		var err error
		defer w.Close()
		part, err := mpw.CreateFormField("resources")
		if err != nil {
			mpw.Close()
			panic(err)
		}
		_, err = io.Copy(part, bytes.NewBuffer([]byte("[]")))
		if err != nil {
			mpw.Close()
			panic(err)
		}
		h := make(textproto.MIMEHeader)
		h.Set("Content-Disposition", `form-data; name="application"; filename="application.zip"`)
		h.Set("Content-Type", "application/zip")
		h.Set("Content-Length", fmt.Sprintf("%d", fileSize))
		h.Set("Content-Transfer-Encoding", "binary")

		part, err = mpw.CreatePart(h)
		if err != nil {
			mpw.Close()
			panic(err)
		}
		if _, err = io.Copy(part, zipFileReader); err != nil {
			mpw.Close()
			panic(err)
		}
		mpw.Close()
	}()
	req, err := http.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return err
	}
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", mpw.Boundary())
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = m.predictPartApp(fileSize, mpw.Boundary())
	req.Body = r

	resp, err := m.client.ExecuteRequest(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return nil
}

func (m BitsManager) predictPartApp(filesize int64, boundary string) int64 {
	buf := new(bytes.Buffer)
	mpw := multipart.NewWriter(buf)

	err := mpw.SetBoundary(boundary)
	if err != nil {
		panic(err)
	}
	part, err := mpw.CreateFormField("resources")
	if err != nil {
		mpw.Close()
		panic(err)
	}
	_, err = io.Copy(part, bytes.NewBuffer([]byte("[]")))
	if err != nil {
		mpw.Close()
		panic(err)
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="application"; filename="application.zip"`)
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", filesize))
	h.Set("Content-Transfer-Encoding", "binary")

	_, err = mpw.CreatePart(h)
	if err != nil {
		mpw.Close()
		panic(err)
	}
	mpw.Close()
	b, _ := ioutil.ReadAll(buf)
	return int64(len(b)) + filesize
}

func (m BitsManager) predictPartBuildpack(filename string, filesize int64, boundary string) int64 {
	buf := new(bytes.Buffer)
	mpw := multipart.NewWriter(buf)

	err := mpw.SetBoundary(boundary)
	if err != nil {
		panic(err)
	}
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="buildpack"; filename="%s"`, filename))
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", filesize))
	h.Set("Content-Transfer-Encoding", "binary")

	_, err = mpw.CreatePart(h)
	if err != nil {
		mpw.Close()
		panic(err)
	}
	mpw.Close()
	b, _ := ioutil.ReadAll(buf)
	return int64(len(b)) + filesize
}

func (m BitsManager) RetrieveZip(path string) (ZipFile, error) {
	path = strings.TrimPrefix(path, "file://")
	baseName := filepath.Base(path)
	if strings.HasPrefix(path, "http") {
		resp, err := m.client.HTTPClient().Get(path)
		if err != nil {
			return ZipFile{}, err
		}
		fileSize := resp.ContentLength
		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return ZipFile{}, fmt.Errorf("Failed to download file %s (%s)", path, strings.Trim(resp.Status, " "))
		}
		_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
		if err == nil {
			var ok bool
			if _, ok = params["filename"]; ok {
				baseName = params["filename"]
			}
		}
		if fileSize > 0 {
			return ZipFile{
				r:        resp.Body,
				baseName: baseName,
				filesize: fileSize,
			}, nil
		} else {
			tempFile, err := ioutil.TempFile("", "")
			if err != nil {
				return ZipFile{}, err
			}
			defer os.Remove(tempFile.Name())
			fileSize, err := io.Copy(tempFile, resp.Body)
			if err != nil {
				return ZipFile{}, err
			}
			_, err = tempFile.Seek(0, 0)
			if err != nil {
				return ZipFile{}, err
			}
			return ZipFile{
				r:        tempFile,
				baseName: baseName,
				filesize: fileSize,
			}, nil
		}
	}
	f, err := os.Open(path)
	if err != nil {
		return ZipFile{}, err
	}
	stat, err := f.Stat()
	if err != nil {
		return ZipFile{}, err
	}
	return ZipFile{
		r:        f,
		baseName: baseName,
		filesize: stat.Size(),
	}, nil
}

// v3

// CreateDockerPackage creates a package from a docker image
func (m *BitsManager) CreateDockerPackage(appGUID string, dockerImage string, dockerUsername string, dockerPassword string) (*resource.Package, error) {
	// Create a Docker package request body using the provided function
	pkgCreateRequest := resource.NewDockerPackageCreate(appGUID, dockerImage, dockerUsername, dockerPassword)

	// Context for the request
	ctx := context.Background()

	// Send the request to create the package
	createdPackage, err := m.client.Packages.Create(ctx, pkgCreateRequest)
	if err != nil {
		return nil, err
	}

	return createdPackage, nil
}

// UploadBits uploads the application bits to the given package
func (m *BitsManager) UploadBits(pkg resource.Package, bitsPath string) (*resource.Package, error) {
	// Open the bits file
	file, err := os.Open(bitsPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Context for the request
	ctx := context.Background()

	// Assuming cfclient has a method for uploading bits to a package
	updatedPackage, err := m.client.Packages.Upload(ctx, pkg.GUID, file)
	if err != nil {
		return nil, err
	}

	return updatedPackage, nil
}

// CreateAndUploadBitsPackage creates a new package and uploads bits to the application
func (m *BitsManager) CreateAndUploadBitsPackage(appGUID string, path string, stageTimeout time.Duration) (*resource.Package, ccv3.Warnings, error) {
	// Create a new package for the application
	pkg, warnings, err := m.CreateBitsPackageByApplication(appGUID)
	if err != nil {
		return nil, warnings, err
	}

	// If path starts with "http", retrieve the zip file from remote artifactory
	if strings.HasPrefix(path, "http") {
		zipFile, err := m.RetrieveZip(path)
		if err != nil {
			return nil, warnings, err
		}

		tempDir, err := ioutil.TempDir("", "temp-")
		if err != nil {
			return nil, warnings, err
		}
		defer func() {
			if err := os.RemoveAll(tempDir); err != nil {
				log.Printf("Error removing temp dir %s: %v", tempDir, err)
			}
		}()

		outputFile, err := os.Create(filepath.Join(tempDir, zipFile.baseName))
		if err != nil {
			return nil, warnings, err
		}
		defer outputFile.Close()

		_, err = io.Copy(outputFile, zipFile.r)
		if err != nil {
			return nil, warnings, err
		}

		if err := outputFile.Sync(); err != nil {
			return nil, warnings, err
		}

		path = outputFile.Name()
	}

	// Upload the bits to the package
	uploadedPackage, err := m.UploadBits(*pkg, path)
	if err != nil {
		return nil, warnings, err
	}

	// Poll the package status until it's ready or failed
	ctx := context.Background()
	err = common.PollingWithTimeout(func() (bool, error) {
		ccPkg, err := m.client.Packages.Get(ctx, pkg.GUID)
		if err != nil {
			return true, err
		}

		switch ccPkg.State {
		case resource.PackageStateReady:
			return true, nil
		case resource.PackageStateFailed:
			return true, fmt.Errorf("Package processing failed")
		case resource.PackageStateExpired:
			return true, fmt.Errorf("Package expired")
		}

		return false, nil
	}, 5*time.Second, stageTimeout)

	if err != nil {
		return nil, warnings, err
	}

	// Return the uploaded package, warnings, and no error
	return uploadedPackage, warnings, nil
}

// CreateBitsPackageByApplication creates a new package for an app to upload bits
func (m *BitsManager) CreateBitsPackageByApplication(appGUID string) (*resource.Package, ccv3.Warnings, error) {
	inputPackage := &resource.PackageCreate{
		Type: string(constant.PackageTypeBits),
		Relationships: resource.AppRelationship{
			App: resource.ToOneRelationship{
				Data: &resource.Relationship{
					GUID: appGUID,
				},
			},
		},
	}

	// Context for the request
	ctx := context.Background()

	// Call the client method to create the package
	createdPackage, err := m.client.Packages.Create(ctx, inputPackage)
	if err != nil {
		return nil, nil, err
	}

	// Return the created package, warnings (if any), and no error
	return createdPackage, nil, nil
}

// CopyAppV3 - Copy one app to another by using only API
func (m *BitsManager) CopyAppV3(origAppGUID string, newAppGUID string) error {
	// Create the context for the requests
	ctx := context.Background()

	// Define the options for listing packages
	packageListOptions := &client.PackageListOptions{
		GUIDs:  client.Filter{Values: []string{origAppGUID}},
		States: client.Filter{Values: []string{"READY"}},
	}

	// Get the packages for the original app
	srcPkgs, err := m.client.Packages.ListAll(ctx, packageListOptions)
	if err != nil {
		return err
	}

	if len(srcPkgs) == 0 {
		return fmt.Errorf("No package found for app %s", origAppGUID)
	}

	latestPkg := srcPkgs[0]

	copiedPkg, err := m.client.Packages.Copy(ctx, latestPkg.GUID, newAppGUID)
	if err != nil {
		return err
	}

	// Wait for the package to be ready
	timeout := 15 * time.Minute
	err = m.PackageWaitReady(copiedPkg.GUID, timeout)
	if err != nil {
		return err
	}

	return nil
}

// PackageWaitReady : Poll only for READY state
func (m BitsManager) PackageWaitReady(packageGUID string, timeout time.Duration) error {
	// Create the context for the requests
	ctx := context.Background()
	return common.PollingWithTimeout(func() (bool, error) {

		ccPkg, err := m.client.Packages.Get(ctx, packageGUID)
		if err != nil {
			return true, err
		}

		if ccPkg.State == resource.PackageState(constant.PackageReady) {
			return true, nil
		}

		if ccPkg.State == resource.PackageState(constant.PackageFailed) || ccPkg.State == resource.PackageState(constant.PackageExpired) {
			return true, fmt.Errorf("Package %s, state: %s", ccPkg.GUID, ccPkg.State)
		}

		// Continue on any other states
		return false, nil
	}, 5*time.Second, timeout)
}
