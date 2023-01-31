package zipper

import (
	"archive/zip"
	"fmt"
	"mime"
	"path/filepath"
	"time"

	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type HttpHandler struct {
}

func (h HttpHandler) Zip(src *Source) (ZipReadCloser, error) {
	processor := NewCompressProcessor(src, h.readCloserFunc)
	zipProc, err := processor.ToZip()
	if err != nil {
		return nil, err
	}
	if zipProc != nil {
		return zipProc, nil
	}
	resp, err := h.doRequest(src)
	if err != nil {
		return nil, err
	}
	err = h.checkRespHttpError(resp)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return h.createZipFile(resp, src)
}

func (h HttpHandler) readCloserFunc(src *Source) (io.ReadCloser, int64, string, error) {
	resp, err := h.doRequest(src)
	if err != nil {
		return nil, 0, "", err
	}
	err = h.checkRespHttpError(resp)
	if err != nil {
		return nil, 0, "", err
	}
	path := src.Path
	_, params, err := mime.ParseMediaType(resp.Header.Get("Content-Disposition"))
	if err == nil {
		var ok bool
		if _, ok = params["filename"]; ok {
			path = params["filename"]
		}
	}
	return resp.Body, resp.ContentLength, path, nil
}

func (h HttpHandler) checkRespHttpError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	content := ""
	if err == nil {
		content = string(b)
	}
	return fmt.Errorf(
		"Error occured when dowloading file: %d %s: \n%s",
		resp.StatusCode,
		http.StatusText(resp.StatusCode),
		content,
	)
}

func (h HttpHandler) isExecutable(src *Source) bool {
	resp, err := h.doRequest(src)
	if err != nil {
		return false
	}

	err = h.checkRespHttpError(resp)
	if err != nil {
		return false
	}
	return IsExecutable(resp.Body)
}

func (h HttpHandler) createZipFile(resp *http.Response, src *Source) (ZipReadCloser, error) {
	zipFile, err := ioutil.TempFile("", "downloads-zipper")
	if err != nil {
		return nil, err
	}
	cleanFunc := func() error {
		return os.Remove(zipFile.Name())
	}
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	size := resp.ContentLength
	fh := &zip.FileHeader{
		Name:               filepath.Base(src.Path),
		UncompressedSize64: uint64(size),
	}
	fh.SetModTime(time.Now())
	if h.isExecutable(src) {
		fh.SetMode(0755)
	} else {
		fh.SetMode(0644)
	}

	if fh.UncompressedSize64 > ((1 << 32) - 1) {
		fh.UncompressedSize = (1 << 32) - 1
	} else {
		fh.UncompressedSize = uint32(fh.UncompressedSize64)
	}
	w, err := zipWriter.CreateHeader(fh)
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(w, resp.Body)
	if err != nil {
		return nil, err
	}
	zipWriter.Close()
	zipFile.Close()

	file, err := os.Open(zipFile.Name())
	if err != nil {
		return nil, err
	}
	fs, _ := file.Stat()
	return NewZipFile(file, fs.Size(), cleanFunc), nil
}

func (h HttpHandler) doRequest(src *Source) (*http.Response, error) {
	client := CtxHttpClient(src)
	u, _ := url.Parse(src.Path)
	username := ""
	password := ""
	if u.User != nil && u.User.Username() != "" {
		username = u.User.Username()
		password, _ = u.User.Password()
	}
	u.User = nil
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	if username != "" {
		req.SetBasicAuth(username, password)
	}
	return client.Do(req)
}

func (h HttpHandler) Detect(src *Source) bool {
	path := src.Path
	return IsWebURL(path)
}

func (h HttpHandler) Sha1(src *Source) (string, error) {
	client := CtxHttpClient(src)
	path := src.Path
	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return "", err
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	err = h.checkRespHttpError(resp)
	if err != nil {
		return "", err
	}
	return GetSha1FromReader(resp.Body)
}

func (h HttpHandler) Name() string {
	return "http"
}
