/*
* 微信 wxpusher 推送例程
* */

const app_token = ""  // token
const uids = [] // uids

function on_message(message) {
    let data = {
        appToken: app_token,
        content: "",
        contentType: 3,
        uids: uids
    }

    // 需要用 Number 将 MessageType 转换一下, 或者用 `==`
    // 否则会为 false
    if (Number(message["MessageType"]) === 1) {
        // 开播
        data.content = `${message["RealRoomId"]} 开播`
    } else if (Number(message["MessageType"]) === 2) {
        // 下播
        data.content = `${message["RealRoomId"]} 下播`
    } else {
        return
    }

    let resp = http_request({
        Url: "https://wxpusher.zjiecode.com/api/send/message",
        Method: "POST",
        Body: string2bytes(JSON.stringify(data)),
        Headers: {
            "Content-Type": "application/json"
        }
    })
    echo(bytes2string(resp))
}
