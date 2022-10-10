package main

import (
	"github.com/Touhou-Freshman-Camp/bingo-server/arrays"
	"github.com/Touhou-Freshman-Camp/bingo-server/myws"
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
	"start_game_cs":       handleStartGame,
	"get_spells_cs":       handleGetSpells,
	"stop_game_cs":        handleStopGame,
}

func handleStopGame(playerConn *PlayerConn, protoName string, _ map[string]interface{}) error {
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
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
		if room.Host != playerConn.token {
			return errors.New("你不是房主")
		} else if !room.Started {
			return errors.New("游戏还没开始")
		}
		room.Started = false
		room.Spells = nil
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	playerConn.NotifyPlayersInRoom(protoName, &myws.Message{
		MsgName: "stop_game_sc",
	})
	return nil
}

func handleGetSpells(playerConn *PlayerConn, protoName string, _ map[string]interface{}) error {
	var spells []*Spell
	err := db.View(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
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
		if !room.Started {
			return errors.New("游戏还未开始")
		}
		spells = room.Spells
		return nil
	})
	if err != nil {
		return err
	}
	playerConn.Send(&myws.Message{
		MsgName: "spell_list_sc",
		Reply:   protoName,
		Data: map[string]interface{}{
			"spells": spells,
		},
	})
	return nil
}

func handleStartGame(playerConn *PlayerConn, protoName string, data map[string]interface{}) error {
	games, err := cast.ToStringSliceE(data["games"])
	if err != nil {
		return errors.WithStack(err)
	}
	var spells []*Spell
	err = db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
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
		if room.Host != playerConn.token {
			return errors.New("你不是房主")
		} else if room.Started {
			return errors.New("游戏已经开始")
		} else if arrays.Any(room.Players, func(s string) bool { return len(s) == 0 }) {
			return errors.New("玩家没满")
		}
		spells, err = RandSpells(games)
		if err != nil {
			return errors.Wrap(err, "随符卡失败")
		}
		room.Started = true
		room.Spells = spells
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	playerConn.NotifyPlayersInRoom(protoName, &myws.Message{
		MsgName: "spell_list_sc",
		Data: map[string]interface{}{
			"spells": spells,
		},
	})
	return nil
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
		player, err := GetPlayer(txn, playerConn.token)
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
		player, err := GetPlayer(txn, playerConn.token)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return errors.New("不在房间里")
		}
		room, err := GetRoom(txn, player.RoomId)
		if err == badger.ErrKeyNotFound {
			return errors.New("不在房间里")
		} else if err != nil {
			return err
		}
		if room.Host != playerConn.token {
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
		player, err := GetPlayer(txn, playerConn.token)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return errors.New("并不在房间里")
		}
		room, err := GetRoom(txn, player.RoomId)
		if err == badger.ErrKeyNotFound {
			return errors.New("不在房间里")
		} else if err != nil {
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
		player, err := GetPlayer(txn, playerConn.token)
		if err != nil {
			return err
		}
		if len(player.RoomId) != 0 {
			return errors.New("已经在房间里了")
		}
		room, err := GetRoom(txn, rid)
		if err == badger.ErrKeyNotFound {
			return errors.New("房间不存在")
		} else if err != nil {
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
		player, err := GetPlayer(txn, playerConn.token)
		if err != nil {
			return err
		}
		if len(player.RoomId) != 0 {
			return errors.New("already in room")
		}
		key := append([]byte("room: "), []byte(rid)...)
		_, err = txn.Get(key)
		if err == nil {
			return errors.New("房间已存在")
		} else if err != badger.ErrKeyNotFound {
			return errors.WithStack(err)
		}
		var room = Room{
			RoomId:   rid,
			RoomType: roomType,
			Host:     playerConn.token,
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
	playerConn.Send(&myws.Message{
		MsgName: "heart_sc",
		Reply:   protoName,
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
		if _, ok := tokenConnMap[tokenStr]; ok {
			return errors.New("already online")
		} else {
			tokenConnMap[tokenStr] = playerConn
		}
		return SetPlayer(txn, player)
	})
	if err != nil {
		return err
	}
	playerConn.token = tokenStr
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}
