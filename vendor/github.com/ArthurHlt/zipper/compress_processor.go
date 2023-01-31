package zipper

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var ZIP_FILE_EXT []string = []string{
	".zip",
	".jar",
	".war",
}
var TAR_FILE_EXT []string = []string{
	".tar",
}
var GZIP_FILE_EXT []string = []string{
	".gz",
	".gzip",
}

var TARGZ_FILE_EXT []string = []string{
	".tgz",
}
var BZ2_FILE_EXT []string = []string{
	".bz2",
}

type readCloserFunc func(src *Source) (io.ReadCloser, int64, string, error)

type CompressProcessor struct {
	src            *Source
	readCloserFunc readCloserFunc
}

func NewCompressProcessor(src *Source, readCloserFunc readCloserFunc) *CompressProcessor {
	return &CompressProcessor{
		src:            src,
		readCloserFunc: readCloserFunc,
	}
}

func (p CompressProcessor) ToZip() (ZipReadCloser, error) {
	reader, _, path, err := p.readCloserFunc(p.src)
	if err != nil {
		return nil, err
	}
	isZipFile, err := p.isZipFile(reader, path)
	if err != nil {
		return nil, err
	}
	reader, dataLen, path, err := p.readCloserFunc(p.src)
	if err != nil {
		return nil, err
	}
	if isZipFile {
		return NewZipFile(reader, dataLen, func() error {
			return nil
		}), nil
	}

	isTarFile, err := p.isTarFile(reader, path)
	if err != nil {
		return nil, err
	}
	reader, dataLen, path, err = p.readCloserFunc(p.src)
	if err != nil {
		return nil, err
	}
	if isTarFile {
		return p.tarToZip(reader)
	}

	isTarGzFile, err := p.isTarGzFile(reader, path)
	if err != nil {
		return nil, err
	}
	reader, dataLen, path, err = p.readCloserFunc(p.src)
	if err != nil {
		return nil, err
	}
	if isTarGzFile {
		return p.tarGzToZip(reader)
	}

	isTarBz2File, err := p.isTarBz2File(reader, path)
	if err != nil {
		return nil, err
	}
	reader, dataLen, path, err = p.readCloserFunc(p.src)
	if err != nil {
		return nil, err
	}
	if isTarBz2File {
		return p.tarBzip2ToZip(reader)
	}
	return nil, nil
}

func (p CompressProcessor) tarGzToZip(r io.ReadCloser) (*ZipFile, error) {
	gzf, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return p.tarToZip(gzf)
}

func (p CompressProcessor) tarBzip2ToZip(r io.ReadCloser) (*ZipFile, error) {
	bz2 := bzip2.NewReader(r)

	return p.tarToZip(ioutil.NopCloser(bz2))
}

func (p CompressProcessor) tarToZip(r io.ReadCloser) (*ZipFile, error) {
	zipFile, err := ioutil.TempFile("", "processor-zipper")
	if err != nil {
		return nil, err
	}
	cleanFunc := func() error {
		return os.Remove(zipFile.Name())
	}
	err = p.writeTarToZip(r, zipFile)
	if err != nil {
		zipFile.Close()
		return nil, err
	}
	zipFile.Close()
	file, err := os.Open(zipFile.Name())
	if err != nil {
		return nil, err
	}
	fs, _ := file.Stat()
	return NewZipFile(file, fs.Size(), cleanFunc), nil
}

func (p CompressProcessor) writeTarToZip(r io.Reader, zipFile *os.File) error {
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()
	tarReader := tar.NewReader(r)
	hasRootFolder := false
	i := 0
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		fileInfo := header.FileInfo()
		if i == 0 && fileInfo.IsDir() {
			hasRootFolder = true
			continue
		}
		zipHeader, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}
		if !hasRootFolder {
			zipHeader.Name = header.Name
		} else {
			splitFile := strings.Split(header.Name, "/")
			zipHeader.Name = strings.Join(splitFile[1:], "/")
		}
		if !fileInfo.IsDir() {
			zipHeader.Method = zip.Deflate
		}
		w, err := zipWriter.CreateHeader(zipHeader)
		if err != nil {
			return err
		}
		i++
		if fileInfo.IsDir() {
			continue
		}
		_, err = io.Copy(w, tarReader)
	}
	return nil
}

