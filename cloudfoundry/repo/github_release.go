package repo

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/github"
)

// GithubRelease -
type GithubRelease struct {
	client *github.Client

	archivePath string

	owner    string
	repoName string

	archiveName string
}

// GetPath -
func (r *GithubRelease) GetPath() string {
	return r.archivePath
}

// SetVersion -
func (r *GithubRelease) SetVersion(version string, versionType VersionType) (err error) {

	var (
		release *github.RepositoryRelease

		url  string
		resp *http.Response

		in  io.ReadCloser
		out io.Writer

		zipIn  *zip.ReadCloser
		zipOut *zip.Writer

		gzipIn    *gzip.Reader
		tarIn     *tar.Reader
		tarHeader *tar.Header

		zipFile *os.File
		fh      *zip.FileHeader
	)

	ctx := context.Background()
	if release, _, err = r.client.Repositories.GetReleaseByTag(ctx, r.owner, r.repoName, version); err != nil {
		return
	}

	if r.archiveName == string(github.Zipball) {

		if resp, err = http.Get(*release.ZipballURL); err != nil {
			return
		}
		in = resp.Body

	} else if r.archiveName == string(github.Tarball) {

		if resp, err = http.Get(*release.TarballURL); err != nil {
			return
		}
		in = resp.Body

	} else {
		for _, asset := range release.Assets {

			if r.archiveName == *asset.Name {
				if in, url, err = r.client.Repositories.DownloadReleaseAsset(ctx, r.owner, r.repoName, *asset.ID); err != nil {
					return
				}
				if len(url) > 0 {
					if resp, err = http.Get(url); err != nil {
						return
					}
					in = resp.Body
				}
				break
			}
		}
	}

	if err = r.createArchive(in); err != nil {
		return
	}
	if r.archiveName == string(github.Zipball) {

		origArchive := r.archivePath + ".orig"
		if err = os.Rename(r.archivePath, origArchive); err != nil {
			return
		}
		defer os.Remove(origArchive)

		if zipIn, err = zip.OpenReader(origArchive); err != nil {
			return
		}
		defer zipIn.Close()

		if zipFile, err = os.Create(r.archivePath); err != nil {
			return
		}
		defer zipFile.Close()

		zipOut = zip.NewWriter(zipFile)
		defer zipOut.Close()

		for _, f := range zipIn.File {

			fh = &zip.FileHeader{}
			*fh = f.FileHeader
			fh.Name = f.Name[strings.Index(f.Name, "/")+1:]

			if len(fh.Name) > 0 {

				if out, err = zipOut.CreateHeader(fh); err != nil {
					return err
				}
				if !strings.HasSuffix(fh.Name, "/") {
					if in, err = f.Open(); err != nil {
						return
					}
					if _, err = io.Copy(out, in); err != nil {
						return
					}
				}
			}
		}

	} else if r.archiveName == string(github.Tarball) {

		origArchive := r.archivePath + ".orig"
		if err = os.Rename(r.archivePath, origArchive); err != nil {
			return
		}
		defer os.Remove(origArchive)

		if in, err = os.Open(origArchive); err != nil {
			return
		}
		defer in.Close()

		if gzipIn, err = gzip.NewReader(in); err != nil {
			return err
		}
		defer gzipIn.Close()
		tarIn = tar.NewReader(gzipIn)

		if zipFile, err = os.Create(r.archivePath); err != nil {
			return
		}
		defer zipFile.Close()

		zipOut = zip.NewWriter(zipFile)
		defer zipOut.Close()

		for {
			tarHeader, err = tarIn.Next()
			if err == io.EOF {
				err = nil
				break
			} else if err != nil {
				return err
			}

			if tarHeader.Name == "pax_global_header" {
				continue
			}

			fi := tarHeader.FileInfo()
			if fh, err = zip.FileInfoHeader(fi); err != nil {
				return

			}
			fh.Name = tarHeader.Name[strings.Index(tarHeader.Name, "/")+1:]

			if len(fh.Name) > 0 {
				if !fi.IsDir() {
					fh.Method = zip.Deflate
				}
				if out, err = zipOut.CreateHeader(fh); err != nil {
					return err
				}
				if !fi.IsDir() {
					if _, err = io.Copy(out, tarIn); err != nil {
						return
					}
				}
			}
		}
	}
	return
}

func (r *GithubRelease) createArchive(in io.ReadCloser) (err error) {

	out, err := os.Create(r.archivePath)
	defer in.Close()
	defer out.Close()

	if err == nil {
		_, err = io.Copy(out, in)
	}
	return
}
