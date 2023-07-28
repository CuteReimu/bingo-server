package main

import (
	"github.com/CuteReimu/bingo-server/myws"
	"github.com/davyxu/cellnet"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
	"golang.org/x/exp/slices"
	"regexp"
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
		if !room.IsAdmin(token) {
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
		if !room.IsAdmin(token) {
			return errors.New("没有权限")
		}
		if !room.Started {
			return errors.New("游戏还没开始，不能暂停")
		}
		if pause && room.StartMs <= now-int64(room.GameTime)*60000-int64(room.Countdown)*1000-room.TotalPauseMs {
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
		if !room.IsAdmin(token) {
			return errors.New("没有权限")
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
		if !room.IsAdmin(token) {
			return errors.New("没有权限")
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
			room.LastGetTime[i] = 0
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
		now := time.Now().UnixMilli()
		tokens, newStatus, err = room.Type().HandleUpdateSpell(token, idx, status, now)
		if err != nil {
			return err
		}
		room.Status[idx] = newStatus
		if playerIndex := slices.Index(room.Players, token); playerIndex >= 0 && status.isGetStatus() {
			room.LastGetTime[playerIndex] = now
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
		if !room.IsAdmin(token) {
			return errors.New("没有权限")
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
	return db.View(func(txn *badger.Txn) error {
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
		if !room.Started {
			return errors.New("游戏还未开始")
		}
		status := make([]int32, 0, len(room.Status))
		for _, st := range room.Status {
			status = append(status, int32(st))
		}
		var whoseTurn, banPick int32
		if room.BpData != nil {
			whoseTurn = room.BpData.WhoseTurn
			banPick = room.BpData.BanPick
		}
		session.Send(&myws.Message{
			Reply:   protoName,
			Trigger: player.Name,
			Data: &SpellListSc{
				Spells:         room.Spells,
				Time:           time.Now().UnixMilli(),
				StartTime:      room.StartMs,
				GameTime:       room.GameTime,
				Countdown:      room.Countdown,
				NeedWin:        room.NeedWin,
				WhoseTurn:      whoseTurn,
				BanPick:        banPick,
				TotalPauseTime: room.TotalPauseMs,
				PauseBeginMs:   room.PauseBeginMs,
				Status:         status,
				Phase:          room.Phase,
				Link:           room.LinkData,
				Difficulty:     room.Difficulty,
				EnableTools:    room.EnableTools,
				LastGetTime:    room.LastGetTime,
			},
		})
		return nil
	})
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
	var lastGetTime []int64
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
		if !room.IsAdmin(token) {
			return errors.New("没有权限")
		} else if room.Started {
			return errors.New("游戏已经开始")
		} else if slices.ContainsFunc(room.Players, func(s string) bool { return len(s) == 0 }) {
			return errors.New("玩家没满")
		}
		var difficulty [3]int
		switch m.Difficulty {
		case 1:
			difficulty = difficultyE
		case 2:
			difficulty = difficultyN
		case 3:
			difficulty = difficultyL
		default:
			difficulty = difficultyRandom()
		}
		spells, err = room.Type().RandSpells(games, ranks, difficulty)
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
		room.Difficulty = m.Difficulty
		room.EnableTools = m.EnableTools
		if roomType, ok := room.Type().(RoomStartHandler); ok {
			roomType.OnStart()
		}
		if room.BpData != nil {
			whoseTurn = room.BpData.WhoseTurn
			banPick = room.BpData.BanPick
		}
		linkData = room.LinkData
		phase = room.Phase
		if !m.IsPrivate && !slices.Contains(room.Players, robotPlayer.Token) { // 单人练习模式不推送
			miraiPusher.Push(room)
		}
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	message := &myws.Message{
		Trigger: trigger,
		Data: &SpellListSc{
			Spells:      spells,
			Time:        startTime,
			StartTime:   startTime,
			GameTime:    gameTime,
			Countdown:   countdown,
			NeedWin:     needWin,
			WhoseTurn:   whoseTurn,
			BanPick:     banPick,
			Link:        linkData,
			Phase:       phase,
			Difficulty:  m.Difficulty,
			EnableTools: m.EnableTools,
			LastGetTime: lastGetTime,
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
		if !room.IsAdmin(token) {
			return errors.New("没有权限")
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
	var tokens []string
	var roomDestroyed bool
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
		if room.Host == token {
			for _, p := range room.Players {
				if len(p) > 0 && p != robotPlayer.Token {
					tokens = append(tokens, p)
					if err = SetPlayer(txn, &Player{Token: p}); err != nil {
						return err
					}
				}
			}
			for _, p := range room.Watchers {
				tokens = append(tokens, p)
				if err = SetPlayer(txn, &Player{Token: p}); err != nil {
					return err
				}
			}
			roomDestroyed = true
		} else {
			if index := slices.Index(room.Players, token); index >= 0 {
				room.Players[index] = ""
			} else {
				if index = slices.Index(room.Watchers, token); index >= 0 {
					room.Watchers = append(room.Watchers[:index], room.Watchers[index+1:]...)
				}
			}
			var roomPlayers []string
			for _, s := range room.Players {
				if len(s) > 0 && s != robotPlayer.Token {
					roomPlayers = append(tokens, s)
				}
			}
			if len(room.Host) > 0 {
				tokens = append(tokens, room.Host)
			}
			tokens = append(tokens, roomPlayers...)
			tokens = append(tokens, room.Watchers...)
			roomDestroyed = len(room.Host) == 0 && len(roomPlayers) == 0
		}
		if roomDestroyed {
			if err = DelRoom(txn, room.RoomId); err != nil {
				return err
			}
		} else {
			if err = SetRoom(txn, room); err != nil {
				return err
			}
		}
		player.RoomId = ""
		player.Name = ""
		return SetPlayer(txn, player)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayerInfo(token, protoName) // 已经退出了，所以这里只能通知到自己
	// 需要再通知房间里的其他人
	var message *myws.Message
	for _, t := range tokens {
		conn := s.tokenConnMap[t]
		if conn != nil {
			if roomDestroyed {
				conn.Send(&myws.Message{
					Trigger: token,
					Data:    &RoomInfoSc{},
				})
			} else {
				if message == nil {
					message, _, err = s.buildPlayerInfo(t)
					if err != nil {
						log.Errorf("db error: %+v", err)
					} else {
						message.Trigger = token
					}
				}
				if message != nil {
					conn.Send(message)
				}
			}
		}
	}
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
		if len(room.Host) > 0 {
			host, err := GetPlayer(txn, room.Host)
			if err != nil {
				return err
			}
			if host.Name == name {
				return errors.New("名字重复")
			}
		}
		hasSameName := func(token1 string) bool {
			if len(token1) == 0 {
				return false
			}
			player, _ := GetPlayer(txn, token1)
			return player != nil && player.Name == name
		}
		if slices.ContainsFunc(room.Players, hasSameName) {
			return errors.New("名字重复")
		}
		if slices.ContainsFunc(room.Watchers, hasSameName) {
			return errors.New("名字重复")
		}
		index := slices.Index(room.Players, "")
		if index >= 0 {
			room.Players[index] = token
		} else {
			room.Watchers = append(room.Watchers, token)
		}
		player.RoomId = rid
		player.Name = name
		if err = SetPlayer(txn, player); err != nil {
			return err
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
	if matched, _ := regexp.MatchString(`^\d{1,16}$`, rid); !matched {
		return errors.New("房间ID不合法")
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
		var host string
		if !m.Solo {
			host = token
		}
		var room = Room{
			RoomId:          rid,
			RoomType:        roomType,
			Host:            host,
			Players:         make([]string, 2),
			Score:           make([]uint32, 2),
			ChangeCardCount: make([]uint32, 2),
			LastGetTime:     make([]int64, 2),
		}
		if m.Solo {
			room.Players[0] = token
		}
		if m.AddRobot {
			room.Players[1] = robotPlayer.Token
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
	if oldChannel := s.tokenConnMap[m.Token]; oldChannel != nil {
		log.Warnln("already online, kick old session")
		oldChannel.Send(&myws.Message{Data: &ErrorSc{Code: -403, Msg: "有另一个客户端登录了此账号"}})
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayerOrNew(txn, tokenStr)
		if err != nil {
			return err
		}
		player.Token = tokenStr
		return SetPlayer(txn, player)
	})
	if err != nil {
		return err
	}
	s.tokenConnMap[tokenStr] = session
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
	if m.Event != 1 && m.Event != 2 && m.Event != 3 {
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
		if !room.IsAdmin(token) {
			return errors.New("没有权限")
		}
		if _, ok := room.Type().(RoomTypeLink); !ok {
			return errors.New("不支持这种操作")
		}
		data := room.LinkData
		if m.Whose == 0 {
			switch m.Event {
			case 1:
				if data.StartMsA == 0 && data.EndMsA == 0 { // 开始
					data.StartMsA = time.Now().UnixMilli()
				} else if data.StartMsA > 0 && data.EndMsA > 0 { // 继续
					data.StartMsA = time.Now().UnixMilli() - (data.EndMsA - data.StartMsA)
					data.EndMsA = 0
				} else {
					return errors.New("已经在计时了，不能开始")
				}
			case 2, 3:
				if data.StartMsA > 0 && data.EndMsA == 0 { // 停止/暂停
					data.EndMsA = time.Now().UnixMilli()
				} else {
					return errors.New("还未开始计时，不能停止")
				}
			}
			data.EventA = m.Event
		} else {
			switch m.Event {
			case 1:
				if data.StartMsB == 0 && data.EndMsB == 0 { // 开始
					data.StartMsB = time.Now().UnixMilli()
				} else if data.StartMsB > 0 && data.EndMsB > 0 { // 继续
					data.StartMsB = time.Now().UnixMilli() - (data.EndMsB - data.StartMsB)
					data.EndMsB = 0
				} else {
					return errors.New("已经在计时了，不能开始")
				}
			case 2, 3:
				if data.StartMsB > 0 && data.EndMsB == 0 { // 停止/暂停
					data.EndMsB = time.Now().UnixMilli()
				} else {
					return errors.New("还未开始计时，不能停止")
				}
			}
			data.EventB = m.Event
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
		if !room.IsAdmin(token) {
			return errors.New("没有权限")
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

func (m *SitDownCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
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
		if room.Started {
			return errors.New("游戏已开始")
		}
		if room.Locked {
			return errors.New("连续比赛还未结束")
		}
		index := slices.Index(room.Players, "")
		if index < 0 {
			return errors.New("房间已满")
		}
		watcherIndex := slices.Index(room.Watchers, token)
		if watcherIndex < 0 {
			return errors.New("你不是观众")
		}
		room.Watchers = append(room.Watchers[:watcherIndex], room.Watchers[watcherIndex+1:]...)
		room.Players[index] = token
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayerInfo(token, protoName)
	return nil
}

func (m *StandUpCs) Handle(s *bingoServer, _ cellnet.Session, token, protoName string) error {
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
		if room.Started {
			return errors.New("游戏已开始")
		}
		if room.Locked {
			return errors.New("连续比赛还未结束")
		}
		index := slices.Index(room.Players, token)
		if index < 0 {
			return errors.New("你不是选手")
		}
		if len(room.Host) == 0 && (len(room.Players[1-index]) == 0 || room.Players[1-index] == robotPlayer.Token) {
			return errors.New("你是房间里的最后一位选手，不能成为观众")
		}
		room.Players[index] = ""
		room.Watchers = append(room.Watchers, token)
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	s.NotifyPlayerInfo(token, protoName)
	return nil
}
