package manager

import (
	"chatExtensionServer/internal/concurrency/concurrentqueue"
	"chatExtensionServer/internal/ratelimiting"
	"chatExtensionServer/internal/types"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func reader(ws *websocket.Conn, jobs *concurrentqueue.SafeQueue, rateLimiter *ratelimiting.RateLimiter) {
	for {
		var m types.Message
		err := ws.ReadJSON(&m)
		if err != nil {
			log.Println(err)
			return
		}

		if rateLimiter.Add(m.UserID) {
			log.Println("Tried to send messages too fast, timing out user with ID: " + fmt.Sprint(m.UserID) + "...")
			return
		}

		jobs.Push(&m)
	}
}

func process(jobs *concurrentqueue.SafeQueue, mgr *PubSubMgr, rateLimiter *ratelimiting.RateLimiter) {
	for true {
		var item *types.Message = jobs.Pop()
		err := mgr.BroadcastMessage(item)
		rateLimiter.Resolve(item.UserID)
		if err != nil {
			println("Error broadcasting message!")
			log.Fatal(err)
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request,
	jobs *concurrentqueue.SafeQueue, mgr *PubSubMgr, rateLimiter *ratelimiting.RateLimiter) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		println("Error upgrading socket!")
		log.Fatal(err)

	}

	var t types.TransactionToken
	err = ws.ReadJSON(&t)
	if err != nil {
		println("Error decoding json body!")
		log.Fatal(err)
	}

	t.UserID = mgr.Connect(t.UserName, t.VideoID, ws)
	defer mgr.Disconnect(t.UserID)

	err = mgr.SendTokenToUser(&t)

	if err != nil {
		println("Error sending back the token!")
		log.Fatal(err)
	}

	reader(ws, jobs, rateLimiter)
}

func setupTestServer(sThreads int, rateLimit uint16) (*httptest.Server, string) {
	var jobs concurrentqueue.SafeQueue
	var rateLimiter ratelimiting.RateLimiter

	jobs.Init()
	rateLimiter.Init(rateLimit)
	mgr := CreateNewPSMgr()

	for i := 0; i < sThreads; i++ {
		go process(&jobs, &mgr, &rateLimiter)
	}

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { wsEndpoint(w, r, &jobs, &mgr, &rateLimiter) }))
	u := "ws" + strings.TrimPrefix(s.URL, "http")
	return s, u
}

func setupTestUser(username string, videoid string, serverURI string) (*websocket.Conn, types.UIDType, error) {
	ws, _, err := websocket.DefaultDialer.Dial(serverURI, nil)

	if err != nil {
		return nil, 0, err
	}
	defer ws.Close()

	t := types.TransactionToken{UserName: username, VideoID: videoid, UserID: 0}
	err = ws.WriteJSON(&t)
	if err != nil {
		return nil, 0, err
	}
	err = ws.ReadJSON(&t)

	if err != nil {
		return nil, 0, err
	}

	return ws, t.UserID, nil
}

func TestConnectDisconnect(t *testing.T) {
	s, uri := setupTestServer(3, 1)
	defer s.Close()

	ws, _, err := setupTestUser("al", "aapl", uri)

	if err != nil {
		t.Fatalf("%v", err)
	}

	defer ws.Close()
}
func TestBroadcastMessage(t *testing.T) {

	s, uri := setupTestServer(3, 1)
	defer s.Close()

	users := [3][2]string{{"u1", "aapl"}, {"u2", "aapl"}, {"u3", "banana"}}
	ids := [3]int64{0, 0, 0}
	sockets := [3]*websocket.Conn{nil, nil, nil}

	for idx, usr := range users {
		ws, id, err := setupTestUser(usr[0], usr[1], uri)
		if err != nil {
			t.Fatalf("%v", err)
		}
		sockets[idx] = ws
		ids[idx] = id

		sockets[0].WriteJSON(types.Message{
			MsgText:   "Message",
			UserName:  users[0][0],
			UserID:    ids[0],
			TimeStamp: "0",
			VideoID:   users[0][1],
		})

		time.Sleep(1000 * time.Millisecond)

		defer ws.Close()
	}

	sockets[0].WriteJSON(types.Message{
		MsgText:   "Message",
		UserName:  users[0][0],
		UserID:    ids[0],
		TimeStamp: "0",
		VideoID:   users[0][1],
	})

	time.Sleep(1000 * time.Millisecond)

}
