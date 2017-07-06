package repo_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/builtin/providers/cf/repo"
)

func TestGithubReleaseRepo(t *testing.T) {

	workspace := getGithubReleaseWorkspace()
	defer os.RemoveAll(workspace)

	testReleaseFileDownload(workspace, t)
	testSourceZipFileDownload(workspace, t)
	testSourceTarFileDownload(workspace, t)
}

func testReleaseFileDownload(workspace string, t *testing.T) {
	fmt.Println("Test: release file download")

	repoManager := repo.NewRepoManager(workspace)
	ghRelease, err := repoManager.GetGithubRelease("mevansam", "test-app", "test_release_file.zip", nil)
	checkError(t, err)

	err = ghRelease.SetVersion("v0.0.1", repo.DefaultVersionType)
	checkError(t, err)

	content := readArchiveZip(ghRelease.GetPath(), t)
	if content != "test release file 0.0.1" {
		t.Fatalf("unexpected archive contents: %s\n", content)
	}

	err = ghRelease.SetVersion("v0.0.2", repo.DefaultVersionType)
	checkError(t, err)

	content = readArchiveZip(ghRelease.GetPath(), t)
	if content != "test release file 0.0.2" {
		t.Fatalf("unexpected archive contents: %s\n", content)
	}
}

func testSourceZipFileDownload(workspace string, t *testing.T) {
	fmt.Println("Test: source zip file download")

	repoManager := repo.NewRepoManager(workspace)
	ghRelease, err := repoManager.GetGithubRelease("mevansam", "test-app", "zipball", nil)
	checkError(t, err)

	err = ghRelease.SetVersion("v0.0.1", repo.DefaultVersionType)
	checkError(t, err)

	validateSourceZip(ghRelease.GetPath(), t)
}

func testSourceTarFileDownload(workspace string, t *testing.T) {
	fmt.Println("Test: source tar file download")

	repoManager := repo.NewRepoManager(workspace)
	ghRelease, err := repoManager.GetGithubRelease("mevansam", "test-app", "tarball", nil)
	checkError(t, err)

	err = ghRelease.SetVersion("v0.0.1", repo.DefaultVersionType)
	checkError(t, err)

	validateSourceZip(ghRelease.GetPath(), t)
}

func readArchiveZip(path string, t *testing.T) (content string) {

	r, err := zip.OpenReader(path)
	defer r.Close()
	checkError(t, err)

	if len(r.File) == 1 {

		rc, err := r.File[0].Open()
		defer rc.Close()
		checkError(t, err)

		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, rc)
		checkError(t, err)

		content = buf.String()

	} else {
		err = fmt.Errorf("expected only 1 file in the test release archive zip '%s'", path)
	}
	return
}

func validateSourceZip(path string, t *testing.T) {

	r, err := zip.OpenReader(path)
	defer r.Close()
	checkError(t, err)

	for _, f := range r.File {
		if f.Name == "README.md" {
			rc, err := f.Open()
			defer rc.Close()
			checkError(t, err)

			buf := bytes.NewBuffer(nil)
			_, err = io.Copy(buf, rc)
			checkError(t, err)

			matched, err := regexp.Match("# Test App - a simple Go webapp\n", buf.Bytes())
			checkError(t, err)
			if matched {
				return
			}
			t.Fatalf("'%s/README.md' content is not consistent with what was expected", path)
		}
	}

	t.Fatalf("'README.md' was not found in source archive")
}

func getGithubReleaseWorkspace() (dir string) {

	var err error

	if dir, err = filepath.Abs(filepath.Dir(os.Args[0])); err == nil {

		dir += "/.test_github_release"
		if err = os.RemoveAll(dir); err == nil {
			if err = os.Mkdir(dir, os.ModePerm); err == nil {
				return
			}
		}
	}
	panic(err.Error())
}
