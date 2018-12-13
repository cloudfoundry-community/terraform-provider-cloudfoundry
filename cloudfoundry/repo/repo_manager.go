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
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
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
	Clean() error
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

func (rm *RepoManager) getAuthMethod(repoURL string, user, password, privateKey *string) (transport.AuthMethod, error) {
	ep, err := transport.NewEndpoint(repoURL)
	if err != nil {
		err = fmt.Errorf("unable to parse repository url : %s", err)
		return nil, err
	}
	proto := ep.Protocol()

	if user == nil {
		return nil, nil
	}

	if proto == "http" || proto == "https" {
		if privateKey != nil {
			return nil, fmt.Errorf("privatekey authentication not available with http(s) protocol")
		}
		if password != nil {
			return http.NewBasicAuth(*user, *password), nil
		}
		return nil, fmt.Errorf("missing password for http(s) authentication")
	}

	if proto == "ssh" {
		if privateKey != nil {
			auth, err := ssh.NewPublicKeys(*user, []byte(*privateKey), "")
			if err != nil {
				return nil, fmt.Errorf("cannot use private key for ssh authentication: %s", err)
			}
			return auth, nil
		}
		if password != nil {
			return &ssh.Password{User: *user, Pass: *password}, nil
		}
		return nil, fmt.Errorf("missing password or private key for ssh authentication")
	}

	return nil, fmt.Errorf("authentication not available for protocol '%s'", proto)
}

// GetGitRepository -
func (rm *RepoManager) GetGitRepository(name string, repoURL string, user, password, privateKey *string) (repo Repository, err error) {
	rm.gitMutex.Lock()
	defer rm.gitMutex.Unlock()
	var r *git.Repository
	var auth transport.AuthMethod

	p, err := ioutil.TempDir("", "terraform-provider-cloudfoundry")
	if err != nil {
		return nil, err
	}
	p = p + "/" + name

	if auth, err = rm.getAuthMethod(repoURL, user, password, privateKey); err != nil {
		return nil, err
	}

	r, err = git.PlainClone(p, false, &git.CloneOptions{
		URL:               repoURL,
		Auth:              auth,
		ReferenceName:     plumbing.Master,
		RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
	})

	if err != nil {
		_ = os.RemoveAll(p)
		return nil, fmt.Errorf("unable to clone repository : %s", err)
	}

	return &GitRepository{
		repoPath: p,
		gitRepo:  r,
		mutex:    rm.gitMutex,
		auth:     auth,
	}, nil
}

// GetGithubRelease -
func (rm *RepoManager) GetGithubRelease(ghOwner, ghRepoName, archiveName string, user *string, password *string) (repo Repository, err error) {
	rm.gitMutex.Lock()
	defer rm.gitMutex.Unlock()

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

	path, err := ioutil.TempDir("", "terraform-provider-cloudfoundry")
	if err != nil {
		return nil, err
	}

	return &GithubRelease{
		client:      ghClient,
		archivePath: path + "/" + archiveName,
		owner:       ghOwner,
		repoName:    ghRepoName,
		archiveName: archiveName,
		mutex:       rm.gitMutex,
	}, nil
}
