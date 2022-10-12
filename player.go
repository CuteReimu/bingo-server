package main

import (
	"github.com/Touhou-Freshman-Camp/bingo-server/myws"
	"github.com/davyxu/cellnet"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"
	"google.golang.org/protobuf/proto"
	"time"
)

type PlayerConn struct {
	token string
	cellnet.Session
	heartTimer *time.Timer
	Limit      *rate.Limiter
}

var tokenConnMap = make(map[string]*PlayerConn)

func (playerConn *PlayerConn) SetHeartTimer() {
	if playerConn.heartTimer != nil {
		playerConn.heartTimer.Stop()
	}
	playerConn.heartTimer = time.AfterFunc(time.Minute, func() {
		log.WithField("conn_id", playerConn.ID()).Warn("长时间没有心跳，断开连接")
		playerConn.Close()
	})
}

func (playerConn *PlayerConn) Handle(name string, data map[string]interface{}) {
	if !playerConn.Limit.Allow() {
		playerConn.Close()
		return
	}
	handler := handlers[name]
	if handler == nil {
		log.Warn("can not find handler: ", name)
		playerConn.SendError(name, 404, "404 not found")
		return
	}
	playerConn.SetHeartTimer()
	if len(playerConn.token) == 0 && name != "login_cs" {
		playerConn.SendError(name, -1, "You haven't login.")
		return
	}
	if err := handler(playerConn, name, data); err != nil {
		playerConn.SendError(name, 500, err.Error())
		log.WithError(err).Error("handle failed: ", name)
	}
}

func (playerConn *PlayerConn) SendError(reply string, code int, msg string) {
	playerConn.Send(&myws.Message{
		MsgName: "error_sc",
		Reply:   reply,
		Data: map[string]interface{}{
			"code": code,
			"msg":  msg,
		},
	})
}

func (playerConn *PlayerConn) SendSuccess(reply string) {
	playerConn.Send(&myws.Message{
		MsgName: "success_sc",
		Reply:   reply,
	})
}

func (playerConn *PlayerConn) OnDisconnect() {
	playerConn.heartTimer.Stop()
	if playerConn.heartTimer != nil {
		playerConn.heartTimer.Stop()
		playerConn.heartTimer = nil
	}
	if len(playerConn.token) == 0 {
		return
	}
	var tokens []string
	var roomDestroyed bool
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
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
				if len(room.Players[i]) != 0 && room.Players[i] != room.Host {
					tokens = append(tokens, room.Players[i])
					p, err := GetPlayer(txn, room.Players[i])
					if IsErrKeyNotFound(err) {
						continue
					} else if err != nil {
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
			if err = DelRoom(txn, room.RoomId); err != nil {
				return err
			}
			roomDestroyed = true
		} else {
			for i := range room.Players {
				if room.Players[i] == player.Token {
					room.Players[i] = ""
				} else {
					tokens = append(tokens, room.Players[i])
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
	delete(tokenConnMap, playerConn.token)
	for _, token := range tokens {
		conn := tokenConnMap[token]
		if conn != nil {
			if roomDestroyed {
				conn.Send(&myws.Message{MsgName: "room_info_sc"})
			} else {
				conn.NotifyPlayerInfo("")
				break
			}
		}
	}
}

func isAlphaNum(s string) bool {
	for _, c := range []byte(s) {
		if (c < '0' || c > '9') && (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}
	return true
}

func (playerConn *PlayerConn) buildPlayerInfo() (*myws.Message, []string, error) {
	var message = &myws.Message{MsgName: "room_info_sc"}
	var tokens []string
	err := db.View(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return nil
		}
		room, err := GetRoom(txn, player.RoomId)
		if IsErrKeyNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}
		message.Data, tokens, err = PackRoomInfo(txn, room)
		if err != nil {
			return err
		}
		message.Data["name"] = player.Name
		message.Data["started"] = room.Started
		message.Data["score"] = room.Score
		if len(room.WinnerName) > 0 {
			message.Data["winner"] = room.WinnerName
		}
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	return message, tokens, nil
}

func (playerConn *PlayerConn) getAllPlayersInRoom() ([]string, error) {
	var tokens []string
	err := db.View(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return nil
		}
		room, err := GetRoom(txn, player.RoomId)
		if IsErrKeyNotFound(err) {
			return nil
		} else if err != nil {
			return err
		}
		tokens = append(room.Players, room.Host)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return tokens, nil
}

func (playerConn *PlayerConn) NotifyPlayerInfo(reply string) {
	message, tokens, err := playerConn.buildPlayerInfo()
	if err != nil {
		log.WithError(err).Error("db error")
	} else {
		for _, token := range tokens {
			if token != playerConn.token {
				if conn, ok := tokenConnMap[token]; ok {
					conn.Send(message)
				}
			}
		}
		playerConn.Send(&myws.Message{
			MsgName: message.MsgName,
			Reply:   reply,
			Data:    message.Data,
		})
	}
}

func (playerConn *PlayerConn) NotifyPlayersInRoom(reply string, message *myws.Message) {
	tokens, err := playerConn.getAllPlayersInRoom()
	if err != nil {
		log.WithError(err).Error("db error")
	} else {
		for _, token := range tokens {
			if token != playerConn.token {
				if conn, ok := tokenConnMap[token]; ok {
					conn.Send(message)
				}
			}
		}
		playerConn.Send(&myws.Message{
			MsgName: message.MsgName,
			Reply:   reply,
			Data:    message.Data,
		})
	}
}

func GetPlayer(txn *badger.Txn, token string) (*Player, error) {
	key := append([]byte("token: "), []byte(token)...)
	item, err := txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, errors.Wrap(err, "cannot find this token")
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
	key := append([]byte("token: "), []byte(token)...)
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
	key := append([]byte("token: "), []byte(player.Token)...)
	val, err := proto.Marshal(player)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(txn.SetEntry(badger.NewEntry(key, val).WithTTL(24 * time.Hour)))
}

func DelPlayer(txn *badger.Txn, token string) error {
	key := append([]byte("token: "), []byte(token)...)
	return errors.WithStack(txn.Delete(key))
}
