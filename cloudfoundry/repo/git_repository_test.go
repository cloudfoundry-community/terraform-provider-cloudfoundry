package repo_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"testing"

	"github.com/terraform-providers/terraform-provider-cloudfoundry/cloudfoundry/repo"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
)

func TestGitRepo(t *testing.T) {
	testClone(t)
	testBranchedContent(t)
	testTaggedContent(t)
}

func testClone(t *testing.T) {

	fmt.Println("Test: get repo master branch")

	gitRepo := getGitRepo(t)
	defer gitRepo.Clean()
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

func testBranchedContent(t *testing.T) {
	fmt.Println("Test: get repo branched content")

	gitRepo := getGitRepo(t)
	defer gitRepo.Clean()
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

func testTaggedContent(t *testing.T) {
	fmt.Println("Test: get repo tagged content")

	gitRepo := getGitRepo(t)
	defer gitRepo.Clean()
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

func getGitRepo(t *testing.T) (gitRepo repo.Repository) {

	repoManager := repo.NewRepoManager()
	gitRepo, err := repoManager.GetGitRepository("", "https://github.com/mevansam/test-app.git", nil, nil, nil)
	checkError(t, err)

	path := gitRepo.GetPath()

	if _, err := os.Stat(path + "/LICENSE"); os.IsNotExist(err) {
		t.Fatalf("file '%s/LICENSE' does not exist", path)
	}
	if _, err := os.Stat(path + "/README.md"); os.IsNotExist(err) {
		t.Fatalf("file '%s/README.md' does not exist", path)
	}
	return gitRepo
}
