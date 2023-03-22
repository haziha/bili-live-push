package main

import (
	"context"
	"testing"
)

func TestNewWsManage(t *testing.T) {
	manage := NewWsManage()
	manage.AddRoomId("21457197")

	<-context.Background().Done()
}
