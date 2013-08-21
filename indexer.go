// rubygems_indexer project indexer.go
package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

//gems相关配置
const (
	GEMS_FROM = "http://ruby.taobao.org/"    //上级rubygems源，注意末尾有斜杠，与官方格式保持一致
	GEMS_TO   = "/data/opensources/rubygems" //本地存放目录
)

var BASE_FILES = []string{
	"latest_specs.4.8.gz",
	"latest_specs.4.8",
	"prerelease_specs.4.8.gz",
	"prerelease_specs.4.8",
	"specs.4.8.gz",
	"specs.4.8",
	"Marshal.4.8.Z",
	"Marshal.4.8",
	"yaml",
	"yaml.Z",
	"quick/latest_index.rz",
}

const GEMSPECS_DIR = "quick/Marshal.4.8/"

func main() {
	fmt.Println("+------------------------------------+")
	fmt.Println("|       Rubygems Indexer v1.0        |")
	fmt.Println("|           by horsleyli             |")
	fmt.Println("+------------------------------------+")

	fetch_basefile()
	fetch_gemspecs()
}

func fetch_basefile() {
	fmt.Println("fetch basefiles start!")

	for _, filename := range BASE_FILES {
		fmt.Println(" -> " + filename)
		fetch(GEMS_FROM+filename, filepath.Join(GEMS_TO, filename))
	}

	fmt.Println("fetch basefiles end!")
}

func fetch_gemspecs() {
	fmt.Println("fetch gemspecs start!")

	os.MkdirAll(filepath.Join(GEMS_TO, GEMSPECS_DIR), 0700)
	filepath.Walk(filepath.Join(GEMS_TO, "gems"), func(path string, info os.FileInfo, err error) error {
		if info.Name() == "." || info.Name() == ".." {
			return nil
		}
		gem_name := info.Name()
		gemspec_name := gem_name[:len(gem_name)-len(filepath.Ext(gem_name))] + ".gemspec.rz"
		fmt.Println(" -> " + gemspec_name)
		fetch(GEMS_FROM+GEMSPECS_DIR+gemspec_name, filepath.Join(GEMS_TO, GEMSPECS_DIR, gemspec_name))
		return nil
	})

	fmt.Println("fetch gemspecs end!")
}

func fetch(url, to string) (err error) {
	var out *os.File
	var resp *http.Response
	if out, err = os.Create(to); err != nil {
		panic(err)
	}
	defer out.Close()

	if resp, err = http.Get(url); err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		_, err = io.Copy(out, resp.Body)
	}

	return
}
