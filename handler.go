package main

import (
	"github.com/CuteReimu/bingo-server/myws"
	"github.com/davyxu/cellnet"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"time"
)

func (m *NextRoundCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	var whoseTurn, banPick int32
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
		if room.Host != token {
			return errors.New("没有权限")
		}
		if r, ok := room.Type().(RoomNextRoundHandler); !ok {
			return errors.New("不支持下一回合的游戏类型")
		} else {
			err = r.HandleNextRound()
			if err != nil {
				return err
			}
		}
		if room.BpData != nil {
			whoseTurn = room.BpData.WhoseTurn
			banPick = room.BpData.BanPick
		}
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayersInRoom(token, protoName, &myws.Message{
		Data: &NextRoundSc{
			WhoseTurn: whoseTurn,
			BanPick:   banPick,
		},
	})
	return nil
}

func (m *PauseCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	pause := m.Pause
	var totalPauseMs, pauseBeginMs int64
	now := time.Now().UnixMilli()
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
		if !room.Type().CanPause() {
			return errors.New("不支持暂停的游戏类型")
		}
		if room.Host != token {
			return errors.New("你不是房主")
		}
		if !room.Started {
			return errors.New("游戏还没开始，不能暂停")
		}
		if room.StartMs <= now-int64(room.GameTime)*60000-int64(room.Countdown)*1000-room.TotalPauseMs {
			return errors.New("游戏时间到，不能暂停")
		}
		if room.StartMs > now-int64(room.Countdown)*1000 {
			return errors.New("倒计时还没结束，不能暂停")
		}
		if pause {
			if room.PauseBeginMs == 0 {
				room.PauseBeginMs = now
			}
		} else {
			if room.PauseBeginMs != 0 {
				room.TotalPauseMs += now - room.PauseBeginMs
				room.PauseBeginMs = 0
			}
		}
		totalPauseMs = room.TotalPauseMs
		pauseBeginMs = room.PauseBeginMs
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayersInRoom(token, protoName, &myws.Message{
		Data: &PauseSc{
			Time:         now,
			TotalPauseMs: totalPauseMs,
			PauseBeginMs: pauseBeginMs,
		},
	})
	return nil
}

