package myws

import (
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/util"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

type WSMessageTransmitter struct {
}

func (WSMessageTransmitter) OnRecvMessage(ses cellnet.Session) (msg interface{}, err error) {

	conn, ok := ses.Raw().(*websocket.Conn)

	// 转换错误，或者连接已经关闭时退出
	if !ok || conn == nil {
		return nil, nil
	}

	var messageType int
	var raw []byte
	messageType, raw, err = conn.ReadMessage()

	if err != nil {
		return
	}

	switch messageType {
	case websocket.TextMessage:
		if !gjson.ValidBytes(raw) {
			return nil, errors.New("json unmarshal failed")
		}

		protoName := gjson.GetBytes(raw, "name").String()
		msg, _, err = codec.DecodeMessage(int(util.StringHash(protoName)), []byte(gjson.GetBytes(raw, "data").Raw))
		if err != nil {
			return nil, err
		}
		msg = &Message{MsgName: protoName, Data: msg}
	}

	return
}

func (WSMessageTransmitter) OnSendMessage(ses cellnet.Session, msg interface{}) error {

	conn, ok := ses.Raw().(*websocket.Conn)

	// 转换错误，或者连接已经关闭时退出
	if !ok || conn == nil {
		return nil
	}

	var pkt []byte

	switch m := msg.(type) {
	case *cellnet.RawPacket: // 发裸包
		pkt = m.MsgData
	case *Message: // 发普通编码包，需要将用户数据转换为字节数组
		buf, meta, err := codec.EncodeMessage(m.Data, nil)
		if err != nil {
			return err
		}
		pkt = make([]byte, 0, 50)
		pkt = append(pkt, []byte(`{"name":"`)...)
		if len(m.MsgName) == 0 {
			pkt = append(pkt, []byte(meta.GetContextAsString("name", ""))...)
		} else {
			pkt = append(pkt, []byte(m.MsgName)...)
		}
		pkt = append(pkt, '"')
		if len(m.Reply) > 0 {
			pkt = append(pkt, []byte(`,"reply":"`)...)
			pkt = append(pkt, []byte(m.Reply)...)
			pkt = append(pkt, '"')
		}
		if len(m.Trigger) > 0 {
			pkt = append(pkt, []byte(`,"trigger":"`)...)
			pkt = append(pkt, []byte(m.Trigger)...)
			pkt = append(pkt, '"')
		}
		if len(buf) > 0 {
			pkt = append(pkt, []byte(`,"data":`)...)
			pkt = append(pkt, buf...)
		}
		pkt = append(pkt, '}')
	default:
		return errors.New("can not send this message type")
	}

	return conn.WriteMessage(websocket.TextMessage, pkt)
}
