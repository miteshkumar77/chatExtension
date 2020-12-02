package manager

import (
	"chatExtensionServer/internal/concurrency/concurrentroomtable"
	"chatExtensionServer/internal/concurrency/concurrentusertable"
	"chatExtensionServer/internal/types"
	"fmt"
	"unsafe"

	"github.com/gorilla/websocket"
)

// PubSubMgr is an object that manages User publishes and subscribes
type PubSubMgr struct {
	Videos concurrentroomtable.ConcurrentHashMap
	Users  concurrentusertable.ConcurrentHashMap
}

// CreateNewPSMgr creates and initializes a new PubSubMgr object
func CreateNewPSMgr() PubSubMgr {
	var manager PubSubMgr
	manager.init()
	return manager
}

func (mgr *PubSubMgr) init() {
	mgr.Videos = concurrentroomtable.CreateNewRoomTable()
	mgr.Users = concurrentusertable.CreateNewUserTable()
}

// Connect connects a user
func (mgr *PubSubMgr) Connect(UserName string, videoID string, sockConn *websocket.Conn) types.UIDType {
	println("New User: " + UserName + " attempting to connect to video room: " + videoID + "...")
	var newUID types.UIDType = mgr.createNewUser(UserName, videoID, sockConn)
	println("New User: " + UserName + " connected with UID: " + fmt.Sprint(newUID) + " to video room: " + videoID + ".")
	println("Total " + fmt.Sprint(mgr.Users.Size()) + " open sockets, and " + fmt.Sprint(mgr.Videos.Size()) + " active rooms.")

	return newUID
}

// Disconnect disconnects a user
func (mgr *PubSubMgr) Disconnect(UserID types.UIDType) {
	println("User with UID: " + fmt.Sprint(UserID) + " attempting to disconnect...")
	mgr.deleteUser(UserID)
	println("User with UID: " + fmt.Sprint(UserID) + " disconnected.")
	println("Total " + fmt.Sprint(mgr.Users.Size()) + " open sockets, and " + fmt.Sprint(mgr.Videos.Size()) + " active rooms.")
}

func (mgr *PubSubMgr) createNewVideoRoom(videoID string) {
	mgr.Videos.Set(videoID, make(map[types.UIDType]bool))
}

func (mgr *PubSubMgr) deleteVideoRoom(videoID string) {
	mgr.Videos.Erase(videoID)
}

func (mgr *PubSubMgr) createNewUser(UserName string, videoID string, sockConn *websocket.Conn) types.UIDType {
	var newUserPtr *types.User = &types.User{UserName: UserName, UserID: 0, VideoID: videoID, SockConn: sockConn}
	newUserPtr.UserID = *(*types.UIDType)(unsafe.Pointer(newUserPtr))
	mgr.Users.Set(newUserPtr.UserID, *newUserPtr)
	if mgr.Videos.Contains(newUserPtr.VideoID) == false {
		mgr.createNewVideoRoom(newUserPtr.VideoID)
	}
	mgr.Videos.CallBackUpdate(newUserPtr.VideoID, func(original map[types.UIDType]bool) map[types.UIDType]bool {
		original[newUserPtr.UserID] = true
		return original
	})
	return newUserPtr.UserID
}

func (mgr *PubSubMgr) deleteUser(UserID types.UIDType) error {

	var err error
	mgr.Users.CallBackActionAndDelete(UserID, func(original types.User) {
		videoID := original.VideoID
		mgr.Videos.CallBackUpdateOrDelete(videoID, func(original map[types.UIDType]bool) (bool, map[types.UIDType]bool) {
			delete(original, UserID)
			return (len(original) == 0), original
		})

		err = original.SockConn.Close()
	})
	return err
}

// BroadcastMessage sends messages to all members within the room it was broadcasted
// in
func (mgr *PubSubMgr) BroadcastMessage(incomingMessage *types.Message) error {

	videoID := incomingMessage.VideoID
	senderID := incomingMessage.UserID
	var err error
	incomingMessage.UserID = 0
	mgr.Videos.CallBackAction(videoID, func(original map[types.UIDType]bool) {
		for uid := range original {
			mgr.Users.CallBackAction(uid, func(user types.User) {

				println("Sending a message to: " + user.UserName + ".")

				if uid == senderID {
					incomingMessage.UserID = senderID
				}
				err = mgr.broadcastMessageToSock(incomingMessage, user.SockConn)

				if uid == senderID {
					incomingMessage.UserID = 0
				}
			})
		}
	})

	incomingMessage.UserID = senderID
	return err
}

// SendTokenToUser sends the connection handshake token to the appropriate user
func (mgr *PubSubMgr) SendTokenToUser(incomingToken *types.TransactionToken) error {
	var err error
	mgr.Users.CallBackAction(incomingToken.UserID, func(user types.User) {
		err = user.SockConn.WriteJSON(incomingToken)
	})
	return err
}

func (mgr *PubSubMgr) broadcastMessageToSock(incomingMessage *types.Message, sock *websocket.Conn) error {
	return sock.WriteJSON(&incomingMessage)
}
