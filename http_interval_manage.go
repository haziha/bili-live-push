package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type HttpIntervalManage struct {
	rwLocker sync.RWMutex

	connects map[string]struct {
		RoomId     string
		RealRoomId string
		Living     bool
	}
	outputChan chan *Message

	ticker *time.Ticker
}

func (c *HttpIntervalManage) GetOutputChan() <-chan *Message {
	return c.outputChan
}

func (c *HttpIntervalManage) AddRoomId(roomId string) {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	c.connects[roomId] = struct {
		RoomId     string
		RealRoomId string
		Living     bool
	}{RoomId: roomId, RealRoomId: roomId, Living: false}
}

func (c *HttpIntervalManage) RemoveRoomId(roomId string) {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	for k, v := range c.connects {
		if v.RoomId == roomId || v.RealRoomId == roomId {
			delete(c.connects, k)
		}
	}
}

func (c *HttpIntervalManage) GetRoomsId() []struct {
	RoomId     string
	RealRoomId string
} {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()

	rooms := make([]struct {
		RoomId     string
		RealRoomId string
	}, 0)

	for _, v := range c.connects {
		rooms = append(rooms, struct {
			RoomId     string
			RealRoomId string
		}{RoomId: v.RoomId, RealRoomId: v.RealRoomId})
	}

	return rooms
}

func (c *HttpIntervalManage) interval() {
	c.rwLocker.RLock()
	ticker := c.ticker
	c.rwLocker.RUnlock()

	for ticker != nil && ticker == c.ticker {
		select {
		case <-ticker.C:
			c.traversal()
		}
	}
}

func (c *HttpIntervalManage) traversal() {
	c.rwLocker.RLock()
	defer c.rwLocker.RUnlock()

	for k := range c.connects {
		c.apiCheck(k)
	}
}

func (c *HttpIntervalManage) apiCheck(roomId string) {
	room := c.connects[roomId]
	u, err := url.Parse("https://api.live.bilibili.com/room/v1/Room/room_init")
	if err != nil {
		log.Printf("(%s)[%s] url parse failure (%v)\n", room.RoomId, room.RealRoomId, err)
		return
	}
	q := u.Query()
	q.Set("id", roomId)
	u.RawQuery = q.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		log.Printf("(%s)[%s] new request failure (%v)\n", room.RoomId, room.RealRoomId, err)
		return
	}
	cli := http.Client{}
	resp, err := cli.Do(req)
	if err != nil {
		log.Printf("(%s)[%s] request failure (%v)\n", room.RoomId, room.RealRoomId, err)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("(%s)[%s] read response failure (%v)\n", room.RoomId, room.RealRoomId, err)
		return
	}

	data := struct {
		Code    json.Number `json:"code"`
		Message string      `json:"message"`
		Data    struct {
			RoomId     json.Number `json:"room_id"`
			ShortId    json.Number `json:"short_id"`
			LiveStatus json.Number `json:"live_status"`
		} `json:"data"`
	}{}

	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Printf("(%s)[%s] body parse failure (%v)\n", room.RoomId, room.RealRoomId, err)
		return
	}

	if data.Code.String() != "0" {
		log.Printf("(%s)[%s] code failure (%v)\n", room.RoomId, room.RealRoomId, data.Message)
		return
	}

	if data.Data.ShortId.String() == "0" {
		data.Data.ShortId = json.Number(roomId)
	}
	if data.Data.RoomId != data.Data.ShortId {
		if _, ok := c.connects[data.Data.ShortId.String()]; ok {
			c.connects[data.Data.RoomId.String()] = c.connects[data.Data.ShortId.String()]
			delete(c.connects, data.Data.ShortId.String())
		}
	}

	room = c.connects[data.Data.RoomId.String()]
	room.RoomId = data.Data.ShortId.String()
	room.RealRoomId = data.Data.RoomId.String()
	if (data.Data.LiveStatus.String() == "0" || data.Data.LiveStatus.String() == "2") && room.Living {
		room.Living = false
		select {
		case c.outputChan <- &Message{
			MessageType: MtPreparing,
			FromType:    FromWs,
			RoomId:      room.RoomId,
			RealRoomId:  room.RealRoomId,
		}:
		default:
		}
		log.Printf("(%s)[%s] api check: preparing", room.RoomId, room.RealRoomId)
	} else if data.Data.LiveStatus.String() == "1" && !room.Living {
		room.Living = true
		select {
		case c.outputChan <- &Message{
			MessageType: MtLive,
			FromType:    FromHttp,
			RoomId:      room.RoomId,
			RealRoomId:  room.RealRoomId,
		}:
		default:
		}
		log.Printf("(%s)[%s] api check: live", room.RoomId, room.RealRoomId)
	}
	c.connects[data.Data.RoomId.String()] = room
}

func (c *HttpIntervalManage) Start() {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	c.connects = make(map[string]struct {
		RoomId     string
		RealRoomId string
		Living     bool
	})

	if c.outputChan != nil {
		close(c.outputChan)
	}
	c.outputChan = make(chan *Message, 10)
	c.ticker = time.NewTicker(time.Second * 60)
	go c.interval()
}

func (c *HttpIntervalManage) Stop() {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	c.connects = nil
	close(c.outputChan)
	c.outputChan = nil
	c.ticker.Stop()
	c.ticker = nil
}

func NewHttpIntervalManage() *HttpIntervalManage {
	manage := &HttpIntervalManage{}
	manage.Start()
	return manage
}
