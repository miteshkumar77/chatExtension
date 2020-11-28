package main

import (
	"chatExtensionServer/internal/types"
	"fmt"
	"unsafe"

	"github.com/gorilla/websocket"
)

// PubSubMgr is an object that manages User publishes and subscribes
type PubSubMgr struct {
	videos map[string]map[*types.User]bool // map of video IDs to collections of User pointers
	Users  map[types.UIDType]*types.User   // map of User IDs to User pointers

}

func (this *PubSubMgr) Connect(UserName string, videoID string, sockConn *websocket.Conn) types.UIDType {
	println("New User: " + UserName + " attempting to connect to video room: " + videoID + "...")
	var newUID types.UIDType = this.createNewUser(UserName, videoID, sockConn)
	println("New User: " + UserName + " connected with UID: " + fmt.Sprint(newUID) + " to video room: " + videoID + ".")
	println("Total " + fmt.Sprint(len(this.Users)) + " open sockets, and " + fmt.Sprint(len(this.videos)) + " active rooms.")

	return newUID
}

func (this *PubSubMgr) Disconnect(UserID types.UIDType) {
	println("User with UID: " + fmt.Sprint(UserID) + " attempting to disconnect...")
	this.deleteUser(UserID)
	println("User with UID: " + fmt.Sprint(UserID) + " disconnected.")
	println("Total " + fmt.Sprint(len(this.Users)) + " open sockets, and " + fmt.Sprint(len(this.videos)) + " active rooms.")
}

func (this *PubSubMgr) createNewVideoRoom(videoID string) {
	this.videos[videoID] = make(map[*types.User]bool)
}

func (this *PubSubMgr) deleteVideoRoom(videoID string) {
	delete(this.videos, videoID)
}

func (this *PubSubMgr) createNewUser(UserName string, videoID string, sockConn *websocket.Conn) types.UIDType {
	var newUserPtr *types.User = &types.User{UserName: UserName, UserID: 0, VideoID: videoID, SockConn: sockConn}
	newUserPtr.UserID = *(*types.UIDType)(unsafe.Pointer(newUserPtr))
	this.Users[newUserPtr.UserID] = newUserPtr
	if _, exists := this.videos[newUserPtr.VideoID]; !exists {
		this.createNewVideoRoom(newUserPtr.VideoID)
	}
	this.videos[newUserPtr.VideoID][newUserPtr] = true
	return newUserPtr.UserID
}

func (this *PubSubMgr) deleteUser(UserID types.UIDType) error {
	var UserPtr = this.Users[UserID]
	var videoID = UserPtr.VideoID
	delete(this.videos[videoID], UserPtr)
	if len(this.videos[videoID]) == 0 {
		this.deleteVideoRoom(videoID)
	}
	delete(this.Users, UserID)
	err := UserPtr.SockConn.Close()
	UserPtr = nil
	return err
}

func (this *PubSubMgr) BroadcastMessage(incomingMessage *types.Message) error {
	var videoID = incomingMessage.VideoID
	var err error
	var senderID = incomingMessage.UserID

	incomingMessage.UserID = 0
	for k := range this.videos[videoID] {
		println("Sending a message to: " + k.UserName + ".")
		if k.UserID == senderID {
			incomingMessage.UserID = senderID
		}
		err = this.broadcastMessageToSock(incomingMessage, k.SockConn)

		if k.UserID == senderID {
			incomingMessage.UserID = 0
		}
	}
	return err
}

func (this *PubSubMgr) SendTokenToUser(incomingToken *types.TransactionToken) error {
	return this.Users[incomingToken.UserID].SockConn.WriteJSON(incomingToken)

}

func (this *PubSubMgr) broadcastMessageToSock(incomingMessage *types.Message, sock *websocket.Conn) error {
	return sock.WriteJSON(&incomingMessage)
}
