package main

import (
	"context"
	"flag"
	"fmt"
	"path/filepath"
)

var jsPath = flag.String("path", "", "存放JS文件的目录")

type stringArray []string

func (s *stringArray) String() string {
	arr := make([]string, len(*s))
	for i, str := range *s {
		arr[i] = str
	}
	return fmt.Sprintf("%v", arr)
}

func (s *stringArray) Set(val string) error {
	*s = append(*s, val)
	return nil
}

var jsFiles stringArray

var jsVmManage *JsVmManage
var wsManage *WsManage
var manage *Manage

func main() {
	flag.Var(&jsFiles, "file", "JS文件路径")
	flag.Parse()

	jsVmManage = NewJsVmManage()
	wsManage = NewWsManage()
	manage = NewManage(jsVmManage, wsManage)

	abs, err := filepath.Abs(*jsPath)
	if err != nil {
		fmt.Printf("abs js path failure: %s (%v)\n", *jsPath, err)
	} else {
		abs = filepath.Join(abs, "*.js")
		files, err := filepath.Glob(abs)
		if err != nil {
			fmt.Printf("get js files failure: %s (%v)\n", abs, err)
		} else {
			for _, file := range files {
				jsFiles = append(jsFiles, file)
			}
		}
	}

	for _, file := range jsFiles {
		err := jsVmManage.AddJsFile(file)
		if err != nil {
			fmt.Printf("add js file failure: %s (%v)\n", file, err)
		} else {
			fmt.Printf("add js file success: %s\n", file)
		}
	}

	manage.ReloadRoomsId()

	<-context.Background().Done()
}
