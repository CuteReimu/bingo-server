package main

import (
	"github.com/Touhou-Freshman-Camp/bingo-server/myws"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
	"time"
)

func (s *bingoServer) buildPlayerInfo(token string) (*myws.Message, []string, error) {
	var message = &myws.Message{}
	var tokens []string
	err := db.View(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
		return nil
	})
	if err != nil {
		return nil, nil, err
	}
	if message.Data == nil {
		message.Data = &RoomInfoSc{}
	}
	return message, tokens, nil
}

func (s *bingoServer) getAllPlayersInRoom(token string) ([]string, error) {
	var tokens []string
	err := db.View(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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

func (s *bingoServer) NotifyPlayerInfo(token, reply string, winnerIdx ...int32) {
	message, tokens, err := s.buildPlayerInfo(token)
	if len(winnerIdx) > 0 {
		message.Data.(*RoomInfoSc).Winner = winnerIdx[0]
	}
	if err != nil {
		log.WithError(err).Error("db error")
	} else {
		for _, token1 := range tokens {
			if conn, ok := s.tokenConnMap[token1]; ok {
				if token1 != token {
					conn.Send(message)
				} else {
					conn.Send(&myws.Message{
						MsgName: message.MsgName,
						Reply:   reply,
						Data:    message.Data,
					})
				}
			}
		}
	}
}

func (s *bingoServer) NotifyPlayersInRoom(token, reply string, message *myws.Message) {
	tokens, err := s.getAllPlayersInRoom(token)
	if err != nil {
		log.WithError(err).Error("db error")
	} else {
		for _, token1 := range tokens {
			if conn, ok := s.tokenConnMap[token1]; ok {
				if token1 != token {
					conn.Send(message)
				} else {
					conn.Send(&myws.Message{
						Reply: reply,
						Data:  message.Data,
					})
				}
			}
		}
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
