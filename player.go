package main

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type Player struct {
	conn       *websocket.Conn
	name       string
	token      string // token唯一确定一个角色
	heartTimer *time.Timer
}

var playerCache = new(sync.Map)

type Message struct {
	Name  string                 `json:"name"`
	Reply string                 `json:"reply,omitempty"`
	Data  map[string]interface{} `json:"data"`
}

func (player *Player) OnConnect(conn *websocket.Conn) {
	player.conn = conn
	closeFunc := func() { _ = player.conn.Close() }
	player.heartTimer = time.AfterFunc(time.Minute, closeFunc)
	defer closeFunc()
	for {
		mt, buf, err := player.conn.ReadMessage()
		if err != nil {
			log.WithError(err).Error("read failed")
			break
		}
		if mt != websocket.TextMessage {
			log.Warn("unsupported message type: ", mt)
			continue
		}
		var message *Message
		if err = json.Unmarshal(buf, &message); err != nil {
			log.WithError(err).Error("unmarshal json failed")
			continue
		}
		if len(message.Name) == 0 {
			log.Error("no proto name")
			continue
		}
		handler := handlers[message.Name]
		if handler == nil {
			log.Warn("can not find handler: ", message.Name)
			continue
		}
		log.Debug("recv@", player.conn.RemoteAddr(), ": ", string(buf))
		handler(player, message.Data)
	}
}

func (player *Player) Send(message *Message) {
	buf, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("marshal json failed")
		return
	}
	if err := player.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
		log.WithError(err).Error("write failed")
		return
	}
	log.Debug("send@", player.conn.RemoteAddr(), ": ", string(buf))
}
