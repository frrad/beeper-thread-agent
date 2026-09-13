package main

import "golang.org/x/oauth2"

type Attachment struct {
	Type          string `json:"type"`
	MIMEType      string `json:"mimeType"`
	FileName      string `json:"fileName"`
	FileSize      int64  `json:"fileSize"`
	SrcURL        string `json:"srcURL"`
	Transcription *struct {
		Text   string `json:"transcription"`
		Engine string `json:"engine"`
	} `json:"transcription"`
}

type BeeperMessage struct {
	ID          string       `json:"id"`
	ChatID      string       `json:"chatID"`
	SenderID    string       `json:"senderID"`
	SenderName  string       `json:"senderName"`
	Timestamp   string       `json:"timestamp"`
	SortKey     string       `json:"sortKey"`
	Type        string       `json:"type"`
	Text        string       `json:"text"`
	IsSender    bool         `json:"isSender"`
	IsDeleted   bool         `json:"isDeleted"`
	Attachments []Attachment `json:"attachments"`
}

type MessagePage struct {
	Items        []BeeperMessage `json:"items"`
	HasMore      bool            `json:"hasMore"`
	OldestCursor string          `json:"oldestCursor"`
	NewestCursor string          `json:"newestCursor"`
}

type ConversationTurn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatState struct {
	Checkpoint string             `json:"checkpoint,omitempty"`
	History    []ConversationTurn `json:"history,omitempty"`
}

type OAuthState struct {
	Config *oauth2.Config `json:"config,omitempty"`
	Token  *oauth2.Token  `json:"token,omitempty"`
}

type PersistedState struct {
	OAuth *OAuthState           `json:"oauth,omitempty"`
	Chats map[string]*ChatState `json:"chats"`
}
