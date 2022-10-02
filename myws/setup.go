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

func (c *Message) Name() string {
	return "json"
}

func (c *Message) MimeType() string {
	return "application/json"
}

func (c *Message) Encode(msgObj interface{}, _ cellnet.ContextSet) (data interface{}, err error) {
	return json.Marshal(msgObj)
}

func (c *Message) Decode(data interface{}, msgObj interface{}) error {
	return json.Unmarshal(data.([]byte), msgObj.(*Message))
}

func (c *Message) String() string {
	buf, _ := json.Marshal(c)
	return string(buf)
}

func init() {

	proc.RegisterProcessor("myws", func(bundle proc.ProcessorBundle, userCallback cellnet.EventCallback) {

		bundle.SetTransmitter(new(WSMessageTransmitter))
		bundle.SetHooker(new(MsgHooker))
		bundle.SetCallback(proc.NewQueuedEventCallback(userCallback))

	})
	cellnet.RegisterMessageMeta(&cellnet.MessageMeta{
		Codec: new(Message),
		Type:  reflect.TypeOf((*Message)(nil)).Elem(),
		ID:    1,
	})
}
