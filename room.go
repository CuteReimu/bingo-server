package main

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

func GetRoom(txn *badger.Txn, roomId string) (*Room, error) {
	key := append([]byte("room: "), []byte(roomId)...)
	item, err := txn.Get(key)
	if err == badger.ErrKeyNotFound {
		return nil, errors.Wrap(err, "cannot find this room")
	} else if err != nil {
		return nil, errors.WithStack(err)
	}
	var room Room
	err = item.Value(func(val []byte) error {
		return proto.Unmarshal(val, &room)
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &room, nil
}

func SetRoom(txn *badger.Txn, room *Room) error {
	key := append([]byte("room: "), []byte(room.RoomId)...)
	val, err := proto.Marshal(room)
	if err != nil {
		return errors.WithStack(err)
	}
	return errors.WithStack(txn.Set(key, val))
}

func PackRoomInfo(txn *badger.Txn, room *Room) (map[string]interface{}, error) {
	players := make([]string, len(room.Players))
	for i := range players {
		if len(room.Players[i]) > 0 {
			player, err := GetPlayer(txn, room.Players[i])
			if err != nil {
				return nil, err
			}
			players[i] = player.Name
		}
	}
	return map[string]interface{}{
		"rid":   room.RoomId,
		"type":  room.RoomType,
		"host":  room.Host,
		"names": players,
	}, nil
}
