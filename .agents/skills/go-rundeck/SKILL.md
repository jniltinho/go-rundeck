---
name: Go-Rundeck Development
description: Guidelines, tech stack, and architectural rules for developing the Go-Rundeck project.
---

# Go-Rundeck Development Skill

This skill contains all the core rules, architectural guidelines, and styling instructions for the **Go-Rundeck** project, as defined in its Product Requirements Document (`DOCUMENTS/PRD-Go-Rundeck.md`).

## 1. Core Tech Stack
- **Backend**: Go (1.26), Echo Framework v5, GORM, MariaDB 10.x+, Cobra CLI, Viper.
- **Frontend**: Tailwind CSS 4.2.0 (standalone CLI, no Node/npm), jQuery 4.0.0, jQuery Toast Plugin, Go `html/template`.
- **Key Feature**: Zero dependencies. All web assets (`web/templates`, `web/static`) are embedded into the Go binary via `embed.go`.

## 2. Architectural Rules
- **Entry Points**: 
  - `main.go` **MUST be at the project root** (no `cmd/main.go`). It calls Cobra's `cmd.Execute()`.
  - `embed.go` **MUST be at the project root** and must declare `TemplatesFS` and `StaticFS` embeds.
- **Project Structure**:
  - `cmd/`: Cobra subcommands (`serve`, `migrate`, etc.).
  - `internal/handler/`: HTTP Echo handlers. They only orchestrate request/response and return errors using `echo.NewHTTPError()`.
  - `internal/service/`: Business logic goes here.
  - `internal/model/`: GORM models.
  - `internal/repository/`: Data access layer.
  - `web/`: All embedded frontend assets.
- **No external assets** should be referenced in production (no external CDNs). Everything must be served from `web/static/`.

## 3. UI/UX Design System (Neo-Brutalist)
- **Borders**: **Square** (`rounded-none` by default). ZERO rounding on any element (cards, buttons, inputs, modals).
- **Shadows**: Thin, solid, and directional (e.g., `2px 2px 0 var(--color-border)` or `4px 4px 0 var(--color-border)`). NO shadow blur.
- **Colors**: High contrast. Background off-white (`#F5F0E8`) or dark (`#0C0C0C`), with solid accents (`#FF5C00`, `#FFD600`, etc.).
- **Tailwind 4.2.0**: The project uses the CSS-first format (`@theme` inside `web/static/css/input.css`), WITHOUT `tailwind.config.js`.

## 4. Build Process
- Compiling CSS **must** precede the Go build.
  1. `tailwindcss -i web/static/css/input.css -o web/static/css/app.css --minify`
  2. `go build -ldflags "..." -o bin/gorundeck .`
- Or use the `make all` command which handles this order. 
- During development, `make dev` will use `air` and run `tailwindcss --watch` in parallel.

## 5. SSH Execution Flow
- Job execution runs `SSHService.Connect` using private keys or passwords from the KeyStorage.
- Commands/Scripts are run via the SSH Session.
- Logs are emitted to a Go channel and streamed to the frontend in real-time using **Server-Sent Events (SSE)** via jQuery in `execution.js`.

## 6. Security
- Handle sessions with secure cookies. (Optional JWT for API).
- Passwords must be hashed with `bcrypt` (cost >= 12).
- SSH keys are encrypted at rest with AES-256-GCM.
- Forms that modify state (POST/PUT/DELETE) use Echo CSRF middleware.
- GORM handles prepared statements to avoid SQL Injection.

When modifying code in the Go-Rundeck project, adhere strictly to these rules, especially regarding the project layout and the Neo-Brutalist design aesthetic.
