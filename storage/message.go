package storage

type Message struct {
	From    string `json:"from,omitempty"`
	Channel string `json:"channel,omitempty"`
	Text    string `json:"text,omitempty"`
}
