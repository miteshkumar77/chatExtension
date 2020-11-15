package main

// Message : variable type
type Message struct {
	MsgText   string `json:"msgText"`
	UserName  string `json:"userName"`
	UserID    uint32 `json:"userID"`
	TimeStamp string `json:"timeStamp"`
	VideoID   string `json:"videoID"`
}

/**
	TransactionToken: First message that client sends to server,
		identifying the client's username, videoID. 

		server sends back the same message with a generated UserID
*/
type TransactionToken struct {
	UserName string `json:"userName"`
	VideoID  string `json:"videoID"`
	UserID   uint32 `json:"userID"`
}