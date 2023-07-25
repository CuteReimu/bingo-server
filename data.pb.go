// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v3.21.12
// source: data.proto

package main

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type SpellStatus int32

const (
	SpellStatus_none         SpellStatus = 0
	SpellStatus_banned       SpellStatus = -1
	SpellStatus_left_select  SpellStatus = 1
	SpellStatus_right_select SpellStatus = 3
	SpellStatus_left_get     SpellStatus = 5
	SpellStatus_right_get    SpellStatus = 7
	SpellStatus_both_select  SpellStatus = 2
	SpellStatus_both_get     SpellStatus = 6
)

// Enum value maps for SpellStatus.
var (
	SpellStatus_name = map[int32]string{
		0:  "none",
		-1: "banned",
		1:  "left_select",
		3:  "right_select",
		5:  "left_get",
		7:  "right_get",
		2:  "both_select",
		6:  "both_get",
	}
	SpellStatus_value = map[string]int32{
		"none":         0,
		"banned":       -1,
		"left_select":  1,
		"right_select": 3,
		"left_get":     5,
		"right_get":    7,
		"both_select":  2,
		"both_get":     6,
	}
)

func (x SpellStatus) Enum() *SpellStatus {
	p := new(SpellStatus)
	*p = x
	return p
}

func (x SpellStatus) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (SpellStatus) Descriptor() protoreflect.EnumDescriptor {
	return file_data_proto_enumTypes[0].Descriptor()
}

func (SpellStatus) Type() protoreflect.EnumType {
	return &file_data_proto_enumTypes[0]
}

func (x SpellStatus) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use SpellStatus.Descriptor instead.
func (SpellStatus) EnumDescriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{0}
}

type Player struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Token  string `protobuf:"bytes,1,opt,name=token,proto3" json:"token,omitempty"`
	Name   string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	RoomId string `protobuf:"bytes,3,opt,name=room_id,json=roomId,proto3" json:"room_id,omitempty"`
}

func (x *Player) Reset() {
	*x = Player{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Player) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Player) ProtoMessage() {}

func (x *Player) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Player.ProtoReflect.Descriptor instead.
func (*Player) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{0}
}

func (x *Player) GetToken() string {
	if x != nil {
		return x.Token
	}
	return ""
}

func (x *Player) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Player) GetRoomId() string {
	if x != nil {
		return x.RoomId
	}
	return ""
}

type Room struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	RoomId          string        `protobuf:"bytes,1,opt,name=room_id,json=roomId,proto3" json:"room_id,omitempty"`
	RoomType        int32         `protobuf:"varint,2,opt,name=room_type,json=roomType,proto3" json:"room_type,omitempty"`
	Host            string        `protobuf:"bytes,3,opt,name=host,proto3" json:"host,omitempty"`
	Players         []string      `protobuf:"bytes,4,rep,name=players,proto3" json:"players,omitempty"`
	Started         bool          `protobuf:"varint,5,opt,name=started,proto3" json:"started,omitempty"`
	Spells          []*Spell      `protobuf:"bytes,6,rep,name=spells,proto3" json:"spells,omitempty"`
	StartMs         int64         `protobuf:"varint,7,opt,name=start_ms,json=startMs,proto3" json:"start_ms,omitempty"`
	GameTime        uint32        `protobuf:"varint,8,opt,name=game_time,json=gameTime,proto3" json:"game_time,omitempty"`      // 比赛时长，分
	Countdown       uint32        `protobuf:"varint,9,opt,name=countdown,proto3" json:"countdown,omitempty"`                    // 倒计时，秒
	Status          []SpellStatus `protobuf:"varint,10,rep,packed,name=status,proto3,enum=SpellStatus" json:"status,omitempty"` // 每个格子的状态
	Score           []uint32      `protobuf:"varint,11,rep,packed,name=score,proto3" json:"score,omitempty"`                    // 比分
	Locked          bool          `protobuf:"varint,12,opt,name=locked,proto3" json:"locked,omitempty"`                         // 连续多局就需要锁上
	NeedWin         uint32        `protobuf:"varint,13,opt,name=need_win,json=needWin,proto3" json:"need_win,omitempty"`        // 需要赢几局才算赢
	ChangeCardCount []uint32      `protobuf:"varint,14,rep,packed,name=change_card_count,json=changeCardCount,proto3" json:"change_card_count,omitempty"`
	LastGetTime     []int64       `protobuf:"varint,15,rep,packed,name=last_get_time,json=lastGetTime,proto3" json:"last_get_time,omitempty"` // 上次收卡时间
	TotalPauseMs    int64         `protobuf:"varint,16,opt,name=total_pause_ms,json=totalPauseMs,proto3" json:"total_pause_ms,omitempty"`     // 累计暂停时长，毫秒
	PauseBeginMs    int64         `protobuf:"varint,17,opt,name=pause_begin_ms,json=pauseBeginMs,proto3" json:"pause_begin_ms,omitempty"`     // 开始暂停时刻，毫秒，0表示没暂停
	LastWinner      int32         `protobuf:"varint,18,opt,name=last_winner,json=lastWinner,proto3" json:"last_winner,omitempty"`             // 上一场是谁赢，1或2
	BpData          *BpData       `protobuf:"bytes,19,opt,name=bp_data,json=bpData,proto3" json:"bp_data,omitempty"`
	LinkData        *LinkData     `protobuf:"bytes,20,opt,name=link_data,json=linkData,proto3" json:"link_data,omitempty"`
	Phase           int32         `protobuf:"varint,21,opt,name=phase,proto3" json:"phase,omitempty"`      // 纯客户端用，服务器只记录
	Watchers        []string      `protobuf:"bytes,22,rep,name=watchers,proto3" json:"watchers,omitempty"` // 观众
	Difficulty      int32         `protobuf:"varint,23,opt,name=difficulty,proto3" json:"difficulty,omitempty"`
	EnableTools     bool          `protobuf:"varint,24,opt,name=enable_tools,json=enableTools,proto3" json:"enable_tools,omitempty"`
}

