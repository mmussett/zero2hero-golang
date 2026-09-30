---
name: project-zero2hero-golang
description: zero2hero-golang is a 34-day self-paced Go curriculum modelled on the zero2hero-rust repo. Created from scratch with 34 day modules and 6 capstone skeletons.
metadata:
  type: project
---

This repository is a complete 34-day Go learning curriculum at C:/Users/mmussett/src/zero2hero-golang.

**Why:** User wanted a Go equivalent of their Rust curriculum (github.com/mmussett/zero2hero-rust), structured the same way — 30 core days + bonus deep-dive days.

**Structure:**
- go.work workspace at root, each day-XX is an independent Go module
- Days 01–30: core curriculum (4 weeks)
- Days 31–34: bonus deep-dive days added mid-creation at user request
- capstone/: 6 skeleton projects (grep, passgen, todo-api, url-shortener, chat-server, kvdb)
- Reference docs: PRIMITIVES.md (Go types), DATA_STRUCTURES.md (collections), CURRICULUM.md

**Bonus days added (days 31–34):**
- Day 31: Go interfaces deep dive
- Day 32: Idiomatic Go project structure
- Day 33: Go channel model
- Day 34: Goroutines and sync (Mutex, WaitGroup, Once, atomic, errgroup)

**How to apply:** If user asks to add more content to this repo, follow the same pattern (day-XX/ with go.mod, main.go, README.md). Update go.work and CURRICULUM.md when adding new days. Update CLAUDE.md module list.
