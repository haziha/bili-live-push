package main

import (
	"os"
	"path/filepath"
	"sync"
)

type JsVmManage struct {
	rwLocker sync.RWMutex
	jsVms    map[string]*JsVmClient
}

func (c *JsVmManage) Broadcast(message *Message) {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()

	for _, vm := range c.jsVms {
		_ = vm.OnMessageSafe(message)
	}
}

func (c *JsVmManage) GetRoomsId() []string {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()

	roomsIdMap := map[string]struct{}{}

	for _, vm := range c.jsVms {
		rsId, err := vm.GetRoomsIdSafe()
		if err != nil {
			continue
		}
		for _, rId := range rsId {
			roomsIdMap[rId] = struct{}{}
		}
	}

	roomsId := make([]string, 0, len(roomsIdMap))
	for rId := range roomsIdMap {
		roomsId = append(roomsId, rId)
	}

	return roomsId
}

func (c *JsVmManage) AddJsFile(fileName string) error {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	fileName, err := filepath.Abs(fileName)
	if err != nil {
		return err
	}
	fileName = filepath.ToSlash(fileName)
	jvc, err := NewJsVmFromFile(fileName)
	if err != nil {
		return err
	}
	c.jsVms[fileName] = jvc
	return nil
}

func (c *JsVmManage) RemoveJsFile(fileName string) error {
	fi, err := os.Stat(fileName)
	if err != nil {
		return err
	}

	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	for k := range c.jsVms {
		kFi, err := os.Stat(k)
		if err != nil {
			continue
		}
		if os.SameFile(fi, kFi) {
			delete(c.jsVms, k)
			break
		}
	}

	return nil
}

func NewJsVmManage() *JsVmManage {
	return &JsVmManage{
		jsVms: make(map[string]*JsVmClient),
	}
}
