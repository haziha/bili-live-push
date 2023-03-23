package main

import "testing"

func TestNewJsVmManage(t *testing.T) {
	manage := NewJsVmManage()
	err := manage.AddJsFile("./example/example_get_rooms_id.js")
	if err != nil {
		panic(err)
	}
	err = manage.AddJsFile("./example/example_on_message.js")
	if err != nil {
		panic(err)
	}
	err = manage.AddJsFile("./example/example_full.js")
	if err != nil {
		panic(err)
	}
	manage.Broadcast(&Message{MessageType: MtLive, FromType: FromWs, RoomId: "21457197", RealRoomId: "21457197"})
	_ = manage.RemoveJsFile("./example/example_on_message.js")
	manage.Broadcast(&Message{MessageType: MtLive, FromType: FromWs, RoomId: "-1", RealRoomId: "-1"})
}
