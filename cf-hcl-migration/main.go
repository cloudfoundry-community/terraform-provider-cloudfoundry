// This program was used for migrating tf files from version 0.9.9 to 0.10.0
// this is no longer useful but i let the code because it can be useful some other day to know how to do
package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl"
	"github.com/hashicorp/hcl/hcl/ast"
	"github.com/hashicorp/hcl/hcl/printer"
	"github.com/hashicorp/hcl/hcl/token"
)

var folderBits = "bits"

func main() {
	allpath := os.Args[1:]
	if len(allpath) == 0 {
		allpath = []string{"."}
	}
	allfiles := make([]string, 0)
	for _, p := range allpath {
		p := strings.TrimSuffix(strings.TrimSpace(p), "/")
		p = filepath.Join(p, "*.tf")
		files, err := filepath.Glob(p)
		if err != nil {
			panic(err)
		}
		allfiles = append(allfiles, files...)
	}
	for _, f := range allfiles {
		err := walkfile(f)
		if err != nil {
			log.Printf("Error for file '%s': %s", f, err.Error())
		}
	}

}

func walkfile(file string) error {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	hclAst, err := hcl.ParseBytes(b)
	if err != nil {
		return err
	}
	objList, ok := hclAst.Node.(*ast.ObjectList)
	if !ok {
		return nil
	}
	for _, item := range objList.Items {
		err := migrateBits(item)
		if err != nil {
			log.Printf("Error for file '%s': %s", file, err.Error())
		}
	}
	buf := &bytes.Buffer{}
	err = printer.Fprint(buf, hclAst)
	if err != nil {
		return err
	}
	info, _ := os.Stat(file)
	return ioutil.WriteFile(file, buf.Bytes(), info.Mode())
}

func migrateBits(item *ast.ObjectItem) error {
	if len(item.Keys) < 2 {
		return nil
	}
	if item.Keys[0].Token.Text != `resource` {
		return nil
	}
	if item.Keys[1].Token.Text != `"cloudfoundry_app"` &&
		item.Keys[1].Token.Text != `"cloudfoundry_buildpack"` {
		return nil
	}
	objList, ok := item.Val.(*ast.ObjectType)
	if !ok {
		return nil
	}

	items := objList.List.Items
	var pos token.Pos
	var assignPos token.Pos
	toDelete := make([]int, 0)
	var finalUrl *url.URL
	for index, item := range items {
		u, err := migrateBitsUrl(item)
		if err != nil {
			return err
		}
		if u != nil {
			finalUrl = u
			pos = item.Keys[0].Pos()
			assignPos = item.Assign
			toDelete = append(toDelete, index)
		}
		u, err = migrateBitsGit(item)
		if err != nil {
			return err
		}
		if u != nil {
			finalUrl = u
			pos = item.Keys[0].Pos()
			assignPos = item.Assign
			toDelete = append(toDelete, index-len(toDelete))
		}
		u, err = migrateBitsGithub(item)
		if err != nil {
			return err
		}
		if u != nil {
			finalUrl = u
			pos = item.Keys[0].Pos()
			assignPos = item.Assign
			toDelete = append(toDelete, index-len(toDelete))
		}
	}
	if finalUrl == nil {
		return nil
	}
	for _, index := range toDelete {
		items = append(items[:index], items[index+1:]...)
	}
	var finalPath string
	if (finalUrl.Scheme == "http" || finalUrl.Scheme == "https") && filepath.Ext(finalUrl.Path) == ".zip" {
		finalPath = finalUrl.String()
	} else {
		finalPath = filepath.Join(folderBits, outputPath(finalUrl))
	}
	items = append(items, &ast.ObjectItem{
		Keys: []*ast.ObjectKey{
			{
				Token: token.Token{
					Type: 4,
					Pos:  pos,
					Text: "path",
					JSON: false,
				},
			},
		},
		Assign: assignPos,
		Val: &ast.LiteralType{
			Token: token.Token{
				Type: 4,
				Pos:  pos,
				Text: quote(finalPath),
				JSON: false,
			},
		},
	})
	objList.List.Items = items

	return nil
}

