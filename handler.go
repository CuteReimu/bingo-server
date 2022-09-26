package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cast"
	"time"
)

var handlers = map[string]func(player *Player, protoName string, result map[string]interface{}) error{
	"heart_cs":       handleHeart,
	"create_room_cs": handleCreateRoom,
}

func handleCreateRoom(player *Player, protoName string, data map[string]interface{}) error {
	token, err := cast.ToStringE(data["token"])
	if err != nil {
		return errors.WithStack(err)
	}
	name, err := cast.ToStringE(data["name"])
	if err != nil {
		return errors.WithStack(err)
	}
	rid, err := cast.ToStringE(data["rid"])
	if err != nil {
		return errors.WithStack(err)
	}
	roomType, err := cast.ToIntE(data["type"])
	if err != nil {
		return errors.WithStack(err)
	}
	room := GetRoom(rid, roomType, true)
	if room == nil {
		return errors.New("create room failed")
	}
	if _, loaded := playerCache.LoadOrStore(player.token, player); loaded {
		player.SendError(protoName, 1, "already online")
		return nil
	}
	player.name = name
	player.token = token
	player.room = room
	ret := CallRoom(room, func() map[string]interface{} {
		defer room.Lock()()
		if room.host.token != player.token {
			return nil
		}
		ret := make(map[string]interface{})
		room.host = player
		ret["rid"] = room.roomId
		ret["type"] = room.roomType
		ret["host"] = room.host
		ret["names"] = room.GetPlayerNames()
		return ret
	})
	if ret == nil {
		player.SendError(protoName, 2, "room already exists")
		return nil
	}
	ret["name"] = name
	player.Send(&Message{
		Name:  "create_room_sc",
		Reply: protoName,
		Data:  ret,
	})
	return nil
}

func handleHeart(player *Player, protoName string, _ map[string]interface{}) error {
	if player.heartTimer != nil {
		player.heartTimer.Stop()
	}
	player.heartTimer = time.AfterFunc(time.Minute, func() { _ = player.conn.Close() })
	player.Send(&Message{
		Name:  "heart_sc",
		Reply: protoName,
		Data: map[string]interface{}{
			"time": time.Now().UnixMilli(),
		},
	})
	return nil
}
