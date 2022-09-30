package main

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"sync"
	"time"
)

type Message struct {
	Name  string                 `json:"name"`
	Reply string                 `json:"reply,omitempty"`
	Data  map[string]interface{} `json:"data,omitempty"`
}

type PlayerConn struct {
	mu         sync.Mutex
	player     string
	conn       *websocket.Conn
	heartTimer *time.Timer
	syncTimer  *time.Ticker
	syncHash   uint32
}

var playerConnCache sync.Map

func (playerConn *PlayerConn) OnConnect(conn *websocket.Conn) {
	playerConn.conn = conn
	closeFunc := func() { _ = playerConn.conn.Close() }
	playerConn.heartTimer = time.AfterFunc(time.Minute, closeFunc)
	defer playerConn.OnDisconnect()
	for {
		mt, buf, err := playerConn.conn.ReadMessage()
		if err != nil {
			log.WithError(err).Error("read failed")
			break
		}
		if mt != websocket.TextMessage {
			log.Warn("unsupported message type: ", mt)
			continue
		}
		log.WithField("addr", playerConn.conn.RemoteAddr().String()).WithField("len", len(buf)).Debug("recv: ", string(buf))
		var message *Message
		if err = json.Unmarshal(buf, &message); err != nil {
			log.WithError(err).Error("unmarshal json failed")
			continue
		}
		if len(message.Name) == 0 {
			log.Error("no proto name")
			continue
		}
		playerConn.Handle(message.Name, message.Data)
	}
}

func (playerConn *PlayerConn) Handle(name string, data map[string]interface{}) {
	defer func() {
		if r := recover(); r != nil {
			if err, ok := r.(error); ok {
				log.WithError(err).Error("panic")
			} else {
				log.Error("panic: ", r)
			}
			playerConn.SendError(name, 500, "internal server error")
		}
	}()
	handler := handlers[name]
	if handler == nil {
		log.Warn("can not find handler: ", name)
		playerConn.SendError(name, 404, "404 not found")
		return
	}
	playerConn.heartTimer.Stop()
	playerConn.heartTimer = time.AfterFunc(time.Minute, func() { _ = playerConn.conn.Close() })
	if len(playerConn.player) == 0 && name != "login_cs" {
		playerConn.SendError(name, -1, "You haven't login.")
		return
	}
	if err := handler(playerConn, name, data); err != nil {
		playerConn.SendError(name, 500, err.Error())
		log.WithError(err).Error("handle failed: ", name)
	}
}

func (playerConn *PlayerConn) Send(message Message) {
	playerConn.mu.Lock()
	defer playerConn.mu.Unlock()
	buf, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("marshal json failed")
		return
	}
	if err := playerConn.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
		log.WithError(err).Error("write failed")
		return
	}
	log.WithField("addr", playerConn.conn.RemoteAddr().String()).WithField("len", len(buf)).Debug("send: ", string(buf))
}

func (playerConn *PlayerConn) SendSync(message Message) {
	playerConn.mu.Lock()
	defer playerConn.mu.Unlock()
	buf, err := json.Marshal(message)
	if err != nil {
		log.WithError(err).Error("marshal json failed")
		return
	}
	hash := stringHash(buf)
	if playerConn.syncHash == hash {
		return
	}
	playerConn.syncHash = hash
	if err := playerConn.conn.WriteMessage(websocket.TextMessage, buf); err != nil {
		log.WithError(err).Error("write failed")
		return
	}
	log.WithField("addr", playerConn.conn.RemoteAddr().String()).Debug("send: ", string(buf))
}

func (playerConn *PlayerConn) SendError(reply string, code int, msg string) {
	playerConn.Send(Message{
		Name:  "error_sc",
		Reply: reply,
		Data: map[string]interface{}{
			"code": code,
			"msg":  msg,
		},
	})
}

func (playerConn *PlayerConn) SendSuccess(reply string) {
	playerConn.Send(Message{
		Name:  "success_sc",
		Reply: reply,
	})
}

