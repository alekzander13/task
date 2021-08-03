package main

type messageText struct {
	MsgID     int64  `json:"msgID"`
	Type      string `json:"type,omitempty"`
	TagSelf   string `json:"tag,omitempty"`
	TagClient string `json:"tagClient,omitempty"`
	Text      string `json:"text,omitempty"`
	Error     string `json:"error,omitempty"`
}

type message struct {
	Count int
	Text  []byte
	Type  string
}
