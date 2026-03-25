# ccmanager

A terminal UI for browsing and resuming Claude Code sessions.

```
┌─────────────────────┬──────────────────────────────┬─────────────────┐
│ Sessions (12)       │ Chat - my-project            │ Files (8)       │
│                     │                              │                 │
│ * my-project  2m ago│ You >                        │ └── src/        │
│   api-server  1h ago│   Fix the login bug          │     ├── auth.go │
│   frontend    3d ago│                              │     └── main.go │
│                     │ Claude >                     │                 │
│                     │   I'll help fix that...      │                 │
└─────────────────────┴──────────────────────────────┴─────────────────┘
```

## Why?

Claude Code stores conversation history in `~/.claude/projects/` as JSONL files. These sessions persist across terminal restarts, but there's no easy way to:

- See all your past sessions at a glance
- Preview what a conversation was about before resuming
- Search across sessions by project or content
- Clean up old sessions you no longer need

ccmanager gives you a visual interface to manage these sessions without digging through files manually.

## Features

- **Three-panel layout** - Sessions list, chat preview, and file changes
- **Instant resume** - Press `r` to jump back into any session with Claude CLI
- **Search** - Find sessions by project name, path, or conversation content
- **Filter by project** - Focus on sessions from a specific codebase
- **Activity indicators** - Green/yellow/gray dots show session age
- **File tree** - See which files were touched in each session
- **Tool usage stats** - Count of Read, Edit, Write, Bash operations

## Install

```bash
# Clone and build
git clone <repo>
cd ccmanager
go build -o ccmanager .

# Or install directly
go install github.com/youruser/ccmanager@latest
```

Requires Go 1.21+ and the `claude` CLI installed.

## Usage

```bash
./ccmanager
```

### Keyboard Shortcuts

| Key | Action |
|-----|--------|
| `j/k` | Navigate up/down |
| `Enter` | Select session, switch to chat panel |
| `r` | Resume session in Claude CLI |
| `d` | Delete session |
| `/` | Search sessions |
| `p` | Filter by project |
| `Tab` | Switch panels |
| `g/G` | Jump to top/bottom |
| `Ctrl+u/d` | Half-page scroll |
| `?` | Show help |
| `q` | Quit |

## How It Works

1. **Scans** `~/.claude/projects/` for JSONL session files
2. **Parses** each file to extract timestamps, messages, and tool usage
3. **Displays** sessions sorted by last activity
4. **Resumes** sessions by calling `claude --resume <session-id>`

The session files are read-only - ccmanager never modifies your conversation history (except when you explicitly delete a session).

## File Structure

```
~/.claude/
└── projects/
    └── -Users-you-myproject/     # Encoded project path
        ├── abc123.jsonl          # Session file
        └── def456.jsonl          # Another session
```

ccmanager decodes these directory names back to readable paths and displays the project name in the UI.

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - UI components

## License

MIT
