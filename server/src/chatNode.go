package main

import (
	"fmt"
	"unsafe"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

type user struct {
	userName string
	userID   uidType
	videoID  string
	sockConn *websocket.Conn
}

// PubSubMgr is an object that manages user publishes and subscribes
type PubSubMgr struct {
	videos   map[string]map[*user]bool // map of video IDs to collections of user pointers
	users    map[uidType]*user         // map of user IDs to user pointers
	receiver *redis.PubSub             // redis channel
	lock     chan bool
}

func (this *PubSubMgr) Connect(userName string, videoID string, sockConn *websocket.Conn) uidType {
	println("New User: " + userName + " attempting to connect to video room: " + videoID + "...")
	var newUID uidType = this.createNewUser(userName, videoID, sockConn)
	println("New User: " + userName + " connected with UID: " + fmt.Sprint(newUID) + " to video room: " + videoID + ".")
	println("Total " + fmt.Sprint(len(this.users)) + " open sockets, and " + fmt.Sprint(len(this.videos)) + " active rooms.")
	return newUID
}

func (this *PubSubMgr) Disconnect(userID uidType) {
	println("User with UID: " + fmt.Sprint(userID) + " attempting to disconnect...")
	this.deleteUser(userID)
	println("User with UID: " + fmt.Sprint(userID) + " disconnected.")
	println("Total " + fmt.Sprint(len(this.users)) + " open sockets, and " + fmt.Sprint(len(this.videos)) + " active rooms.")
}

func (this *PubSubMgr) createNewVideoRoom(videoID string) {
	this.videos[videoID] = make(map[*user]bool)
}

func (this *PubSubMgr) deleteVideoRoom(videoID string) {
	delete(this.videos, videoID)
}

func (this *PubSubMgr) createNewUser(userName string, videoID string, sockConn *websocket.Conn) uidType {
	this.lock <- true
	var newUserPtr *user = &user{userName: userName, userID: 0, videoID: videoID, sockConn: sockConn}
	newUserPtr.userID = *(*uidType)(unsafe.Pointer(newUserPtr))
	this.users[newUserPtr.userID] = newUserPtr
	if _, exists := this.videos[newUserPtr.videoID]; !exists {
		this.createNewVideoRoom(newUserPtr.videoID)
	}
	this.videos[newUserPtr.videoID][newUserPtr] = true
	<-this.lock
	return newUserPtr.userID
}

func (this *PubSubMgr) deleteUser(userID uidType) error {
	this.lock <- true
	var userPtr = this.users[userID]
	var videoID = userPtr.videoID
	delete(this.videos[videoID], userPtr)
	if len(this.videos[videoID]) == 0 {
		this.deleteVideoRoom(videoID)
	}
	delete(this.users, userID)
	err := userPtr.sockConn.Close()
	userPtr = nil
	<-this.lock
	return err
}

func (this *PubSubMgr) BroadcastMessage(incomingMessage *Message) error {
	this.lock <- true
	<-this.lock

	var videoID = incomingMessage.VideoID
	var err error
	var senderID = incomingMessage.UserID

	incomingMessage.UserID = 0
	for k := range this.videos[videoID] {
		println("Sending a message to: " + k.userName + ".")
		if k.userID == senderID {
			incomingMessage.UserID = senderID
		}
		err = this.broadcastMessageToSock(incomingMessage, k.sockConn)

		if k.userID == senderID {
			incomingMessage.UserID = 0
		}
	}
	return err
}

func (this *PubSubMgr) SendTokenToUser(incomingToken *TransactionToken) error {
	return this.users[incomingToken.UserID].sockConn.WriteJSON(incomingToken)

}

func (this *PubSubMgr) broadcastMessageToSock(incomingMessage *Message, sock *websocket.Conn) error {
	return sock.WriteJSON(&incomingMessage)
}
