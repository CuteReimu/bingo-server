package main

import (
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"time"
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
	return errors.WithStack(txn.SetEntry(badger.NewEntry(key, val).WithTTL(6 * time.Hour)))
}

func DelRoom(txn *badger.Txn, roomId string) error {
	key := append([]byte("room: "), []byte(roomId)...)
	return errors.WithStack(txn.Delete(key))
}

func PackRoomInfo(txn *badger.Txn, room *Room) (map[string]interface{}, error) {
	host, err := GetPlayer(txn, room.Host)
	if err == badger.ErrKeyNotFound {
		return nil, DelRoom(txn, room.RoomId)
	} else if err != nil {
		return nil, err
	}
	var updated bool
	players := make([]string, len(room.Players))
	for i := range players {
		if len(room.Players[i]) > 0 {
			player, err := GetPlayer(txn, room.Players[i])
			if err == badger.ErrKeyNotFound {
				room.Players[i] = ""
				updated = true
				continue
			} else if err != nil {
				return nil, err
			}
			players[i] = player.Name
		}
	}
	ret := map[string]interface{}{
		"rid":   room.RoomId,
		"type":  room.RoomType,
		"host":  host.Name,
		"names": players,
	}
	if updated {
		return ret, SetRoom(txn, room)
	} else {
		return ret, nil
	}
}
