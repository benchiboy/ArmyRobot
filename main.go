// ArmFight project ArmFight.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

<<<<<<< HEAD
var addr = flag.String("addr", "localhost:8080", "http service address")
=======
var (
	GId2ConnMap = &sync.Map{}
	GId2IdMap   = &sync.Map{}
	GConn2IdMap = &sync.Map{}
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

/*
	玩家签到处理
	1:从数据库查询用户的信息 同步到内存中

*/
func signIn(c *websocket.Conn, playerType int, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SignIn============>")
	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Type = SIGN_IN_RESP
	cmdMsgResp.Success = true
	_, ok := GId2ConnMap.Load(cmdMsg.NickName)
	if ok {
		log.Println(cmdMsg.NickName + "用户已经在线")
		cmdMsgResp.Success = false
		cmdMsgResp.Message = "用户已经在线"
		return cmdMsgResp
	}
	GId2ConnMap.Store(cmdMsg.NickName, Player{CurrConn: c, SignInTime: time.Now(),
		NickName: cmdMsg.NickName, Status: STATUS_ONLIN_IDLE, PlayerType: playerType})
	GConn2IdMap.Store(c, cmdMsg.NickName)

	cmdMsgResp.FromId = cmdMsg.FromId
	cmdMsgResp.Message = "Sign In Success!"
	return cmdMsgResp
}

/*
	玩家发牌处理
*/

func playCard(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>PlayCard============>")

	cmdMsg.Role = getRole(cmdMsg.FromId)
	proxyMsg(c, cmdMsg)

	setCard(cmdMsg.FromId, cmdMsg.SCore, cmdMsg.Message)
	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Message = cmdMsg.Message
	cmdMsgResp.Role = getRole(cmdMsg.FromId)
	cmdMsgResp.Type = PLAY_CARD_RESP

	return cmdMsgResp
}

/*
	玩家发牌处理
*/

func reqPlayCard(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlayCard============>")

	proxyMsg(c, cmdMsg)

	var cmdMsgResp CommandMsgResp
	cmdMsgResp.Message = cmdMsg.Message
	cmdMsgResp.Type = REQ_PLAY_CARD_RESP

	return cmdMsgResp
}

/*
	玩家发牌对比结果
*/

func queryResult(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>queryResult============>")
	sScore, sCard := getCard(cmdMsg.FromId)
	mScore, mCard := getCard(cmdMsg.ToId)

	var cmdMsgResp CommandMsgResp
	if sScore > mScore {
		//出现炸弹，地址 和人员相碰的情况
		if sScore == 101 || sScore == 100 {
			if mScore == 0 {
				cmdMsgResp.Winner = "M"
			} else {
				cmdMsgResp.Winner = "B"
			}
		} else {
			cmdMsgResp.Winner = "S"
			if mScore == 0 {
				cmdMsgResp.Status = "E"
			}
		}
	} else if sScore == mScore {
		cmdMsgResp.Winner = "B"
	} else {
		//出现炸弹，地雷 和人员相碰的情况
		if mScore == 101 || mScore == 100 {
			if sScore == 0 {
				cmdMsgResp.Winner = "S"
			} else {
				cmdMsgResp.Winner = "B"
			}

		} else {
			cmdMsgResp.Winner = "M"
			if sScore == 0 {
				cmdMsgResp.Status = "E"
			}
		}
	}
	//出工兵和地理的情况
	if sScore == 1 && mScore == 100 {
		cmdMsgResp.Winner = "S"
	}
	if mScore == 1 && sScore == 100 {
		cmdMsgResp.Winner = "M"
	}
	//出炸弹和地雷的情况
	if sScore == 100 && mScore == 101 {
		cmdMsgResp.Winner = "B"
	}
	if mScore == 100 && sScore == 101 {
		cmdMsgResp.Winner = "B"
	}

	cmdMsgResp.Type = QUERY_RESULT_RESP
	cmdMsgResp.Message = mCard
	cmdMsgResp.AnotherMsg = sCard

	cmdMsgResp.Role = "M"

	proxyMsgResp(cmdMsg.ToId, cmdMsgResp)

	cmdMsgResp.Role = "S"
	cmdMsgResp.Message = sCard
	cmdMsgResp.AnotherMsg = mCard
	return cmdMsgResp
}

/*
	玩家发送消息处理
*/
func sendMsg(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SendMsg============>")

	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = SEND_MSG_RESP
	return cmdMsgResp
}

/*
	玩家发送语音
*/
func sendVoice(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>SendVoice============>")
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = SEND_VOICE_RESP
	return cmdMsgResp
}

/*
	得到用户列表
*/
func getUsers(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>getUsers============>")
	uList := make([]User, 0)
	var cmdMsgResp CommandMsgResp
	GId2ConnMap.Range(func(k, v interface{}) bool {
		p, _ := v.(Player)
		var u User
		log.Println(p.NickName)
		if cmdMsg.NickName != p.NickName {
			u.UserId = fmt.Sprintf("%s", k)
			u.NickName = p.NickName
			u.Avatar = p.Avatar
			u.Candy = p.Candy
			u.PlayerType = p.PlayerType
			u.Decoration = p.Decoration
			u.Icecream = p.Icecream
			u.LoginTime = p.SignInTime
			u.Memo = p.Memo
			u.Status = p.Status
			uList = append(uList, u)
		}
		return true
	})
	if userBuf, err := json.Marshal(uList); err != nil {
		log.Println(err)
	} else {
		cmdMsgResp.Message = string(userBuf)
		cmdMsgResp.Success = true
	}
	cmdMsgResp.Type = GET_USERS_RESP
	return cmdMsgResp
}

/*
	消息转发
*/
func proxyMsg(c *websocket.Conn, cmdMsg CommandMsg) {
	log.Println("=========>proxyMsg============>")
	log.Println(cmdMsg.FromId, cmdMsg.ToId)
	playerObj, ok := GId2ConnMap.Load(cmdMsg.ToId)
	if !ok {
		log.Println(cmdMsg.ToId + "缓存信息没有获取到")
		return
	}
	toPlayer, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
		return
	}
	if reqBuf, err := json.Marshal(cmdMsg); err != nil {
		log.Println(err)
	} else {
		if err := toPlayer.CurrConn.WriteMessage(websocket.TextMessage, []byte(string(reqBuf))); err != nil {
			log.Println("发送出错")
		}
	}
}

