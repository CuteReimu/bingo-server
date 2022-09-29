package main

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"time"
)

var handlers = map[string]func(player *PlayerConn, protoName string, result map[string]interface{}) error{
	"login_cs":       handleLogin,
	"heart_cs":       handleHeart,
	"create_room_cs": handleCreateRoom,
	"join_room_cs":   handleJoinRoom,
}

func handleJoinRoom(playerConn *PlayerConn, protoName string, data map[string]interface{}) error {
	name, err := cast.ToStringE(data["name"])
	if err != nil {
		return errors.WithStack(err)
	}
	rid, err := cast.ToStringE(data["rid"])
	if err != nil {
		return errors.WithStack(err)
	}
	var roomInfo map[string]interface{}
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
			if len(room.Players[i]) == 0 {
				ok = true
				room.Players[i] = player.Token
				break
			}
		}
		if !ok {
			return errors.New("房间满了")
		}
		roomInfo, err = PackRoomInfo(txn, room)
		if err != nil {
			return err
		}
		var count int
		for _, name2 := range roomInfo["names"].([]string) {
			if name2 == name {
				count++
			}
		}
		if count >= 2 {
			return errors.New("该名字已被使用")
		}
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	roomInfo["name"] = name
	playerConn.Send(Message{
		Name:  "join_room_sc",
		Reply: protoName,
		Data:  roomInfo,
	})
	return nil
}

func handleCreateRoom(playerConn *PlayerConn, protoName string, data map[string]interface{}) error {
	name, err := cast.ToStringE(data["name"])
	if err != nil {
		return errors.WithStack(err)
	}
	rid, err := cast.ToStringE(data["rid"])
	if err != nil {
		return errors.WithStack(err)
	}
	roomType, err := cast.ToInt32E(data["type"])
	if err != nil {
		return errors.WithStack(err)
	}
	var roomInfo map[string]interface{}
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
			Host:     name,
			Players:  make([]string, 2),
		}
		player.RoomId = rid
		player.Name = name
		if err = SetPlayer(txn, player); err != nil {
			return err
		}
		roomInfo, err = PackRoomInfo(txn, &room)
		if err != nil {
			return err
		}
		return SetRoom(txn, &room)
	})
	if err != nil {
		return err
	}
	roomInfo["name"] = name
	playerConn.Send(Message{
		Name:  "join_room_sc",
		Reply: protoName,
		Data:  roomInfo,
	})
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
		playerConn.player = tokenStr
		return SetPlayer(txn, player)
	})
	if err != nil {
		return err
	}
	playerConn.Send(Message{
		Name:  "login_sc",
		Reply: protoName,
		Data: map[string]interface{}{
			"time": time.Now().UnixMilli(),
		},
	})
	return nil
}
