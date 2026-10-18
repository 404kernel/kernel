# KERNEL ⚡

> **High-performance, keyboard-driven terminal AI agent built with Go and Bubble Tea.**

[![Go Report Card](https://goreportcard.com/badge/github.com/404kernel/kernel)](https://goreportcard.com/report/github.com/404kernel/kernel)
[![CI & Release](https://github.com/404kernel/kernel/actions/workflows/ci.yml/badge.svg)](https://github.com/404kernel/kernel/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/go-1.27%2B-cyan.svg)](https://golang.org)

```
╭────────────────────────────────────────────────────────────────────────────╮
│ ⚡ KERNEL   [GEMINI :: gemini-2.5-flash]   cwd: /workspace/project         │
╰────────────────────────────────────────────────────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────────╮
│ [ USER ] 15:04                                                             │
│                                                                            │
│ Write a high-concurrency worker pool in Go with channel backpressure       │
╰────────────────────────────────────────────────────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────────╮
│ [ KERNEL ] ⠋ 15:04                                                         │
│                                                                            │
│ Here is a production-grade worker pool utilizing bounded buffer channels: │
│                                                                            │
│ func NewWorkerPool(workers, queueSize int) *Pool { ... }                   │
╰────────────────────────────────────────────────────────────────────────────╯

╭────────────────────────────────────────────────────────────────────────────╮
│ > Ask kernel to explain code, generate commands, or solve problems...      │
╰────────────────────────────────────────────────────────────────────────────╯
[Enter] send · [Esc] cancel/exit · [q] quit
```

`kernel` is a fast, standalone terminal workspace orchestrator designed for developers who live in their CLI. It features a reactive Terminal User Interface (TUI) powered by **Charmbracelet's Bubble Tea** and **Lip Gloss**, real-time token streaming, and multi-provider AI model orchestration.

---

## ✨ Features

- 🏎️ **Single Compiled Binary**: Zero node_modules, zero external runtimes. Single fast `.exe` / ELF binary.
- 🎨 **Modern Cyber TUI**: Rounded neon box models, reactive viewport scrolling, streaming token indicators, and syntax-aware formatting.
- 🧠 **Multi-Provider AI Engine**:
  - **Google Gemini** (`gemini-2.5-flash`, `gemini-2.5-pro`)
  - **Anthropic Claude** (`claude-3-5-sonnet`, `claude-3-5-haiku`)
  - **OpenAI** (`gpt-4o`, `o3-mini`)
  - **Local Ollama** (`llama3`, `deepseek-r1`, `qwen2.5-coder`)
  - **Built-in Interactive Demo Mode** (zero setup required to run)
- ⚡ **Cross-Platform**: Compiles natively for Linux, macOS (Apple Silicon & Intel), and Windows.

---

## 🚀 Quick Start

### 1. Prebuilt Binaries

Download the latest release for your OS from the [Releases](https://github.com/404kernel/kernel/releases) page.

### 2. From Source

Ensure you have **Go 1.22+** installed:

```bash
# Clone the repository
git clone https://github.com/404kernel/kernel.git
cd kernel

# Run immediately
go run main.go

# Or compile your own standalone binary
go build -o kernel main.go
```

---

## 🎮 Usage

### Interactive Demo Mode (No API keys needed)
```bash
./kernel
```

### With Google Gemini
```bash
export GEMINI_API_KEY="your-gemini-key"
./kernel -provider gemini -model gemini-2.5-flash
```

### With Anthropic Claude
```bash
export ANTHROPIC_API_KEY="your-anthropic-key"
./kernel -provider claude -model claude-3-5-sonnet-latest
```

### With Local Ollama
```bash
./kernel -provider ollama -model llama3
```

---

## ⌨️ Keybindings

| Key | Action |
|---|---|
| `Enter` | Submit prompt / dispatch task |
| `Ctrl+C` / `Esc` | Cancel current stream / Exit |
| `q` | Exit application |

---

## 🛠️ Architecture

```
kernel/
├── main.go                # Application CLI entrypoint & flag parsing
├── internal/
│   ├── ui/
│   │   ├── model.go       # Bubble Tea state machine & Elm loop
│   │   └── styles.go      # Lip Gloss color tokens & box geometries
│   └── ai/
│       └── client.go      # Multi-provider streaming HTTP engine
└── .github/workflows/
    └── ci.yml             # Automated cross-compilation pipeline
```

---

## 🤝 Contributing

Contributions, issues, and feature requests are welcome!
Feel free to check the [issues page](https://github.com/404kernel/kernel/issues).

---

## 📄 License

Distributed under the **MIT License**. See `LICENSE` for more information.
