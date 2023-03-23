package main

import (
	"fmt"
	"testing"
)

func TestNewJsVmFromFile(t *testing.T) {
	cli, err := NewJsVmFromFile("./example/example_get_rooms_id.js")
	if err != nil {
		panic(err)
	}
	roomsId, err := cli.GetRoomsIdSafe()
	if err != nil {
		panic(err)
	}
	fmt.Println(roomsId)
}

func TestNewJsVmFromFile2(t *testing.T) {
	cli, err := NewJsVmFromFile("./example/example_on_message.js")
	if err != nil {
		panic(err)
	}
	err = cli.OnMessageSafe(&Message{MessageType: MtLive, FromType: FromWs, RoomId: "21457197", RealRoomId: "21457197"})
	if err != nil {
		panic(err)
	}
}

func TestNewJsVmFromFile3(t *testing.T) {
	cli, err := NewJsVmFromFile("./example/example_full.js")
	if err != nil {
		panic(err)
	}
	err = cli.OnMessageSafe(&Message{MessageType: MtLive, FromType: FromWs, RoomId: "21457197", RealRoomId: "21457197"})
	if err != nil {
		panic(err)
	}
	err = cli.OnMessageSafe(&Message{MessageType: MtLive, FromType: FromWs, RoomId: "-1", RealRoomId: "-1"})
	if err != nil {
		panic(err)
	}
}
