package myws

import (
	"encoding/json"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/codec"
	"github.com/davyxu/cellnet/proc"
	"reflect"
)

type Message struct {
	MsgName string      `json:"name"`
	Reply   string      `json:"reply,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

func (m *Message) Message() interface{} {
	return m.Data
}

func (m *Message) String() string {
	buf, _ := json.Marshal(m.Data)
	return string(buf)
}

type jsonCodec struct {
}

func (c *jsonCodec) Name() string {
	return "json"
}

func (c *jsonCodec) MimeType() string {
	return "application/json"
}

func (c *jsonCodec) Encode(msgObj interface{}, _ cellnet.ContextSet) (data interface{}, err error) {
	return json.Marshal(msgObj)
}

func (c *jsonCodec) Decode(data interface{}, msgObj interface{}) error {
	buf := data.([]byte)
	if len(buf) == 0 {
		buf = []byte{'{', '}'}
	}
	return json.Unmarshal(buf, msgObj)
}

func init() {
	c := new(jsonCodec)
	codec.RegisterCodec(c)
	proc.RegisterProcessor("myws", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {
		bundle.SetTransmitter(new(WSMessageTransmitter))
		bundle.SetHooker(new(MsgHooker))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))
	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: c,
		Type:  reflect.TypeOf((*Message)(nil)).Elem(),
		ID:    1,
	})
}