func (x *Room) Reset() {
	*x = Room{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Room) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Room) ProtoMessage() {}

func (x *Room) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Room.ProtoReflect.Descriptor instead.
func (*Room) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{1}
}

func (x *Room) GetRoomId() string {
	if x != nil {
		return x.RoomId
	}
	return ""
}

func (x *Room) GetRoomType() int32 {
	if x != nil {
		return x.RoomType
	}
	return 0
}

func (x *Room) GetHost() string {
	if x != nil {
		return x.Host
	}
	return ""
}

func (x *Room) GetPlayers() []string {
	if x != nil {
		return x.Players
	}
	return nil
}

func (x *Room) GetStarted() bool {
	if x != nil {
		return x.Started
	}
	return false
}

func (x *Room) GetSpells() []*Spell {
	if x != nil {
		return x.Spells
	}
	return nil
}

func (x *Room) GetStartMs() int64 {
	if x != nil {
		return x.StartMs
	}
	return 0
}

func (x *Room) GetGameTime() uint32 {
	if x != nil {
		return x.GameTime
	}
	return 0
}

func (x *Room) GetCountdown() uint32 {
	if x != nil {
		return x.Countdown
	}
	return 0
}

func (x *Room) GetStatus() []SpellStatus {
	if x != nil {
		return x.Status
	}
	return nil
}

func (x *Room) GetScore() []uint32 {
	if x != nil {
		return x.Score
	}
	return nil
}

func (x *Room) GetLocked() bool {
	if x != nil {
		return x.Locked
	}
	return false
}

func (x *Room) GetNeedWin() uint32 {
	if x != nil {
		return x.NeedWin
	}
	return 0
}

func (x *Room) GetChangeCardCount() []uint32 {
	if x != nil {
		return x.ChangeCardCount
	}
	return nil
}

func (x *Room) GetLastGetTime() []int64 {
	if x != nil {
		return x.LastGetTime
	}
	return nil
}

func (x *Room) GetTotalPauseMs() int64 {
	if x != nil {
		return x.TotalPauseMs
	}
	return 0
}

func (x *Room) GetPauseBeginMs() int64 {
	if x != nil {
		return x.PauseBeginMs
	}
	return 0
}

func (x *Room) GetLastWinner() int32 {
	if x != nil {
		return x.LastWinner
	}
	return 0
}

func (x *Room) GetBpData() *BpData {
	if x != nil {
		return x.BpData
	}
	return nil
}

func (x *Room) GetLinkData() *LinkData {
	if x != nil {
		return x.LinkData
	}
	return nil
}

func (x *Room) GetPhase() int32 {
	if x != nil {
		return x.Phase
	}
	return 0
}

