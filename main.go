package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"os"
	"strconv"
)

var port = flag.Int("p", 9999, "listening port")
var address = flag.String("a", "/ws", "ws address endpoint")

var upGrader = websocket.Upgrader{}

func main() {
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(2)
	}
	if (*address)[0] != '/' {
		flag.Usage()
		log.Fatalln("ws address endpoint的格式不对")
	}
	http.HandleFunc(*address, func(w http.ResponseWriter, r *http.Request) {
		ws, err := upGrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		fmt.Println("连接成功：", r.RemoteAddr)
		defer func(ws *websocket.Conn) { _ = ws.Close() }(ws)
		for {
			mt, message, err := ws.ReadMessage()
			if err != nil {
				log.Println(err)
				break
			}
			fmt.Printf("mt: %d, message: %s\n", mt, string(message))
		}
	})
	fmt.Printf("请访问：ws://127.0.0.1:%d%s\n", *port, *address)
	_ = http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
