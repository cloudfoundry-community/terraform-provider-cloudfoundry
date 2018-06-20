package repo_test

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-cf/cloudfoundry/repo"
)

func TestGithubReleaseRepo(t *testing.T) {

	testReleaseFileDownload(t)
	testSourceZipFileDownload(t)
	testSourceTarFileDownload(t)
}

func testReleaseFileDownload(t *testing.T) {
	fmt.Println("Test: release file download")

	repoManager := repo.NewManager(workspace)
	testUser := os.Getenv("GITHUB_USER")
	testPassword := os.Getenv("GITHUB_TOKEN")
	ghRelease, err := repoManager.GetGithubRelease("mevansam", "test-app", "test_release_file.zip", &testUser, &testPassword)
	checkError(t, err)
	defer ghRelease.Clean()

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

func testSourceZipFileDownload(t *testing.T) {
	fmt.Println("Test: source zip file download")

	repoManager := repo.NewManager(workspace)
	testUser := os.Getenv("GITHUB_USER")
	testPassword := os.Getenv("GITHUB_TOKEN")
	ghRelease, err := repoManager.GetGithubRelease("mevansam", "test-app", "zipball", &testUser, &testPassword)

	checkError(t, err)
	defer ghRelease.Clean()

	err = ghRelease.SetVersion("v0.0.1", repo.DefaultVersionType)
	checkError(t, err)

	validateSourceZip(ghRelease.GetPath(), t)
}

func testSourceTarFileDownload(t *testing.T) {
	fmt.Println("Test: source tar file download")

	repoManager := repo.NewManager(workspace)
	testUser := os.Getenv("GITHUB_USER")
	testPassword := os.Getenv("GITHUB_TOKEN")
	ghRelease, err := repoManager.GetGithubRelease("mevansam", "test-app", "tarball", &testUser, &testPassword)

	checkError(t, err)
	defer ghRelease.Clean()

	err = ghRelease.SetVersion("v0.0.1", repo.DefaultVersionType)
	checkError(t, err)

	validateSourceZip(ghRelease.GetPath(), t)
}

func readArchiveZip(path string, t *testing.T) (content string) {

	r, err := zip.OpenReader(path)
	checkError(t, err)
	defer r.Close()

	if len(r.File) == 1 {
		var rc io.ReadCloser
		rc, err = r.File[0].Open()
		checkError(t, err)
		defer rc.Close()

		buf := bytes.NewBuffer(nil)
		_, err = io.Copy(buf, rc)
		checkError(t, err)

		content = buf.String()

	} else {
		err = fmt.Errorf("expected only 1 file in the test release archive zip '%s'", path)
		checkError(t, err)
	}
	return content
}

func validateSourceZip(path string, t *testing.T) {

	r, err := zip.OpenReader(path)
	checkError(t, err)
	defer r.Close()

	var matcher *regexp.Regexp
	matcher, err = regexp.Compile("# Test App - a simple Go webapp\n")
	checkError(t, err)

	for _, f := range r.File {

		if f.Name == "README.md" {
			rc, err := f.Open()
			checkError(t, err)
			defer rc.Close()

			buf := bytes.NewBuffer(nil)
			_, err = io.Copy(buf, rc)
			checkError(t, err)

			matched := matcher.Match(buf.Bytes())
			if matched {
				return
			}
			t.Fatalf("'%s/README.md' content is not consistent with what was expected", path)
		}
	}

	t.Fatalf("'README.md' was not found in source archive")
}
