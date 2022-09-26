package main

import (
	"sync"
)

type Room struct {
	mu       sync.RWMutex
	roomId   string
	roomType int
	host     *Player
	players  []*Player
	ch       chan func()
}

var roomCache = new(sync.Map)

func GetRoom(roomId string, roomType int, create bool) *Room {
	if room, ok := roomCache.Load(roomId); ok {
		return room.(*Room)
	}
	if !create {
		return nil
	}
	room := &Room{
		roomId:   roomId,
		roomType: roomType,
		players:  []*Player{nil, nil},
		ch:       make(chan func(), 1024),
	}
	if _, loaded := roomCache.LoadOrStore(roomId, room); loaded {
		close(room.ch)
		return nil
	}
	go func(room *Room) {
		for {
			f, ok := <-room.ch
			if !ok {
				break
			}
			f()
		}
	}(room)
	return room
}

func CallRoom[T any](room *Room, callback func() T) T {
	ch := make(chan T, 1)
	room.ch <- func() {
		defer func() {
			if r := recover(); r != nil {
				if err, ok := r.(error); ok {
					log.WithError(err).Error("panic")
				} else {
					log.Error("panic: ", r)
				}
			}
			close(ch)
		}()
		ch <- callback()
	}
	return <-ch
}

// CloseRoom 关闭房间，应该由房主线程调用
func CloseRoom(room *Room) {
	roomCache.Delete(room.roomId)
	close(room.ch)
}

func (room *Room) Lock() func() {
	room.mu.Lock()
	return room.mu.Unlock
}

func (room *Room) RLock() func() {
	room.mu.RLock()
	return room.mu.RUnlock
}

func (room *Room) GetPlayerNames() []string {
	names := make([]string, len(room.players))
	for i, p := range room.players {
		names[i] = p.name
	}
	return names
}
