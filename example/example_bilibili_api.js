// test 函数在加载js代码后自动执行一次
function test() {
    get_room_title()
}

function get_room_title() {
    const u = "https://api.live.bilibili.com/xlive/web-room/v1/index/getInfoByRoom"
    let resp = http_request({
        Url: u,
        Method: "GET",
        Params: {
            room_id: ["21457197"],
        }
    })
    if (!resp) {
        return
    }
    let data = JSON.parse(bytes2string(resp))
    if (data.code !== 0) {
        return
    }
    data = data.data
    let room_info = data["room_info"]
    let anchor_info = data["anchor_info"]
    let base_info = anchor_info["base_info"]

    // 主播信息
    let uid = room_info["uid"] // 主播id
    let uname = base_info["uname"]  // 主播名
    let face = base_info["face"] // 头像url
    let gender = base_info["gender"] // 性别

    // 直播间信息
    let room_id = room_info["room_id"] // 直播间id
    let short_id = room_info["short_id"] // 直播间短id, 无短id则为0
    let live_status = room_info["live_status"] // 直播状态 0: 下播, 1: 直播, 2: 轮播
    let title = room_info["title"] // 直播间标题
    let cover = room_info["cover"] // 直播封面url
    let keyframe = room_info["keyframe"] // 直播关键帧url

    echo({
        uid: uid,
        uname: uname,
        face: face,
        gender: gender,
        room_id: room_id,
        short_id: short_id,
        live_status: live_status,
        title: title,
        cover: cover,
        keyframe: keyframe,
    })
}
