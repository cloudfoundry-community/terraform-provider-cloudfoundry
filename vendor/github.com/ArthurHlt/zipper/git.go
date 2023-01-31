package zipper

import (
	"fmt"
	"github.com/whilp/git-urls"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/client"
	githttp "gopkg.in/src-d/go-git.v4/plumbing/transport/http"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type GitHandler struct {
	client *http.Client
}

var scpSyntax = regexp.MustCompile(`^([a-zA-Z0-9_]+@)?([a-zA-Z0-9._-]+):(.*)$`)

func NewGitHandler() *GitHandler {
	customClient := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		},
	}
	client.InstallProtocol("https", githttp.NewClient(customClient))
	client.InstallProtocol("http", githttp.NewClient(customClient))
	return &GitHandler{customClient}
}

func (h GitHandler) Zip(src *Source) (ZipReadCloser, error) {
	h.setHttpClient(src)
	path := src.Path
	tmpDir, err := ioutil.TempDir("", "git-zipper")
	if err != nil {
		return nil, err
	}
	gitUtils := h.makeGitUtils(tmpDir, path)
	err = gitUtils.Clone()
	if err != nil {
		return nil, err
	}
	err = os.RemoveAll(filepath.Join(tmpDir, ".git"))
	if err != nil {
		return nil, err
	}
	newSrc := NewSource(tmpDir + gitUtils.SubPath)
	newSrc.WithContext(src.Context())
	lh := &LocalHandler{}
	localFh, err := lh.Zip(newSrc)
	if err != nil {
		return nil, err
	}
	cleanFunc := func() error {
		err := localFh.Close()
		if err != nil {
			return err
		}
		return os.RemoveAll(tmpDir)
	}
	return NewZipFile(localFh, localFh.Size(), cleanFunc), nil
}

func (h GitHandler) makeGitUtils(tmpDir, path string) *GitUtils {
	u, err := giturls.Parse(path)
	if err != nil {
		u, _ = giturls.Parse("ssh://" + path)
	}

	// workaround of a bug when scheme wasn'nt set, fragment is not well parsed
	if u.Scheme == "ssh" && !strings.HasPrefix(path, "ssh") {
		u, _ = giturls.Parse("ssh://" + path)
	}

	subPath := ""
	if strings.Contains(u.Path, ".git/") {
		subPathSplit := strings.SplitN(u.Path, ".git/", 2)
		u.Path = subPathSplit[0] + ".git"
		if len(subPathSplit) == 2 {
			subPath = "/" + subPathSplit[1]
		}
	}

	refName := "master"
	if u.Fragment != "" {
		refName = u.Fragment
		u.Fragment = ""
	}
	authMethod, _ := createGitAuthMethod(u)
	if u.RawQuery != "" {
		u.RawQuery = ""
	}
	finalUrl := u.String()
	if u.Scheme == "ssh" && scpSyntax.MatchString(strings.TrimPrefix(u.String(), "ssh://")) {
		u.Scheme = ""
		finalUrl = strings.TrimPrefix(u.String(), "//")
	}
	gitUtils := &GitUtils{
		Url:        finalUrl,
		Folder:     tmpDir,
		RefName:    refName,
		AuthMethod: authMethod,
		SubPath:    subPath,
	}
	return gitUtils
}

func (h GitHandler) Sha1(src *Source) (string, error) {
	h.setHttpClient(src)
	path := src.Path
	tmpDir, err := ioutil.TempDir("", "git-zipper")
	if err != nil {
		return "", err
	}
	defer os.RemoveAll(tmpDir)
	gitUtils := h.makeGitUtils(tmpDir, path)
	return gitUtils.CommitSha1()
}

func (h GitHandler) Detect(src *Source) bool {
	path := src.Path
	u, err := giturls.Parse(path)
	if err != nil {
		return false
	}
	if !IsWebURL(path) && u.Scheme != "ssh" {
		return false
	}
	return HasExtFile(u.Path, ".git")
}

func (h *GitHandler) setHttpClient(src *Source) {
	*h.client = *CtxHttpClient(src)
}

func (h GitHandler) Name() string {
	return "git"
}

type GitUtils struct {
	Folder     string
	Url        string
	RefName    string
	AuthMethod transport.AuthMethod
	SubPath    string
}

var refTypes []string = []string{"heads", "tags"}

func createGitAuthMethod(uri *url.URL) (transport.AuthMethod, error) {
	privKeyPath := uri.Query().Get("private-key")
	passKey := uri.Query().Get("password-key")

	if uri.Scheme == "ssh" && privKeyPath == "" {
		return ssh.NewSSHAgentAuth(uri.User.Username())
	}

	if uri.Scheme == "ssh" && privKeyPath != "" {
		return ssh.NewPublicKeysFromFile(uri.User.Username(), privKeyPath, passKey)
	}
	if uri.User == nil || uri.User.Username() == "" {
		return nil, nil
	}
	password, _ := uri.User.Password()
	return &githttp.BasicAuth{
		Username: uri.User.Username(),
		Password: password,
	}, nil
}

func (g GitUtils) Clone() error {
	_, err := g.findRepo(false)
	if err != nil {
		return err
	}
	return nil
}
func (g GitUtils) CommitSha1() (string, error) {
	if g.refNameIsHash() {
		return g.RefName, nil
	}
	repo, err := g.findRepo(true)
	if err != nil {
		return "", err
	}
	iter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return "", err
	}
	defer iter.Close()
	commit, err := iter.Next()
	if err != nil {
		return "", err
	}
	return commit.Hash.String(), nil
}
func (g GitUtils) refNameIsHash() bool {
	return len(g.RefName) == 40
}
func (g GitUtils) findRepoFromHash(isBare bool) (*git.Repository, error) {
	repo, err := git.PlainClone(g.Folder, isBare, &git.CloneOptions{
		URL:  g.Url,
		Auth: g.AuthMethod,
	})
	if err != nil {
		return nil, err
	}
	tree, err := repo.Worktree()
	if err != nil {
		return nil, err
	}
	err = tree.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(g.RefName),
		Force: true,
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}
func (g GitUtils) findRepo(isBare bool) (*git.Repository, error) {
	if g.refNameIsHash() {
		return g.findRepoFromHash(isBare)
	}
	var repo *git.Repository
	var err error
	for _, refType := range refTypes {
		repo, err = git.PlainClone(g.Folder, isBare, &git.CloneOptions{
			URL:          g.Url,
			SingleBranch: true,
			Auth:         g.AuthMethod,
			ReferenceName: plumbing.ReferenceName(fmt.Sprintf(
				"refs/%s/%s",
				refType,
				strings.ToLower(g.RefName),
			)),
			Depth: 1,
		})
		if err == nil {
			return repo, nil
		}
		if strings.Contains(err.Error(), "couldn't find remote ref") {
			os.RemoveAll(g.Folder)
			os.Mkdir(g.Folder, 0777)
			continue
		}
		return nil, err
	}
	return repo, err
}
