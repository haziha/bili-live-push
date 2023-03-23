package main

import (
	"encoding/json"
	"github.com/haziha/bili_live_ws_codec"
	"log"
	"sync"
	"time"
)

type WsClient struct {
	client *bili_live_ws_codec.Client
}

type WsManage struct {
	rwLocker   sync.RWMutex
	connects   map[string]*WsClient
	outputChan chan *Message

	ticker *time.Ticker
}

// GetOutputChan 获取输出信息
func (c *WsManage) GetOutputChan() <-chan *Message {
	return c.outputChan
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

func (c *WsManage) readGoroutine(wsCli *WsClient) {
	packet := bili_live_ws_codec.Packet{}
	cli := wsCli.client
	defer func() {
		if cli != nil {
			_ = cli.Close()
		}
	}()
	for cli == wsCli.client && cli != nil {
		err := cli.ReadPacket(&packet)
		if err != nil {
			log.Printf("%s[%s]: 读取数据失败, 断开连接\n", cli.RoomId(), cli.RealRoomId().String())
			return
		}

		c.filter(cli, packet.DeepCopy())
	}
}

func (c *WsManage) filter(cli *bili_live_ws_codec.Client, packet *bili_live_ws_codec.Packet) {
	if packet.IsPvZlib() || packet.IsPvBrotli() {
		for {
			next, err := packet.DecompressNext()
			if !next || err != nil {
				break
			}
			c.filter(cli, packet.DeepCopy())
		}
	} else {
		if popularity, ok := packet.IsOpHeartbeatReply(); ok {
			log.Printf("%s[%s]: 人气值 %d\n", cli.RoomId(), cli.RealRoomId().String(), popularity)
		} else if packet.Operation == bili_live_ws_codec.OpNormal {
			body := struct {
				Cmd string `json:"cmd"`
			}{}
			if err := json.Unmarshal(packet.Body, &body); err == nil && body.Cmd != "" {
				log.Printf("%s[%s]: cmd=\"%s\"\n", cli.RoomId(), cli.RealRoomId().String(), body.Cmd)
				if body.Cmd == "LIVE" {
					select {
					case c.outputChan <- &Message{
						MessageType: MtLive,
						FromType:    FromWs,
						RoomId:      cli.RoomId().String(),
						RealRoomId:  cli.RealRoomId().String(),
					}:
					default:
					}
				} else if body.Cmd == "PREPARING" {
					select {
					case c.outputChan <- &Message{
						MessageType: MtPreparing,
						FromType:    FromWs,
						RoomId:      cli.RoomId().String(),
						RealRoomId:  cli.RealRoomId().String(),
					}:
					default:
					}
				}
			}
		} else if packet.Operation == bili_live_ws_codec.OpHeartbeatReply {
			log.Printf("%s[%s]: 心跳回应\n", cli.RoomId(), cli.RealRoomId().String())
		} else if packet.Operation == bili_live_ws_codec.OpJoinRoomReply {
			log.Printf("%s[%s]: 成功进入房间\n", cli.RoomId(), cli.RealRoomId().String())
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
		log.Printf("%s[%s]: 建立连接失败 (%v)\n", roomId, roomId, err)
	} else {
		c.connects[conn.client.RealRoomId().String()] = conn
		go c.readGoroutine(conn)
		log.Printf("%s[%s]: 建立连接成功\n", roomId, conn.client.RealRoomId().String())
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
		connects:   make(map[string]*WsClient),
		outputChan: make(chan *Message, 10),
		ticker:     time.NewTicker(time.Second * 15),
	}

	go wm.heartbeat()
	return wm
}
