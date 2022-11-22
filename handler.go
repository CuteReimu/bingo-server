package main

import (
	"time"

	"github.com/CuteReimu/goutil/slices"
	"github.com/Touhou-Freshman-Camp/bingo-server/myws"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
)

var handlers = map[string]func(player *PlayerConn, protoName string, result interface{}) error{
	"login_cs":             handleLogin,
	"heart_cs":             handleHeart,
	"create_room_cs":       handleCreateRoom,
	"join_room_cs":         handleJoinRoom,
	"leave_room_cs":        handleLeaveRoom,
	"update_room_type_cs":  handleUpdateRoomType,
	"update_name_cs":       handleUpdateName,
	"start_game_cs":        handleStartGame,
	"get_spells_cs":        handleGetSpells,
	"stop_game_cs":         handleStopGame,
	"update_spell_cs":      handleUpdateSpell,
	"reset_room_cs":        handleResetRoom,
	"change_card_count_cs": handleChangeCardCount,
	"pause_cs":             handlePause,
	"next_round_cs":        handleNextRound,
}

func handleNextRound(playerConn *PlayerConn, protoName string, _ interface{}) error {
	var whoseTurn, banPick int32
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
	playerConn.NotifyPlayersInRoom(protoName, &myws.Message{
		Data: &NextRoundSc{
			WhoseTurn: whoseTurn,
			BanPick:   banPick,
		},
	})
	return nil
}

func handlePause(playerConn *PlayerConn, protoName string, data interface{}) error {
	pause := data.(*PauseCs).Pause
	var totalPauseMs, pauseBeginMs int64
	now := time.Now().UnixMilli()
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
		if !room.Type().CanPause() {
			return errors.New("不支持暂停的游戏类型")
		}
		if room.Host != playerConn.token {
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
	msgData := map[string]interface{}{
		"time": now,
	}
	if totalPauseMs > 0 {
		msgData["total_pause_ms"] = totalPauseMs
	}
	if pauseBeginMs > 0 {
		msgData["pause_begin_ms"] = pauseBeginMs
	}
	playerConn.NotifyPlayersInRoom(protoName, &myws.Message{
		Data: &PauseSc{
			Time:         now,
			TotalPauseMs: totalPauseMs,
			PauseBeginMs: pauseBeginMs,
		},
	})
	return nil
}

func handleChangeCardCount(playerConn *PlayerConn, protoName string, data interface{}) error {
	counts := data.(*ChangeCardCountCs).Counts
	if len(counts) != 2 || counts[0] > 9999 || counts[1] > 9999 {
		return errors.New("cnt参数错误")
	}
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
		}
		room.ChangeCardCount[0] = counts[0]
		room.ChangeCardCount[1] = counts[1]
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}

func handleResetRoom(playerConn *PlayerConn, protoName string, _ interface{}) error {
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
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}

func handleUpdateSpell(playerConn *PlayerConn, protoName string, data interface{}) error {
	data0 := data.(*UpdateSpellCs)
	idx := data0.Idx
	if idx >= 25 {
		return errors.New("idx超出范围")
	}
	status := data0.Status
	if status == SpellStatus_both_select {
		return errors.New("status不合法")
	}
	var newStatus SpellStatus
	var tokens []string
	var whoseTurn, banPick int32
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
		if !room.Started {
			return errors.New("游戏还没开始")
		}
		tokens, newStatus, err = room.Type().HandleUpdateSpell(playerConn, idx, status)
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
	for _, token := range tokens {
		if len(token) > 0 {
			if conn, ok := tokenConnMap[token]; ok {
				message := &myws.Message{
					Data: &UpdateSpellSc{
						Idx:       idx,
						Status:    newStatus,
						WhoseTurn: whoseTurn,
						BanPick:   banPick,
					},
				}
				if token == playerConn.token {
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

func handleStopGame(playerConn *PlayerConn, protoName string, data interface{}) error {
	winnerIdx := data.(*StopGameCs).Winner
	if winnerIdx != -1 && winnerIdx != 0 && winnerIdx != 1 {
		return errors.New("winner不正确")
	}
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
		playerConn.NotifyPlayerInfo(protoName)
	} else {
		playerConn.NotifyPlayerInfo(protoName, winnerIdx)
	}
	return nil
}

func handleGetSpells(playerConn *PlayerConn, protoName string, _ interface{}) error {
	var spells []*Spell
	var startTime int64
	var gameTime, countdown uint32
	var status []int32
	var needWin uint32
	var totalPauseMs, pauseBeginMs int64
	var whoseTurn, banPick int32
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
		startTime = room.StartMs
		countdown = room.Countdown
		gameTime = room.GameTime
		needWin = room.NeedWin
		totalPauseMs = room.TotalPauseMs
		pauseBeginMs = room.PauseBeginMs
		if playerConn.token == room.Players[0] {
			status = slices.Map(len(room.Status), func(i int) (int32, bool) { return int32(room.Status[i].hideRightSelect()), true })
		} else if playerConn.token == room.Players[1] {
			status = slices.Map(len(room.Status), func(i int) (int32, bool) { return int32(room.Status[i].hideLeftSelect()), true })
		} else {
			status = slices.Map(len(room.Status), func(i int) (int32, bool) { return int32(room.Status[i]), true })
		}
		if room.BpData != nil {
			whoseTurn = room.BpData.WhoseTurn
			banPick = room.BpData.BanPick
		}
		if roomType, ok := room.Type().(interface{ GetWhoseTurnAndBanPick() (int32, int32) }); ok {
			whoseTurn, banPick = roomType.GetWhoseTurnAndBanPick()
		}
		return nil
	})
	if err != nil {
		return err
	}
	playerConn.Send(&myws.Message{
		Reply: protoName,
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
		},
	})
	return nil
}

func handleStartGame(playerConn *PlayerConn, protoName string, data interface{}) error {
	data0 := data.(*StartGameCs)
	gameTime := data0.GameTime
	if gameTime == 0 {
		return errors.New("游戏时间不能为0")
	}
	if gameTime > 1440 {
		return errors.New("游戏时间太长")
	}
	countdown := data0.Countdown
	if countdown > 86400 {
		return errors.New("倒计时太长")
	}
	games := data0.Games
	if len(games) > 99 {
		return errors.New("选择的作品数太多")
	}
	ranks := data0.Ranks
	if len(ranks) > 6 {
		return errors.New("选择的难度数太多")
	}
	needWin := data0.NeedWin
	if needWin > 99 {
		return errors.New("需要胜场的数值不正确")
	}
	if needWin == 0 {
		needWin = 1
	}
	startTime := time.Now().UnixMilli()
	var spells []*Spell
	var whoseTurn, banPick int32
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
		} else if room.Started {
			return errors.New("游戏已经开始")
		} else if slices.Any(len(room.Players), func(i int) bool { return len(room.Players[i]) == 0 }) {
			return errors.New("玩家没满")
		}
		spells, err = RandSpells(games, ranks, room.Type().CardCount())
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
		return SetRoom(txn, room)
	})
	if err != nil {
		return err
	}
	message := &myws.Message{
		Data: &SpellListSc{
			Spells:    spells,
			Time:      startTime,
			StartTime: startTime,
			GameTime:  gameTime,
			Countdown: countdown,
			NeedWin:   needWin,
			WhoseTurn: whoseTurn,
			BanPick:   banPick,
		},
	}
	playerConn.NotifyPlayersInRoom(protoName, message)
	return nil
}

