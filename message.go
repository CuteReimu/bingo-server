package main

import (
	"encoding/json"
	_ "github.com/Touhou-Freshman-Camp/bingo-server/myws"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/util"
	"github.com/ozgio/strutil"
	"reflect"
	"strings"
)

type MessageHandler interface {
	Handle(s *bingoServer, session cellnet.Session, token, protoName string) error
}

func init() {
	msglog.SetCurrMsgLogMode(msglog.MsgLogMode_BlackList)
	c := codec.MustGetCodec("json")
	initMessage(c, (*LoginCs)(nil))
	initMessage(c, (*ErrorSc)(nil))
	initMessage(c, (*RoomInfoSc)(nil))
	initMessage(c, (*HeartCs)(nil))
	initMessage(c, (*HeartSc)(nil))
	initMessage(c, (*CreateRoomCs)(nil))
	initMessage(c, (*JoinRoomCs)(nil))
	initMessage(c, (*LeaveRoomCs)(nil))
	initMessage(c, (*UpdateRoomTypeCs)(nil))
	initMessage(c, (*UpdateNameCs)(nil))
	initMessage(c, (*StartGameCs)(nil))
	initMessage(c, (*SpellListSc)(nil))
	initMessage(c, (*GetSpellsCs)(nil))
	initMessage(c, (*StopGameCs)(nil))
	initMessage(c, (*UpdateSpellCs)(nil))
	initMessage(c, (*UpdateSpellSc)(nil))
	initMessage(c, (*ResetRoomCs)(nil))
	initMessage(c, (*ChangeCardCountCs)(nil))
	initMessage(c, (*PauseCs)(nil))
	initMessage(c, (*PauseSc)(nil))
	initMessage(c, (*NextRoundCs)(nil))
	initMessage(c, (*NextRoundSc)(nil))
	initMessage(c, (*LinkTimeCs)(nil))
	initMessage(c, (*LinkDataSc)(nil))
	initMessage(c, (*SetPhaseCs)(nil))
	initMessage(c, (*SetPhaseSc)(nil))
}

func initMessage(c cellnet.Codec, i interface{}) {
	t := reflect.TypeOf(i).Elem()
	name := strings.Replace(t.Name(), "Msg", "", 1)
	name = strutil.ToSnakeCase(strings.Join(strutil.SplitCamelCase(name), " "))
	meta := &cellnet.MessageMeta{
		Codec: c,
		Type:  t,
		ID:    int(util.StringHash(name)),
	}
	meta.SetContext("name", name)
	cellnet.RegisterMessageMeta(meta)
	if name == "heart_sc" || name == "heart_cs" {
		if err := msglog.SetMsgLogRule(meta.FullName(), msglog.MsgLogRule_BlackList); err != nil {
			panic(err)
		}
	}
}

type LoginCs struct {
	Token string `json:"token"`
}

func (m *LoginCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type ErrorSc struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func (m *ErrorSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type RoomInfoSc struct {
	RoomId          string   `json:"rid,omitempty"`
	Type            int32    `json:"type,omitempty"`
	HostName        string   `json:"host,omitempty"`
	PlayerNames     []string `json:"names,omitempty"`
	ChangeCardCount []uint32 `json:"change_card_count,omitempty"`
	Started         bool     `json:"started,omitempty"`
	Score           []uint32 `json:"score,omitempty"`
	Winner          int32    `json:"winner,omitempty"`
}

func (m *RoomInfoSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type CreateRoomCs struct {
	Name   string `json:"name"`
	RoomId string `json:"rid"`
	Type   int32  `json:"type"`
}

func (m *CreateRoomCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type JoinRoomCs struct {
	Name   string `json:"name"`
	RoomId string `json:"rid"`
}

func (m *JoinRoomCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type LeaveRoomCs struct {
}

func (m *LeaveRoomCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type UpdateRoomTypeCs struct {
	Type int32 `json:"type"`
}

func (m *UpdateRoomTypeCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type UpdateNameCs struct {
	Name string `json:"name"`
}

func (m *UpdateNameCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type HeartCs struct {
}

func (m *HeartCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type HeartSc struct {
	Time int64 `json:"time"`
}

func (m *HeartSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type StartGameCs struct {
	GameTime  uint32   `json:"game_time"` // 游戏总时间（不含倒计时），单位：分
	Countdown uint32   `json:"countdown"` // 倒计时，单位：秒
	Games     []string `json:"games"`
	Ranks     []string `json:"ranks"`
	NeedWin   uint32   `json:"need_win"`
}

func (m *StartGameCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type SpellListSc struct {
	Spells         []*Spell  `json:"spells"`
	Time           int64     `json:"time"`
	StartTime      int64     `json:"start_time"`
	GameTime       uint32    `json:"game_time"` // 游戏总时间（不含倒计时），单位：分
	Countdown      uint32    `json:"countdown"` // 倒计时，单位：秒
	NeedWin        uint32    `json:"need_win"`
	WhoseTurn      int32     `json:"whose_turn"`
	BanPick        int32     `json:"ban_pick"`
	TotalPauseTime int64     `json:"total_pause_time,omitempty"`
	PauseBeginMs   int64     `json:"pause_begin_ms,omitempty"`
	Status         []int32   `json:"status,omitempty"`
	Phase          int32     `json:"phase"`
	Link           *LinkData `json:"link_data,omitempty"`
}

func (m *SpellListSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type GetSpellsCs struct {
}

func (m *GetSpellsCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type StopGameCs struct {
	Winner int32 `json:"winner"`
}

func (m *StopGameCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type UpdateSpellCs struct {
	Idx    uint32      `json:"idx"`
	Status SpellStatus `json:"status"`
}

func (m *UpdateSpellCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type UpdateSpellSc struct {
	Idx       uint32      `json:"idx"`
	Status    SpellStatus `json:"status"`
	WhoseTurn int32       `json:"whose_turn"`
	BanPick   int32       `json:"ban_pick"`
}

func (m *UpdateSpellSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type ResetRoomCs struct {
}

func (m *ResetRoomCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type ChangeCardCountCs struct {
	Counts []uint32 `json:"cnt"`
}

func (m *ChangeCardCountCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type PauseCs struct {
	Pause bool `json:"pause"`
}

func (m *PauseCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type PauseSc struct {
	Time         int64 `json:"time"`
	TotalPauseMs int64 `json:"total_pause_ms,omitempty"`
	PauseBeginMs int64 `json:"pause_begin_ms,omitempty"`
}

func (m *PauseSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type NextRoundCs struct {
}

func (m *NextRoundCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type NextRoundSc struct {
	WhoseTurn int32 `json:"whose_turn"`
	BanPick   int32 `json:"ban_pick"`
}

func (m *NextRoundSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type LinkTimeCs struct {
	Whose int32 `json:"whose"`
	Start bool  `json:"start"`
}

func (m *LinkTimeCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type LinkDataSc LinkData

func (m *LinkDataSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type SetPhaseCs struct {
	Phase int32 `json:"phase"`
}

func (m *SetPhaseCs) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}

type SetPhaseSc struct {
	Phase int32 `json:"phase"`
}

func (m *SetPhaseSc) String() string {
	buf, _ := json.Marshal(m)
	return string(buf)
}
