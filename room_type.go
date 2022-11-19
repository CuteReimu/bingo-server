package main

import (
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

type RoomStartHandler interface {
	OnStart()
}

type RoomUndoHandler interface {
	OnStart()
	SaveSnapshot(room *Room, idx uint32)
	HandleUndo(room *Room) (idx uint32, err error)
}

type RoomType interface {
	CanPause() bool
	CardCount() [3]int
	HandleUpdateSpell(playerConn *PlayerConn, idx uint32, status SpellStatus) (tokens []string, newStatus SpellStatus, err error)
}

func (x *Room) Type() RoomType {
	switch x.RoomType {
	case 1:
		return RoomTypeNormal{room: x}
	case 2:
		return RoomTypeBP{room: x}
	default:
		panic("不支持的游戏类型")
	}
}

type RoomTypeNormal struct {
	room *Room
}

func (r RoomTypeNormal) CanPause() bool {
	return true
}

func (r RoomTypeNormal) CardCount() [3]int {
	return [3]int{10, 10, 5}
}

func (r RoomTypeNormal) HandleUpdateSpell(playerConn *PlayerConn, idx uint32, status SpellStatus) (tokens []string, newStatus SpellStatus, err error) {
	room := r.room
	st := room.Status[idx]
	if st == SpellStatus_banned {
		return nil, st, errors.New("游戏时间到")
	}
	now := time.Now().UnixMilli()
	if room.PauseBeginMs != 0 && playerConn.token != room.Host {
		return nil, status, errors.New("暂停中，不能操作")
	}
	if room.StartMs <= now-int64(room.GameTime)*60000-int64(room.Countdown)*1000-room.TotalPauseMs {
		return nil, st, errors.New("游戏时间到")
	}
	if room.StartMs > now-int64(room.Countdown)*1000 && !status.isSelectStatus() && !(status == SpellStatus_none && st.isSelectStatus()) {
		return nil, st, errors.New("倒计时还没结束")
	}
	tokens = append(tokens, room.Host)
	switch playerConn.token {
	case room.Host:
		newStatus = status
		tokens = append(tokens, room.Players...)
	case room.Players[0]:
		if status.isRightStatus() || st == SpellStatus_left_get && status != SpellStatus_left_get {
			return nil, st, errors.New("权限不足")
		}
		if st == SpellStatus_right_get {
			return nil, st, errors.New("对方已经打完")
		}
		switch status {
		case SpellStatus_left_get:
			newStatus = status
		case SpellStatus_left_select:
			if st == SpellStatus_right_select {
				newStatus = SpellStatus_both_select
			} else {
				newStatus = status
			}
		case SpellStatus_none:
			if st == SpellStatus_both_select {
				newStatus = SpellStatus_right_select
			} else {
				newStatus = status
			}
		}
		tokens = append(tokens, room.Players[0])
		if status != SpellStatus_left_select && status != SpellStatus_none {
			tokens = append(tokens, room.Players[1])
		}
	case room.Players[1]:
		if status.isLeftStatus() || st == SpellStatus_right_get && status != SpellStatus_right_get {
			return nil, st, errors.New("权限不足")
		}
		if st == SpellStatus_left_get {
			return nil, st, errors.New("对方已经打完")
		}
		switch status {
		case SpellStatus_right_get:
			newStatus = status
		case SpellStatus_right_select:
			if st == SpellStatus_left_select {
				newStatus = SpellStatus_both_select
			} else {
				newStatus = status
			}
		case SpellStatus_none:
			if st == SpellStatus_both_select {
				newStatus = SpellStatus_left_select
			} else {
				newStatus = status
			}
		}
		tokens = append(tokens, room.Players[1])
		if status != SpellStatus_right_select && status != SpellStatus_none {
			tokens = append(tokens, room.Players[0])
		}
	}
	return
}

type RoomTypeBP struct {
	room *Room
}

func (r RoomTypeBP) CanPause() bool {
	return false
}

func (r RoomTypeBP) OnStart() {
	if r.room.LastWinner > 0 {
		r.room.BpData = &BpData{
			WhoseTurn: r.room.LastWinner - 1,
			BanPick:   1,
		}
	} else {
		r.room.BpData = &BpData{
			WhoseTurn: rand.Int31n(2),
			BanPick:   1,
		}
	}
}

func (r RoomTypeBP) CardCount() [3]int {
	return [3]int{5, 15, 5}
}

func (r RoomTypeBP) HandleUpdateSpell(playerConn *PlayerConn, idx uint32, status SpellStatus) (tokens []string, newStatus SpellStatus, err error) {
	room := r.room
	st := room.Status[idx]
	switch playerConn.token {
	case room.Host:
		if st.isSelectStatus() && status != SpellStatus_banned {
			r.nextRound()
		}
	case room.Players[0]:
		if room.BpData.WhoseTurn != 0 {
			return nil, st, errors.New("不是你的回合")
		}
		if st != SpellStatus_none ||
			room.BpData.BanPick == 0 && status != SpellStatus_left_select ||
			room.BpData.BanPick == 1 && status != SpellStatus_banned {
			return nil, st, errors.New("权限不足")
		}
		if status == SpellStatus_banned {
			r.nextRound()
		}
	case room.Players[1]:
		if room.BpData.WhoseTurn != 1 {
			return nil, st, errors.New("不是你的回合")
		}
		if st != SpellStatus_none ||
			room.BpData.BanPick == 0 && status != SpellStatus_right_select ||
			room.BpData.BanPick == 1 && status != SpellStatus_banned {
			return nil, st, errors.New("权限不足")
		}
		if status == SpellStatus_banned {
			r.nextRound()
		}
	}
	newStatus = status
	tokens = append(tokens, room.Host)
	tokens = append(tokens, room.Players...)
	return
}

func (r RoomTypeBP) nextRound() {
	bp := r.room.BpData
	bp.Round++
	switch bp.Round {
	case 1:
		bp.WhoseTurn = 1 - bp.WhoseTurn
	case 2:
		bp.WhoseTurn = 1 - bp.WhoseTurn
		bp.BanPick = 0
	case 3:
		bp.WhoseTurn = 1 - bp.WhoseTurn
	case 4:
	case 5:
		bp.WhoseTurn = 1 - bp.WhoseTurn
		bp.BanPick = 1
	case 6:
		bp.WhoseTurn = 1 - bp.WhoseTurn
	case 7:
	case 8:
		bp.WhoseTurn = 1 - bp.WhoseTurn
		bp.BanPick = 0
	case 9:
	case 10:
		bp.WhoseTurn = 1 - bp.WhoseTurn
	case 11:
		bp.WhoseTurn = 1 - bp.WhoseTurn
		bp.BanPick = 1
	case 12:
	case 13:
		bp.WhoseTurn = 1 - bp.WhoseTurn
	case 14:
		bp.BanPick = 0
	default:
		if !bp.LessThan4 && bp.Round%4 == 2 {
			count := 0
			for _, status := range r.room.Status {
				if status.isGetStatus() {
					count++
				}
			}
			if 25-count < 4 {
				bp.LessThan4 = true
			}
		}
		if bp.LessThan4 || bp.Round%4 == 0 {
			bp.WhoseTurn = 1 - bp.WhoseTurn
		}
	}
}

func (r RoomTypeBP) SaveSnapshot(room *Room, idx uint32) {
	room.BpData.Snapshots = append(room.BpData.Snapshots, &Snapshot{
		Idx:       idx,
		Status:    room.Status[idx],
		WhoseTurn: room.BpData.WhoseTurn,
		BanPick:   room.BpData.BanPick,
		Round:     room.BpData.Round,
		LessThan4: room.BpData.LessThan4,
	})
}

func (r RoomTypeBP) HandleUndo(room *Room) (idx uint32, err error) {
	if len(room.BpData.Snapshots) == 0 {
		return 0, errors.New("当前是第一回合，不能再撤销了")
	}
	snapshot := room.BpData.Snapshots[len(room.BpData.Snapshots)-1]
	room.Status[idx] = snapshot.Status
	room.BpData.WhoseTurn = snapshot.WhoseTurn
	room.BpData.BanPick = snapshot.BanPick
	room.BpData.Round = snapshot.Round
	room.BpData.LessThan4 = snapshot.LessThan4
	room.BpData.Snapshots = room.BpData.Snapshots[:len(room.BpData.Snapshots)-1]
	return idx, nil
}