func (m *ChangeCardCountCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	counts := m.Counts
	if len(counts) != 2 || counts[0] > 9999 || counts[1] > 9999 {
		return errors.New("cnt参数错误")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
		if room.Host != token {
			return errors.New("你不是房主")
		}
		room.ChangeCardCount[0] = counts[0]
		room.ChangeCardCount[1] = counts[1]
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayerInfo(token, protoName)
	return nil
}

func (m *ResetRoomCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
		if room.Host != token {
			return errors.New("你不是房主")
		}
		if room.Started {
			return errors.New("游戏已开始，不能重置房间")
		}
		for i := range room.Score {
			room.Score[i] = 0
		}
		room.Locked = false
		for i := range room.ChangeCardCount {
			room.ChangeCardCount[i] = 0
		}
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayerInfo(token, protoName)
	return nil
}

func (m *UpdateSpellCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	idx := m.Idx
	if idx >= 25 {
		return errors.New("idx超出范围")
	}
	status := m.Status
	if status == SpellStatus_both_select {
		return errors.New("status不合法")
	}
	var newStatus SpellStatus
	var tokens []string
	var whoseTurn, banPick int32
	var trigger string
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
		if err != nil {
			return err
		}
		trigger = player.Name
		if len(player.RoomId) == 0 {
			return errors.New("不在房间里")
		}
		room, err := GetRoom(txn, player.RoomId)
		if err != nil {
			return err
		}
		if !room.Started {
			return errors.New("游戏还没开始")
		}
		tokens, newStatus, err = room.Type().HandleUpdateSpell(token, idx, status)
		if err != nil {
			return err
		}
		room.Status[idx] = newStatus
		if room.BpData != nil {
			whoseTurn = room.BpData.WhoseTurn
			banPick = room.BpData.BanPick
		}
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	for _, token1 := range tokens {
		if len(token1) > 0 {
			if conn, ok := s.tokenConnMap[token1]; ok {
				message := &myws.Message{
					Reply:   protoName,
					Trigger: trigger,
					Data: &UpdateSpellSc{
						Idx:       idx,
						Status:    newStatus,
						WhoseTurn: whoseTurn,
						BanPick:   banPick,
					},
				}
				if token1 == token {
					message.Reply = protoName
				} else {
					message.Reply = ""
				}
				conn.Send(message)
			}
		}
	}
	return nil
}

func (m *StopGameCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	winnerIdx := m.Winner
	if winnerIdx != -1 && winnerIdx != 0 && winnerIdx != 1 {
		return errors.New("winner不正确")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
		if room.Host != token {
			return errors.New("你不是房主")
		} else if !room.Started {
			return errors.New("游戏还没开始")
		}
		if winnerIdx >= 0 {
			room.Score[winnerIdx]++
			if room.Score[winnerIdx] >= room.NeedWin {
				room.Locked = false
			}
			room.LastWinner = winnerIdx + 1
		}
		room.Started = false
		room.Spells = nil
		room.StartMs = 0
		room.GameTime = 0
		room.Countdown = 0
		room.Status = nil
		room.TotalPauseMs = 0
		room.PauseBeginMs = 0
		room.BpData = nil
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	if winnerIdx == -1 {
		s.NotifyPlayerInfo(token, protoName)
	} else {
		s.NotifyPlayerInfo(token, protoName, winnerIdx)
	}
	return nil
}

func (m *GetSpellsCs) Handle(_ *bingoServer, session cellnet.Session, token, protoName string) error {
	var spells []*Spell
	var startTime int64
	var gameTime, countdown uint32
	var status []int32
	var needWin uint32
	var totalPauseMs, pauseBeginMs int64
	var whoseTurn, banPick, phase int32
	var linkData *LinkData
	var trigger string
	err := db.View(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
		if err != nil {
			return err
		}
		trigger = player.Name
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
		startTime = room.StartMs
		countdown = room.Countdown
		gameTime = room.GameTime
		needWin = room.NeedWin
		totalPauseMs = room.TotalPauseMs
		pauseBeginMs = room.PauseBeginMs
		for _, st := range room.Status {
			if token == room.Players[0] {
				status = append(status, int32(st.hideRightSelect()))
			} else if token == room.Players[1] {
				status = append(status, int32(st.hideLeftSelect()))
			} else {
				status = append(status, int32(st))
			}
		}
		if room.BpData != nil {
			whoseTurn = room.BpData.WhoseTurn
			banPick = room.BpData.BanPick
		}
		linkData = room.LinkData
		phase = room.Phase
		return nil
	})
	if err != nil {
		return err
	}
	session.Send(&myws.Message{
		Reply:   protoName,
		Trigger: trigger,
		Data: &SpellListSc{
			Spells:         spells,
			Time:           time.Now().UnixMilli(),
			StartTime:      startTime,
			GameTime:       gameTime,
			Countdown:      countdown,
			NeedWin:        needWin,
			WhoseTurn:      whoseTurn,
			BanPick:        banPick,
			TotalPauseTime: totalPauseMs,
			PauseBeginMs:   pauseBeginMs,
			Status:         status,
			Link:           linkData,
			Phase:          phase,
		},
	})
	return nil
}

func (m *StartGameCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	gameTime := m.GameTime
	if gameTime == 0 {
		return errors.New("游戏时间不能为0")
	}
	if gameTime > 1440 {
		return errors.New("游戏时间太长")
	}
	countdown := m.Countdown
	if countdown > 86400 {
		return errors.New("倒计时太长")
	}
	games := m.Games
	if len(games) > 99 {
		return errors.New("选择的作品数太多")
	}
	ranks := m.Ranks
	if len(ranks) > 6 {
		return errors.New("选择的难度数太多")
	}
	needWin := m.NeedWin
	if needWin > 99 {
		return errors.New("需要胜场的数值不正确")
	}
	if needWin == 0 {
		needWin = 1
	}
	startTime := time.Now().UnixMilli()
	var spells []*Spell
	var whoseTurn, banPick, phase int32
	var linkData *LinkData
	var trigger string
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
		if err != nil {
			return err
		}
		trigger = player.Name
		if len(player.RoomId) == 0 {
			return errors.New("不在房间里")
		}
		room, err := GetRoom(txn, player.RoomId)
		if err != nil {
			return err
		}
		if room.Host != token {
			return errors.New("你不是房主")
		} else if room.Started {
			return errors.New("游戏已经开始")
		} else if slices.ContainsFunc(room.Players, func(s string) bool { return len(s) == 0 }) {
			return errors.New("玩家没满")
		}
		spells, err = room.Type().RandSpells(games, ranks)
		if err != nil {
			return errors.Wrap(err, "随符卡失败")
		}
		room.Started = true
		room.Spells = spells
		room.StartMs = startTime
		room.Countdown = countdown
		room.GameTime = gameTime
		room.Status = make([]SpellStatus, len(spells))
		room.NeedWin = needWin
		room.Locked = true
		if roomType, ok := room.Type().(RoomStartHandler); ok {
			roomType.OnStart()
		}
		if room.BpData != nil {
			whoseTurn = room.BpData.WhoseTurn
			banPick = room.BpData.BanPick
		}
		linkData = room.LinkData
		phase = room.Phase
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	message := &myws.Message{
		Reply:   protoName,
		Trigger: trigger,
		Data: &SpellListSc{
			Spells:    spells,
			Time:      startTime,
			StartTime: startTime,
			GameTime:  gameTime,
			Countdown: countdown,
			NeedWin:   needWin,
			WhoseTurn: whoseTurn,
			BanPick:   banPick,
			Link:      linkData,
			Phase:     phase,
		},
	}
	s.NotifyPlayersInRoom(token, protoName, message)
	return nil
}

func (m *UpdateNameCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	name := m.Name
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	if len(name) > 48 {
		return errors.New("名字太长")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
	s.NotifyPlayerInfo(token, protoName)
	return nil
}

func (m *UpdateRoomTypeCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	roomType := m.Type
	if roomType < 1 || roomType > 3 {
		return errors.New("不支持的游戏类型")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return errors.New("不在房间里")
		}
		room, err := GetRoom(txn, player.RoomId)
		if IsErrKeyNotFound(err) {
			return errors.New("不在房间里")
		} else if err != nil {
			return err
		}
		if room.Host != token {
			return errors.New("不是房主")
		}
		room.RoomType = roomType
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayerInfo(token, protoName)
	return nil
}

func (m *LeaveRoomCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
		if err != nil {
			return err
		}
		if len(player.RoomId) == 0 {
			return errors.New("并不在房间里")
		}
		room, err := GetRoom(txn, player.RoomId)
		if IsErrKeyNotFound(err) {
			return errors.New("不在房间里")
		} else if err != nil {
			return err
		}
		if room.GetStarted() {
			return errors.New("比赛已经开始了，不能退出")
		}
		if room.Host != player.Token && room.Locked {
			return errors.New("连续比赛没结束，不能退出")
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
	s.NotifyPlayerInfo(token, protoName)
	return nil
}

func (m *JoinRoomCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	name := m.Name
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	if len(name) > 48 {
		return errors.New("名字太长")
	}
	rid := m.RoomId
	if len(rid) == 0 {
		return errors.New("房间ID为空")
	}
	if len(rid) > 16 {
		return errors.New("房间ID太长")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
		if err != nil {
			return err
		}
		if len(player.RoomId) != 0 {
			_, err = GetRoom(txn, player.RoomId)
			if err == nil {
				return errors.New("已经在房间里了")
			} else if !IsErrKeyNotFound(err) {
				return errors.WithStack(err)
			}
		}
		room, err := GetRoom(txn, rid)
		if IsErrKeyNotFound(err) {
			return errors.New("房间不存在")
		} else if err != nil {
			return err
		}
		host, err := GetPlayer(txn, room.Host)
		if err != nil {
			return err
		}
		if host.Name == name {
			return errors.New("名字重复")
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
	s.NotifyPlayerInfo(token, protoName)
	return nil
}

func (m *CreateRoomCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	name := m.Name
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	if len(name) > 48 {
		return errors.New("名字太长")
	}
	rid := m.RoomId
	if len(rid) == 0 {
		return errors.New("房间ID为空")
	}
	if len(rid) > 16 {
		return errors.New("房间ID太长")
	}
	roomType := m.Type
	if roomType < 1 || roomType > 3 {
		return errors.New("不支持的游戏类型")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
		if err != nil {
			return err
		}
		if len(player.RoomId) != 0 {
			_, err = GetRoom(txn, player.RoomId)
			if err == nil {
				return errors.New("已经在房间里了")
			} else if !IsErrKeyNotFound(err) {
				return errors.WithStack(err)
			}
		}
		key := append([]byte("room: "), []byte(rid)...)
		_, err = txn.Get(key)
		if err == nil {
			return errors.New("房间已存在")
		} else if err != badger.ErrKeyNotFound {
			return errors.WithStack(err)
		}
		var room = Room{
			RoomId:          rid,
			RoomType:        roomType,
			Host:            token,
			Players:         make([]string, 2),
			Score:           make([]uint32, 2),
			ChangeCardCount: make([]uint32, 2),
		}
		player.RoomId = rid
		player.Name = name
		if err = SetPlayer(txn, player); err != nil {
			return err
		}
		return SetRoom(txn, &room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayerInfo(token, protoName)
	return nil
}

func (m *HeartCs) Handle(_ *bingoServer, session cellnet.Session, _, protoName string) error {
	session.Send(&myws.Message{
		Reply: protoName,
		Data: &HeartSc{
			Time: time.Now().UnixMilli(),
		},
	})
	return nil
}

func isAlphaNum(s string) bool {
	for _, c := range []byte(s) {
		if (c < '0' || c > '9') && (c < 'A' || c > 'Z') && (c < 'a' || c > 'z') {
			return false
		}
	}
	return true
}

func (m *LoginCs) Handle(s *bingoServer, session cellnet.Session, _, protoName string) error {
	tokenStr := m.Token
	if len(tokenStr) == 0 || len(tokenStr) > 128 || !isAlphaNum(tokenStr) {
		session.Send(&myws.Message{Reply: protoName, Data: &ErrorSc{Code: 400, Msg: "invalid token"}})
		return nil
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayerOrNew(txn, tokenStr)
		if err != nil {
			return err
		}
		player.Token = tokenStr
		if _, ok := s.tokenConnMap[tokenStr]; ok {
			return errors.New("already online")
		} else {
			s.tokenConnMap[tokenStr] = session
		}
		return SetPlayer(txn, player)
	})
	if err != nil {
		return err
	}
	session.(cellnet.ContextSet).SetContext(playerConnToken, tokenStr)
	message, _, err := s.buildPlayerInfo(tokenStr)
	if err != nil {
		return err
	}
	message.Reply = protoName
	session.Send(message)
	return nil
}

func (m *LinkTimeCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	if m.Whose != 0 && m.Whose != 1 {
		return errors.New("参数错误")
	}
	var message *LinkDataSc
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
		if room.Host != token {
			return errors.New("你不是房主")
		}
		if _, ok := room.Type().(RoomTypeLink); !ok {
			return errors.New("不支持这种操作")
		}
		data := room.LinkData
		if m.Whose == 0 {
			if m.Start {
				if data.StartMsA == 0 && data.EndMsA == 0 { // 开始
					data.StartMsA = time.Now().UnixMilli()
				} else if data.StartMsA > 0 && data.EndMsA > 0 { // 继续
					data.StartMsA = time.Now().UnixMilli() - (data.EndMsA - data.StartMsA)
					data.EndMsA = 0
				} else {
					return errors.New("已经在计时了，不能开始")
				}
			} else {
				if data.StartMsA > 0 && data.EndMsA == 0 { // 停止/暂停
					data.EndMsA = time.Now().UnixMilli()
				} else {
					return errors.New("还未开始计时，不能停止")
				}
			}
		} else {
			if m.Start {
				if data.StartMsB == 0 && data.EndMsB == 0 { // 开始
					data.StartMsB = time.Now().UnixMilli()
				} else if data.StartMsB > 0 && data.EndMsB > 0 { // 继续
					data.StartMsB = time.Now().UnixMilli() - (data.EndMsB - data.StartMsB)
					data.EndMsB = 0
				} else {
					return errors.New("已经在计时了，不能开始")
				}
			} else {
				if data.StartMsB > 0 && data.EndMsB == 0 { // 停止/暂停
					data.EndMsB = time.Now().UnixMilli()
				} else {
					return errors.New("还未开始计时，不能停止")
				}
			}
		}
		message = (*LinkDataSc)(data)
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayersInRoom(token, protoName, &myws.Message{Data: message})
	return nil
}

func (m *SetPhaseCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, token)
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
		if room.Host != token {
			return errors.New("你不是房主")
		}
		room.Phase = m.Phase
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayersInRoom(token, protoName, &myws.Message{Data: &SetPhaseSc{Phase: m.Phase}})
	return nil
}
