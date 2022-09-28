package main

import (
	"flag"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"
)

var port = flag.Int("p", 9999, "listening port")
var address = flag.String("a", "/ws", "ws address endpoint")

var upGrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	Error: func(w http.ResponseWriter, r *http.Request, status int, reason error) {
		log.WithError(reason).Error("status: ", status)
	},
}

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
	defer initDB()()
	go initWS()
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
}

var db *badger.DB

func initDB() func() {
	var err error
	db, err = badger.Open(badger.DefaultOptions("db"))
	if err != nil {
		log.Error(err)
		panic(err)
	}
	go func() {
		ticker := time.NewTicker(time.Hour)
		for range ticker.C {
		again:
			err := db.RunValueLogGC(0.5)
			if err == nil {
				goto again
			}
		}
	}()
	return func() {
		err := db.Close()
		if err != nil {
			log.WithError(err).Error("close db failed")
		}
	}
}

func initWS() {
	http.HandleFunc(*address, func(w http.ResponseWriter, r *http.Request) {
		ws, err := upGrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		log.Info("连接成功：", r.RemoteAddr)
		(&PlayerConn{}).OnConnect(ws)
	})
	fmt.Printf("请访问：ws://127.0.0.1:%d%s\n", *port, *address)
	_ = http.ListenAndServe(":"+strconv.Itoa(*port), nil)
}
