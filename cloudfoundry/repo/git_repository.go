package repo

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"os"
	"sync"
)

const (
	// GitVersionTypeBranch -
	GitVersionTypeBranch = 1 + iota
	// GitVersionTypeTag -
	GitVersionTypeTag
)

// GitRepository -
type GitRepository struct {
	repoPath string
	gitRepo  *git.Repository
	auth     transport.AuthMethod
	mutex    *sync.Mutex
}

// GetPath -
func (r *GitRepository) GetPath() string {
	return r.repoPath
}

// SetVersion -
func (r *GitRepository) SetVersion(version string, versionType VersionType) (err error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	var (
		w       *git.Worktree
		refName plumbing.ReferenceName
	)

	if w, err = r.gitRepo.Worktree(); err != nil {
		return err
	}

	switch versionType {
	case GitVersionTypeBranch:
		refName = plumbing.ReferenceName("refs/heads/" + version)
		options := &git.PullOptions{
			ReferenceName:     refName,
			SingleBranch:      true,
			Auth:              r.auth,
			RecurseSubmodules: git.DefaultSubmoduleRecursionDepth,
		}
		if err = r.gitRepo.Pull(options); err != nil && err != git.NoErrAlreadyUpToDate {
			return err
		}

	case GitVersionTypeTag:
		refName = plumbing.ReferenceName("refs/tags/" + version)
		options := &git.CheckoutOptions{
			Branch: refName,
			Force:  true,
		}
		if err = w.Checkout(options); err != nil && err != git.NoErrAlreadyUpToDate {
			return err
		}

	default:
		return fmt.Errorf("invalid git version type")
	}

	return nil
}

// String -
func (r *GitRepository) String() string {

	ref, err := r.gitRepo.Head()
	if err != nil {
		panic(err.Error())
	}
	return ref.Hash().String()
}

// Clean -
func (r *GitRepository) Clean() error {
	return os.RemoveAll(r.repoPath)
}
