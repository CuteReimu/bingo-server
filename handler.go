package main

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"time"
)

var handlers = map[string]func(player *PlayerConn, protoName string, result map[string]interface{}) error{
	"login_cs":            handleLogin,
	"heart_cs":            handleHeart,
	"create_room_cs":      handleCreateRoom,
	"join_room_cs":        handleJoinRoom,
	"leave_room_cs":       handleLeaveRoom,
	"update_room_type_cs": handleUpdateRoomType,
	"update_name_cs":      handleUpdateName,
}

func handleUpdateName(playerConn *PlayerConn, protoName string, data map[string]interface{}) error {
	name, err := cast.ToStringE(data["name"])
	if err != nil {
		return errors.WithStack(err)
	}
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	err = db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.player)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return errors.New("不在房间里")
		}
		player.Name = name
		return SetPlayer(txn, player)
	})
	if err != nil {
		return err
	}
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}

func handleUpdateRoomType(playerConn *PlayerConn, protoName string, data map[string]interface{}) error {
	roomType, err := cast.ToInt32E(data["type"])
	if err != nil {
		return errors.WithStack(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.player)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return errors.New("不在房间里")
		}
		room, err := GetRoom(txn, player.RoomId)
		if err != nil {
			return err
		}
		if room.Host != playerConn.player {
			return errors.New("不是房主")
		}
		room.RoomType = roomType
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}

func handleLeaveRoom(playerConn *PlayerConn, protoName string, _ map[string]interface{}) error {
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.player)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return errors.New("并不在房间里")
		}
		room, err := GetRoom(txn, player.RoomId)
		if err != nil {
			return err
		}
		if room.GetStarted() {
			return errors.New("比赛已经开始了，不能退出")
		}
		player.RoomId = ""
		player.Name = ""
		if room.Host == player.Token {
			for i := range room.Players {
				if len(room.Players[i]) != 0 && room.Players[i] != room.Host {
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
			if err = DelRoom(txn, room.RoomId); err != nil {
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
		return SetPlayer(txn, player)
	})
	if err != nil {
		return err
	}
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}

func handleJoinRoom(playerConn *PlayerConn, protoName string, data map[string]interface{}) error {
	name, err := cast.ToStringE(data["name"])
	if err != nil {
		return errors.WithStack(err)
	}
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	rid, err := cast.ToStringE(data["rid"])
	if err != nil {
		return errors.WithStack(err)
	}
	if len(rid) == 0 {
		return errors.New("房间ID为空")
	}
	err = db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.player)
		if err != nil {
			return err
		}
		if len(player.RoomId) != 0 {
			return errors.New("已经在房间里了")
		}
		room, err := GetRoom(txn, rid)
		if err != nil {
			return err
		}
		player.RoomId = rid
		player.Name = name
		if err = SetPlayer(txn, player); err != nil {
			return err
		}
		var ok bool
		for i := range room.Players {
			if ok {
				if len(room.Players[i]) != 0 {
					player2, err := GetPlayer(txn, room.Players[i])
					if err != nil {
						return err
					}
					if player2.Name == name {
						return errors.New("名字重复")
					}
				}
			} else if len(room.Players[i]) == 0 {
				ok = true
				room.Players[i] = player.Token
			}
		}
		if !ok {
			return errors.New("房间满了")
		}
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}

func handleCreateRoom(playerConn *PlayerConn, protoName string, data map[string]interface{}) error {
	name, err := cast.ToStringE(data["name"])
	if err != nil {
		return errors.WithStack(err)
	}
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	rid, err := cast.ToStringE(data["rid"])
	if err != nil {
		return errors.WithStack(err)
	}
	if len(rid) == 0 {
		return errors.New("房间ID为空")
	}
	roomType, err := cast.ToInt32E(data["type"])
	if err != nil {
		return errors.WithStack(err)
	}
	err = db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.player)
		if err != nil {
			return err
		}
		if len(player.RoomId) != 0 {
			return errors.New("already in room")
		}
		key := append([]byte("room: "), []byte(rid)...)
		_, err = txn.Get(key)
		if err == nil {
			return errors.New("room already exists")
		} else if err != badger.ErrKeyNotFound {
			return errors.WithStack(err)
		}
		var room = Room{
			RoomId:   rid,
			RoomType: roomType,
			Host:     playerConn.player,
			Players:  make([]string, 2),
		}
		player.RoomId = rid
		player.Name = name
		if err = SetPlayer(txn, player); err != nil {
			return err
		}
		if err != nil {
			return err
		}
		return SetRoom(txn, &room)
	})
	if err != nil {
		return err
	}
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}

func handleHeart(playerConn *PlayerConn, protoName string, _ map[string]interface{}) error {
	playerConn.Send(Message{
		Name:  "heart_sc",
		Reply: protoName,
		Data: map[string]interface{}{
			"time": time.Now().UnixMilli(),
		},
	})
	return nil
}

func handleLogin(playerConn *PlayerConn, protoName string, data map[string]interface{}) error {
	token, ok := data["token"]
	if !ok {
		playerConn.SendError(protoName, 400, "no token")
		return nil
	}
	tokenStr, _ := token.(string)
	if len(tokenStr) == 0 || !isAlphaNum(tokenStr) {
		playerConn.SendError(protoName, 400, "invalid token")
		return nil
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayerOrNew(txn, tokenStr)
		if err != nil {
			return err
		}
		player.Token = tokenStr
		_, loaded := playerConnCache.LoadOrStore(tokenStr, playerConn)
		if loaded {
			return errors.New("already online")
		}
		return SetPlayer(txn, player)
	})
	if err != nil {
		return err
	}
	playerConn.player = tokenStr
	playerConn.NotifyPlayerInfo(protoName)
	playerConn.StartNotifyPlayerInfo()
	return nil
}