func (x *Room) GetWatchers() []string {
	if x != nil {
		return x.Watchers
	}
	return nil
}

func (x *Room) GetDifficulty() int32 {
	if x != nil {
		return x.Difficulty
	}
	return 0
}

func (x *Room) GetEnableTools() bool {
	if x != nil {
		return x.EnableTools
	}
	return false
}

type BpData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	WhoseTurn int32  `protobuf:"varint,1,opt,name=whoseTurn,proto3" json:"whoseTurn,omitempty"` // 用于bp赛
	BanPick   int32  `protobuf:"varint,2,opt,name=banPick,proto3" json:"banPick,omitempty"`     // 用于bp赛
	Round     uint32 `protobuf:"varint,3,opt,name=round,proto3" json:"round,omitempty"`         // 用于bp赛
	LessThan4 bool   `protobuf:"varint,4,opt,name=lessThan4,proto3" json:"lessThan4,omitempty"` // 用于bp赛
}

func (x *BpData) Reset() {
	*x = BpData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *BpData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*BpData) ProtoMessage() {}

func (x *BpData) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use BpData.ProtoReflect.Descriptor instead.
func (*BpData) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{2}
}

func (x *BpData) GetWhoseTurn() int32 {
	if x != nil {
		return x.WhoseTurn
	}
	return 0
}

func (x *BpData) GetBanPick() int32 {
	if x != nil {
		return x.BanPick
	}
	return 0
}

func (x *BpData) GetRound() uint32 {
	if x != nil {
		return x.Round
	}
	return 0
}

func (x *BpData) GetLessThan4() bool {
	if x != nil {
		return x.LessThan4
	}
	return false
}

type LinkData struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	LinkIdxA []uint32 `protobuf:"varint,1,rep,packed,name=link_idx_a,json=linkIdxA,proto3" json:"link_idx_a,omitempty"`
	LinkIdxB []uint32 `protobuf:"varint,2,rep,packed,name=link_idx_b,json=linkIdxB,proto3" json:"link_idx_b,omitempty"`
	StartMsA int64    `protobuf:"varint,3,opt,name=start_ms_a,json=startMsA,proto3" json:"start_ms_a,omitempty"`
	EndMsA   int64    `protobuf:"varint,4,opt,name=end_ms_a,json=endMsA,proto3" json:"end_ms_a,omitempty"`
	EventA   int32    `protobuf:"varint,5,opt,name=event_a,json=eventA,proto3" json:"event_a,omitempty"`
	StartMsB int64    `protobuf:"varint,6,opt,name=start_ms_b,json=startMsB,proto3" json:"start_ms_b,omitempty"`
	EndMsB   int64    `protobuf:"varint,7,opt,name=end_ms_b,json=endMsB,proto3" json:"end_ms_b,omitempty"`
	EventB   int32    `protobuf:"varint,8,opt,name=event_b,json=eventB,proto3" json:"event_b,omitempty"`
}

func (x *LinkData) Reset() {
	*x = LinkData{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LinkData) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LinkData) ProtoMessage() {}

func (x *LinkData) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LinkData.ProtoReflect.Descriptor instead.
func (*LinkData) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{3}
}

func (x *LinkData) GetLinkIdxA() []uint32 {
	if x != nil {
		return x.LinkIdxA
	}
	return nil
}

func (x *LinkData) GetLinkIdxB() []uint32 {
	if x != nil {
		return x.LinkIdxB
	}
	return nil
}

func (x *LinkData) GetStartMsA() int64 {
	if x != nil {
		return x.StartMsA
	}
	return 0
}

func (x *LinkData) GetEndMsA() int64 {
	if x != nil {
		return x.EndMsA
	}
	return 0
}

func (x *LinkData) GetEventA() int32 {
	if x != nil {
		return x.EventA
	}
	return 0
}

func (x *LinkData) GetStartMsB() int64 {
	if x != nil {
		return x.StartMsB
	}
	return 0
}

func (x *LinkData) GetEndMsB() int64 {
	if x != nil {
		return x.EndMsB
	}
	return 0
}

func (x *LinkData) GetEventB() int32 {
	if x != nil {
		return x.EventB
	}
	return 0
}

