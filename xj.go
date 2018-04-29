package main

/*
#include "stdio.h"
#include "stdlib.h"
#include <windows.h>

typedef int(*hand)(char*,char*);
HINSTANCE dll;
hand hisinterface;
char* output;

void cInit(char* path){
 dll = LoadLibrary(path);
 hisinterface = (hand)GetProcAddress(dll,"hisinterface");
}
void Free(){
    hisinterface = NULL;
    FreeLibrary(dll);
}
int cFunc( char * input , char * output){
  if(hisinterface==NULL){
    printf("%s","output;;;;;;; hisinterface is NULL");
    return -1;
  }
  return hisinterface(input,output);
}
*/
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

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var (
	debugData = flag.String("d", "", "Command data")
	serverT   = flag.String("s", "0.0.0.0:8000", "Address, if none , run in termail!")
	dllPath   = flag.String("p", "./", "Dll path .")
)

func StrPtr(s string) uintptr {
	return uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(s)))
}
func loadFunc(body string) string {
	output := C.CString(" ")
	//defer C.free(unsafe.Pointer(output))
	data, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(body)), simplifiedchinese.GBK.NewEncoder()))
	stdin := C.CString(string(data))
	defer C.free(unsafe.Pointer(output))
	defer C.free(unsafe.Pointer(stdin))
	C.cFunc(stdin, output)

	outdata, _ := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(C.GoString(output))), simplifiedchinese.GBK.NewDecoder()))
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
	flag.Parse()
	C.cInit(C.CString(*dllPath + "hisinterface.dll"))

	if *serverT == "" {
		fmt.Print(loadFunc(*debugData))
		//C.Free()
	} else {
		defer C.Free()
		log.Println("Host:", *serverT)
		http.HandleFunc("/api", runHttp)
		log.Fatal(http.ListenAndServe(*serverT, nil))
	}
}
