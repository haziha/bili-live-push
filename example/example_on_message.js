function echo(...arg) {
    // 加载JS代码时会重设该函数
}

function on_message(message_obj) {
    let text = `message_type: ${message_obj["MessageType"]}, ` +
        `from_type: ${message_obj["FromType"]}, ` +
        `room_id: ${message_obj["RoomId"]}, ` +
        `real_room_id: ${message_obj["RealRoomId"]}`
    echo(text, message_obj)
}