func (playerConn *PlayerConn) OnDisconnect() {
	playerConn.heartTimer.Stop()
	_ = playerConn.conn.Close()
	if playerConn.syncTimer != nil {
		playerConn.syncTimer.Stop()
	}
	if len(playerConn.player) == 0 {
		return
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.player)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return nil
		}
		room, err := GetRoom(txn, player.RoomId)
		if err != nil {
			return err
		}
		if room.GetStarted() {
			return nil
		}
		if room.Host == player.Token {
			for i := range room.Players {
				if room.Players[i] != room.Host {
					p, err := GetPlayer(txn, room.Players[i])
					if err != nil {
						return err
					}
					p.RoomId = ""
					p.Name = ""
					err = SetPlayer(txn, p)
					if err != nil {
						return err
					}
				}
			}
			if err = DelRoom(txn, player.RoomId); err != nil {
				return err
			}
		} else {
			for i := range room.Players {
				if room.Players[i] == player.Token {
					room.Players[i] = ""
				}
			}
			err = SetRoom(txn, room)
			if err != nil {
				return err
			}
		}
		return DelPlayer(txn, player.Token)
	})
	if err != nil {
		log.WithError(err).Error("on disconnect error")
	}
	playerConnCache.Delete(playerConn.player)
}

func isAlphaNum(s string) bool {
	for _, c := range []byte(s) {
		if (c < '0' || c > '9') && (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}
	return true
}

func stringHash(s []byte) (hash uint32) {
	for _, c := range s {
		ch := uint32(c)
		hash = hash + ((hash) << 5) + ch + (ch << 7)
	}
	return
}

func (playerConn *PlayerConn) buildPlayerInfo() (Message, error) {
	var message Message
	err := db.View(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.player)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			message.Name = "global_info_sc"
			return nil
		}
		room, err := GetRoom(txn, player.RoomId)
		if err != nil {
			return err
		}
		message.Data, err = PackRoomInfo(txn, room)
		if err != nil {
			return err
		}
		message.Data["name"] = player.Name
		message.Name = "room_info_sc"
		return nil
	})
	return message, err
}

func (playerConn *PlayerConn) StartNotifyPlayerInfo() {
	playerConn.syncTimer = time.NewTicker(time.Second)
	go func() {
		for {
			_, ok := <-playerConn.syncTimer.C
			if !ok {
				break
			}
			message, err := playerConn.buildPlayerInfo()
			if err != nil {
				log.WithError(err).Error("db error")
			} else {
				playerConn.SendSync(message)
			}
		}
	}()
}

func (playerConn *PlayerConn) NotifyPlayerInfo(reply string) {
	message, err := playerConn.buildPlayerInfo()
	if err != nil {
		log.WithError(err).Error("db error")
	} else {
		message.Reply = reply
		playerConn.Send(message)
	}
}

func GetPlayer(txn *badger.Txn, token string) (*Player, error) {
	key := append([]byte("player: "), []byte(token)...)
	item, err := txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, errors.Wrap(err, "cannot find this player")
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	var player Player
	err = item.Value(func(val []byte) error {
		return proto.Unmarshal(val, &player)
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &player, nil
}

func GetPlayerOrNew(txn *badger.Txn, token string) (*Player, error) {
	key := append([]byte("player: "), []byte(token)...)
	item, err := txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return new(Player), nil
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	var player Player
	err = item.Value(func(val []byte) error {
		return proto.Unmarshal(val, &player)
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &player, nil
}

func SetPlayer(txn *badger.Txn, player *Player) error {
	key := append([]byte("player: "), []byte(player.Token)...)
	val, err := proto.Marshal(player)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(txn.Set(key, val))
}

func DelPlayer(txn *badger.Txn, token string) error {
	key := append([]byte("player: "), []byte(token)...)
	return errors.WithStack(txn.Delete(key))
}