func (p CompressProcessor) isTarFile(reader io.ReadCloser, path string) (bool, error) {
	defer reader.Close()
	if HasExtFile(path, TAR_FILE_EXT...) {
		return true, nil
	}
	buf, err := Chunk(reader, 8)
	if err != nil {
		return false, err
	}
	if len(buf) < 8 {
		return false, nil
	}
	if string(buf[:5]) != "ustar" {
		return false, nil
	}
	return (buf[5] == 0x00 || buf[5] == 20) && (buf[6] == 0x30 || buf[6] == 20) && (buf[7] == 0x30 || buf[7] == 00), nil
}

func (p CompressProcessor) isGzFile(reader io.ReadCloser, path string) (bool, error) {
	defer reader.Close()
	if HasExtFile(path, GZIP_FILE_EXT...) {
		return true, nil
	}
	buf, err := Chunk(reader, 2)
	if err != nil {
		return false, err
	}
	if len(buf) < 2 {
		return false, nil
	}
	return buf[0] == 0x1F && buf[1] == 0x8B, nil
}

func (p CompressProcessor) isBz2File(reader io.ReadCloser, path string) (bool, error) {
	defer reader.Close()
	if HasExtFile(path, BZ2_FILE_EXT...) {
		return true, nil
	}
	buf, err := Chunk(reader, 2)
	if err != nil {
		return false, err
	}
	if len(buf) < 2 {
		return false, nil
	}
	return buf[0] == 0x42 && buf[1] == 0x5A, nil
}

func (p CompressProcessor) isZipFile(reader io.ReadCloser, path string) (bool, error) {
	defer reader.Close()
	if HasExtFile(path, ZIP_FILE_EXT...) {
		return true, nil
	}
	buf, err := Chunk(reader, 4)
	if err != nil {
		return false, err
	}
	if len(buf) < 4 {
		return false, nil
	}
	return buf[0] == 0x50 && buf[1] == 0x4b && (buf[2] == 0x03 || buf[2] == 0x05 || buf[2] == 0x07) && (buf[3] == 0x04 || buf[3] == 0x06 || buf[3] == 0x08), nil
}

func (p CompressProcessor) isTarGzFile(reader io.ReadCloser, path string) (bool, error) {
	defer reader.Close()
	isTgz := HasExtFile(path, TARGZ_FILE_EXT...)
	if isTgz {
		return true, nil
	}
	isGz := HasExtFile(path, GZIP_FILE_EXT...)
	if isGz {
		if IsTarFile(filepath.Ext(strings.TrimSuffix(path, filepath.Ext(path)))) {
			return true, nil
		}
	}
	isGz, err := p.isGzFile(reader, path)
	if err != nil {
		return false, err
	}
	if !isGz {
		return false, nil
	}
	newReader, _, path, err := p.readCloserFunc(p.src)
	if err != nil {
		return false, err
	}
	defer newReader.Close()
	gzf, err := gzip.NewReader(newReader)
	if err != nil {
		return false, err
	}
	return p.isTarFile(gzf, path)
}

func (p CompressProcessor) isTarBz2File(reader io.ReadCloser, path string) (bool, error) {
	defer reader.Close()
	isBz2 := HasExtFile(path, BZ2_FILE_EXT...)
	if isBz2 {
		if IsTarFile(filepath.Ext(strings.TrimSuffix(path, filepath.Ext(path)))) {
			return true, nil
		}
	}
	isBz2, err := p.isBz2File(reader, path)
	if err != nil {
		return false, err
	}
	if !isBz2 {
		return false, nil
	}
	newReader, _, path, err := p.readCloserFunc(p.src)
	if err != nil {
		return false, err
	}
	defer newReader.Close()
	bz2 := bzip2.NewReader(newReader)
	return p.isTarFile(ioutil.NopCloser(bz2), path)
}
