/*
* Telegram bot 推送例程
* */

const telegram_bot_token = ""  // 机器人token
const chat_id = ""  // 账号id

function on_message(message) {
    let data = {
        chat_id: chat_id,
        text: "",
    }

    // 需要用 Number 将 MessageType 转换一下, 或者用 `==`
    // 否则会为 false
    if (Number(message["MessageType"]) === 1) {
        // 开播
        data.text = `${message["RealRoomId"]} 开播`
    } else if (Number(message["MessageType"]) === 2) {
        // 下播
        data.text = `${message["RealRoomId"]} 下播`
    } else {
        return
    }

    let resp = http_request({
        Url: `https://api.telegram.org/bot${telegram_bot_token}/sendMessage`,
        Method: "POST",
        Body: string2bytes(JSON.stringify(data)),
        Headers: {
            "Content-Type": "application/json"
        }
    })
    echo(bytes2string(resp))
}
