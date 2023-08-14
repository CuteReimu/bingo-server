package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func init() {
	buf, err := os.ReadFile("application.json")
	if err == nil {
		if err = json.Unmarshal(buf, &miraiPusher); err != nil {
			panic(err)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if buf, err = json.Marshal(miraiPusher); err == nil {
			_ = os.WriteFile("application.json", buf, 0666)
		}
	} else {
		panic(err)
	}
}

type MiraiPusher struct {
	EnablePush     bool    `json:"enablePush"`
	SelfRoomAddr   string  `json:"selfRoomAddr"`
	MiraiHttpUrl   string  `json:"miraiHttpUrl"`
	MiraiVerifyKey string  `json:"miraiVerifyKey"`
	RobotQQ        int64   `json:"robotQQ"`
	PushQQGroups   []int64 `json:"pushQQGroups"`
	PushInterval   int64   `json:"pushInterval"`
	lastPushTime   int64
}

func (p *MiraiPusher) Push(room *Room) {
	if !p.EnablePush {
		return
	}
	now := time.Now().UnixMilli()
	if now-p.lastPushTime < p.PushInterval*60000 {
		return
	}
	p.lastPushTime = now
	text := fmt.Sprintf("Bingo %s正在激烈进行，快来围观吧：\n%s/%s", room.Type().Name(), p.SelfRoomAddr, room.RoomId)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Errorln("panic recover: ", r)
			}
		}()
		session, err := p.verify()
		if err != nil {
			log.Errorf("%+v", err)
			return
		}
		if err = p.bind(session); err != nil {
			log.Errorf("%+v", err)
			return
		}
		defer func() {
			if err = p.release(session); err != nil {
				log.Errorf("%+v", err)
				return
			}
		}()
		for _, group := range p.PushQQGroups {
			if err = p.sendGroupMessage(session, group, text); err != nil {
				log.Errorf("%+v", err)
			}
		}
	}()
}

func (p *MiraiPusher) verify() (string, error) {
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Post(
		p.MiraiHttpUrl+"/verify",
		"application/json; charset=utf-8",
		strings.NewReader(fmt.Sprintf(`{"verifyKey":"%s"}`, p.MiraiVerifyKey)),
	)
	if err != nil {
		return "", errors.Wrap(err, "verify failed")
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.Wrap(err, "verify failed")
	}
	if !gjson.ValidBytes(buf) {
		return "", errors.Errorf("invalid json: %s", string(buf))
	}
	code := gjson.GetBytes(buf, "code")
	if !code.Exists() || code.Int() != 0 {
		return "", errors.Errorf("bind failed: %d", code.Int())
	}
	return gjson.GetBytes(buf, "session").String(), nil
}

func (p *MiraiPusher) bind(sessionKey string) error {
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Post(
		p.MiraiHttpUrl+"/bind",
		"application/json; charset=utf-8",
		strings.NewReader(fmt.Sprintf(`{"sessionKey":"%s","qq":%d}`, sessionKey, p.RobotQQ)),
	)
	if err != nil {
		return errors.Wrap(err, "verify failed")
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "verify failed")
	}
	if !gjson.ValidBytes(buf) {
		return errors.Errorf("invalid json: %s", string(buf))
	}
	code := gjson.GetBytes(buf, "code")
	if !code.Exists() || code.Int() != 0 {
		return errors.Errorf("bind failed: %d", code.Int())
	}
	return nil
}

func (p *MiraiPusher) sendGroupMessage(sessionKey string, groupId int64, message string) error {
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Post(
		p.MiraiHttpUrl+"/sendGroupMessage",
		"application/json; charset=utf-8",
		strings.NewReader(fmt.Sprintf(`{"sessionKey":"%s","target":%d,"messageChain":[{"type":"Plain","text":"%s"}]}`, sessionKey, groupId, message)),
	)
	if err != nil {
		return errors.Wrap(err, "verify failed")
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "verify failed")
	}
	if !gjson.ValidBytes(buf) {
		return errors.Errorf("invalid json: %s", string(buf))
	}
	code := gjson.GetBytes(buf, "code")
	if !code.Exists() || code.Int() != 0 {
		return errors.Errorf("sendGroupMessage failed: %d", code.Int())
	}
	return nil
}

func (p *MiraiPusher) release(sessionKey string) error {
	resp, err := (&http.Client{Timeout: 20 * time.Second}).Post(
		p.MiraiHttpUrl+"/release",
		"application/json; charset=utf-8",
		strings.NewReader(fmt.Sprintf(`{"sessionKey":"%s","qq":%d}`, sessionKey, p.RobotQQ)),
	)
	if err != nil {
		return errors.Wrap(err, "verify failed")
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Wrap(err, "verify failed")
	}
	if !gjson.ValidBytes(buf) {
		return errors.Errorf("invalid json: %s", string(buf))
	}
	code := gjson.GetBytes(buf, "code")
	if !code.Exists() || code.Int() != 0 {
		return errors.Errorf("release failed: %d", code.Int())
	}
	return nil
}

var miraiPusher = MiraiPusher{
	EnablePush:     false,
	SelfRoomAddr:   "http://127.0.0.1:9961/room",
	MiraiHttpUrl:   "http://127.0.0.1:8080",
	MiraiVerifyKey: "",
	RobotQQ:        12345678,
	PushQQGroups:   nil,
	PushInterval:   10,
}
