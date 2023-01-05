package main

import (
	"math/rand"
	"time"

	"github.com/pkg/errors"
)

type RoomStartHandler interface {
	OnStart()
}

type RoomNextRoundHandler interface {
	HandleNextRound() error
}

type RoomType interface {
	CanPause() bool
	RandSpells(games, ranks []string) ([]*Spell, error)
	HandleUpdateSpell(token string, idx uint32, status SpellStatus) (tokens []string, newStatus SpellStatus, err error)
}

func (x *Room) Type() RoomType {
	switch x.RoomType {
	case 1:
		return RoomTypeNormal{room: x}
	case 2:
		return RoomTypeBP{room: x}
	case 3:
		return RoomTypeLink{room: x}
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

func (r RoomTypeNormal) RandSpells(games, ranks []string) ([]*Spell, error) {
	return RandSpells(games, ranks, [3]int{10, 10, 5})
}

func (r RoomTypeNormal) HandleUpdateSpell(token string, idx uint32, status SpellStatus) (tokens []string, newStatus SpellStatus, err error) {
	room := r.room
	st := room.Status[idx]
	if status == SpellStatus_banned {
		return nil, st, errors.New("不支持的操作")
	}
	now := time.Now().UnixMilli()
	if room.PauseBeginMs != 0 && token != room.Host {
		return nil, status, errors.New("暂停中，不能操作")
	}
	if room.StartMs <= now-int64(room.GameTime)*60000-int64(room.Countdown)*1000-room.TotalPauseMs {
		return nil, st, errors.New("游戏时间到")
	}
	if room.StartMs > now-int64(room.Countdown)*1000 && !status.isSelectStatus() && !(status == SpellStatus_none && st.isSelectStatus()) {
		return nil, st, errors.New("倒计时还没结束")
	}
	tokens = append(tokens, room.Host)
	switch token {
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

func (r RoomTypeBP) RandSpells(games, ranks []string) ([]*Spell, error) {
	return RandSpells(games, ranks, [3]int{5, 15, 5})
}

func (r RoomTypeBP) HandleUpdateSpell(token string, idx uint32, status SpellStatus) (tokens []string, newStatus SpellStatus, err error) {
	room := r.room
	st := room.Status[idx]
	switch token {
	case room.Players[0]:
		if room.BpData.WhoseTurn != 0 {
			return nil, st, errors.New("不是你的回合")
		}
		if st != SpellStatus_none ||
			room.BpData.BanPick == 0 && status != SpellStatus_left_select ||
			room.BpData.BanPick == 1 && status != SpellStatus_banned {
			return nil, st, errors.New("权限不足")
		}
		r.nextRound()
	case room.Players[1]:
		if room.BpData.WhoseTurn != 1 {
			return nil, st, errors.New("不是你的回合")
		}
		if st != SpellStatus_none ||
			room.BpData.BanPick == 0 && status != SpellStatus_right_select ||
			room.BpData.BanPick == 1 && status != SpellStatus_banned {
			return nil, st, errors.New("权限不足")
		}
		r.nextRound()
	}
	newStatus = status
	tokens = append(tokens, room.Host)
	tokens = append(tokens, room.Players...)
	return
}

func (r RoomTypeBP) HandleNextRound() error {
	if r.room.BpData.BanPick != 2 {
		return errors.New("现在不是这个操作的时候")
	}
	r.nextRound()
	return nil
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
		bp.BanPick = 2
	case 6:
		bp.BanPick = 1
	case 7:
		bp.WhoseTurn = 1 - bp.WhoseTurn
	case 8:
	case 9:
		bp.WhoseTurn = 1 - bp.WhoseTurn
		bp.BanPick = 0
	case 10:
	case 11:
		bp.WhoseTurn = 1 - bp.WhoseTurn
	case 12:
		bp.WhoseTurn = 1 - bp.WhoseTurn
		bp.BanPick = 2
	case 13:
		bp.BanPick = 1
	case 14:
	case 15:
		bp.WhoseTurn = 1 - bp.WhoseTurn
	case 16:
		bp.BanPick = 0
	default:
		if !bp.LessThan4 && bp.Round%5 == 1 {
			count := 0
			for _, status := range r.room.Status {
				if status == SpellStatus_none {
					count++
				}
			}
			if count < 4 {
				bp.LessThan4 = true
			}
		}
		if bp.LessThan4 {
			if bp.BanPick == 2 {
				bp.WhoseTurn = 1 - bp.WhoseTurn
				bp.BanPick = 0
			} else {
				bp.BanPick = 2
			}
		} else {
			switch bp.Round % 5 {
			case 0:
				bp.WhoseTurn = 1 - bp.WhoseTurn
				bp.BanPick = 2
			case 1:
				bp.WhoseTurn = 1 - bp.WhoseTurn
				bp.BanPick = 0
			case 3:
				bp.WhoseTurn = 1 - bp.WhoseTurn
			}
		}
	}
}

type RoomTypeLink struct {
	room *Room
}

func (r RoomTypeLink) CanPause() bool {
	return false
}

func (r RoomTypeLink) OnStart() {
	r.room.Status[0] = SpellStatus_left_select
	r.room.Status[4] = SpellStatus_right_select
	r.room.LinkData = &LinkData{LinkIdxA: []uint32{0}, LinkIdxB: []uint32{4}}
}

func (r RoomTypeLink) RandSpells(games, ranks []string) ([]*Spell, error) {
	return RandSpells2(games, ranks, [3]int{5, 15, 5})
}

func (r RoomTypeLink) HandleUpdateSpell(token string, idx uint32, status SpellStatus) (tokens []string, newStatus SpellStatus, err error) {
	room := r.room
	st := room.Status[idx]
	if status == SpellStatus_banned {
		return nil, st, errors.New("不支持的操作")
	}
	tokens = append(tokens, room.Host)
	switch token {
	case room.Host:
		newStatus = status
		tokens = append(tokens, room.Players...)
	case room.Players[0]:
		if r.room.LinkData.FinishSelect {
			return nil, st, errors.New("选卡已结束")
		}
		switch status {
		case SpellStatus_left_select:
			for _, idx1 := range r.room.LinkData.LinkIdxA {
				if idx1 == idx {
					return nil, st, errors.New("已经选了这张卡")
				}
			}
			if len(r.room.LinkData.LinkIdxA) == 0 {
				if idx != 0 {
					return nil, st, errors.New("不合理的选卡")
				}
			} else {
				idx0 := r.room.LinkData.LinkIdxA[len(r.room.LinkData.LinkIdxA)-1]
				diff := int32(idx) - int32(idx0)
				if idx0 == 24 || diff != -6 && diff != -5 && diff != -4 && diff != -1 && diff != 1 && diff != 4 && diff != 5 && diff != 6 {
					return nil, st, errors.New("不合理的选卡")
				}
			}
			if st == SpellStatus_right_select {
				newStatus = SpellStatus_both_select
			} else {
				newStatus = status
			}
			room.LinkData.LinkIdxA = append(room.LinkData.LinkIdxA, idx)
		case SpellStatus_none:
			if len(room.LinkData.LinkIdxA) <= 1 {
				return nil, st, errors.New("初始选卡不能删除")
			}
			if room.LinkData.LinkIdxA[len(room.LinkData.LinkIdxA)-1] != idx {
				return nil, st, errors.New("只能删除最后一张卡")
			}
			if st == SpellStatus_both_select {
				newStatus = SpellStatus_right_select
			} else {
				newStatus = status
			}
			room.LinkData.LinkIdxA = room.LinkData.LinkIdxA[:len(room.LinkData.LinkIdxA)-1]
		default:
			return nil, st, errors.New("权限不足")
		}
		tokens = append(tokens, room.Players[0])
	case room.Players[1]:
		if r.room.LinkData.FinishSelect {
			return nil, st, errors.New("选卡已结束")
		}
		switch status {
		case SpellStatus_right_select:
			for _, idx1 := range r.room.LinkData.LinkIdxB {
				if idx1 == idx {
					return nil, st, errors.New("已经选了这张卡")
				}
			}
			if len(r.room.LinkData.LinkIdxB) == 0 {
				if idx != 4 {
					return nil, st, errors.New("不合理的选卡")
				}
			} else {
				idx0 := r.room.LinkData.LinkIdxB[len(r.room.LinkData.LinkIdxB)-1]
				diff := int32(idx) - int32(idx0)
				if idx0 == 20 || diff != -6 && diff != -5 && diff != -4 && diff != -1 && diff != 1 && diff != 4 && diff != 5 && diff != 6 {
					return nil, st, errors.New("不合理的选卡")
				}
			}
			if st == SpellStatus_left_select {
				newStatus = SpellStatus_both_select
			} else {
				newStatus = status
			}
			room.LinkData.LinkIdxB = append(room.LinkData.LinkIdxB, idx)
		case SpellStatus_none:
			if len(room.LinkData.LinkIdxB) <= 1 {
				return nil, st, errors.New("初始选卡不能删除")
			}
			if room.LinkData.LinkIdxB[len(room.LinkData.LinkIdxB)-1] != idx {
				return nil, st, errors.New("只能删除最后一张卡")
			}
			if st == SpellStatus_both_select {
				newStatus = SpellStatus_left_select
			} else {
				newStatus = status
			}
			room.LinkData.LinkIdxB = room.LinkData.LinkIdxB[:len(room.LinkData.LinkIdxB)-1]
		default:
			return nil, st, errors.New("权限不足")
		}
		tokens = append(tokens, room.Players[1])
	}
	return
}
