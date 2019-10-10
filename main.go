// ArmFight project ArmFight.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"sync"

	"github.com/gorilla/websocket"
)

var GId2DataMap = &sync.Map{}

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")
var upgrader = websocket.Upgrader{} // use default options

/*
	功能：机器人签到
	结果：出错返回错误
		 成功返回空
*/
func initSiginin(c *websocket.Conn, index int) error {
	log.Println("=========>Init Robot Signin============>")
	var cmdMsg CommandMsg
	cmdMsg.Type = ROBOT_SIGN_IN
	cmdMsg.FromId = ROBOT_PREFIX + fmt.Sprintf("%d", index)
	cmdMsg.NickName = ROBOT_PREFIX + fmt.Sprintf("%d", index)
	cmdMsg.Message = cmdMsg.NickName + " Signin..."
	writeMsg, err := json.Marshal(cmdMsg)
	err = c.WriteMessage(websocket.TextMessage, writeMsg)
	if err != nil {
		log.Println("write:", err)
		return err
	}
	return nil
}

/*
	功能：获取初始化数据
	结果：出错返回错误
		 成功返回空
*/
func reqInitData(c *websocket.Conn, cmdMsg CommandMsg) error {
	var newcmdMsg CommandMsg
	newcmdMsg.Type = REQ_INIT_DATA
	newcmdMsg.FromId = cmdMsg.ToId
	newcmdMsg.ToId = cmdMsg.FromId
	newcmdMsg.NickName = cmdMsg.NickName
	newcmdMsg.Message = "请求初始化数据..."
	msg, err := json.Marshal(newcmdMsg)
	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return err
	}
	return nil
}

/*
	功能：初始化本地数据
	结果：出错返回错误
		 成功返回空
*/
func initData(userId string, msgBuf string) error {
	log.Println("=======>", userId, msgBuf)
	m := make(map[string]CardInfo)
	if err := json.Unmarshal([]byte(msgBuf), &m); err != nil {
		log.Println(err)
		return err
	}
	cardSlice := make([]CardInfo, 0)
	for k, v := range m {
		for i := 0; i < v.Count; i++ {
			v.CardId = k
			cardSlice = append(cardSlice, v)
		}
	}
	GId2DataMap.Store(userId, cardSlice)
	return nil
}

/*
	功能：根据策略获取出牌
	结果：出错返回错误
		 成功返回空
*/
func getCard(userId string) CardInfo {
	cardObj, ok := GId2DataMap.Load(userId)
	if !ok {
		log.Println(userId, "缓存信息没有获取到")
	}
	cardList, ret := cardObj.([]CardInfo)
	if !ret {
		log.Println("类型断言错误")
	}
	index := rand.Intn(len(cardList))
	card := cardList[index]
	log.Println("出牌之前,牌列表长度==>", len(cardList))
	cardList = append(cardList[:index], cardList[index+1:]...)
	log.Println("出牌之后,牌列表长度==>", len(cardList), card)
	GId2DataMap.Store(userId, cardList)
	return card
}

/*
	功能：回答同意
	结果：出错返回错误
		 成功返回空
*/
func replyYes(c *websocket.Conn, cmdMsg CommandMsg) error {
	var newcmdMsg CommandMsg
	newcmdMsg.Type = REQ_PLAY_YES
	newcmdMsg.FromId = cmdMsg.ToId
	newcmdMsg.ToId = cmdMsg.FromId
	newcmdMsg.NickName = cmdMsg.NickName
	newcmdMsg.Message = "同意你的请求"
	msg, err := json.Marshal(newcmdMsg)
	err = c.WriteMessage(websocket.TextMessage, msg)
	if err != nil {
		log.Println("write:", err)
		return err
	}
	return nil
}

/*
	机器人出牌
*/
func playCard(c *websocket.Conn, card CardInfo, cmdMsg CommandMsg) {
	var newcmdMsg CommandMsg
	newcmdMsg.Type = PLAY_CARD
	newcmdMsg.FromId = cmdMsg.ToId
	newcmdMsg.ToId = cmdMsg.FromId
	newcmdMsg.NickName = cmdMsg.NickName
	newcmdMsg.SCore = card.SCore
	newcmdMsg.Message = card.Name
	msg, _ := json.Marshal(newcmdMsg)
	c.WriteMessage(websocket.TextMessage, msg)
}

/*
	主流程处理
*/
func procHandle(c *websocket.Conn) {
	log.Println("=========>mainHandle============>")
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			continue
		}
		var cmdMsg CommandMsg
		if err = json.Unmarshal(message, &cmdMsg); err != nil {
			log.Println("Unmarshal:", err)
			continue
		}
		switch cmdMsg.Type {
		case PLAY_CARD_RESP:
			log.Println("收到出牌指令的应答", cmdMsg.Message)
		case SIGN_IN_RESP:
			log.Println(cmdMsg.FromId + "签到成功...")
			reqInitData(c, cmdMsg)
		case REQ_PLAY:
			log.Println("收到求战请求")
			replyYes(c, cmdMsg)
		case REQ_PLAY_YES_RESP:
			log.Println("收到求战同意的应答")
		case REQ_PLAY_CARD:
			log.Println("收到出牌指令:", cmdMsg.Message, cmdMsg.ToId)
			card := getCard(cmdMsg.ToId)
			playCard(c, card, cmdMsg)
		case REQ_INIT_DATA_RESP:
			log.Println("接受服务器的初始数据:", cmdMsg.FromId, cmdMsg.ToId)
			initData(cmdMsg.ToId, cmdMsg.Message)
		case START_GAME:
			log.Println("开始游戏:", cmdMsg)
		case OFFLINE_MSG:
			log.Println("下线通知:", cmdMsg)
		}
	}
}

func main() {
	log.Println("=============>ArmyRobot Starting....")
	wg := sync.WaitGroup{}
	wg.Add(1)
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
