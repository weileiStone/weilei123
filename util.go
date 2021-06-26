package utils

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"git-biz.qianxin-inc.cn/dlp/web/ceb/datadetective/tools/logsteward.git/log"
	"github.com/shirou/gopsutil/v3/process"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"time"
)

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

func CopyFile(dstName, srcName string) (written int64, err error) {
	src, err := os.Open(srcName)
	if err != nil {
		return 0, err
	}
	defer src.Close()
	dst, err := os.OpenFile(dstName, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return 0, err
	}
	defer dst.Close()
	return io.Copy(dst, src)
}

func WriteToFile(fileName string, content string) error {
	f, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	if err != nil {
		return err
	} else {
		// offset
		//os.Truncate(filename, 0) //clear
		n, _ := f.Seek(0, os.SEEK_END)
		_, err = f.WriteAt([]byte(content), n)
		defer f.Close()
	}
	return err
}

func ReadLine(lineNumber, offset int, filePath string) []string {
	var content []string
	if offset == 0 {
		return content
	}
	file, _ := os.Open(filePath)
	defer file.Close()
	fileScanner := bufio.NewScanner(file)
	lineCount := 1
	for fileScanner.Scan() {
		//只读lineNumber之后的行
		if lineCount >= lineNumber {
			content = append(content, fileScanner.Text())
		}
		//当读取的行超过offset 退出循环
		if lineCount >= offset {
			break
		}
		lineCount++
	}
	return content
}

func FileIsExist(filepath string) bool {
	_, err := os.Lstat(filepath)
	return !os.IsNotExist(err)
}

func IsNil(i interface{}) bool {

	vi := reflect.ValueOf(i)

	if vi.Kind() == reflect.Ptr {
		return vi.IsNil()
	}

	return false
}

func RemoveFile(dirs []string) {
	for _, dir := range dirs {
		if (dir != `./`) && (dir != `../`) {
			err := os.Remove(dir)
			if err != nil {
				log.Errorf("This is error message: %v", err)
			}
		}

	}

}

func Post(url string, data interface{}, contentType string) (string, error) {
	jsonStr, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Add("content-type", contentType)
	if err != nil {
		log.Errorf(`http post err %s`, err)
		return ``, err
	}
	tr := &http.Transport{
		//TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer req.Body.Close()

	client := &http.Client{Timeout: 2 * time.Minute, Transport: tr}
	resp, error := client.Do(req)
	if error != nil {
		log.Errorf(`http post err %s`, error)
		return ``, error
	}
	defer resp.Body.Close()

	result, _ := ioutil.ReadAll(resp.Body)
	content := string(result)
	return content, nil
}

func processName() (pname []string) {
	pids, _ := process.Pids()
	for _, pid := range pids {
		pn, _ := process.NewProcess(pid)
		pName, _ := pn.Name()
		pname = append(pname, pName)
	}
	return pname
}

func ProcessIsRepat() bool {
	var count int
	for _, name := range processName() {
		if name == BaseProcessNme() {
			count++
		}
	}
	if count > 1 {
		return true
	}
	return false
}

func BaseProcessNme() string {
	pid := os.Getpid()
	prc, _ := process.NewProcess(int32(pid))
	pName, _ := prc.Name()
	return pName
}
