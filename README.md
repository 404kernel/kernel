<div align="center">

```
██╗  ██╗███████╗██████╗ ███╗   ██╗███████╗██╗     
██║ ██╔╝██╔════╝██╔══██╗████╗  ██║██╔════╝██║     
█████╔╝ █████╗  ██████╔╝██╔██╗ ██║█████╗  ██║     
██╔═██╗ ██╔══╝  ██╔══██╗██║╚██╗██║██╔══╝  ██║     
██║  ██╗███████╗██║  ██║██║ ╚████║███████╗███████╗
╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝╚══════╝╚══════╝
```

### ⚡ The Cyberpunk Terminal AI Agent with an Animated Desktop Pet Companion 🐶

[![GitHub Release](https://img.shields.io/github/v/release/404kernel/kernel?color=00D7FE&logo=github&style=for-the-badge)](https://github.com/404kernel/kernel/releases)
[![CI & Release](https://img.shields.io/github/actions/workflow/status/404kernel/kernel/ci.yml?branch=main&label=CI%20BUILD&logo=githubactions&style=for-the-badge)](https://github.com/404kernel/kernel/actions)
[![Go Version](https://img.shields.io/github/go-mod/go-version/404kernel/kernel?color=7D56F4&logo=go&style=for-the-badge)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-04B575.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

<br/>

[Quick Start](#-quick-start) • [Features](#-features) • [Slash Commands](#-slash-commands--tools) • [Companion Sparky](#-pet-companion-gamification) • [Keybindings](#%EF%B8%8F-keybindings)

<br/>

```text
╭─────────────────────────────────────────────────────────────────────────────────────────────╮
│ ⚡ KERNEL   GEMINI :: gemini-3.5-flash   cwd: C:\Users\Dev\workspace                       │
╰─────────────────────────────────────────────────────────────────────────────────────────────╯
╭─────────────────────────────────────────────────────────────────────────────────────────────╮
│   /^ ^\     COMPANION: Sparky  [Lv. 2 Terminal Scout]                                       │
│  / • • \    Status: listening                                                               │
│  V\ Y /V    Affection: 100%  |  XP: 45/60                                                   │
│   / - \  ~                                                                                  │
│  (____)m                                                                                    │
╰─────────────────────────────────────────────────────────────────────────────────────────────╯

╭─────────────────────────────────────────────────────────────────────────────────────────────╮
│  USER   15:04                                                                               │
│                                                                                             │
│ Write a concurrent crawler in Go using worker pools and backpressure channels               │
╰─────────────────────────────────────────────────────────────────────────────────────────────╯

╭─────────────────────────────────────────────────────────────────────────────────────────────╮
│  KERNEL  ● ○ ○ 15:04                                                                        │
│                                                                                             │
│ # Concurrent Worker Pool with Bounded Channels                                              │
│                                                                                             │
│ ```go                                                                                       │
│ type Job func(ctx context.Context) error                                                    │
│                                                                                             │
│ type Pool struct {                                                                          │
│     workers int                                                                             │
│     jobs    chan Job                                                                        │
│ }                                                                                           │
│ ```                                                                                         │
╰─────────────────────────────────────────────────────────────────────────────────────────────╯

╭─────────────────────────────────────────────────────────────────────────────────────────────╮
│ > Ask kernel to explain code, generate a game, or run tasks...                             │
╰─────────────────────────────────────────────────────────────────────────────────────────────╯
[Enter] submit · [Ctrl+C] cancel stream · [:q] quit
```

</div>

---

## 🌟 Why Kernel?

Most CLI AI tools are either bloated Python scripts or plain text wrappers that dump raw text without style. 

**Kernel** is built from the ground up in pure **Go** using **Charm's Bubble Tea, Lip Gloss, and Glamour**. It gives developers a keyboard-driven terminal IDE assistant equipped with:
- 🏎️ **Instant 1-second startup** with zero external dependencies (no Node, no Python).
- 🐶 **Sparky the ASCII Dog Companion** who reacts to your coding, levels up with XP, and watches your tasks.
- 🎨 **Glamour Syntax Highlighting** for Go, Python, Rust, JavaScript, TypeScript, C++, and Bash.
- 🌐 **Zero-Key Web Search** to inspect documentation or debug issues without leaving your terminal.
- 🛠️ **Native Shell Execution** with `$ <cmd>` directly inside the prompt.
- 🧠 **Multi-Turn Context & Multi-Provider AI Engine** supporting Gemini, Claude, OpenAI, Ollama, and Custom endpoints.

---

## ✨ Features

| Feature | Description |
|---|---|
| 🐶 **Sparky Companion** | Animated canine that blinks, sniffs, sleeps when idle, wags tail on completion, and levels up from *Code Puppy* to *Kernel Archon*. |
| ⚡ **Glamour Syntax Cards** | Terminal code blocks render with rich dark-theme syntax highlighting, line wrapping, and bordered cards. |
| 🌐 **Live Web Search** | Built-in `/web` command queries DuckDuckGo for live docs & solutions without needing any search API keys. |
| 💻 **Native Shell Tool** | Run `$ git status`, `$ npm test`, or `/exec <cmd>` directly inside the TUI with exit code & timing telemetry. |
| 💾 **Session History** | Automatic conversation archiving to `~/.kernel/history.json` and one-click Markdown exporting via `/save`. |
| 🛡️ **Interactive Setup Wizard** | Guided onboarding menu to configure API keys with persistent storage in `~/.kernel/config.json`. |
| 🪟 **Embedded Windows Icon** | Native `>_` terminal application icon embedded directly into `kernel.exe` using PE resource compilation. |

---

## 🚀 Quick Start

### Option 1: One-Line Install (Fastest)

**Windows (PowerShell):**
```powershell
iwr -useb https://raw.githubusercontent.com/404kernel/kernel/main/install.ps1 | iex
```

**Linux & macOS (Bash/Zsh):**
```bash
curl -sSL https://raw.githubusercontent.com/404kernel/kernel/main/install.sh | bash
```

### Option 2: Pre-compiled Binaries

Download from the [Releases](https://github.com/404kernel/kernel/releases) page:
- **Windows**: `kernel-windows-amd64.exe` / `kernel-windows-arm64.exe`
- **macOS**: `kernel-darwin-arm64` / `kernel-darwin-amd64`
- **Linux**: `kernel-linux-amd64` / `kernel-linux-arm64`

### Option 3: Go Install
```bash
go install github.com/404kernel/kernel@latest
```

### Option 4: Build From Source
```bash
# Clone the repository
git clone https://github.com/404kernel/kernel.git
cd kernel

# Run immediately
go run main.go

# Or compile your own binary with embedded icon (Windows)
go build -ldflags="-s -w" -o kernel.exe .
```

---

## 🎮 Slash Commands & Tools

Kernel includes a suite of instant commands accessible directly from the input bar:

| Command | Action |
|---|---|
| `/help` | Displays the interactive command and shortcuts reference |
| `/clear` | Clears the viewport for a clean conversation |
| `/pet` | Pet companion (+20 XP, restores affection to 100% ♥) |
| `/skin <name>` | Change companion avatar (`cat`, `dragon`, `drone`, `dog`) |
| `/stats` | Live system telemetry, RAM & CPU dashboard |
| `/tree [dir]` | Visual directory tree explorer with file type icons |
| `/diff <file>` | Inspect git diffs directly in the terminal |
| `/auto <goal>` | Autonomous coding agent workflow |
| `/voice [on/off]` | Toggle sound effects, barks, and chimes |
| `/speak <text>` | Text-to-speech audio reader |
| `/model <id>` | Dynamically switch models (e.g. `/model gemini-3.5-flash`, `/model gpt-4o`) |
| `/web <query>` | Performs a zero-key live web search and returns markdown summaries |
| `/read <path>` | Reads and displays syntax-highlighted code from any project file |
| `/save [filename]` | Exports the entire chat session to a formatted `.md` file |
| `$ <cmd>` or `/exec <cmd>` | Executes a local shell command and embeds stdout/stderr in chat |
| `:q`, `/quit`, `exit` | Saves conversation to history and cleanly exits |

---

## 🐶 Pet Companion Gamification

Sparky isn't just decoration — he is an interactive developer companion with his own RPG progression:

```
  /^ ^\      COMPANION: Sparky [Lv. 3 Senior Dev Doggo]
 / • • \     Status: listening
 V\ Y /V     Affection: 100%  |  XP: 30/60
  / - \  ~
 (____)m
```

- **XP Progression**: Sparky earns XP whenever you write code, ask questions, or run terminal tasks.
- **Affection System**: Keep your companion happy by petting him with `/pet` or solving bugs together.
- **Titles**:
  - `Level 1` — **Code Puppy**
  - `Level 2` — **Terminal Scout**
  - `Level 3` — **Senior Dev Doggo**
  - `Level 4` — **Cyber Sentinel**
  - `Level 5+` — **Kernel Archon**
- **Dynamic States**:
  - `Idle` — Blinks and scans the environment.
  - `Thinking` — Sniffs and gears turn while your model streams code.
  - `Happy` — Tail wags excitedly when solutions finish.
  - `Sleeping` — Curls up and snores (`zzZ...`) when you step away from the keyboard.

---

## 🧠 Supported AI Providers

Kernel supports virtually all modern LLM providers:

```bash
# Google Gemini (Default - Ultra fast ~1.5s latency)
./kernel -provider gemini -model gemini-3.5-flash

# Anthropic Claude
./kernel -provider claude -model claude-3-5-sonnet-latest

# OpenAI
./kernel -provider openai -model gpt-4o

# Local Ollama (Free & 100% Offline)
./kernel -provider ollama -model deepseek-r1:8b

# Custom OpenAI-Compatible (DeepSeek, OpenRouter, Groq, Together AI)
./kernel -provider custom -model deepseek-chat
```

---

## ⌨️ Keybindings

| Key | Action |
|---|---|
| `Enter` | Submit prompt / execute slash command |
| `Ctrl+C` | Cancel active streaming generation without exiting |
| `Esc` | Cancel stream / return to prompt |
| `:q` | Clean quit and archive session |

---

## 🛠️ Project Structure

```text
kernel/
├── main.go                # Application entrypoint & crash guard
├── rsrc.syso              # Embedded Windows application icon resource
├── internal/
│   ├── ui/
│   │   ├── model.go       # Bubble Tea state machine & interactive loop
│   │   ├── styles.go      # Lip Gloss neon theme & typography tokens
│   │   └── render.go      # Glamour terminal markdown syntax renderer
│   ├── pet/
│   │   └── pet.go         # Animated ASCII dog, states, leveling & XP
│   ├── ai/
│   │   └── client.go      # Multi-turn streaming engine (Gemini, Claude, OpenAI)
│   ├── tools/
│   │   ├── web.go         # Zero-key live DuckDuckGo web search
│   │   └── file.go        # Workspace file inspection & unified diffs
│   ├── exec/
│   │   └── exec.go        # Native PowerShell/Bash subprocess execution
│   ├── history/
│   │   └── history.go     # Session persistence & markdown exporter
│   └── config/
│       └── config.go      # JSON persistent configuration loader
└── .github/workflows/
    └── ci.yml             # Automated cross-compilation & release pipeline
```

---

## 🤝 Contributing

We welcome contributions, issues, and feature suggestions!
Feel free to open an issue or submit a pull request on the [GitHub Repository](https://github.com/404kernel/kernel).

---

## 📄 License

Distributed under the **MIT License**. Built with care by [404kernel](https://github.com/404kernel).
