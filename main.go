package main

import (
	"flag"
	"fmt"
	"github.com/Touhou-Freshman-Camp/bingo-server/myws"
	"github.com/davyxu/cellnet"
	"github.com/davyxu/cellnet/msglog"
	"github.com/davyxu/cellnet/peer"
	_ "github.com/davyxu/cellnet/peer/gorillaws"
	"github.com/davyxu/cellnet/proc"
	"github.com/dgraph-io/badger/v3"
	"golang.org/x/time/rate"
	"os"
	"os/signal"
	"time"
)

var port = flag.Int("p", 9999, "listening port")
var address = flag.String("a", "/ws", "ws address endpoint")
var tcpMute = flag.Bool("m", false, "mute tcp debug log")

var eventQueue = cellnet.NewEventQueue()

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
	if *tcpMute {
		msglog.SetCurrMsgLogMode(msglog.MsgLogMode_Mute)
	}
	// 创建一个tcp的侦听器，名称为server，所有连接将事件投递到queue队列,单线程的处理
	p := peer.NewGenericPeer("gorillaws.Acceptor", "server", fmt.Sprintf("ws://0.0.0.0:%d%s", *port, *address), eventQueue)
	idConnMap := make(map[int64]*PlayerConn)
	proc.BindProcessorHandler(p, "myws", func(ev cellnet.Event) {
		id := ev.Session().ID()
		switch pb := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			log.Info("session connected: ", id)
			playerConn := &PlayerConn{Session: ev.Session(), Limit: rate.NewLimiter(5, 5)}
			idConnMap[id] = playerConn
			playerConn.SetHeartTimer()
		case *myws.Message:
			if player, ok := idConnMap[id]; ok {
				player.Handle(pb.MsgName, pb.Data)
			}
		case *cellnet.SessionClosed:
			if player, ok := idConnMap[id]; ok {
				player.OnDisconnect()
				delete(idConnMap, id)
			}
		}
	})
	p.Start()
	eventQueue.EnableCapturePanic(true)
	eventQueue.StartLoop()
	fmt.Printf("请访问：ws://127.0.0.1:%d%s\n", *port, *address)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	eventQueue.StopLoop()
	eventQueue.Wait()
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
