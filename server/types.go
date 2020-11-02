package main

// Message : variable type
type Message struct {
	MsgText   string `json:"msgText"`
	UserName  string `json:"userName"`
	UserID    uint32 `json:"userID"`
	TimeStamp string `json:"timeStamp"`
	VideoID   string `json:"videoID"`
}

// TransactionToken : variable type
type TransactionToken struct {
	UserName string `json:"userName"`
	VideoID  string `json:"videoID"`
	UserID   uint32 `json:"userID"`
}
