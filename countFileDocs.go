package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"github.com/axgle/mahonia"
)

var sema = make(chan struct{}, 20)

const LineFeed  = "\r\n"

type NewFile struct {
	File os.FileInfo
	Dir string
}

func main()  {
	flag.Parse()
	roots := flag.Args() //需要统计的目录
	basePath, _ := GetCurrentPath()
	if len(roots) == 0 {
		roots = []string{basePath}
	}
	//一组协程
	var wg sync.WaitGroup
	//文件信息chan
	fileInfo := make(chan NewFile)
	for _, root := range roots{
		//一个根目录 一个协程
		wg.Add(1)
		//&wg 类似引用
		go WalkDir(root, &wg, fileInfo)
	}
	go func() {
		//等待协程 关闭协程
		wg.Wait()
		close(fileInfo)
	}()

	var fileCount, fileSize int64
	loop:
	for  {
		select {
		case  info, ok := <-fileInfo:
			if !ok {
				break loop
			}
			fStr := FormatInfo(info)
			err := WriteToFile(basePath+`2.csv`,fStr )
			if err != nil {
				fmt.Fprintf(os.Stderr, "write err %s \n", err)
			}
			fileCount+=1
			fileSize+=info.File.Size()
		}
	}
	enc:= mahonia.NewEncoder("gbk")
	fileCountInfo := fmt.Sprintf(`%s,%d,%.3f MB`,enc.ConvertString(`总计`), fileCount, float64(fileSize/(1024*1024)))
	err :=WriteToFile(basePath+`2.csv`, fileCountInfo)
	if err != nil {
		fmt.Fprintf(os.Stderr, "write err %s \n", err)
	}
}

func FormatInfo(fileInfo NewFile) string {
	enc:= mahonia.NewEncoder("gbk")
	var file  string
	file = fmt.Sprintf("%s,%s,%.2f KB,%s", enc.ConvertString(fileInfo.Dir), enc.ConvertString(fileInfo.File.Name()),float64(fileInfo.File.Size()/1024),fileInfo.File.ModTime().Format("2006-01-02 03:04:00"))
	return file
}


func WalkDir(dir string, wg *sync.WaitGroup, fileInfo chan<- NewFile)  {
	//退出前wg.done
	defer wg.Done()
	//打开多个目录列表
	for _, file := range ReadDir(dir){
		//fmt.Fprintf(os.Stdout, "dir %s, file %s \n", dir, file.Name())
		if file.IsDir() {
			//如果是目录 添加 go routine 多协程 继续递归调用
			wg.Add(1)
			subName := filepath.Join(dir, file.Name())
			go WalkDir(subName, wg, fileInfo)
		}else {
			fileNew := NewFile{
				File: file,
				Dir:  dir,
			}
			fileInfo <- fileNew
		}
	}
}

func ReadDir(dir string)[]os.FileInfo  {
	//这个不太清楚
	sema <- struct{}{}
	defer func() { <-sema }()

	fileInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "dir %s open err  %s", dir,err)
		return nil
	}
	return fileInfo
}

func WriteToFile(fileName string, content string) error {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	_, err = io.WriteString(f, LineFeed+content)

	defer f.Close()
	return err
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\".`)
	}
	return string(path[0 : i+1]), nil
}
