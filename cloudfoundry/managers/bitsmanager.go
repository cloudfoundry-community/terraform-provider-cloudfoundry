package managers

import (
	"bytes"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv2"
	"code.cloudfoundry.org/cli/api/cloudcontroller/ccv3"
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
)

// Manage upload bits like app and buildpack in full stream
type BitsManager struct {
	clientV2   *ccv2.Client
	clientV3   *ccv3.Client
	rawClient  *RawClient
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
func NewBitsManager(sess *Session) *BitsManager {
	return &BitsManager{
		clientV2:   sess.ClientV2,
		clientV3:   sess.ClientV3,
		rawClient:  sess.RawClient,
		httpClient: sess.HttpClient,
	}
}

// CopyApp - Copy one app to another by using only api
func (m BitsManager) CopyApp(origAppGuid string, newAppGuid string) error {
	path := fmt.Sprintf("/v2/apps/%s/copy_bits", newAppGuid)
	data := bytes.NewReader([]byte(fmt.Sprintf(`{"source_app_guid":"%s"}`, origAppGuid)))
	req, err := m.rawClient.NewRequest("POST", path, ioutil.NopCloser(data))
	if err != nil {
		return err
	}
	resp, err := m.rawClient.Do(req)
	if err != nil {
		return err
	}
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
	request, err := m.rawClient.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return err
	}
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", mpw.Boundary())
	request.Header.Set("Content-Type", contentType)
	request.ContentLength = int64(m.predictPartBuildpack(baseName, int64(fileSize), mpw.Boundary()))
	request.Body = r

	_, err = m.rawClient.Do(request)
	if err != nil {
		panic(err)
	}
	return nil
}

// UploadApp - Upload an zip file containing app code to cloud foundry in full stream
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
	request, err := m.rawClient.NewRequest("PUT", apiURL, nil)
	if err != nil {
		return err
	}
	contentType := fmt.Sprintf("multipart/form-data; boundary=%s", mpw.Boundary())
	request.Header.Set("Content-Type", contentType)
	request.ContentLength = int64(m.predictPartApp(int64(fileSize), mpw.Boundary()))
	request.Body = r

	_, err = m.rawClient.Do(request)
	if err != nil {
		panic(err)
	}
	return nil
}

func (m BitsManager) predictPartApp(filesize int64, boundary string) int64 {
	buf := new(bytes.Buffer)
	mpw := multipart.NewWriter(buf)

	mpw.SetBoundary(boundary)
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

	part, err = mpw.CreatePart(h)
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

	mpw.SetBoundary(boundary)
	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="buildpack"; filename="%s"`, filename))
	h.Set("Content-Type", "application/zip")
	h.Set("Content-Length", fmt.Sprintf("%d", filesize))
	h.Set("Content-Transfer-Encoding", "binary")

	_, err := mpw.CreatePart(h)
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
		return ZipFile{
			r:        resp.Body,
			baseName: baseName,
			filesize: fileSize,
		}, nil
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
