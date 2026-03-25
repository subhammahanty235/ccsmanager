package session

import (
	"encoding/json"
	"strings"
	"time"
)

func ParseJSONLLine(line []byte) []*Message {
	if len(line) == 0 {
		return nil
	}

	defer func() { recover() }()

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(line, &raw); err != nil {
		return nil
	}

	var msgType string
	if typeRaw, ok := raw["type"]; ok {
		json.Unmarshal(typeRaw, &msgType)
	}

	if msgType == "file-history-snapshot" || msgType == "" {
		return nil
	}

	timestamp := parseTimestamp(raw)

	var messages []*Message

	switch msgType {
	case "user":
		content := extractUserContent(raw)
		if content == "" {
			return nil
		}
		if isMeta(raw) || isCommandOutput(content) {
			return nil
		}
		messages = append(messages, &Message{
			Type:      MessageTypeHuman,
			Content:   content,
			Timestamp: timestamp,
			Raw:       line,
		})

	case "assistant":
		messages = append(messages, extractAssistantMessages(raw, timestamp, line)...)

	default:
		content := extractUserContent(raw)
		if content != "" {
			messages = append(messages, &Message{
				Type:      MessageTypeSystem,
				Content:   content,
				Timestamp: timestamp,
				Raw:       line,
			})
		}
	}

	return messages
}

func parseTimestamp(raw map[string]json.RawMessage) time.Time {
	tsRaw, ok := raw["timestamp"]
	if !ok {
		return time.Time{}
	}

	var tsStr string
	if err := json.Unmarshal(tsRaw, &tsStr); err != nil {
		return time.Time{}
	}

	if ts, err := time.Parse(time.RFC3339, tsStr); err == nil {
		return ts
	}
	if ts, err := time.Parse(time.RFC3339Nano, tsStr); err == nil {
		return ts
	}
	return time.Time{}
}

func isMeta(raw map[string]json.RawMessage) bool {
	metaRaw, ok := raw["isMeta"]
	if !ok {
		return false
	}
	var meta bool
	json.Unmarshal(metaRaw, &meta)
	return meta
}

func isCommandOutput(content string) bool {
	return strings.HasPrefix(content, "<command-") || strings.HasPrefix(content, "<local-command")
}

func extractUserContent(raw map[string]json.RawMessage) string {
	msgRaw, ok := raw["message"]
	if !ok {
		return ""
	}

	var msgObj map[string]json.RawMessage
	if err := json.Unmarshal(msgRaw, &msgObj); err != nil {
		return ""
	}

	contentRaw, ok := msgObj["content"]
	if !ok {
		return ""
	}

	var contentStr string
	if err := json.Unmarshal(contentRaw, &contentStr); err == nil {
		return contentStr
	}

	var contentArr []map[string]interface{}
	if err := json.Unmarshal(contentRaw, &contentArr); err == nil {
		var parts []string
		for _, block := range contentArr {
			blockType, _ := block["type"].(string)
			if blockType == "tool_result" {
				if c, ok := block["content"].(string); ok {
					parts = append(parts, c)
				}
			} else if blockType == "text" {
				if t, ok := block["text"].(string); ok {
					parts = append(parts, t)
				}
			}
		}
		return strings.Join(parts, "\n")
	}

	return ""
}

