// 提供直播间ID函数 和 接收直播间消息函数 可以写同一个文件

function echo(...arg) {
}

function get_rooms_id() {
    return [
        "21457197"
    ]
}


function on_message(message_obj) {
    if (get_rooms_id().indexOf(message_obj["RoomId"]) === -1 &&
        get_rooms_id().indexOf(message_obj["RealRoomId"]) === -1) {
        echo("room id not in list:", message_obj["RoomId"], message_obj["RealRoomId"])
        return
    }
    let text = `message_type: ${message_obj["MessageType"]}, ` +
        `from_type: ${message_obj["FromType"]}, ` +
        `room_id: ${message_obj["RoomId"]}, ` +
        `real_room_id: ${message_obj["RealRoomId"]}`
    echo(text, message_obj)
}