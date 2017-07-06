package repo

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
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
	workspace string
}

// NewRepoManager -
func NewRepoManager(workspace string) *RepoManager {
	return &RepoManager{
		workspace: workspace,
	}
}

// GetGitRepository -
func (rm *RepoManager) GetGitRepository(repoURL string, user, password, privateKey *string) (repo Repository, err error) {

	var r *git.Repository

	urlPath, err := url.Parse(repoURL)
	if err != nil {
		return
	}

	baseName := filepath.Base(urlPath.Path)
	extName := filepath.Ext(urlPath.Path)
	p := fmt.Sprintf("%s/%s", rm.workspace, baseName[:len(baseName)-len(extName)])

	if _, err = os.Stat(p); os.IsNotExist(err) {
		err = nil

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
	} else {
		r, err = git.PlainOpen(p)
	}
	if err != nil {
		os.RemoveAll(p)
		return
	}

	err = nil
	repo = &GitRepository{
		repoPath: p,
		gitRepo:  r,
	}
	return
}

// GetGithubRelease -
func (rm *RepoManager) GetGithubRelease(ghOwner, ghRepoName, archiveName string, token *string) (repo Repository, err error) {

	var ghClient *github.Client
	ctx := context.Background()

	if token == nil {
		ghClient = github.NewClient(nil)
	} else {
		ts := oauth2.StaticTokenSource(
			&oauth2.Token{AccessToken: *token},
		)
		tc := oauth2.NewClient(ctx, ts)
		ghClient = github.NewClient(tc)
	}

	if _, _, err = ghClient.Repositories.Get(ctx, ghOwner, ghRepoName); err != nil {
		return
	}

	path := rm.workspace + "/github_releases/" + ghOwner + "/" + ghRepoName
	if err = os.MkdirAll(path, os.ModePerm); err != nil {
		return
	}

	repo = &GithubRelease{
		client: ghClient,

		archivePath: path + "/" + archiveName,
		owner:       ghOwner,
		repoName:    ghRepoName,

		archiveName: archiveName,
	}
	return
}
