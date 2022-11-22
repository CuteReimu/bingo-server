package myws

import (
	"encoding/json"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/proc"
	"reflect"
)

type Message struct {
	MsgName string                 `json:"name"`
	Reply   string                 `json:"reply,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

func (m *Message) String() string {
	buf, _ := json.Marshal(m)
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

	return json.Unmarshal(data.([]byte), msgObj)
}

func init() {

	proc.RegisterProcessor("myws", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {

		bundle.SetTransmitter(new(WSMessageTransmitter))
		bundle.SetHooker(new(MsgHooker))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))

	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: new(jsonCodec),
		Type:  reflect.TypeOf((*Message)(nil)).Elem(),
		ID:    1,
	})
}
