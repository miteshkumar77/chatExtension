package types

import (
	"encoding/json"

	"github.com/gorilla/websocket"
)

// UIDType is the underlying type that the user id uses
type UIDType = int64

// Message is a variable type
type Message struct {
	MsgText   string  `json:"msgText"`
	UserName  string  `json:"userName"`
	UserID    UIDType `json:"userID"`
	TimeStamp string  `json:"timeStamp"`
	VideoID   string  `json:"videoID"`
}

// MarshalBinary -
func (msg *Message) MarshalBinary() ([]byte, error) {
	return json.Marshal(msg)
}

// UnmarshalBinary -
func (msg *Message) UnmarshalBinary(data []byte) error {
	if err := json.Unmarshal(data, &msg); err != nil {
		return err
	}

	return nil
}

// TransactionToken is the First message that client sends to server,
// identifying the clients username, videoID.
// server sends back the same message with a generated UserID
type TransactionToken struct {
	UserName string  `json:"userName"`
	VideoID  string  `json:"videoID"`
	UserID   UIDType `json:"userID"`
}

// User is the object that represents a connected user
type User struct {
	UserName string
	UserID   UIDType
	VideoID  string
	SockConn *websocket.Conn
}

// RedisChatChannel is the redis channel that all nodes use to exchange messages
const RedisChatChannel = "livechat"
