package main

import (
	"context"
	"testing"
)

func TestNewManage(t *testing.T) {
	jsManage := NewJsVmManage()
	wsManage := NewWsManage()
	manage := NewManage(jsManage, wsManage)
	err := jsManage.AddJsFile("./example/example_full.js")
	if err != nil {
		panic(err)
	}
	manage.ReloadRoomsId()
	<-context.Background().Done()
}
