package myws

import (
	"encoding/json"
	"github.com/davyxu/cellnet"
	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
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
		var message *Message
		err = json.Unmarshal(raw, &message)
		msg = message
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
		var err error
		pkt, err = json.Marshal(msg)
		if err != nil {
			return err
		}
	default:
		return errors.New("can not send this message type")
	}

	return conn.WriteMessage(websocket.TextMessage, pkt)
}
