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
	room       *Room
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
	defer func() {
		_ = player.conn.Close()
		if len(player.token) > 0 {
			playerCache.Delete(player.token)
		}
	}()
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
		log.WithField("addr", player.conn.RemoteAddr().String()).Debug("recv: ", string(buf))
		var message *Message
		if err = json.Unmarshal(buf, &message); err != nil {
			log.WithError(err).Error("unmarshal json failed")
			continue
		}
		if len(message.Name) == 0 {
			log.Error("no proto name")
			continue
		}
		player.Handle(message.Name, message.Data)
	}
}

func (player *Player) Handle(name string, data map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.WithError(err).Error("panic")
			} else {
				log.Error("panic: ", r)
			}
			player.SendError(name, 500, "internal server error")
		}
	}()
	handler := handlers[name]
	if handler == nil {
		log.Warn("can not find handler: ", name)
		player.SendError(name, 404, "404 not found")
		return
	}
	if err := handler(player, name, data); err != nil {
		log.WithError(err).Error("handle failed: ", name)
		return
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
	log.WithField("addr", player.conn.RemoteAddr().String()).Debug("send: ", string(buf))
}

func (player *Player) SendError(reply string, code int, msg string) {
	player.Send(&Message{
		Name:  "error_sc",
		Reply: reply,
		Data: map[string]interface{}{
			"code": code,
			"msg":  msg,
		},
	})
}
