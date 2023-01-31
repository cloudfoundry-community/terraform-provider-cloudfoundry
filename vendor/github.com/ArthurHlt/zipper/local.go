package zipper

import (
	"archive/zip"
	"bufio"
	"bytes"
	"code.cloudfoundry.org/gofileutils/fileutils"
	"fmt"
	"github.com/ArthurHlt/zipper/dirfiles"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
)

type LocalHandler struct {
}

func (h LocalHandler) Zip(src *Source) (ZipReadCloser, error) {
	stat, err := os.Stat(src.Path)
	if err != nil {
		return nil, err
	}
	if !stat.IsDir() {
		processor := NewCompressProcessor(src, func(src *Source) (io.ReadCloser, int64, string, error) {
			file, err := os.Open(src.Path)
			if err != nil {
				return nil, 0, "", err
			}
			stat, err := file.Stat()
			if err != nil {
				file.Close()
				return nil, 0, "", err
			}
			return file, stat.Size(), src.Path, nil
		})
		zipProc, err := processor.ToZip()
		if err != nil {
			return nil, err
		}
		if zipProc != nil {
			return zipProc, nil
		}
	}
	path := src.Path
	zipFile, err := ioutil.TempFile("", "uploads-zipper")
	if err != nil {
		return nil, err
	}
	err = h.ZipFiles(path, zipFile)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(zipFile.Name())
	if err != nil {
		return nil, err
	}
	cleanFunc := func() error {
		return os.Remove(zipFile.Name())
	}
	fs, _ := file.Stat()
	return NewZipFile(file, fs.Size(), cleanFunc), nil
}
func (h LocalHandler) Detect(src *Source) bool {
	path := src.Path
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}
func (h LocalHandler) Sha1(src *Source) (string, error) {
	zipFile, err := h.Zip(src)
	if err != nil {
		return "", err
	}
	defer zipFile.Close()
	return GetSha1FromReader(zipFile)
}

func (h LocalHandler) Name() string {
	return "local"
}

func (h LocalHandler) ZipFiles(dirOrZipFilePath string, targetFile *os.File) error {
	err := h.writeZipFile(dirOrZipFilePath, targetFile)
	if err != nil {
		return err
	}

	_, err = targetFile.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	return nil
}

func (h LocalHandler) GetZipSize(zipFile *os.File) (int64, error) {
	zipFileSize := int64(0)

	stat, err := zipFile.Stat()
	if err != nil {
		return 0, err
	}

	zipFileSize = int64(stat.Size())

	return zipFileSize, nil
}

func (h LocalHandler) writeZipFile(dir string, targetFile *os.File) error {
	isEmpty, err := fileutils.IsDirEmpty(dir)
	if err != nil {
		return err
	}

	if isEmpty {
		return fmt.Errorf("%s is empty", dir)
	}

	writer := zip.NewWriter(targetFile)
	defer writer.Close()

	appfiles := dirfiles.DirFiles{}
	return appfiles.WalkAppFiles(dir, func(fileName string, fullPath string) error {
		fileInfo, err := os.Stat(fullPath)
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(fileInfo)
		if err != nil {
			return err
		}

		if runtime.GOOS == "windows" {
			header.SetMode(header.Mode() | 0700)
		}

		header.Name = filepath.ToSlash(fileName)
		header.Method = zip.Deflate

		if fileInfo.IsDir() {
			header.Name += "/"
		}

		zipFilePart, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}

		if fileInfo.IsDir() {
			return nil
		}

		file, err := os.Open(fullPath)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(zipFilePart, file)
		if err != nil {
			return err
		}

		return nil
	})
}

func (h LocalHandler) zipFileHeaderLocation(name string) (int64, error) {
	f, err := os.Open(name)
	if err != nil {
		return -1, err
	}

	defer f.Close()

	// zip file header signature, 0x04034b50, reversed due to little-endian byte order
	firstByte := byte(0x50)
	restBytes := []byte{0x4b, 0x03, 0x04}
	count := int64(-1)
	foundAt := int64(-1)

	reader := bufio.NewReader(f)

	keepGoing := true
	for keepGoing {
		count++

		b, err := reader.ReadByte()
		if err != nil {
			keepGoing = false
			break
		}

		if b == firstByte {
			nextBytes, err := reader.Peek(3)
			if err != nil {
				keepGoing = false
			}
			if bytes.Compare(nextBytes, restBytes) == 0 {
				foundAt = count
				keepGoing = false
				break
			}
		}
	}

	return foundAt, nil
}

func (h LocalHandler) isZipWithOffsetFileHeaderLocation(name string) bool {
	loc, err := h.zipFileHeaderLocation(name)
	if err != nil {
		return false
	}

	if loc > int64(-1) {
		f, err := os.Open(name)
		if err != nil {
			return false
		}

		defer f.Close()

		fi, err := f.Stat()
		if err != nil {
			return false
		}

		readerAt := io.NewSectionReader(f, loc, fi.Size())
		_, err = zip.NewReader(readerAt, fi.Size())
		if err == nil {
			return true
		}
	}

	return false
}

func (h LocalHandler) extractFile(f *zip.File, destDir string) error {
	if f.FileInfo().IsDir() {
		err := os.MkdirAll(filepath.Join(destDir, f.Name), os.ModeDir|os.ModePerm)
		if err != nil {
			return err
		}
		return nil
	}

	src, err := f.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	destFilePath := filepath.Join(destDir, f.Name)

	err = os.MkdirAll(filepath.Dir(destFilePath), os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	destFile, err := os.Create(destFilePath)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, src)
	if err != nil {
		return err
	}

	err = os.Chmod(destFilePath, f.FileInfo().Mode())
	if err != nil {
		return err
	}

	return nil
}
