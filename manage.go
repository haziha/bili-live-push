package main

type Manage struct {
	jsVmManage *JsVmManage
	wsManage   *WsManage
	closedChan chan struct{}
}

func (c *Manage) readMessage() {
	for {
		select {
		case message := <-c.wsManage.GetOutputChan():
			c.jsVmManage.Broadcast(message)
		case <-c.closedChan:
			return
		}
	}
}

func (c *Manage) Close() {
	close(c.closedChan)
}

func (c *Manage) ReloadRoomsId() {
	roomsId := c.jsVmManage.GetRoomsId()
	connRoomsId := c.wsManage.GetRoomsId()
	for roomId := range roomsId {
		c.wsManage.AddRoomId(roomId)
	}
	for _, roomId := range connRoomsId {
		if _, ok := roomsId[roomId.RoomId]; ok {
			continue
		} else if _, ok = roomsId[roomId.RealRoomId]; ok {
			continue
		} else {
			c.wsManage.RemoveRoomId(roomId.RealRoomId)
		}
	}
}

func NewManage(jsVmManage *JsVmManage, wsManage *WsManage) *Manage {
	manage := &Manage{
		jsVmManage: jsVmManage,
		wsManage:   wsManage,
		closedChan: make(chan struct{}),
	}
	manage.ReloadRoomsId()
	go manage.readMessage()
	return manage
}
