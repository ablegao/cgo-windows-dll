package main

import "C"

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	_ "net/http/pprof"
	"syscall"
	"unsafe"

	"runtime"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	debugData = flag.String("d", "", "Command data")
	serverT   = flag.String("s", "0.0.0.0:8000", "Address, if none , run in termail!")
	dllPath   = flag.String("p", "./", "Dll path .")
	cFunc     *syscall.Proc
)

func StrPtr(s string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}
func loadFunc(body string) string {

	// 声明一快内存空间, 必须明确大小，否则会因为编码规范差异造成溢出。
	out := make([]byte, 1024)
	// 取出指针地址
	p := &out[0]
	//输入数据转换成GBK编码
	input, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(body)), simplifiedchinese.GBK.NewEncoder()))

	//调用dll 反馈，  uintptr(unsafe.Pointer(p)) 将指针传给dll ,dll会给指针赋值.
	cFunc.Call(StrPtr(string(input)), uintptr(unsafe.Pointer(p)))
	output := []byte{}
	for _, x := range out {
		if x != 0 {
			output = append(output, x)
		} else {
			break
		}
	}
	//将结果转成UTF-8输出
	outdata, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader(output), simplifiedchinese.GBK.NewDecoder()))
	return string(outdata)
}

func runHttp(w http.ResponseWriter, r *http.Request) {
	buf, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	out := loadFunc(string(buf))
	log.Println("Request:", string(buf))
	w.WriteHeader(200)
	w.Write([]byte(out))
}

func main() {
	runtime.GOMAXPROCS(1)
	flag.Parse()

	//加载dll
	dll := syscall.MustLoadDLL(*dllPath + "hisinterface.dll")
	//加载 dll 函数
	cFunc = dll.MustFindProc("hisinterface")
	if *serverT == "" {
		fmt.Print(loadFunc(*debugData))
	} else {
		log.Println("Host:", *serverT)
		http.HandleFunc("/api", runHttp)
		log.Fatal(http.ListenAndServe(*serverT, nil))
	}
}