func extractAssistantMessages(raw map[string]json.RawMessage, timestamp time.Time, line []byte) []*Message {
	var messages []*Message

	msgRaw, ok := raw["message"]
	if !ok {
		return nil
	}

	var msgObj map[string]json.RawMessage
	if err := json.Unmarshal(msgRaw, &msgObj); err != nil {
		return nil
	}

	contentRaw, ok := msgObj["content"]
	if !ok {
		return nil
	}

	var contentArr []map[string]json.RawMessage
	if err := json.Unmarshal(contentRaw, &contentArr); err != nil {
		var contentStr string
		if err := json.Unmarshal(contentRaw, &contentStr); err == nil && contentStr != "" {
			messages = append(messages, &Message{
				Type:      MessageTypeAssistant,
				Content:   contentStr,
				Timestamp: timestamp,
				Raw:       line,
			})
		}
		return messages
	}

	for _, block := range contentArr {
		var blockType string
		if typeRaw, ok := block["type"]; ok {
			json.Unmarshal(typeRaw, &blockType)
		}

		switch blockType {
		case "text":
			var text string
			if textRaw, ok := block["text"]; ok {
				json.Unmarshal(textRaw, &text)
			}
			if text != "" {
				messages = append(messages, &Message{
					Type:      MessageTypeAssistant,
					Content:   text,
					Timestamp: timestamp,
					Raw:       line,
				})
			}

		case "tool_use":
			var name string
			if nameRaw, ok := block["name"]; ok {
				json.Unmarshal(nameRaw, &name)
			}

			var input map[string]interface{}
			if inputRaw, ok := block["input"]; ok {
				json.Unmarshal(inputRaw, &input)
			}

			if name != "" {
				messages = append(messages, &Message{
					Type:      MessageTypeToolUse,
					ToolName:  name,
					ToolInput: input,
					Timestamp: timestamp,
					Raw:       line,
				})
			}
		}
	}

	return messages
}

func ExtractFileChanges(messages []*Message) []FileChange {
	fileOps := make(map[string]*FileChange)

	for _, msg := range messages {
		if msg == nil || msg.Type != MessageTypeToolUse {
			continue
		}

		var path, operation string

		switch msg.ToolName {
		case "Edit":
			path = getStringField(msg.ToolInput, "file_path")
			operation = "edited"
		case "Write":
			path = getStringField(msg.ToolInput, "file_path")
			operation = "created"
		case "Read":
			path = getStringField(msg.ToolInput, "file_path")
			operation = "read"
		}

		if path == "" {
			continue
		}

		if existing, ok := fileOps[path]; ok {
			existing.Count++
			if operation == "created" && existing.Operation == "read" {
				existing.Operation = "created"
			} else if operation == "edited" && existing.Operation == "read" {
				existing.Operation = "edited"
			}
		} else {
			fileOps[path] = &FileChange{
				Path:      path,
				Operation: operation,
				Count:     1,
			}
		}
	}

	var changes []FileChange
	for _, fc := range fileOps {
		if fc.Operation != "read" {
			changes = append(changes, *fc)
		}
	}
	return changes
}

func getStringField(m map[string]interface{}, key string) string {
	if m == nil {
		return ""
	}
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func CountToolUsage(messages []*Message) map[string]int {
	counts := make(map[string]int)
	for _, msg := range messages {
		if msg != nil && msg.Type == MessageTypeToolUse && msg.ToolName != "" {
			counts[msg.ToolName]++
		}
	}
	return counts
}

func FormatToolUse(msg *Message) string {
	if msg == nil || msg.Type != MessageTypeToolUse {
		return ""
	}

	path := getStringField(msg.ToolInput, "file_path")
	if path != "" {
		parts := strings.Split(path, "/")
		path = parts[len(parts)-1]
	}

	switch msg.ToolName {
	case "Edit":
		if path != "" {
			return "[Tool: Edit " + path + "]"
		}
		return "[Tool: Edit]"

	case "Write":
		if path != "" {
			return "[Tool: Write " + path + "]"
		}
		return "[Tool: Write]"

	case "Read":
		if path != "" {
			return "[Tool: Read " + path + "]"
		}
		return "[Tool: Read]"

	case "Bash":
		cmd := getStringField(msg.ToolInput, "command")
		if cmd != "" {
			if len(cmd) > 30 {
				cmd = cmd[:27] + "..."
			}
			return "[Tool: Bash \"" + cmd + "\"]"
		}
		return "[Tool: Bash]"

	case "Grep":
		pattern := getStringField(msg.ToolInput, "pattern")
		if pattern != "" {
			if len(pattern) > 20 {
				pattern = pattern[:17] + "..."
			}
			return "[Tool: Grep \"" + pattern + "\"]"
		}
		return "[Tool: Grep]"

	case "Glob":
		pattern := getStringField(msg.ToolInput, "pattern")
		if pattern != "" {
			return "[Tool: Glob " + pattern + "]"
		}
		return "[Tool: Glob]"

	case "TodoWrite":
		return "[Tool: TodoWrite]"

	default:
		return "[Tool: " + msg.ToolName + "]"
	}
}
