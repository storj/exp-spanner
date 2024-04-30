package main

//go:generate go run . --root ..

import (
	"bytes"
	"flag"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
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
		must0(os.RemoveAll(entry.Name()))
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
			return os.MkdirAll(filepath.Join(dstdir, base), 0755)
		}

		dstfile := filepath.Join(dstdir, base)
		data := must(os.ReadFile(fpath))

		if filepath.Ext(fpath) == ".mod" {
			data = bytes.Replace(data,
				[]byte("module cloud.google.com/go/spanner"),
				[]byte("module "+targetRepo),
				1,
			)
		} else {
			for _, vendor := range vendors {
				dst := path.Join(targetRepo, vendor.To)
				dst = strings.TrimSuffix(dst, "/")

				data = replaceImport(data,
					path.Join("cloud.google.com/go/", vendor.From),
					dst,
				)
			}
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

func replaceImport(data []byte, imp, rep string) []byte {
	// really hacky way to rewrite imports without parsing the code
	data = bytes.ReplaceAll(data, []byte(`	"`+imp), []byte(`	"`+rep))
	data = bytes.ReplaceAll(data, []byte(`, "`+imp), []byte(`<COMMASPACE>"`+imp))
	data = bytes.ReplaceAll(data, []byte(` "`+imp), []byte(` "`+rep))
	data = bytes.ReplaceAll(data, []byte(`<COMMASPACE>"`+imp), []byte(`, "`+imp))
	data = bytes.ReplaceAll(data, []byte(`import "`+imp), []byte(`import "`+rep))
	return data
}
