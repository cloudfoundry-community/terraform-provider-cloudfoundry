package repo_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/builtin/providers/cf/repo"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestGitRepo(t *testing.T) {

	workspace := getGitWorkspaceDirectory()
	defer os.RemoveAll(workspace)

	testClone(workspace, t)
	testBranchedContent(workspace, t)
	testTaggedContent(workspace, t)
}

func testClone(workspace string, t *testing.T) {

	fmt.Println("Test: get repo master branch")

	gitRepo := getGitRepo(workspace, t)
	path := gitRepo.GetPath()

	readMeContent, err := ioutil.ReadFile(path + "/README.md")
	checkError(t, err)

	matched, err := regexp.Match("# Test App - a simple Go webapp\\n", readMeContent)
	checkError(t, err)
	if !matched {
		fmt.Printf("Content of '%s/README.md'\n==>\n%s\n<==\n", path, string(readMeContent))
		t.Fatalf("'%s/README.md' content is not consistent with what was expected", path)
	}

	testRepo, err := git.PlainOpen(path)
	checkError(t, err)
	testRef, err := testRepo.Head()
	checkError(t, err)

	if testRef.Name() != plumbing.Master {
		t.Fatalf("expected git repo to be on '%s' but instead it was on '%s'", plumbing.Master, testRef.Name())
	}
	if testRef.Hash().String() != fmt.Sprintf("%s", gitRepo) {
		t.Fatalf("inconsistent repo hash")
	}
}

func testBranchedContent(workspace string, t *testing.T) {
	fmt.Println("Test: get repo branched content")

	gitRepo := getGitRepo(workspace, t)
	path := gitRepo.GetPath()

	err := gitRepo.SetVersion("test1", repo.GitVersionTypeBranch)
	checkError(t, err)

	readMeContent, err := ioutil.ReadFile(path + "/README.md")
	checkError(t, err)

	matched, err := regexp.Match("# Test App - a simple Go webapp in branch test1\\n", readMeContent)
	checkError(t, err)
	if !matched {
		fmt.Printf("Content of '%s/README.md'\n==>\n%s\n<==\n", path, string(readMeContent))
		t.Fatalf("'%s/README.md' content is not consistent with what was expected", path)
	}

	err = gitRepo.SetVersion("test2", repo.GitVersionTypeBranch)
	checkError(t, err)

	readMeContent, err = ioutil.ReadFile(path + "/README.md")
	checkError(t, err)

	matched, err = regexp.Match("# Test App - a simple Go webapp in branch test2\\n", readMeContent)
	checkError(t, err)
	if !matched {
		fmt.Printf("Content of '%s/README.md'\n==>\n%s\n<==\n", path, string(readMeContent))
		t.Fatalf("'%s/README.md' content is not consistent with what was expected", path)
	}
}

func testTaggedContent(workspace string, t *testing.T) {
	fmt.Println("Test: get repo tagged content")

	gitRepo := getGitRepo(workspace, t)
	path := gitRepo.GetPath()

	err := gitRepo.SetVersion("v0.0_test2", repo.GitVersionTypeTag)
	checkError(t, err)

	readMeContent, err := ioutil.ReadFile(path + "/README.md")
	checkError(t, err)

	matched, err := regexp.Match("# Test App - a simple Go webapp in branch test2 - v0.0_test2\\n", readMeContent)
	checkError(t, err)
	if !matched {
		fmt.Printf("Content of '%s/README.md'\n==>\n%s\n<==\n", path, string(readMeContent))
		t.Fatalf("'%s/README.md' content is not consistent with what was expected", path)
	}
}

func getGitRepo(workspace string, t *testing.T) (gitRepo repo.Repository) {

	repoManager := repo.NewRepoManager(workspace)
	gitRepo, err := repoManager.GetGitRepository("https://github.com/mevansam/test-app.git", nil, nil, nil)
	checkError(t, err)

	path := gitRepo.GetPath()
	if filepath.Base(path) != "test-app" {
		t.Fatalf("repo path '%s' does not have a base folder 'test-app'\n", path)
	}

	if _, err := os.Stat(path + "/LICENSE"); os.IsNotExist(err) {
		t.Fatalf("file '%s/LICENSE' does not exist", path)
	}
	if _, err := os.Stat(path + "/README.md"); os.IsNotExist(err) {
		t.Fatalf("file '%s/README.md' does not exist", path)
	}
	return
}

func getGitWorkspaceDirectory() (dir string) {

	var err error

	if dir, err = filepath.Abs(filepath.Dir(os.Args[0])); err == nil {

		dir += "/.test_git"
		if err = os.RemoveAll(dir); err == nil {
			if err = os.Mkdir(dir, os.ModePerm); err == nil {
				return
			}
		}
	}
	panic(err.Error())
}
