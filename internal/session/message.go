package session

import (
	"encoding/json"
	"time"
)

type MessageType string

const (
	MessageTypeHuman      MessageType = "human"
	MessageTypeAssistant  MessageType = "assistant"
	MessageTypeToolUse    MessageType = "tool_use"
	MessageTypeToolResult MessageType = "tool_result"
	MessageTypeSystem     MessageType = "system"
)

type Message struct {
	Type       MessageType
	Content    string
	Timestamp  time.Time
	ToolName   string
	ToolInput  map[string]interface{}
	ToolOutput string
	Raw        json.RawMessage
}

type FileChange struct {
	Path      string
	Operation string
	Count     int
}