func migrateBitsUrl(item *ast.ObjectItem) (*url.URL, error) {
	if len(item.Keys) == 0 {
		return nil, nil
	}
	if item.Keys[0].Token.Text != `url` {
		return nil, nil
	}
	tok := item.Val.(*ast.LiteralType).Token
	rawUrl := unquote(tok.Text)
	rawUrl = strings.TrimPrefix(rawUrl, "file://")
	return url.Parse(rawUrl)
}

func migrateBitsGit(item *ast.ObjectItem) (*url.URL, error) {
	if len(item.Keys) == 0 {
		return nil, nil
	}
	if item.Keys[0].Token.Text != `git` {
		return nil, nil
	}
	objList, ok := item.Val.(*ast.ObjectType)
	if !ok {
		return nil, nil
	}

	items := objList.List.Items
	for _, item := range items {
		if len(item.Keys) == 0 {
			continue
		}
		if item.Keys[0].Token.Text != `url` {
			continue
		}
		tok := item.Val.(*ast.LiteralType).Token
		rawUrl := unquote(tok.Text)
		return url.Parse(rawUrl)
	}
	return nil, nil
}

func migrateBitsGithub(item *ast.ObjectItem) (*url.URL, error) {
	if len(item.Keys) == 0 {
		return nil, nil
	}
	if item.Keys[0].Token.Text != `github_release` {
		return nil, nil
	}
	objList, ok := item.Val.(*ast.ObjectType)
	if !ok {
		return nil, nil
	}

	items := objList.List.Items
	var owner string
	var repo string
	var filename string
	var version string
	for _, item := range items {
		if len(item.Keys) == 0 {
			continue
		}
		if item.Keys[0].Token.Text == `owner` {
			owner = unquote(item.Val.(*ast.LiteralType).Token.Text)
		}
		if item.Keys[0].Token.Text == `repo` {
			repo = unquote(item.Val.(*ast.LiteralType).Token.Text)
		}
		if item.Keys[0].Token.Text == `filename` {
			filename = unquote(item.Val.(*ast.LiteralType).Token.Text)
		}
		if item.Keys[0].Token.Text == `version` {
			version = unquote(item.Val.(*ast.LiteralType).Token.Text)
		}
	}
	if repo == "" || owner == "" {
		return nil, nil
	}

	if version != "" && filename == "zipball" {
		return url.Parse(fmt.Sprintf("https://github.com/%s/%s/archive/%s.zip", owner, repo, version))
	} else if version != "" && filename == "tarball" {
		return url.Parse(fmt.Sprintf("https://github.com/%s/%s/archive/%s.tar.gz", owner, repo, version))
	} else if version != "" {
		return url.Parse(fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, version, filename))
	} else if filename == "zipball" {
		return url.Parse(fmt.Sprintf("https://glare.now.sh/%s/%s/zip", owner, repo))
	} else if filename == "tarball" {
		return url.Parse(fmt.Sprintf("https://glare.now.sh/%s/%s/tar", owner, repo))
	}

	return url.Parse(fmt.Sprintf("https://glare.now.sh/%s/%s/%s", owner, repo, filename))
}

func unquote(s string) string {
	s = strings.TrimSuffix(s, `"`)
	s = strings.TrimPrefix(s, `"`)
	return s
}

func quote(s string) string {
	return fmt.Sprintf(`"%s"`, s)
}

func outputPath(u *url.URL) string {

	path := u.Path
	if filepath.Base(path) == "zip" || filepath.Base(path) == "tar" {
		path = filepath.Join(filepath.Dir(path), "archive.zip")
	}
	if path != "" {
		path = strings.TrimSuffix(path, filepath.Ext(path)) + ".zip"
	}
	host := u.Host
	if host == "glare.now.sh" {
		host = "github.com"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if host == "" {
		return filepath.Base(path)
	}
	if path == "" {
		return filepath.FromSlash(host) + ".zip"
	}

	return filepath.FromSlash(strings.TrimSuffix(host, "/") + path)
}