type Spell struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Game string `protobuf:"bytes,1,opt,name=game,proto3" json:"game,omitempty"`
	Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Rank string `protobuf:"bytes,3,opt,name=rank,proto3" json:"rank,omitempty"`
	Star int32  `protobuf:"varint,4,opt,name=star,proto3" json:"star,omitempty"`
	Desc string `protobuf:"bytes,5,opt,name=desc,proto3" json:"desc,omitempty"`
	Id   int32  `protobuf:"varint,6,opt,name=id,proto3" json:"id,omitempty"`
}

func (x *Spell) Reset() {
	*x = Spell{}
	if protoimpl.UnsafeEnabled {
		mi := &file_data_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Spell) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Spell) ProtoMessage() {}

func (x *Spell) ProtoReflect() protoreflect.Message {
	mi := &file_data_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Spell.ProtoReflect.Descriptor instead.
func (*Spell) Descriptor() ([]byte, []int) {
	return file_data_proto_rawDescGZIP(), []int{4}
}

func (x *Spell) GetGame() string {
	if x != nil {
		return x.Game
	}
	return ""
}

func (x *Spell) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *Spell) GetRank() string {
	if x != nil {
		return x.Rank
	}
	return ""
}

func (x *Spell) GetStar() int32 {
	if x != nil {
		return x.Star
	}
	return 0
}

func (x *Spell) GetDesc() string {
	if x != nil {
		return x.Desc
	}
	return ""
}

func (x *Spell) GetId() int32 {
	if x != nil {
		return x.Id
	}
	return 0
}

var File_data_proto protoreflect.FileDescriptor

