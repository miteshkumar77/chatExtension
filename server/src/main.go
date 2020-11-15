package main

import (
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

func reader(ws *websocket.Conn, jobs *SafeQueue) {
	for {
		var m Message

		err := ws.ReadJSON(&m)
		if err != nil {
			log.Println(err)
			return
		}

		jobs.Push(&m)
	}
}

func process(jobs *SafeQueue, mgr *PubSubMgr) {
	for true {
		var item *Message = jobs.Pop()
		err := mgr.BroadcastMessage(item)
		if err != nil {
			println("Error broadcasting message!")
			log.Fatal(err)
		}
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request, jobs *SafeQueue, mgr *PubSubMgr) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		println("Error upgrading socket!")
		log.Fatal(err)

	}

	var t TransactionToken
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

	reader(ws, jobs)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h>Welcome to the Home Page</h>")
}

func main() {

	var sPORT int = 5678
	var sThreads int = 3

	var variable, exists = os.LookupEnv("PORT")
	if exists {
		sPORT, _ = strconv.Atoi(variable)
	}

	variable, exists = os.LookupEnv("THREADS")
	if exists {
		sThreads, _ = strconv.Atoi(variable)
	}

	var jobs SafeQueue
	jobs.Init()
	var mgr PubSubMgr = PubSubMgr{make(map[string]map[*user]bool), make(map[uint32]*user)}

	for i := 0; i < sThreads; i++ {
		go process(&jobs, &mgr)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws",
		func(w http.ResponseWriter, r *http.Request) { wsEndpoint(w, r, &jobs, &mgr) })

	var portStr string = ":" + strconv.Itoa(sPORT)
	log.Fatal(http.ListenAndServe(portStr, nil))
}
