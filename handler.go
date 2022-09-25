package main

import "time"

var handlers = map[string]func(player *Player, result map[string]interface{}){
	"heart_cs": handleHeart,
}

func handleHeart(player *Player, _ map[string]interface{}) {
	if player.heartTimer != nil {
		player.heartTimer.Stop()
	}
	player.heartTimer = time.AfterFunc(time.Minute, func() { _ = player.conn.Close() })
	player.Send(&Message{
		Name:  "heart_sc",
		Reply: "heart_cs",
		Data: map[string]interface{}{
			"time": time.Now().UnixMilli(),
		},
	})
}
