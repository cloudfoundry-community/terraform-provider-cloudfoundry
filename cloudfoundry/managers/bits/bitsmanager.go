package bits

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3/constant"
	"code.cloudfoundry.org/cli/resources"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/common"
	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/managers/raw"
)

// Manage upload bits like app and buildpack in full stream
type BitsManager struct {
	clientV2   *ccv2.Client
	clientV3   *ccv3.Client
	rawClient  *raw.RawClient
	httpClient *http.Client
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
func NewBitsManager(clientV2 *ccv2.Client, clientV3 *ccv3.Client, rawClient *raw.RawClient, httpClient *http.Client) *BitsManager {
	return &BitsManager{
		clientV2:   clientV2,
		clientV3:   clientV3,
		rawClient:  rawClient,
		httpClient: httpClient,
	}
}

// CopyApp - Copy one app to another by using only api
func (m BitsManager) CopyApp(origAppGuid string, newAppGuid string) error {
	path := fmt.Sprintf("/v2/apps/%s/copy_bits", newAppGuid)
	data := []byte(fmt.Sprintf(`{"source_app_guid":"%s"}`, origAppGuid))

	req, err := m.rawClient.NewRequest("POST", path, data)
	if err != nil {
		return err
	}
	resp, err := m.rawClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	var j job
	err = json.NewDecoder(resp.Body).Decode(&j)
	if err != nil {
		return err
	}
	_, err = m.clientV2.PollJob(ccv2.Job{
		GUID: j.Entity.GUID,
	})
	if err != nil {
		return err
	}
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

	req, err := m.rawClient.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return err
	}
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", mpw.Boundary())
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = m.predictPartBuildpack(baseName, fileSize, mpw.Boundary())
	req.Body = r

	resp, err := m.rawClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	return nil
}

// GetAppEnvironmentVariables - Get app environment variables
func (m BitsManager) GetAppEnvironmentVariables(appGUID string) (map[string]string, error) {
	apiURL := fmt.Sprintf("/v3/apps/%s/environment_variables", appGUID)

	req, err := m.rawClient.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	var responseBody = struct {
		Var map[string]string `json:"var"`
	}{}
	resp, err := m.rawClient.Do(req)
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

// SetAppEnvironmentVariables - Remove app environment variables
func (m BitsManager) SetAppEnvironmentVariables(appGUID string, env map[string]interface{}) error {
	apiURL := fmt.Sprintf("/v3/apps/%s/environment_variables", appGUID)

	req, err := m.rawClient.NewRequest("PATCH", apiURL, nil)
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

	resp, err := m.rawClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
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
	req, err := m.rawClient.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return err
	}
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", mpw.Boundary())
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = m.predictPartApp(fileSize, mpw.Boundary())
	req.Body = r

	resp, err := m.rawClient.Do(req)
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
		resp, err := m.httpClient.Get(path)
		if err != nil {
			return ZipFile{}, err
		}
		fileSize := resp.ContentLength
		if resp.StatusCode < 200 || resp.StatusCode >= 400 {
			return ZipFile{}, fmt.Errorf(resp.Status)
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
func (m BitsManager) CreateDockerPackage(appGUID string, dockerImage string, dockerUsername string, dockerPassword string) (resources.Package, ccv3.Warnings, error) {
	pkg, warnings, err := m.clientV3.CreatePackage(resources.Package{
		Type: constant.PackageTypeDocker,
		Relationships: resources.Relationships{
			constant.RelationshipTypeApplication: resources.Relationship{GUID: appGUID},
		},
		DockerImage:    dockerImage,
		DockerUsername: dockerUsername,
		DockerPassword: dockerPassword,
	})

	if err != nil {
		return resources.Package{}, warnings, err
	}

	return pkg, warnings, nil
}

// CreateAndUploadBitsPackage creates a new package and upload bits to the application
func (m BitsManager) CreateAndUploadBitsPackage(appGUID string, path string, stageTimeout time.Duration) (resources.Package, ccv3.Warnings, error) {
	pkg, warnings, err := m.CreateBitsPackageByApplication(appGUID)

	if err != nil {
		return resources.Package{}, warnings, err
	}

	_, warnings, err = m.clientV3.UploadPackage(pkg, path)
	if err != nil {
		return resources.Package{}, warnings, err
	}

	// Poll once every 5 sec, timeout ${stageTimeout} fixed by appDeploy
	err = common.PollingWithTimeout(func() (bool, error) {

		ccPkg, _, err := m.clientV3.GetPackage(pkg.GUID)
		if err != nil {
			return true, err
		}

		if ccPkg.State == constant.PackageReady {
			return true, nil
		}

		if ccPkg.State == constant.PackageFailed {
			return true, fmt.Errorf("Package processing failed")
		} else if ccPkg.State == constant.PackageExpired {
			return true, fmt.Errorf("Package expired")
		}

		return false, nil
	}, 5*time.Second, stageTimeout)

	if err != nil {
		return resources.Package{}, warnings, err
	}

	return pkg, warnings, nil
}

// CreateBitsPackageByApplication creates a new package for an app to upload bits
func (m BitsManager) CreateBitsPackageByApplication(appGUID string) (resources.Package, ccv3.Warnings, error) {
	inputPackage := resources.Package{
		Type: constant.PackageTypeBits,
		Relationships: resources.Relationships{
			constant.RelationshipTypeApplication: resources.Relationship{GUID: appGUID},
		},
	}

	pkg, warnings, err := m.clientV3.CreatePackage(inputPackage)
	if err != nil {
		return resources.Package{}, warnings, err
	}

	return pkg, warnings, err
}

// CopyAppV3 - Copy one app to another by using only api
func (m BitsManager) CopyAppV3(origAppGUID string, newAppGUID string) error {
	srcPkgs, _, err := m.clientV3.GetPackages(
		ccv3.Query{Key: ccv3.AppGUIDFilter, Values: []string{origAppGUID}},
		ccv3.Query{Key: ccv3.StatesFilter, Values: []string{"READY"}},
		ccv3.Query{Key: ccv3.OrderBy, Values: []string{"-created_at"}},
	)

	if err != nil {
		return err
	}

	if len(srcPkgs) == 0 {
		return fmt.Errorf("No package found for app %s", origAppGUID)
	}

	latestPkg := srcPkgs[0]

	_, _, err = m.clientV3.CopyPackage(latestPkg.GUID, newAppGUID)

	if err != nil {
		return err
	}
	return nil
}
