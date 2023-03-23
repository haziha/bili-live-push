package main

import (
	"fmt"
	"github.com/dop251/goja"
	"log"
	"os"
)

type JsVmClient struct {
	runtime *goja.Runtime
	value   goja.Value

	GetRoomsId func() []string
	OnMessage  func(*Message)
}

func (c *JsVmClient) OnMessageSafe(message *Message) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("OnMessage panic: %v", e)
		}
	}()
	if c.OnMessage == nil {
		return nil
	}
	c.OnMessage(message)
	return
}

func (c *JsVmClient) GetRoomsIdSafe() (rooms []string, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("GetRoomsId panic: %v", e)
		}
	}()
	if c.GetRoomsId == nil {
		return []string{}, nil
	}
	rooms = c.GetRoomsId()
	return
}

func NewJsVmFromFile(fileName string) (*JsVmClient, error) {
	content, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	return NewJsVmFromString(string(content))
}

func NewJsVmFromString(content string) (*JsVmClient, error) {
	rt := goja.New()
	value, err := rt.RunString(content)
	if err != nil {
		return nil, err
	}

	jsVm := &JsVmClient{
		runtime: rt,
		value:   value,
	}

	getRoomsId := rt.Get("get_rooms_id")
	if getRoomsId != nil {
		if err := rt.ExportTo(getRoomsId, &jsVm.GetRoomsId); err != nil {
			jsVm.GetRoomsId = nil
		}
	}

	onMessage := rt.Get("on_message")
	if onMessage != nil {
		if err := rt.ExportTo(onMessage, &jsVm.OnMessage); err != nil {
			jsVm.OnMessage = nil
		}
	}

	if err = rt.Set("echo", log.Println); err != nil {
		return nil, err
	}

	return jsVm, nil
}
