package main

import (
	"chatExtensionServer/internal/concurrency/concurrentqueue"
	"chatExtensionServer/internal/manager"
	"chatExtensionServer/internal/ratelimiting"
	"chatExtensionServer/internal/types"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var ctx = context.Background()

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func reader(ws *websocket.Conn, jobs *concurrentqueue.SafeQueue, rateLimiter *ratelimiting.RateLimiter, rdb *redis.Client) {
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

		println("Publishing message to channel!")
		// Post the incoming message to channel
		binaryMessage, err := m.MarshalBinary()

		if err != nil {
			log.Fatalf("Error marshalling incoming message")
		}

		err = rdb.Publish(ctx, types.RedisChatChannel, binaryMessage).Err()
		if err != nil {
			log.Fatalf("Error publishing message: %v", err)
		}
	}
}

func redisReader(ps *redis.PubSub, jobs *concurrentqueue.SafeQueue) {

	for true {
		ch := ps.Channel()

		for msg := range ch {

			fmt.Println("Got a message: ")
			fmt.Println(msg.Channel, msg.Payload)

			var marshalledMessage types.Message
			err := marshalledMessage.UnmarshalBinary([]byte(msg.Payload))

			if err != nil {
				log.Fatalf("%v", err)
			}
			jobs.Push(&marshalledMessage)
		}
	}
}

func process(jobs *concurrentqueue.SafeQueue, mgr *manager.PubSubMgr, rateLimiter *ratelimiting.RateLimiter) {
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
	jobs *concurrentqueue.SafeQueue, mgr *manager.PubSubMgr, rateLimiter *ratelimiting.RateLimiter, rdb *redis.Client, appID int) {
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

	println("User connected to room: [" + fmt.Sprint(appID) + "]")

	defer mgr.Disconnect(t.UserID)

	err = mgr.SendTokenToUser(&t)

	if err != nil {
		println("Error sending back the token!")
		log.Fatal(err)
	}

	reader(ws, jobs, rateLimiter, rdb)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "<h>Welcome to the Home Page</h>")
}

func main() {

	var sPORT int = 5678
	var sThreads int = 3
	var sRateLimit uint16 = 1
	var sAPPID int = 1

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
		sRateLimit = uint16(tmp)
	}

	variable, exists = os.LookupEnv("APPID")
	if exists {
		sAPPID, _ = strconv.Atoi(variable)
	}

	var jobs concurrentqueue.SafeQueue
	var rateLimiter ratelimiting.RateLimiter
	jobs.Init()
	rateLimiter.Init(sRateLimit)
	rdb := redis.NewClient(&redis.Options{
		Addr:     "rds:6379",
		Password: "mypassword",
		DB:       0,
	})

	ps := rdb.Subscribe(ctx, types.RedisChatChannel)

	_, err := ps.Receive(ctx)
	if err != nil {
		log.Fatalf("%v", err)
	}

	mgr := manager.CreateNewPSMgr()

	go redisReader(ps, &jobs)

	for i := 0; i < sThreads; i++ {
		go process(&jobs, &mgr, &rateLimiter)
	}

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/ws",
		func(w http.ResponseWriter, r *http.Request) {
			wsEndpoint(w, r, &jobs, &mgr, &rateLimiter, rdb, sAPPID)
		})
	print("Running on port: [" + fmt.Sprint(sPORT) + "]\nThreads: [" + fmt.Sprint(sThreads) + "]\nMessages at a time: [" + fmt.Sprint(sRateLimit) + "]\nAPPID: [" + fmt.Sprint(sAPPID) + "]\n")

	var portStr string = ":" + strconv.Itoa(sPORT)
	log.Fatal(http.ListenAndServe(portStr, nil))
}
