// rubygems_indexer project indexer.go
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//gems相关配置默认值
const (
	GEMS_FROM    = "http://ruby.taobao.org/"    //上级rubygems源，注意末尾有斜杠，与官方格式保持一致
	GEMS_TO      = "/data/opensources/rubygems" //本地存放目录
	GEMSPECS_DIR = "quick/Marshal.4.8/"
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

var from, to string
var force bool

func main() {
	fmt.Println("+------------------------------------+")
	fmt.Println("|       Rubygems Indexer v1.0        |")
	fmt.Println("|           by horsleyli             |")
	fmt.Println("+------------------------------------+")

	flag.Usage = func() {
		fmt.Println("Usage: " + os.Args[0] + " [options]\n\nOptions:")
		fmt.Println("  -d, -destination\tthe local path which contains index files and the \"gems\" sub-directory.")
		fmt.Println("  -s, -source	\tthe upstream rubygems source url.")
		fmt.Println("  -f	\tforce download all(default update only)")
		//flag.PrintDefaults()
	}
	flag.StringVar(&from, "source", GEMS_FROM, "upstream rubygems source url.")
	flag.StringVar(&from, "s", GEMS_FROM, "shorthand for source")
	flag.StringVar(&to, "destination", GEMS_TO, "the local path which contains index files and the \"gems\" sub-directory.")
	flag.StringVar(&to, "d", GEMS_TO, "shorthand for destination")
	flag.BoolVar(&force, "f", false, "force download all")
	flag.Parse()

	if len(os.Args) == 1 {
		flag.Usage()
		return
	}
	fetch_basefile()
	fetch_gemspecs()
}

func fetch_basefile() {
	fmt.Println("fetch basefiles start!")

	for _, filename := range BASE_FILES {
		fmt.Print(" -> " + filename)
		status, _ := fetch(from+filename, filepath.Join(to, filename))
		fmt.Println("\t..." + status)
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

		fmt.Print(" -> " + gemspec_name)
		status, _ := fetch(from+GEMSPECS_DIR+gemspec_name, filepath.Join(to, GEMSPECS_DIR, gemspec_name))
		fmt.Println("\t..." + status)
		return nil
	})

	fmt.Println("fetch gemspecs end!")
}

func fetch(url, to string) (status string, err error) {
	var out *os.File
	var fileinfo os.FileInfo
	var resp *http.Response

	if !force { //检查本地文件的修改时间和上游服务器上的last-modified
		if out, err = os.Open(to); err != nil {
			panic(err)
		}
		defer out.Close()
		if fileinfo, err = out.Stat(); err != nil {
			panic(err)
		}

		if resp, err = http.Head(url); err != nil {
			panic(err)
		}

		last_mod, _ := time.Parse(time.RFC1123, resp.Header.Get("Last-Modified"))
		if fileinfo.ModTime().After(last_mod) {
			status = "skip"
			return
		}
	}
	if out, err = os.Create(to); err != nil {
		panic(err)
	}
	defer out.Close()

	if resp, err = http.Get(url); err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		status = resp.Status
		return
	}
	if _, err = io.Copy(out, resp.Body); err != nil {
		panic(err)
	}
	status = "ok"
	return
}
