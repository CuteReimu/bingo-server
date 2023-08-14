package main

import (
	"flag"
	"fmt"
	"github.com/CuteReimu/bingo-server/myws"
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

const ( // 用于context的key
	lastHeartTime     = 1
	playerConnToken   = 2
	playerConnLimiter = 3
)

var port = flag.Int("p", 9999, "listening port")
var address = flag.String("a", "/ws", "ws address endpoint")
var tcpMute = flag.Bool("m", false, "mute tcp debug log")

func main() {
	flag.Parse()
	if !flag.Parsed() {
		flag.Usage()
		os.Exit(2)
	}
	if (*address)[0] != '/' {
		flag.Usage()
		log.Error("ws address endpoint的格式不对")
		os.Exit(1)
	}
	defer initDB()()
	if *tcpMute {
		msglog.SetCurrMsgLogMode(msglog.MsgLogMode_Mute)
	}
	(&bingoServer{timeout: time.Minute}).start()
}

type bingoServer struct {
	timeout      time.Duration
	peer         cellnet.WSAcceptor
	tokenConnMap map[string]cellnet.Session
}

func (s *bingoServer) start() {
	s.tokenConnMap = make(map[string]cellnet.Session)
	eventQueue := cellnet.NewEventQueue()
	// 创建一个tcp的侦听器，名称为server，所有连接将事件投递到queue队列,单线程的处理
	s.peer = peer.NewGenericPeer("gorillaws.Acceptor", "server", fmt.Sprintf("ws://0.0.0.0:%d%s", *port, *address), eventQueue).(cellnet.WSAcceptor)
	proc.BindProcessorHandler(s.peer, "myws", func(ev cellnet.Event) {
		switch pb := ev.Message().(type) {
		case *cellnet.SessionAccepted:
			log.Info("session connected: ", ev.Session().ID())
			ev.Session().(cellnet.ContextSet).SetContext(playerConnLimiter, rate.NewLimiter(5, 5))
			ev.Session().(cellnet.ContextSet).SetContext(lastHeartTime, time.Now())
		case *myws.Message:
			limiter, ok := ev.Session().(cellnet.ContextSet).GetContext(playerConnLimiter)
			if !ok || !limiter.(*rate.Limiter).Allow() {
				ev.Session().Close()
				break
			}
			if handler, ok := pb.Data.(MessageHandler); ok {
				ev.Session().(cellnet.ContextSet).SetContext(lastHeartTime, time.Now())
				var token string
				if _, ok := handler.(*LoginCs); !ok {
					if !ev.Session().(cellnet.ContextSet).FetchContext(playerConnToken, &token) {
						ev.Session().Send(&myws.Message{Reply: pb.MsgName, Data: &ErrorSc{Code: -1, Msg: "You haven't login."}})
						break
					}
				}
				if err := handler.Handle(s, ev.Session(), token, pb.MsgName); err != nil {
					log.Error(fmt.Sprintf("handle failed: %s, error: %+v", pb.MsgName, err))
					ev.Session().Send(&myws.Message{Reply: pb.MsgName, Data: &ErrorSc{Code: 500, Msg: err.Error()}})
				}
			} else {
				log.Warn("can not find handler: ", pb.MsgName)
				ev.Session().Send(&myws.Message{Reply: pb.MsgName, Data: &ErrorSc{Code: 404, Msg: "404 not found"}})
			}
		case *cellnet.SessionClosed:
			var token string
			if ev.Session().(cellnet.ContextSet).FetchContext(playerConnToken, &token) {
				delete(s.tokenConnMap, token)
				_ = (&LeaveRoomCs{}).Handle(s, ev.Session(), token, "")
			}
		}
	})
	s.peer.Start()
	eventQueue.EnableCapturePanic(true)
	eventQueue.StartLoop()
	go s.startRemoveTimeoutTimer()
	fmt.Printf("请访问：ws://127.0.0.1:%d%s\n", *port, *address)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	<-ch
	eventQueue.StopLoop()
	eventQueue.Wait()
}

func (s *bingoServer) startRemoveTimeoutTimer() {
	ch := time.Tick(15 * time.Second)
	for {
		<-ch
		s.peer.Queue().Post(s.removeTimeoutRoom)
	}
}

func (s *bingoServer) removeTimeoutRoom() {
	if s.peer.SessionCount() == 0 {
		return
	}
	now := time.Now()
	s.peer.VisitSession(func(session cellnet.Session) bool {
		lt, _ := session.(cellnet.ContextSet).GetContext(lastHeartTime)
		if lt.(time.Time).Add(s.timeout).Before(now) {
			session.Close()
		}
		return true
	})
}

var db *badger.DB

func initDB() func() {
	var err error
	db, err = badger.Open(badger.DefaultOptions("db"))
	if err != nil {
		log.Error(fmt.Sprintf("%+v", err))
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
			log.Error(fmt.Sprintf("close db failed: %+v", err))
		}
	}
}
