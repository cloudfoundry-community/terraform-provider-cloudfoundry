# Zipper [![Build Status](https://travis-ci.org/ArthurHlt/zipper.svg?branch=master)](https://travis-ci.org/ArthurHlt/zipper) [![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT) [![GoDoc](https://godoc.org/github.com/ArthurHlt/zipper?status.svg)](https://godoc.org/github.com/ArthurHlt/zipper)

Library to create a zip file from different kind of source, this sources can be:
- [A git repository](#git)
- [An http url](#http)
- [A local folder](#local)

It also provides a fast way to know when your source change by using signature mechanism. 
(signatures are not made to verify validity they often try to only make a signature from a small chunk)

You can find a a cli tool directly on this project (see [cli section](#cli)).

## Install

Run `go get -u github.com/ArthurHlt/zipper/...`

## Usage

```go
package main

import (
    "github.com/ArthurHlt/zipper"
    "os"
    "io"
)

func main(){
    s, _ := zipper.CreateSession("/a/local/dir", "local") // instead of local you can choose: git or http 
    
    // zipper can auto detect what you want
    s, _ = zipper.CreateSession("/a/local/dir")
    // http support zip (useful for signature), tar or tgz
    // when not any tar, tgz or zip file will be store as it is inside a zi[
    s, _ = zipper.CreateSession("http://user:password@url.com/anexecutable") 
    s, _ = zipper.CreateSession("http://user:password@url.com/afile.zip")
    s, _ = zipper.CreateSession("http://user:password@url.com/afile.tar")
    s, _ = zipper.CreateSession("http://user:password@url.com/afile.tar.gz")
    
    // git support is powerful to target different branch or tag or a commit
    s, _ = zipper.CreateSession("http://user:password@github.com/ArthurHlt/zipper.git")
    s, _ = zipper.CreateSession("http://user:password@github.com/ArthurHlt/zipper.git#branch-or-tag-or-commit")
    
    // create the zip
    zipFile, _ := s.Zip() // zipFile implement io.ReadCloser
    defer zipFile.Close()
    
    zipFile.Size() // get the zip size
    
    // let's create the zip file on your fs
    f, _ := os.Create("myfile.zip")
    defer f.Close()
    io.Copy(f, zipFile)
    
    // Create the signature and use it to see if file change
    sig, _ := s.Sha1()
    
    isDiff, sourceSig, _ := s.IsDiff(sig)
    if isDiff {
        // source files have changed
        // let's update current signature
        sig = sourceSig
    }
}
```

## Source types

### Local

Zip from a local directory

- **Type Name**: `local`
- **Auto detection**: on an existing folder
- **Valid path**:
  - `/path/to/a/folder`
  - `/path/to/a/file.zip`
  - `/path/to/a/file.tar`
  - `/path/to/a/file.tar.gz`
  - `/path/to/a/file.tgz`
  - `/path/to/a/file.tar.bz2`
  - `/path/to/a/file.jar`
  - `/path/to/a/file.war`
- **Signature creation**: Create signature from the first 5kb of the final zip file.
  
**Tips**: 
- Creating a `.cfignore`, `.zipignore` or/and `.cloudignore` in `.gitignore` style will make 
zipper ignoring files which match pattern when zipping.
- Any valid zip content will be interpreted as zip (no need extension on a zip file to recognize it)
- Any valid `tar`, `tar.gz` or `tar.bz2` files will be converted as zip (no need extension recognize their types)

### Http

Zip from a `zip` (will be a full http stream in this case), `tar` or `tgz` file. 

- **Type Name**: `http`
- **Auto detection**: on an url with protocol `http` or `https`.
- **Valid path**:
  - `http://url.com/anyfile`
  - `http://url.com/afile.zip`
  - `http://url.com/afile.jar`
  - `http://url.com/afile.war`
  - `http://url.com/afile.tar`
  - `http://url.com/afile.tgz`
  - `http://url.com/afile.tar.gz`
- **Signature creation**: Create signature from the first 5kb on the remote file 
(this mean that when calling sha1 this will not download the entire file).

**Tips**: 
- You can pass user and password for basic auth.
- Excutable file (elf - linux executable, windows executable, macho - osx executable or file containing shebang) 
will be store with executable permission
- Any valid zip content will be interpreted as zip (no need extension on a zip file to recognize it)
- Any valid `tar`, `tar.gz` or `tar.bz2` files will be converted as zip (no need extension recognize their types)

### Git

Zip from a git repository

- **Type Name**: `local`
- **Auto detection**: on an url with protocol `http`, `https` or `ssh` and path ending with `.git`
- **Valid path**:
  - `http://github.com/ArthurHlt/zipper.git`
  - `ssh://git@github.com:ArthurHlt/zipper-fixture.git`
  - `ssh://git@github.com:ArthurHlt/zipper-fixture.git/folder/in/repo`
  - `ssh://git@github.com:ArthurHlt/zipper-fixture.git/folder/in/repo?private-key=/pass/to/pem/key&password-key=`
  - `http://github.com/ArthurHlt/zipper.git#branch-or-tag-or-commit`
  - `ssh://git@github.com:ArthurHlt/zipper-fixture.git#branch-or-tag-or-commit`
- **Signature creation**: From commit sha1 (use a [bare repo](http://www.saintsjd.com/2011/01/what-is-a-bare-git-repository/) to retrieve it faster).
  
**Tips**:
- Creating a `.cfignore` or/and a `.zipignore` in `.gitignore` style will make 
zipper ignoring files which match pattern when zipping. 
- You can pass user and password for basic auth.

## Cli

You can use a command line to use zipper directly.

If you have set your `PATH` with `$GOPATH/bin` you should have zipper available (run `zipper -h`).

See doc:

```
NAME:
   zipper - use zipper in cli

USAGE:
   zipper [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
     zip, z   create zip from a source
     sha1, s  Get sha1 signature for the file from source
     diff, s  Check if file from source is different from your stored sha1
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --type value, -t value  Choose source type
   --insecure, -k          Ignore certificate validation
   --help, -h              show help
   --version, -v           print the version
```