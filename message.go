package main

type MessageType int

const (
	MtLive      MessageType = 1 + iota // 开播
	MtPreparing                        // 下播
)

type FromType int

const (
	FromWs FromType = 1 + iota
	FromHttp
)

type Message struct {
	MessageType MessageType
	FromType    FromType
	RoomId      string
	RealRoomId  string
}
