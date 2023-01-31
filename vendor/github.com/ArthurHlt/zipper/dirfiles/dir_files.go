/*
come from
*/
package dirfiles

import (
	"crypto/sha1"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"strings"

	"code.cloudfoundry.org/gofileutils/fileutils"
)

const windowsPathPrefix = `\\?\`

type FileFields struct {
	Path string
	Sha1 string
	Size int64
	Mode string
}

type Files interface {
	FilesInDir(dir string) (appFiles []FileFields, err error)
	CopyFiles(appFiles []FileFields, fromDir, toDir string) (err error)
	CountFiles(directory string) int64
	WalkAppFiles(dir string, onEachFile func(string, string) error) (err error)
}

type DirFiles struct{}

func (appfiles DirFiles) AppFilesInDir(dir string) ([]FileFields, error) {
	appFiles := []FileFields{}

	fullDirPath, toplevelErr := filepath.Abs(dir)
	if toplevelErr != nil {
		return appFiles, toplevelErr
	}

	toplevelErr = appfiles.WalkAppFiles(fullDirPath, func(fileName string, fullPath string) error {
		fileInfo, err := os.Lstat(fullPath)
		if err != nil {
			return err
		}

		appFile := FileFields{
			Path: filepath.ToSlash(fileName),
			Size: fileInfo.Size(),
		}

		if fileInfo.IsDir() {
			appFile.Sha1 = "0"
			appFile.Size = 0
		} else {
			sha, err := appfiles.shaFile(fullPath)
			if err != nil {
				return err
			}
			appFile.Sha1 = sha
		}

		appFiles = append(appFiles, appFile)

		return nil
	})

	return appFiles, toplevelErr
}

func (appfiles DirFiles) shaFile(fullPath string) (string, error) {
	hash := sha1.New()
	file, err := os.Open(fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (appfiles DirFiles) CopyFiles(appFiles []FileFields, fromDir, toDir string) error {
	for _, file := range appFiles {
		err := func() error {
			fromPath, err := filepath.Abs(filepath.Join(fromDir, file.Path))
			if err != nil {
				return err
			}

			if runtime.GOOS == "windows" {
				fromPath = windowsPathPrefix + fromPath
			}

			srcFileInfo, err := os.Stat(fromPath)
			if err != nil {
				return err
			}

			toPath, err := filepath.Abs(filepath.Join(toDir, file.Path))
			if err != nil {
				return err
			}

			if runtime.GOOS == "windows" {
				toPath = windowsPathPrefix + toPath
			}

			if srcFileInfo.IsDir() {
				err = os.MkdirAll(toPath, srcFileInfo.Mode())
				if err != nil {
					return err
				}
				return nil
			}

			return appfiles.copyFile(fromPath, toPath, srcFileInfo.Mode())
		}()

		if err != nil {
			return err
		}
	}

	return nil
}

func (appfiles DirFiles) copyFile(srcPath string, dstPath string, fileMode os.FileMode) error {
	dst, err := fileutils.Create(dstPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if runtime.GOOS != "windows" {
		err = dst.Chmod(fileMode)
		if err != nil {
			return err
		}
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return nil
}

func (appfiles DirFiles) CountFiles(directory string) int64 {
	var count int64
	appfiles.WalkAppFiles(directory, func(_, _ string) error {
		count++
		return nil
	})
	return count
}

func (appfiles DirFiles) WalkAppFiles(dir string, onEachFile func(string, string) error) error {
	cfIgnore := loadIgnoreFile(dir)
	walkFunc := func(fullPath string, f os.FileInfo, err error) error {
		fileRelativePath, _ := filepath.Rel(dir, fullPath)
		fileRelativeUnixPath := filepath.ToSlash(fileRelativePath)

		if err != nil && runtime.GOOS == "windows" {
			f, err = os.Lstat(windowsPathPrefix + fullPath)
			if err != nil {
				return err
			}
			fullPath = windowsPathPrefix + fullPath
		}

		if fullPath == dir {
			return nil
		}

		if cfIgnore.FileShouldBeIgnored(fileRelativeUnixPath) {
			if err == nil && f.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if err != nil {
			return err
		}

		if !f.Mode().IsRegular() && !f.IsDir() {
			return nil
		}

		return onEachFile(fileRelativePath, fullPath)
	}

	return filepath.Walk(dir, walkFunc)
}

var ignorerFiles []string = []string{
	".cfignore",
	".cloudignore",
	".zipignore",
}

func loadIgnoreFile(dir string) IgnoreFiles {
	fileContents := make([]string, 0)
	for _, ignorer := range ignorerFiles {
		b, _ := ioutil.ReadFile(filepath.Join(dir, ignorer))
		fileContents = append(fileContents, string(b))
	}
	if len(fileContents) == 0 {
		return NewIgnoreFiles("")
	}
	return NewIgnoreFiles(strings.Join(fileContents, "\n"))
}
