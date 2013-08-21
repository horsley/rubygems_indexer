// rubygems_indexer project indexer.go
package main

import (
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

//gems相关配置默认值
const (
	GEMS_FROM     = "http://ruby.taobao.org/"    //上级rubygems源，注意末尾有斜杠，与官方格式保持一致
	GEMS_TO       = "/data/opensources/rubygems" //本地存放目录
	GEMSPECS_DIR  = "quick/Marshal.4.8/"
	NUM_GOROUTINE = 50
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

var from, to, cur_file string
var force bool
var httpClient *http.Client

func main() {
	// 使用多核运行程序
	runtime.GOMAXPROCS(runtime.NumCPU())

	fmt.Println("+------------------------------------+")
	fmt.Println("|       Rubygems Indexer v1.2        |")
	fmt.Println("|           by horsleyli             |")
	fmt.Println("+------------------------------------+")

	flag.Usage = func() {
		fmt.Println("Usage: " + os.Args[0] + " [options]\n\nOptions:")
		fmt.Println("  -d, -destination\tthe local path which contains index files and the \"gems\" sub-directory.")
		fmt.Println("  -s, -source	\tthe upstream rubygems source url.")
		fmt.Println("  -f	\tforce download all(default update only)")
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

	httpClient = &http.Client{Transport: &http.Transport{MaxIdleConnsPerHost: 10}}

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

	os.MkdirAll(filepath.Join(to, GEMSPECS_DIR), 0700)
	//遍历目录得到文件列表
	filelist := make([]string, 0)
	filepath.Walk(filepath.Join(to, "gems"), func(path string, info os.FileInfo, err error) error {
		if info.Name() == "." || info.Name() == ".." {
			return nil
		}
		gem_name := info.Name()
		gemspec_name := gem_name[:len(gem_name)-len(filepath.Ext(gem_name))] + ".gemspec.rz"
		filelist = append(filelist, gemspec_name)
		return nil
	})

	//准备并发开跑
	filecount := len(filelist)
	finish := make(chan bool, filecount)
	for i := 0; i < NUM_GOROUTINE; i++ {
		go func(finish chan bool, i, n int) {
			if n > filecount {
				n = filecount //防止最后一次下标越界
			}
			for ; i < n; i++ {
				cur_file = filelist[i]
				fetch(from+GEMSPECS_DIR+filelist[i], filepath.Join(to, GEMSPECS_DIR, filelist[i]))
				finish <- true
			}

		}(finish, i*filecount/NUM_GOROUTINE, (i+1)*filecount/NUM_GOROUTINE)
	}

	max_statuslen := 0
	for i := 0; i < filecount; i++ {
		<-finish
		status_line := "[" + strconv.Itoa(i+1) + "/" + strconv.Itoa(filecount) + "] -> " + cur_file
		cur_statuslen := len(status_line)
		if cur_statuslen > max_statuslen {
			max_statuslen = len(status_line)
		}
		status_line += strings.Repeat(" ", max_statuslen-cur_statuslen) + "\r"
		fmt.Print(status_line)

	}

	fmt.Println("\nfetch gemspecs end!")
}

func fetch(url, to string) (status string, err error) {
	var out *os.File
	var fileinfo os.FileInfo
	var resp *http.Response

	defer func() {
		panic_err := recover()
		if panic_err != nil {
			if panic_err == "Head_Error" {
				time.Sleep(500 * time.Millisecond)
				if resp, err = httpClient.Head(url); err != nil { //retry
					panic("Head_Error")
				}
			} else {
				panic(panic_err)
			}
		}
	}()

	if !force { //检查本地文件的修改时间和上游服务器上的last-modified
		if out, err = os.Open(to); err != nil {
			panic(err)
		}
		defer out.Close()
		if fileinfo, err = out.Stat(); err != nil {
			panic(err)
		}

		if resp, err = httpClient.Head(url); err != nil {
			panic("Head_Error")
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

	if resp, err = httpClient.Get(url); err != nil {
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
