package main

import (
	"chatExtensionServer/internal/concurrency/concurrentroomtable"
	"chatExtensionServer/internal/concurrency/concurrentusertable"
	"chatExtensionServer/internal/types"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func reader(ws *websocket.Conn, jobs *SafeQueue, rateLimiter *RateLimiter) {
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

func process(jobs *SafeQueue, mgr *PubSubMgr, rateLimiter *RateLimiter) {
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
	jobs *SafeQueue, mgr *PubSubMgr, rateLimiter *RateLimiter) {
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

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h>Welcome to the Home Page</h>")
}

func main() {

	var sPORT int = 5678
	var sThreads int = 3
	var rateLimit uint16 = 1
	var variable, exists = os.LookupEnv("PORT")
	if exists {
		sPORT, _ = strconv.Atoi(variable)
	}

	variable, exists = os.LookupEnv("THREADS")
	if exists {
		sThreads, _ = strconv.Atoi(variable)
	}

	variable, exists = os.LookupEnv("RATELIMIT")
	if exists {
		tmp, _ := strconv.ParseUint(variable, 16, 16)
		rateLimit = uint16(tmp)
	}
	var jobs SafeQueue
	var rateLimiter RateLimiter
	jobs.Init()
	rateLimiter.Init(rateLimit)
	var mgr PubSubMgr = PubSubMgr{concurrentroomtable.CreateNewRoomTable(), concurrentusertable.CreateNewUserTable()}

	for i := 0; i < sThreads; i++ {
		go process(&jobs, &mgr, &rateLimiter)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws",
		func(w http.ResponseWriter, r *http.Request) { wsEndpoint(w, r, &jobs, &mgr, &rateLimiter) })

	var portStr string = ":" + strconv.Itoa(sPORT)
	log.Fatal(http.ListenAndServe(portStr, nil))
}
