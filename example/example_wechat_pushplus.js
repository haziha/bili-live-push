/*
* 微信 pushplus 推送例程
* */

const token = "" // token

function on_message(message) {
    let data = {
        token: token,
        content: "",
        template: "markdown"
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
        Url: "http://www.pushplus.plus/send",
        Method: "POST",
        Body: string2bytes(JSON.stringify(data)),
        Headers: {
            "Content-Type": "application/json"
        }
    })
    echo(bytes2string(resp))
}