func handleUpdateName(playerConn *PlayerConn, protoName string, data interface{}) error {
	name := data.(*UpdateNameCs).Name
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	if len(name) > 48 {
		return errors.New("名字太长")
	}
	err := db.Update(func(txn *badger.Txn) error {
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

func handleUpdateRoomType(playerConn *PlayerConn, protoName string, data interface{}) error {
	roomType := data.(*UpdateRoomTypeCs).Type
	if roomType < 1 || roomType > 3 {
		return errors.New("不支持的游戏类型")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
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

func handleLeaveRoom(playerConn *PlayerConn, protoName string, _ interface{}) error {
	var tokens []string
	var roomDestroyed bool
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
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
					tokens = append(tokens, room.Players[i])
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
			roomDestroyed = true
		} else {
			for i := range room.Players {
				if room.Players[i] == player.Token {
					room.Players[i] = ""
				} else {
					tokens = append(tokens, room.Players[i])
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
	for _, token := range tokens {
		conn := tokenConnMap[token]
		if conn != nil {
			if roomDestroyed {
				conn.Send(&myws.Message{Data: &RoomInfoSc{}})
			} else {
				conn.NotifyPlayerInfo("")
				break
			}
		}
	}
	return nil
}

func handleJoinRoom(playerConn *PlayerConn, protoName string, data interface{}) error {
	data0 := data.(*JoinRoomCs)
	name := data0.Name
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	if len(name) > 48 {
		return errors.New("名字太长")
	}
	rid := data0.RoomId
	if len(rid) == 0 {
		return errors.New("房间ID为空")
	}
	if len(rid) > 16 {
		return errors.New("房间ID太长")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
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
	playerConn.NotifyPlayerInfo(protoName)
	return nil
}

func handleCreateRoom(playerConn *PlayerConn, protoName string, data interface{}) error {
	data0 := data.(*CreateRoomCs)
	name := data0.Name
	if len(name) == 0 {
		return errors.New("名字为空")
	}
	if len(name) > 48 {
		return errors.New("名字太长")
	}
	rid := data0.RoomId
	if len(rid) == 0 {
		return errors.New("房间ID为空")
	}
	if len(rid) > 16 {
		return errors.New("房间ID太长")
	}
	roomType := data0.Type
	if roomType < 1 || roomType > 3 {
		return errors.New("不支持的游戏类型")
	}
	err := db.Update(func(txn *badger.Txn) error {
		player, err := GetPlayer(txn, playerConn.token)
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
			Host:            playerConn.token,
			Players:         make([]string, 2),
			Score:           make([]uint32, 2),
			ChangeCardCount: make([]uint32, 2),
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

func handleHeart(playerConn *PlayerConn, protoName string, _ interface{}) error {
	playerConn.Send(&myws.Message{
		Reply: protoName,
		Data: &HeartSc{
			Time: time.Now().UnixMilli(),
		},
	})
	return nil
}

func handleLogin(playerConn *PlayerConn, protoName string, data interface{}) error {
	tokenStr := data.(*LoginCs).Token
	if len(tokenStr) == 0 || len(tokenStr) > 128 || !isAlphaNum(tokenStr) {
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
	message, _, err := playerConn.buildPlayerInfo()
	message.Reply = protoName
	if err != nil {
		return err
	}
	playerConn.Send(message)
	return nil
}
