package main

type Manage struct {
	jsVmManage *JsVmManage
	wsManage   *WsManage
	httpManage *HttpIntervalManage
	closedChan chan struct{}

	roomsId map[string]bool
}

func (c *Manage) readMessage() {
	for {
		select {
		case message := <-c.wsManage.GetOutputChan():
			c.notify(message)
		case message := <-c.httpManage.GetOutputChan():
			c.notify(message)
		case <-c.closedChan:
			return
		}
	}
}

func (c *Manage) notify(message *Message) {
	if _, ok := c.roomsId[message.RealRoomId]; !ok {
		c.roomsId[message.RealRoomId] = false
	}

	if c.roomsId[message.RealRoomId] && message.MessageType == MtPreparing {
		c.roomsId[message.RealRoomId] = false
		c.jsVmManage.Broadcast(message)
	} else if !c.roomsId[message.RealRoomId] && message.MessageType == MtLive {
		c.roomsId[message.RealRoomId] = true
		c.jsVmManage.Broadcast(message)
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
		c.httpManage.AddRoomId(roomId)
	}
	for _, roomId := range connRoomsId {
		if _, ok := roomsId[roomId.RoomId]; ok {
			continue
		} else if _, ok = roomsId[roomId.RealRoomId]; ok {
			continue
		} else {
			c.wsManage.RemoveRoomId(roomId.RealRoomId)
			c.httpManage.RemoveRoomId(roomId.RealRoomId)
		}
	}
}

func NewManage(jsVmManage *JsVmManage, wsManage *WsManage, httpManage *HttpIntervalManage) *Manage {
	manage := &Manage{
		jsVmManage: jsVmManage,
		wsManage:   wsManage,
		httpManage: httpManage,
		closedChan: make(chan struct{}),
		roomsId:    make(map[string]bool),
	}
	manage.ReloadRoomsId()
	go manage.readMessage()
	return manage
}
