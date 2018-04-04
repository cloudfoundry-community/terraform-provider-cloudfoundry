package repo

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/google/go-github/github"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
)

// VersionType -
type VersionType uint

const (
	// DefaultVersionType -
	DefaultVersionType = 0
)

// Repository -
type Repository interface {
	GetPath() string
	SetVersion(version string, versionType VersionType) (err error)
}

// RepoManager -
type RepoManager struct {
	gitMutex *sync.Mutex
}

// NewRepoManager -
func NewRepoManager() *RepoManager {
	return &RepoManager{
		gitMutex: &sync.Mutex{},
	}
}

// GetGitRepository -
func (rm *RepoManager) GetGitRepository(repoURL string, user, password, privateKey *string) (repo Repository, err error) {

	rm.gitMutex.Lock()
	defer rm.gitMutex.Unlock()

	var r *git.Repository

	p, err := ioutil.TempDir("", "terraform-provider-cf")
	if err != nil {
		return nil, err
	}

	if user != nil {

		var auth transport.AuthMethod

		if password != nil {

			if privateKey != nil {
				auth, err = ssh.NewPublicKeys(*user, []byte(*privateKey), *password)
			} else {
				auth = &ssh.Password{
					User: *user,
					Pass: *password,
				}
			}
		} else if privateKey != nil {
			auth, err = ssh.NewPublicKeys(*user, []byte(*privateKey), "")
		} else {
			err = fmt.Errorf("authentication password or key was not provided for user '%s'\n", *user)
		}
		if err != nil {
			return
		}
		r, err = git.PlainClone(p, false,
			&git.CloneOptions{
				URL:               repoURL,
				Auth:              auth,
				ReferenceName:     plumbing.Master,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
	} else {
		r, err = git.PlainClone(p, false,
			&git.CloneOptions{
				URL:               repoURL,
				ReferenceName:     plumbing.Master,
				RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
			})
	}
	if err != nil {
		_ = os.RemoveAll(p)
		return nil, err
	}

	return &GitRepository{
		repoPath: p,
		gitRepo:  r,
		mutex:    rm.gitMutex,
	}, nil
}

// GetGithubRelease -
func (rm *RepoManager) GetGithubRelease(ghOwner, ghRepoName, archiveName string, user *string, password *string) (repo Repository, err error) {
	var ghClient *github.Client
	ctx := context.Background()

	if user == nil || password == nil {
		ghClient = github.NewClient(nil)
	} else {
		tp := github.BasicAuthTransport{
			Username: *user,
			Password: *password,
		}
		ghClient = github.NewClient(tp.Client())
	}

	if _, _, err = ghClient.Repositories.Get(ctx, ghOwner, ghRepoName); err != nil {
		return nil, err
	}

	path, err := ioutil.TempDir("", "terraform-provider-cf")
	if err != nil {
		return nil, err
	}

	return &GithubRelease{
		client:      ghClient,
		archivePath: path + "/" + archiveName,
		owner:       ghOwner,
		repoName:    ghRepoName,
		archiveName: archiveName,
	}, nil
}
