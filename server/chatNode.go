package main

import (
	"fmt"
	"unsafe"

	"github.com/gorilla/websocket"
)

type user struct {
	userName string
	userID   uint32
	videoID  string
	sockConn *websocket.Conn
}

// PubSubMgr : Publish Subscribe Manager
type PubSubMgr struct {
	videos map[string]map[*user]bool // map of video IDs to collections of user pointers
	users  map[uint32]*user          // map of user IDs to user pointers
}

func (this *PubSubMgr) Connect(userName string, videoID string, sockConn *websocket.Conn) uint32 {
	println("New User: " + userName + " attempting to connect to video room: " + videoID + "...")
	var newUID uint32 = this.createNewUser(userName, videoID, sockConn)
	println("New User: " + userName + " connected with UID: " + fmt.Sprint(newUID) + " to video room: " + videoID + ".")
	println("Total " + fmt.Sprint(len(this.users)) + " open sockets, and " + fmt.Sprint(len(this.videos)) + " active rooms.")

	return newUID
}

func (this *PubSubMgr) Disconnect(userID uint32) {
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

func (this *PubSubMgr) createNewUser(userName string, videoID string, sockConn *websocket.Conn) uint32 {
	var newUserPtr *user = &user{userName: userName, userID: 0, videoID: videoID, sockConn: sockConn}
	newUserPtr.userID = *(*uint32)(unsafe.Pointer(newUserPtr))
	this.users[newUserPtr.userID] = newUserPtr
	if _, exists := this.videos[newUserPtr.videoID]; !exists {
		this.createNewVideoRoom(newUserPtr.videoID)
	}
	this.videos[newUserPtr.videoID][newUserPtr] = true
	return newUserPtr.userID
}

func (this *PubSubMgr) deleteUser(userID uint32) error {
	var userPtr = this.users[userID]
	var videoID = userPtr.videoID
	delete(this.videos[videoID], userPtr)
	if len(this.videos[videoID]) == 0 {
		this.deleteVideoRoom(videoID)
	}
	delete(this.users, userID)
	err := userPtr.sockConn.Close()
	userPtr = nil
	return err
}

func (this *PubSubMgr) BroadcastMessage(incomingMessage *Message) error {
	var videoID = incomingMessage.VideoID
	var err error
	incomingMessage.UserID = 0
	for k := range this.videos[videoID] {
		println("Sending a message to: " + k.userName + ".")
		err = this.broadcastMessageToSock(incomingMessage, k.sockConn)
	}
	return err
}

func (this *PubSubMgr) SendTokenToUser(incomingToken *TransactionToken) error {
	return this.users[incomingToken.UserID].sockConn.WriteJSON(incomingToken)

}

func (this *PubSubMgr) broadcastMessageToSock(incomingMessage *Message, sock *websocket.Conn) error {
	return sock.WriteJSON(&incomingMessage)
}
