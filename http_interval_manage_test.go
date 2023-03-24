package main

import (
	"fmt"
	"testing"
)

func TestNewHttpIntervalManage(t *testing.T) {
	manage := NewHttpIntervalManage()
	manage.AddRoomId("21457197")
	for {
		fmt.Println(*<-manage.GetOutputChan())
	}
}