var file_data_proto_rawDesc = []byte{
	0x0a, 0x0a, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0x4b, 0x0a, 0x06,
	0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x12, 0x14, 0x0a, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x18,
	0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x74, 0x6f, 0x6b, 0x65, 0x6e, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x17, 0x0a, 0x07, 0x72, 0x6f, 0x6f, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x03, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x06, 0x72, 0x6f, 0x6f, 0x6d, 0x49, 0x64, 0x22, 0xe8, 0x05, 0x0a, 0x04, 0x72, 0x6f,
	0x6f, 0x6d, 0x12, 0x17, 0x0a, 0x07, 0x72, 0x6f, 0x6f, 0x6d, 0x5f, 0x69, 0x64, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x6f, 0x6f, 0x6d, 0x49, 0x64, 0x12, 0x1b, 0x0a, 0x09, 0x72,
	0x6f, 0x6f, 0x6d, 0x5f, 0x74, 0x79, 0x70, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x08,
	0x72, 0x6f, 0x6f, 0x6d, 0x54, 0x79, 0x70, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x68, 0x6f, 0x73, 0x74,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x68, 0x6f, 0x73, 0x74, 0x12, 0x18, 0x0a, 0x07,
	0x70, 0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x18, 0x04, 0x20, 0x03, 0x28, 0x09, 0x52, 0x07, 0x70,
	0x6c, 0x61, 0x79, 0x65, 0x72, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65,
	0x64, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x65, 0x64,
	0x12, 0x1e, 0x0a, 0x06, 0x73, 0x70, 0x65, 0x6c, 0x6c, 0x73, 0x18, 0x06, 0x20, 0x03, 0x28, 0x0b,
	0x32, 0x06, 0x2e, 0x73, 0x70, 0x65, 0x6c, 0x6c, 0x52, 0x06, 0x73, 0x70, 0x65, 0x6c, 0x6c, 0x73,
	0x12, 0x19, 0x0a, 0x08, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x6d, 0x73, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x03, 0x52, 0x07, 0x73, 0x74, 0x61, 0x72, 0x74, 0x4d, 0x73, 0x12, 0x1b, 0x0a, 0x09, 0x67,
	0x61, 0x6d, 0x65, 0x5f, 0x74, 0x69, 0x6d, 0x65, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x08,
	0x67, 0x61, 0x6d, 0x65, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x1c, 0x0a, 0x09, 0x63, 0x6f, 0x75, 0x6e,
	0x74, 0x64, 0x6f, 0x77, 0x6e, 0x18, 0x09, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x09, 0x63, 0x6f, 0x75,
	0x6e, 0x74, 0x64, 0x6f, 0x77, 0x6e, 0x12, 0x25, 0x0a, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73,
	0x18, 0x0a, 0x20, 0x03, 0x28, 0x0e, 0x32, 0x0d, 0x2e, 0x73, 0x70, 0x65, 0x6c, 0x6c, 0x5f, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x52, 0x06, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x14, 0x0a,
	0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x0b, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x05, 0x73, 0x63,
	0x6f, 0x72, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x18, 0x0c, 0x20,
	0x01, 0x28, 0x08, 0x52, 0x06, 0x6c, 0x6f, 0x63, 0x6b, 0x65, 0x64, 0x12, 0x19, 0x0a, 0x08, 0x6e,
	0x65, 0x65, 0x64, 0x5f, 0x77, 0x69, 0x6e, 0x18, 0x0d, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x07, 0x6e,
	0x65, 0x65, 0x64, 0x57, 0x69, 0x6e, 0x12, 0x2a, 0x0a, 0x11, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65,
	0x5f, 0x63, 0x61, 0x72, 0x64, 0x5f, 0x63, 0x6f, 0x75, 0x6e, 0x74, 0x18, 0x0e, 0x20, 0x03, 0x28,
	0x0d, 0x52, 0x0f, 0x63, 0x68, 0x61, 0x6e, 0x67, 0x65, 0x43, 0x61, 0x72, 0x64, 0x43, 0x6f, 0x75,
	0x6e, 0x74, 0x12, 0x22, 0x0a, 0x0d, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x67, 0x65, 0x74, 0x5f, 0x74,
	0x69, 0x6d, 0x65, 0x18, 0x0f, 0x20, 0x03, 0x28, 0x03, 0x52, 0x0b, 0x6c, 0x61, 0x73, 0x74, 0x47,
	0x65, 0x74, 0x54, 0x69, 0x6d, 0x65, 0x12, 0x24, 0x0a, 0x0e, 0x74, 0x6f, 0x74, 0x61, 0x6c, 0x5f,
	0x70, 0x61, 0x75, 0x73, 0x65, 0x5f, 0x6d, 0x73, 0x18, 0x10, 0x20, 0x01, 0x28, 0x03, 0x52, 0x0c,
	0x74, 0x6f, 0x74, 0x61, 0x6c, 0x50, 0x61, 0x75, 0x73, 0x65, 0x4d, 0x73, 0x12, 0x24, 0x0a, 0x0e,
	0x70, 0x61, 0x75, 0x73, 0x65, 0x5f, 0x62, 0x65, 0x67, 0x69, 0x6e, 0x5f, 0x6d, 0x73, 0x18, 0x11,
	0x20, 0x01, 0x28, 0x03, 0x52, 0x0c, 0x70, 0x61, 0x75, 0x73, 0x65, 0x42, 0x65, 0x67, 0x69, 0x6e,
	0x4d, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x6c, 0x61, 0x73, 0x74, 0x5f, 0x77, 0x69, 0x6e, 0x6e, 0x65,
	0x72, 0x18, 0x12, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x6c, 0x61, 0x73, 0x74, 0x57, 0x69, 0x6e,
	0x6e, 0x65, 0x72, 0x12, 0x21, 0x0a, 0x07, 0x62, 0x70, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x13,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x08, 0x2e, 0x62, 0x70, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x52, 0x06,
	0x62, 0x70, 0x44, 0x61, 0x74, 0x61, 0x12, 0x27, 0x0a, 0x09, 0x6c, 0x69, 0x6e, 0x6b, 0x5f, 0x64,
	0x61, 0x74, 0x61, 0x18, 0x14, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x0a, 0x2e, 0x6c, 0x69, 0x6e, 0x6b,
	0x5f, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6c, 0x69, 0x6e, 0x6b, 0x44, 0x61, 0x74, 0x61, 0x12,
	0x14, 0x0a, 0x05, 0x70, 0x68, 0x61, 0x73, 0x65, 0x18, 0x15, 0x20, 0x01, 0x28, 0x05, 0x52, 0x05,
	0x70, 0x68, 0x61, 0x73, 0x65, 0x12, 0x1a, 0x0a, 0x08, 0x77, 0x61, 0x74, 0x63, 0x68, 0x65, 0x72,
	0x73, 0x18, 0x16, 0x20, 0x03, 0x28, 0x09, 0x52, 0x08, 0x77, 0x61, 0x74, 0x63, 0x68, 0x65, 0x72,
	0x73, 0x12, 0x1e, 0x0a, 0x0a, 0x64, 0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74, 0x79, 0x18,
	0x17, 0x20, 0x01, 0x28, 0x05, 0x52, 0x0a, 0x64, 0x69, 0x66, 0x66, 0x69, 0x63, 0x75, 0x6c, 0x74,
	0x79, 0x12, 0x21, 0x0a, 0x0c, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x5f, 0x74, 0x6f, 0x6f, 0x6c,
	0x73, 0x18, 0x18, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x65, 0x6e, 0x61, 0x62, 0x6c, 0x65, 0x54,
	0x6f, 0x6f, 0x6c, 0x73, 0x22, 0x75, 0x0a, 0x07, 0x62, 0x70, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x12,
	0x1c, 0x0a, 0x09, 0x77, 0x68, 0x6f, 0x73, 0x65, 0x54, 0x75, 0x72, 0x6e, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x05, 0x52, 0x09, 0x77, 0x68, 0x6f, 0x73, 0x65, 0x54, 0x75, 0x72, 0x6e, 0x12, 0x18, 0x0a,
	0x07, 0x62, 0x61, 0x6e, 0x50, 0x69, 0x63, 0x6b, 0x18, 0x02, 0x20, 0x01, 0x28, 0x05, 0x52, 0x07,
	0x62, 0x61, 0x6e, 0x50, 0x69, 0x63, 0x6b, 0x12, 0x14, 0x0a, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x72, 0x6f, 0x75, 0x6e, 0x64, 0x12, 0x1c, 0x0a,
	0x09, 0x6c, 0x65, 0x73, 0x73, 0x54, 0x68, 0x61, 0x6e, 0x34, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x09, 0x6c, 0x65, 0x73, 0x73, 0x54, 0x68, 0x61, 0x6e, 0x34, 0x22, 0xe9, 0x01, 0x0a, 0x09,
	0x6c, 0x69, 0x6e, 0x6b, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x12, 0x1c, 0x0a, 0x0a, 0x6c, 0x69, 0x6e,
	0x6b, 0x5f, 0x69, 0x64, 0x78, 0x5f, 0x61, 0x18, 0x01, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x08, 0x6c,
	0x69, 0x6e, 0x6b, 0x49, 0x64, 0x78, 0x41, 0x12, 0x1c, 0x0a, 0x0a, 0x6c, 0x69, 0x6e, 0x6b, 0x5f,
	0x69, 0x64, 0x78, 0x5f, 0x62, 0x18, 0x02, 0x20, 0x03, 0x28, 0x0d, 0x52, 0x08, 0x6c, 0x69, 0x6e,
	0x6b, 0x49, 0x64, 0x78, 0x42, 0x12, 0x1c, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f, 0x6d,
	0x73, 0x5f, 0x61, 0x18, 0x03, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x73, 0x74, 0x61, 0x72, 0x74,
	0x4d, 0x73, 0x41, 0x12, 0x18, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x5f, 0x6d, 0x73, 0x5f, 0x61, 0x18,
	0x04, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x65, 0x6e, 0x64, 0x4d, 0x73, 0x41, 0x12, 0x17, 0x0a,
	0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x61, 0x18, 0x05, 0x20, 0x01, 0x28, 0x05, 0x52, 0x06,
	0x65, 0x76, 0x65, 0x6e, 0x74, 0x41, 0x12, 0x1c, 0x0a, 0x0a, 0x73, 0x74, 0x61, 0x72, 0x74, 0x5f,
	0x6d, 0x73, 0x5f, 0x62, 0x18, 0x06, 0x20, 0x01, 0x28, 0x03, 0x52, 0x08, 0x73, 0x74, 0x61, 0x72,
	0x74, 0x4d, 0x73, 0x42, 0x12, 0x18, 0x0a, 0x08, 0x65, 0x6e, 0x64, 0x5f, 0x6d, 0x73, 0x5f, 0x62,
	0x18, 0x07, 0x20, 0x01, 0x28, 0x03, 0x52, 0x06, 0x65, 0x6e, 0x64, 0x4d, 0x73, 0x42, 0x12, 0x17,
	0x0a, 0x07, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x5f, 0x62, 0x18, 0x08, 0x20, 0x01, 0x28, 0x05, 0x52,
	0x06, 0x65, 0x76, 0x65, 0x6e, 0x74, 0x42, 0x22, 0x7b, 0x0a, 0x05, 0x73, 0x70, 0x65, 0x6c, 0x6c,
	0x12, 0x12, 0x0a, 0x04, 0x67, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x67, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x12, 0x0a, 0x04, 0x72, 0x61, 0x6e, 0x6b,
	0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x72, 0x61, 0x6e, 0x6b, 0x12, 0x12, 0x0a, 0x04,
	0x73, 0x74, 0x61, 0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x05, 0x52, 0x04, 0x73, 0x74, 0x61, 0x72,
	0x12, 0x12, 0x0a, 0x04, 0x64, 0x65, 0x73, 0x63, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04,
	0x64, 0x65, 0x73, 0x63, 0x12, 0x0e, 0x0a, 0x02, 0x69, 0x64, 0x18, 0x06, 0x20, 0x01, 0x28, 0x05,
	0x52, 0x02, 0x69, 0x64, 0x2a, 0x8c, 0x01, 0x0a, 0x0c, 0x73, 0x70, 0x65, 0x6c, 0x6c, 0x5f, 0x73,
	0x74, 0x61, 0x74, 0x75, 0x73, 0x12, 0x08, 0x0a, 0x04, 0x6e, 0x6f, 0x6e, 0x65, 0x10, 0x00, 0x12,
	0x13, 0x0a, 0x06, 0x62, 0x61, 0x6e, 0x6e, 0x65, 0x64, 0x10, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
	0xff, 0xff, 0xff, 0x01, 0x12, 0x0f, 0x0a, 0x0b, 0x6c, 0x65, 0x66, 0x74, 0x5f, 0x73, 0x65, 0x6c,
	0x65, 0x63, 0x74, 0x10, 0x01, 0x12, 0x10, 0x0a, 0x0c, 0x72, 0x69, 0x67, 0x68, 0x74, 0x5f, 0x73,
	0x65, 0x6c, 0x65, 0x63, 0x74, 0x10, 0x03, 0x12, 0x0c, 0x0a, 0x08, 0x6c, 0x65, 0x66, 0x74, 0x5f,
	0x67, 0x65, 0x74, 0x10, 0x05, 0x12, 0x0d, 0x0a, 0x09, 0x72, 0x69, 0x67, 0x68, 0x74, 0x5f, 0x67,
	0x65, 0x74, 0x10, 0x07, 0x12, 0x0f, 0x0a, 0x0b, 0x62, 0x6f, 0x74, 0x68, 0x5f, 0x73, 0x65, 0x6c,
	0x65, 0x63, 0x74, 0x10, 0x02, 0x12, 0x0c, 0x0a, 0x08, 0x62, 0x6f, 0x74, 0x68, 0x5f, 0x67, 0x65,
	0x74, 0x10, 0x06, 0x42, 0x09, 0x5a, 0x07, 0x2e, 0x2f, 0x3b, 0x6d, 0x61, 0x69, 0x6e, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_data_proto_rawDescOnce sync.Once
	file_data_proto_rawDescData = file_data_proto_rawDesc
)

