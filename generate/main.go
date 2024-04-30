package main

//go:generate go run . --root ..

import (
	"bytes"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

const targetRepo = "github.com/egonelbre/spanner"

type Vendor struct {
	From, To string
}

var vendors = []Vendor{
	{"spanner", ""},
	{"internal/fields", "internal/fields"},
	{"internal/protostruct", "internal/protostruct"},
	{"internal/trace", "internal/trace"},
	{"internal/testutil", "internal/testutil"},
	{"internal/uid", "internal/uid"},
}

func main() {
	rootdir := flag.String("root", "", "module root directory")
	flag.Parse()

	preserve := map[string]bool{
		".git":                  true,
		".gitmodules":           true,
		"generate":              true,
		"go.mod":                true,
		"value_jsonprovider.go": true,
		"google-cloud-go":       true,
		"LICENSE":               true,
	}

	for _, entry := range must(os.ReadDir(*rootdir)) {
		if preserve[entry.Name()] {
			continue
		}
		must0(os.RemoveAll(filepath.Join(*rootdir, entry.Name())))
	}

	for _, vendor := range vendors {
		copydir(
			filepath.Join(*rootdir, "google-cloud-go", filepath.FromSlash(vendor.From)),
			filepath.Join(*rootdir, filepath.FromSlash(vendor.To)),
		)
	}

	edit(filepath.Join(*rootdir, "value.go"), func(data []byte) []byte {
		data = bytes.ReplaceAll(data, []byte("\tjsoniter \"github.com/json-iterator/go\"\n"), []byte(""))
		data = bytes.ReplaceAll(data, []byte("jsoniter.Config"), []byte("jsoniter_Config"))
		data = bytes.ReplaceAll(data, []byte("jsoniter.ConfigCompatibleWithStandardLibrary"), []byte("jsoniter_ConfigCompatibleWithStandardLibrary"))
		return data
	})

	must(runat(*rootdir, "go", "get", "cloud.google.com/go/spanner@v1.61.0"))
	must(runat(*rootdir, "go", "mod", "tidy"))
}

func must0(err error) {
	if err != nil {
		panic(err)
	}
}

func must[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}

func runat(dir, cmd string, args ...string) (string, error) {
	c := exec.Command(cmd, args...)
	c.Dir = dir
	out, err := c.Output()
	return string(out), err
}

func copydir(srcdir, dstdir string) {
	must0(os.MkdirAll(dstdir, 0755))
	must0(filepath.WalkDir(srcdir, func(fpath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		base := must(filepath.Rel(srcdir, fpath))

		if d.IsDir() {
			if strings.HasSuffix(base, "pb") {
				return filepath.SkipDir
			}
			return os.MkdirAll(filepath.Join(dstdir, base), 0755)
		}

		dstfile := filepath.Join(dstdir, base)
		data := must(os.ReadFile(fpath))

		switch filepath.Ext(fpath) {
		case ".mod":
			data = bytes.Replace(data,
				[]byte("module cloud.google.com/go/spanner"),
				[]byte("module "+targetRepo),
				1,
			)
		case ".go":
			data = replaceImports(data, func(in string) (string, bool) {
				if nested, ok := strings.CutPrefix(in, "cloud.google.com/go/"); ok {
					if strings.HasSuffix(nested, "pb") {
						return in, false
					}

					for _, vendor := range vendors {
						if rest, ok := strings.CutPrefix(nested, vendor.From); ok {
							if vendor.To == "" {
								return targetRepo + rest, true
							}
							return targetRepo + "/" + vendor.To + rest, true
						}
					}
				}
				return in, false
			})
		}

		return os.WriteFile(dstfile, data, 0755)
	}))
}

func edit(path string, fn func(data []byte) []byte) {
	stat := must(os.Stat(path))
	data := must(os.ReadFile(path))
	data = fn(data)
	must0(os.WriteFile(path, data, stat.Mode()))
}

func replaceImports(data []byte, replace func(in string) (string, bool)) []byte {
	fset := token.NewFileSet() // positions are relative to fset
	f := must(parser.ParseFile(fset, "src.go", data, parser.ParseComments))

	changed := false
	for _, imp := range f.Imports {
		if newPath, ok := replace(importPath(imp)); ok {
			changed = true
			imp.EndPos = imp.End()
			imp.Path.Value = strconv.Quote(newPath)
		}
	}

	for _, comgroup := range f.Comments {
		for _, com := range comgroup.List {
			if canImp, ok := strings.CutPrefix(com.Text, "// import "); ok {
				if newPath, ok := replace(must(strconv.Unquote(canImp))); ok {
					changed = true
					com.Text = "// import " + strconv.Quote(newPath)
				}
			}
		}
	}

	if !changed {
		return data
	}

	var out bytes.Buffer
	must0(format.Node(&out, fset, f))

	return out.Bytes()
}

// importPath returns the unquoted import path of s,
// or "" if the path is not properly quoted.
func importPath(s *ast.ImportSpec) string {
	t, err := strconv.Unquote(s.Path.Value)
	if err == nil {
		return t
	}
	return ""
}
