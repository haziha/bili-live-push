// 提供直播间ID函数 和 接收直播间消息函数 可以写同一个文件

function echo(...arg) {
}

function http_request(req_obj) {
    // 加载JS代码时会重设该函数
    /*
    * req_obj: {
    *   Url: "",
    *   Method: "",
    *   Params: {k1: [v1, v2, ...], ...},
    *   Body: array,
    *   Headers: {k1: v1, k2: v2, ...},
    *   Proxy: {
    *       "Http": "...",
    *       "Https": "...",
    *   }
    * }
    * */
}

function bytes2string() {
    // 加载JS代码时会重设该函数
}

function string2bytes() {
    // 加载JS代码时会重设该函数
}

function get_rooms_id() {
    return [
        "21457197"
    ]
}


function on_message(message_obj) {
    /*
    * message_obj结构 在 message.go 中有定义
    *
    * message_obj: {
    *   MessageType: 1: 开播, 2: 下播
    *   FromType: 1: WS方式通知, 2: HTTP轮询方式通知
    *   RoomId: 直播间ID
    *   RealRoomId: 直播间真实ID
    * }
    * */
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

    let resp = http_request({
        Url: "https://api.bilibili.com/x/web-interface/nav",
        Method: "GET",
        Proxy: {
            Http: "http://127.0.0.1:7890",
        }
    })
    echo(bytes2string(resp))
}