func file_data_proto_rawDescGZIP() []byte {
	file_data_proto_rawDescOnce.Do(func() {
		file_data_proto_rawDescData = protoimpl.X.CompressGZIP(file_data_proto_rawDescData)
	})
	return file_data_proto_rawDescData
}

var file_data_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_data_proto_msgTypes = make([]protoimpl.MessageInfo, 5)
var file_data_proto_goTypes = []interface{}{
	(SpellStatus)(0), // 0: spell_status
	(*Player)(nil),   // 1: player
	(*Room)(nil),     // 2: room
	(*BpData)(nil),   // 3: bp_data
	(*LinkData)(nil), // 4: link_data
	(*Spell)(nil),    // 5: spell
}
var file_data_proto_depIdxs = []int32{
	5, // 0: room.spells:type_name -> spell
	0, // 1: room.status:type_name -> spell_status
	3, // 2: room.bp_data:type_name -> bp_data
	4, // 3: room.link_data:type_name -> link_data
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_data_proto_init() }
func file_data_proto_init() {
	if File_data_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_data_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Player); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Room); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*BpData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LinkData); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
		file_data_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Spell); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_data_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   5,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_data_proto_goTypes,
		DependencyIndexes: file_data_proto_depIdxs,
		EnumInfos:         file_data_proto_enumTypes,
		MessageInfos:      file_data_proto_msgTypes,
	}.Build()
	File_data_proto = out.File
	file_data_proto_rawDesc = nil
	file_data_proto_goTypes = nil
	file_data_proto_depIdxs = nil
}