/*
	消息转发
*/
func proxyMsgResp(toId string, cmdMsgResp CommandMsgResp) {
	log.Println("=========>proxyMsgResp============>")
	playerObj, ok := GId2ConnMap.Load(toId)
	if !ok {
		log.Println("缓存信息没有获取到")
	}
	toPlayer, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	if reqBuf, err := json.Marshal(cmdMsgResp); err != nil {
		log.Println(err)
	} else {
		if err := toPlayer.CurrConn.WriteMessage(websocket.TextMessage, []byte(string(reqBuf))); err != nil {
			log.Println("发送出错")
		}
	}
}

/*
	改变用户通知
*/
func changeUser(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>changeUser============>", cmdMsg.FromId, cmdMsg.ToId)
	var cmdMsgResp CommandMsgResp

	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = CHANGE_USER_RESP
	return cmdMsgResp
}

/*
	请求玩家
*/
func reqPlay(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlay============>")
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Type = REQ_PLAY_RESP
	return cmdMsgResp
}

/*
	开始游戏
*/
func startGame(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>startGame============>")
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)

	cmdMsgResp.Type = START_GAME_RESP
	return cmdMsgResp
}

/*
	另一玩家答应请求
*/
func setRole(playerId string, role string) {
	log.Println("=========>setRole============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	player.Role = role
	GId2ConnMap.Store(playerId, player)
	//log.Println("=========>setRole============>", playerId)
}

/*
	另一玩家答应请求
*/
func getRole(playerId string) string {
	log.Println("=========>setRole============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	return player.Role
}
>>>>>>> 15af34ba88809cb2dcdc83c13fa744a378e935f1

/*
	登录成功后初始化数据
*/
func initSiginin(c *websocket.Conn, i int) {
	log.Println("=========>init============>")
	var cmdMsg CommandMsg
	cmdMsg.Type = ROBOT_SIGN_IN
	cmdMsg.FromId = "robot" + fmt.Sprintf("%d", i)
	cmdMsg.ToId = ""
	cmdMsg.NickName = "robot" + fmt.Sprintf("%d", i)
	cmdMsg.Message = "MSG..."
	msg, err := json.Marshal(cmdMsg)
	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
	return
}

/*
	登录成功后初始化数据
*/
<<<<<<< HEAD
func initData(c *websocket.Conn, i int, cmdMsg CommandMsg) {
	var newcmdMsg CommandMsg
	newcmdMsg.Type = REQ_INIT_DATA
	newcmdMsg.FromId = cmdMsg.ToId
	newcmdMsg.ToId = cmdMsg.FromId
	newcmdMsg.NickName = cmdMsg.NickName
	newcmdMsg.Message = "reply yes...."
	msg, err := json.Marshal(newcmdMsg)
	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
	return
=======
func setStatus(playerId string, status int) {
	//log.Println("=========>setStatus============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	player.Status = status

	GId2ConnMap.Store(playerId, player)
	//log.Println("=========>setStatus============>", playerId)
}

/*
	存储另一玩家的信息
*/
func setToNickName(playerId string, toNickName string) {
	//log.Println("=========>setStatus============>")
	playerObj, ok := GId2ConnMap.Load(playerId)
	if !ok {
		log.Println(playerId, "缓存信息没有获取到")
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
	}
	player.ToNickName = toNickName

	GId2ConnMap.Store(playerId, player)
	//log.Println("=========>setStatus============>", playerId)
}

/*
	另一玩家答应请求
*/
func reqPlayYes(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlayYes============>")
	var cmdMsgResp CommandMsgResp
	cmdMsg.Message = "对方同意了"

	proxyMsg(c, cmdMsg)

	//初始玩家双方的状态
	setStatus(cmdMsg.FromId, STATUS_ONLIE_READY)
	setStatus(cmdMsg.ToId, STATUS_ONLIE_READY)

	setToNickName(cmdMsg.FromId, cmdMsg.ToId)
	setToNickName(cmdMsg.ToId, cmdMsg.FromId)

	setRole(cmdMsg.FromId, "S")
	setRole(cmdMsg.ToId, "M")

	cmdMsgResp.Type = REQ_PLAY_YES_RESP
	return cmdMsgResp
}

/*
	另一玩家拒绝请求
*/
func reqPlayNo(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>reqPlayNo============>")
	var cmdMsgResp CommandMsgResp
	cmdMsg.Message = "对方拒绝了"
	proxyMsg(c, cmdMsg)

	cmdMsgResp.Type = REQ_PLAY_NO_RESP
	return cmdMsgResp
>>>>>>> 15af34ba88809cb2dcdc83c13fa744a378e935f1
}

/*
	登录成功后初始化数据
*/
<<<<<<< HEAD
func reqPlayYes(c *websocket.Conn, i int, cmdMsg CommandMsg) {
	var newcmdMsg CommandMsg
	newcmdMsg.Type = REQ_PLAY_YES
	newcmdMsg.FromId = cmdMsg.ToId
	newcmdMsg.ToId = cmdMsg.FromId
	newcmdMsg.NickName = cmdMsg.NickName
	newcmdMsg.Message = "reply yes...."
	msg, err := json.Marshal(newcmdMsg)
	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
	return
=======
func initData(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>initData============>")
	var cmdMsgResp CommandMsgResp
	cardMap := map[string]int{"gongbing": 3, "paizhang": 2, "lianzhang": 2, "yingzhang": 2,
		"tuanzhang": 2, "lvzhang": 2, "shizhang": 2, "junzhang": 2, "siling": 1, "junqi": 1, "dilei": 3, "zhadan": 2}
	mjson, _ := json.Marshal(cardMap)
	mString := string(mjson)
	fmt.Printf("print mString:%s", mString)
	cmdMsgResp.Message = mString
	cmdMsgResp.Type = REQ_INIT_DATA_RESP
	return cmdMsgResp
}

/*
	发起认输
*/
func giveUp(c *websocket.Conn, cmdMsg CommandMsg) CommandMsgResp {
	log.Println("=========>giveUp============>")
	var cmdMsgResp CommandMsgResp
	proxyMsg(c, cmdMsg)
	cmdMsgResp.Message = "放弃认输"
	cmdMsgResp.Type = REQ_GIVEUP_RESP
	return cmdMsgResp
}

/*
	另一玩家确认请求
*/
func disconnClear(c *websocket.Conn) {
	log.Println("=========>disconnClear============>")
	playIdObj, ok := GConn2IdMap.Load(c)
	if !ok {
		log.Println("缓存信息没有获取到")
		return
	}
	playId, ret := playIdObj.(string)
	if !ret {
		log.Println("类型断言错误")
		return
	}

	playerObj, ok := GId2ConnMap.Load(playId)
	if !ok {
		log.Println(playId, "缓存信息没有获取到")
		return
	}
	player, ret := playerObj.(Player)
	if !ret {
		log.Println("类型断言错误")
		return
	}

	//删除SN对应的缓存
	GId2ConnMap.Delete(playId)
	GConn2IdMap.Delete(c)
	log.Println("处理断开之后的清理")
	//如果此用户有关联用户，需要提醒对方
	if player.ToNickName != "" {
		var cmdMsg CommandMsg
		cmdMsg.Type = OFFLINE_MSG
		cmdMsg.FromId = playId
		cmdMsg.Message = "下线通知"
		cmdMsg.ToId = player.ToNickName
		proxyMsg(nil, cmdMsg)
	}

>>>>>>> 15af34ba88809cb2dcdc83c13fa744a378e935f1
}

/*
	机器人出牌
*/
<<<<<<< HEAD
func playCard(c *websocket.Conn, i int, cmdMsg CommandMsg) {
	var newcmdMsg CommandMsg
	newcmdMsg.Type = PLAY_CARD
	newcmdMsg.FromId = cmdMsg.ToId
	newcmdMsg.ToId = cmdMsg.FromId
	newcmdMsg.NickName = cmdMsg.NickName
	newcmdMsg.SCore = 1
	newcmdMsg.Message = "工兵"
	msg, err := json.Marshal(newcmdMsg)
	err = c.WriteMessage(websocket.TextMessage, msg)
=======

func gameHandle(w http.ResponseWriter, r *http.Request) {
	log.Println("==================>")
	c, err := upgrader.Upgrade(w, r, nil)

>>>>>>> 15af34ba88809cb2dcdc83c13fa744a378e935f1
	if err != nil {
		log.Println("write:", err)
		return
	}
	return
}

func procHandle(c *websocket.Conn) {
	log.Println("=========>mainHandle============>")
	for {
<<<<<<< HEAD
		_, message, err := c.ReadMessage()
=======

		mt, message, err := c.ReadMessage()
		log.Println(string(message))
>>>>>>> 15af34ba88809cb2dcdc83c13fa744a378e935f1
		if err != nil {
			log.Println("read:", err)
			return
		}
		var cmdMsg CommandMsg
		if err = json.Unmarshal(message, &cmdMsg); err != nil {
			log.Println("Unmarshal:", err)
		}
		switch cmdMsg.Type {
		case PLAY_CARD:
			log.Println("Play success...")
		case SIGN_IN_RESP:
			log.Println("SigninResp success...")
		case SIGN_IN:
<<<<<<< HEAD
			log.Println("SignIn success...")
=======
			cmdMsgResp = signIn(c, HUMAN_TYPE, cmdMsg)
		case ROBOT_SIGN_IN:
			cmdMsgResp = signIn(c, ROBOT_TYPE, cmdMsg)
>>>>>>> 15af34ba88809cb2dcdc83c13fa744a378e935f1
		case SEND_MSG:
			log.Println("SendMsg success...")
		case REQ_PLAY:
<<<<<<< HEAD
			log.Println("ReqPlay success...")
			reqPlayYes(c, 0, cmdMsg)
		case REQ_PLAY_YES_RESP:
			log.Println("ReqPlayYes Resp success...")
		case START_GAME:
			log.Println("START_GAME.....")
			initData(c, 0, cmdMsg)
		case REQ_INIT_DATA_RESP:
			log.Println("REQ_INIT_DATA_RESP.....", cmdMsg.Message)
		case REQ_PLAY_CARD:
			log.Println("REQ_PLAY_CARD.....", cmdMsg.Message)
			playCard(c, 0, cmdMsg)
		case PLAY_CARD_RESP:
			log.Println("PLAY_CARD_RESP.....", cmdMsg.Message)
=======
			cmdMsgResp = reqPlay(c, cmdMsg)
		case REQ_PLAY_YES:
			cmdMsgResp = reqPlayYes(c, cmdMsg)
		case REQ_PLAY_NO:
			cmdMsgResp = reqPlayNo(c, cmdMsg)
		case REQ_INIT_DATA:
			cmdMsgResp = initData(c, cmdMsg)
		case REQ_GIVEUP:
			cmdMsgResp = giveUp(c, cmdMsg)
		case START_GAME:
			cmdMsgResp = startGame(c, cmdMsg)
		case CHANGE_USER:
			cmdMsgResp = changeUser(c, cmdMsg)
		case REQ_PLAY_CARD:
			cmdMsgResp = reqPlayCard(c, cmdMsg)
		}
		msg, err := json.Marshal(cmdMsgResp)
		err = c.WriteMessage(mt, msg)
		log.Println("发送的消息：", mt, cmdMsgResp)
		if err != nil {
			log.Println("write:", err)
			break
>>>>>>> 15af34ba88809cb2dcdc83c13fa744a378e935f1
		}
	}
}

func main() {
	wg := sync.WaitGroup{}

	wg.Add(3)
	for i := 0; i < 1; i++ {
		u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
		c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
		if err != nil {
			log.Fatal("dial:", err)
		}
		log.Println("......")
		defer c.Close()

		go func(c *websocket.Conn, i int) {
			initSiginin(c, i)
			procHandle(c)
			wg.Done()
		}(c, i)
	}

	wg.Wait()

}
