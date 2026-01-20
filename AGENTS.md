# AGENTS.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Name: Simple System Monitor

Description: Lightweight Go CLI that collects CPU/memory/disk metrics and sends Telegram alerts/summary messages.

- **NOTE:** `.env` is loaded automatically when present.
- **NOTE:** Telegram schedules use standard cron format in UTC; empty or invalid schedules disable periodic Telegram metrics.
- **NOTE:** Thresholds are percentages; alert windows are durations.

## Global Rules

- **NEVER** use emojis!
- **NEVER** try to run the dev server unless the user explicitly tells you that you are the user and instructs you to run it.
- **NEVER** try to build in the project directory; always build in the `/tmp` directory unless the user explicitly tells you that you are the user and instructs you to build in the project directory.
- **NEVER** use comments in code - code should be self-explanatory
- **NEVER** cut corners, don't leave comments like `TODO: Implement X in the future here`! Always fully implement everything!
- **NEVER** revert/delete any changes that you don't know about! Always assume that we are in the middle of a task and that the changes are intentional!
- **ALWAYS** at the end of your turn, ask a follow-up question for the next logical step (**DON'T** ask questions like "Should I run tests?" or "Should I lint?", only ask questions that are relevant to the task at hand)
## Refactor Using Established Engineering Principles

After generating or editing code, you must always refactor your changes using well-established software engineering principles. These apply every time, without relying on diff inspection.

### Core Principles

- **DRY (Don't Repeat Yourself)**: Eliminate duplicate or repetitive logic by consolidating shared behavior into common functions or helpers.  
- **KISS (Keep It Simple, Stupid)**: Prefer simple, straightforward solutions over unnecessarily complex or abstract designs.  
- **YAGNI (You Aren't Gonna Need It)**: Only implement what is required for the current task; avoid speculative features or abstractions.

### Refactoring Requirements

1. Ensure the intent of your change is clear, explicit, and easy to understand.  
2. Maintain consistency with existing patterns, naming, and structure in the codebase.  
3. Remove duplication and merge similar logic following DRY.  
4. Simplify complex code paths or structures following KISS.  
5. Avoid adding features, hooks, or abstractions that the current task does not need, following YAGNI.  

**Principle:**  
> Every change must simplify the codebase, reduce duplication, clarify intent, and make the system easier to maintain.

## High-Level Architecture

- **Entrypoint:** `cmd/simple-system-monitor/main.go` orchestrates config loading, scheduling, and alerting.
- **Config Source of Truth:** `internal/config/config.go` defines env/flag defaults and parsing behavior.
- **Metrics Collection:** `internal/monitor` pulls CPU/memory/disk data via `gopsutil` and filters mounts/fstypes.
- **Alerting:** `internal/alerts` tracks threshold windows and prevents repeated alerts without recovery.
- **Notifications:** `internal/telegram` posts HTML-formatted messages to the Telegram Bot API.
- **Scheduling:** `robfig/cron/v3` schedules periodic Telegram summaries (tracked with next-run time).

## Project Guidelines

### simple-system-monitor (root)

- Language: Go
- Framework/Runtime: CLI application (Go stdlib)
- Package Manager: Go modules (`go.mod`/`go.sum`)
- Important Packages: `github.com/shirou/gopsutil/v4`, `github.com/robfig/cron/v3`, `github.com/joho/godotenv`, `go.uber.org/zap`
- Checks:
  - Format: `gofmt -l .`
  - Static analysis: `go vet ./...`
  - **ALWAYS** run these after you are done making changes
- Rules / conventions:
  - **ALWAYS** run the checks listed above after changes
  - **ALWAYS** keep configuration defaults in `internal/config/config.go` and update `.env.example` and `README.md` when changing config keys or defaults
  - **ALWAYS** HTML-escape Telegram content using the existing helpers before sending messages
- Useful files:
  - `.env.example`
  - `README.md`
  - `cmd/simple-system-monitor/main.go`
  - `internal/config/config.go`
  - `internal/monitor/metrics.go`
  - `internal/alerts/alerts.go`
  - `internal/telegram/client.go`

## Key Architectural Patterns

- **Configuration:** Environment variables and flags are parsed together; defaults live in `internal/config/config.go`.
- **Alert Windows:** Alerts only trigger after remaining above thresholds for the configured window; state is preserved across ticks.
- **Metrics Formatting:** Telegram output uses HTML formatting with explicit escaping to avoid malformed payloads.
- **Disk Filtering:** Mount include/exclude rules support `*` suffix matches, plus fstype exclusion lists.
