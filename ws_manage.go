package main

import (
	"encoding/json"
	"fmt"
	"github.com/haziha/bili_live_ws_codec"
	"sync"
	"time"
)

type WsClient struct {
	client *bili_live_ws_codec.Client
}

type WsManage struct {
	rwLocker sync.RWMutex
	connects map[string]*WsClient

	ticker *time.Ticker
}

// AddRoomId 添加房间号
func (c *WsManage) AddRoomId(roomId string) {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	if c.connects == nil {
		return
	}

	if _, ok := c.connects[roomId]; ok {
		return
	}
	go c.reconnect(roomId)
}

// RemoveRoomId 删除房间号
func (c *WsManage) RemoveRoomId(roomId string) {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	if c.connects == nil {
		return
	}

	if conn, ok := c.connects[roomId]; ok {
		delete(c.connects, roomId)
		_ = conn.client.Close()
	} else {
		for rId, conn := range c.connects {
			if conn.client.RealRoomId().String() == roomId || conn.client.RoomId().String() == roomId {
				delete(c.connects, rId)
				break
			}
		}
	}
}

func (c *WsManage) readGoroutine(rId string, wsCli *WsClient) {
	packet := bili_live_ws_codec.Packet{}
	cli := wsCli.client
	for cli == wsCli.client {
		err := cli.ReadPacket(&packet)
		if err != nil {
			return
		}

		if packet.IsPvZlib() || packet.IsPvBrotli() {
			fmt.Println(rId, packet.PacketHeader)
			for {
				next, err := packet.DecompressNext()
				if !next || err != nil {
					break
				}
				if packet.Operation == bili_live_ws_codec.OpNormal {
					fmt.Println(rId, packet.PacketHeader, string(packet.Body))
				} else {
					fmt.Println(rId, packet.PacketHeader)
				}
			}
		} else {
			if popularity, ok := packet.IsOpHeartbeatReply(); ok {
				fmt.Println(rId, packet.PacketHeader, popularity)
			} else if packet.Operation == bili_live_ws_codec.OpNormal {
				fmt.Println(rId, packet.PacketHeader, string(packet.Body))
			} else {
				fmt.Println(rId, packet.PacketHeader)
			}
		}
	}
}

// reconnect 断开重连
func (c *WsManage) reconnect(roomId string) {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	var conn *WsClient
	var ok bool

	if c.connects == nil {
		return
	}

	if conn, ok = c.connects[roomId]; ok {
		delete(c.connects, roomId)
		_ = conn.client.Close()
	} else {
		conn = &WsClient{}
	}

	conn.client = bili_live_ws_codec.NewClient(json.Number(roomId))
	err := conn.client.Connect()
	if err != nil {
		c.connects[roomId] = conn
	} else {
		c.connects[conn.client.RealRoomId().String()] = conn
		go c.readGoroutine(conn.client.RealRoomId().String(), conn)
	}
}

func (c *WsManage) heartbeat() {
	for {
		_, ok := <-c.ticker.C
		if !ok {
			return
		}

		func() {
			c.rwLocker.RLock()
			defer c.rwLocker.RUnlock()

			if c.connects == nil {
				return
			}

			packet := bili_live_ws_codec.Packet{}
			packet.Heartbeat()

			for k, v := range c.connects {
				err := v.client.WritePacket(&packet)
				if err != nil {
					go c.reconnect(k)
				}
			}
		}()
	}
}

func (c *WsManage) Stop() {
	c.rwLocker.Lock()
	defer c.rwLocker.Unlock()

	c.ticker.Stop()

	for rId, conn := range c.connects {
		_ = conn.client.Close()
		conn.client = nil
		delete(c.connects, rId)
	}

	c.connects = nil
}

func NewWsManage() *WsManage {
	wm := &WsManage{
		connects: make(map[string]*WsClient),
		ticker:   time.NewTicker(time.Second * 15),
	}

	go wm.heartbeat()
	return wm
}
