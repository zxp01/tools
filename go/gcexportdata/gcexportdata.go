// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gcexportdata provides functions for locating, reading, and
// writing export data files containing type information produced by the
// gc compiler.  This package supports go1.7 export data format and all
// later versions.
//
// This package replaces the deprecated golang.org/x/tools/go/gcimporter15
// package, which will be deleted in October 2017.
//
// Although it might seem convenient for this package to live alongside
// go/types in the standard library, this would cause version skew
// problems for developer tools that use it, since they must be able to
// consume the outputs of the gc compiler both before and after a Go
// update such as from Go 1.7 to Go 1.8.  Because this package lives in
// golang.org/x/tools, sites can update their version of this repo some
// time before the Go 1.8 release and rebuild and redeploy their
// developer tools, which will then be able to consume both Go 1.7 and
// Go 1.8 export data files, so they will work before and after the
// Go update. (See discussion at https://github.com/golang/go/issues/15651.)
//
package gcexportdata // import "golang.org/x/tools/go/gcexportdata"

import (
	"bufio"
	"fmt"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"

	gcimporter "golang.org/x/tools/go/gcimporter15"
)

// Find returns the name of an object (.o) or archive (.a) file
// containing type information for the specified import path,
// using the workspace layout conventions of go/build.
// If no file was found, an empty filename is returned.
//
// A relative srcDir is interpreted relative to the current working directory.
//
// Find also returns the package's resolved (canonical) import path,
// reflecting the effects of srcDir and vendoring on importPath.
func Find(importPath string, srcDir string) (filename, path string) {
	return gcimporter.FindPkg(importPath, srcDir)
}

// NewReader returns a reader for the export data section of an object
// (.o) or archive (.a) file read from r.  The new reader may provide
// additional trailing data beyond the end of the export data.
func NewReader(r io.Reader) (io.Reader, error) {
	buf := bufio.NewReader(r)
	_, err := gcimporter.FindExportData(buf)
	// If we ever switch to a zip-like archive format with the ToC
	// at the end, we can return the correct portion of export data,
	// but for now we must return the entire rest of the file.
	return buf, err
}

// Read reads export data from in, decodes it, and returns type
// information for the package.
// The package name is specified by path.
// File position information is added to fset.
//
// Read may inspect and add to the imports map to ensure that references
// within the export data to other packages are consistent.  The caller
// must ensure that imports[path] does not exist, or exists but is
// incomplete (see types.Package.Complete), and Read inserts the
// resulting package into this map entry.
//
// On return, the state of the reader is undefined.
func Read(in io.Reader, fset *token.FileSet, imports map[string]*types.Package, path string) (*types.Package, error) {
	data, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, fmt.Errorf("reading export data for %q: %v", path, err)
	}
	_, pkg, err := gcimporter.BImportData(fset, imports, data, path)
	return pkg, err
}

// Write writes encoded type information for the specified package to out.
// The FileSet provides file position information for named objects.
func Write(out io.Writer, fset *token.FileSet, pkg *types.Package) error {
	_, err := out.Write(gcimporter.BExportData(fset, pkg))
	return err
}
