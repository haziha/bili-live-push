package main

import (
	"log"
	"testing"
)

func TestNewWsManage(t *testing.T) {
	manage := NewWsManage()
	manage.AddRoomId("21457197")

	for {
		message := <-manage.GetOutputChan()
		log.Printf("message: %v\n", message)
	}
}
