# PRD — Go-Rundeck
**Product Requirements Document**
**Version:** 1.1.0
**Date:** 2026-03-17
**Status:** Draft — Rev 2 (square borders, binary embed, root main.go, Tailwind 4.2.0)

---

## 1. Overview

### 1.1 Executive Summary

**Go-Rundeck** is a runbook automation and task orchestration web platform inspired by Rundeck, developed natively in **Go 1.26**. The goal is to offer a lightweight, performant, and modern alternative to the original Rundeck (Java), keeping the layout and workflow familiarity, but with a leaner and more efficient tech stack.

### 1.2 Motivation

- The original Rundeck consumes a lot of memory and requires a JVM, making deployment heavy.
- Go offers a single binary, low footprint, high concurrency, and simplified deployment.
- The interface will be modernized with a **Neo-Brutalist** aesthetic — square borders (`rounded-none`), thin directional shadows, strong typography, and marked contrasts.
- Initial focus on automation via **SSH** on Linux/Unix servers.

### 1.3 References

| Resource | URL |
|---|---|
| Rundeck OSS | https://github.com/rundeck/rundeck |
| go-task/task | https://github.com/go-task/task |
| Rundeck Docs | https://docs.rundeck.com/docs/ |

---

## 2. Tech Stack

### 2.1 Backend

| Component | Version | Purpose |
|---|---|---|
| **Go** | 1.26 | Main language |
| **Echo Framework** | v5 | HTTP server, routing, middleware |
| **GORM** | latest | ORM for persistence |
| **MySQL** | 8.x | Relational database |
| **Cobra CLI** | latest | CLI for administration and initialization |
| **Viper / go-toml** | latest | Configuration via `.toml` file |
| **golang.org/x/crypto/ssh** | latest | SSH connection with remote servers |

### 2.2 Frontend

| Component | Version | Purpose |
|---|---|---|
| **Tailwind CSS** | 4.2.0 | Utility-first styling |
| **jQuery** | 4.0.0 | DOM manipulation, AJAX, interactions |
| **jQuery Toast Plugin** | latest | User notifications/messages |
| **html/template** | Go stdlib | Server-side rendering |

### 2.3 Design System

- **Style:** Neo-Brutalist with industrial/terminal UI influences.
- **Borders:** **Square** (`rounded-none` by default) — zero rounding on cards, buttons, inputs, and modals. This rigidity is intentional and a central part of the visual identity.
- **Shadows:** Thin and directional — fixed offset `2px 2px 0` or `4px 4px 0` with a solid color (no blur), creating a geometric depth effect characteristic of Neo-Brutalism.
- **Typography:** Strong, condensed display font (e.g., `DM Mono`, `IBM Plex Mono` or `JetBrains Mono`) for titles and labels; clean sans-serif for body text.
- **Colors:** High-contrast palette — light background (`#F5F0E8`) or dark (`#0C0C0C`) with strong, saturated accents (orange `#FF5C00`, yellow `#FFD600`, green `#00FF87` or blue `#0066FF`).
- **Animations:** Geometric micro-interactions on hover (slight translate), state transitions, visual feedback on actions.

### 2.4 Embedded Assets in Binary

All static files (CSS, JS) and HTML templates are **embedded directly in the Go binary** via `//go:embed`. The final binary is completely self-sufficient — no external files are required in production.

| Embedded Directory | Content |
|---|---|
| `web/templates/` | All `*.html` files (Go templates) |
| `web/static/css/` | `app.css` (compiled Tailwind 4.2.0 + custom styles) |
| `web/static/js/` | `jquery-4.0.0.min.js`, `jquery.toast.min.js`, `app.js`, `execution.js` |
| `web/static/img/` | Icons, logo, favicons |

---

## 3. Project Architecture

### 3.1 Directory Structure

> **Layout Rules:**
> - `main.go` stays in the **project root** (no `cmd/main.go` subdirectory).
> - Cobra subcommands go in `cmd/`.
> - Web assets (`templates/`, `static/`) go in `web/` at the root and are **embedded in the binary**.
> - The central embed file `embed.go` is placed at the root, declaring the `//go:embed`.

```
go-rundeck/
├── main.go                  # ← Main entrypoint (Cobra root + serve)
├── embed.go                 # ← //go:embed declarations for templates and static
├── config.toml              # Main configuration (NOT embedded — read at runtime)
├── go.mod
├── go.sum
├── Makefile
├── .air.toml                # air config (live reload in dev)
├── README.md
│
├── cmd/                     # Cobra subcommands
│   ├── serve.go             # gorundeck serve
│   ├── migrate.go           # gorundeck migrate
│   ├── user.go              # gorundeck user create
│   ├── version.go           # gorundeck version
│   └── config_check.go      # gorundeck config check
│
├── config/
│   └── config.go            # Reading and parsing config.toml (go-toml)
│
├── internal/
│   ├── handler/             # Echo handlers (HTTP)
│   │   ├── auth.go
│   │   ├── project.go
│   │   ├── node.go
│   │   ├── job.go
│   │   ├── execution.go
│   │   ├── schedule.go
│   │   └── activity.go
│   │
│   ├── model/               # GORM models
│   │   ├── user.go
│   │   ├── project.go
│   │   ├── node.go
│   │   ├── job.go
│   │   ├── job_step.go
│   │   ├── execution.go
│   │   ├── execution_log.go
│   │   ├── schedule.go
│   │   └── key_storage.go
│   │
│   ├── repository/          # Data access layer
│   │   ├── project_repo.go
│   │   ├── node_repo.go
│   │   ├── job_repo.go
│   │   └── execution_repo.go
│   │
│   ├── service/             # Business logic
│   │   ├── project_service.go
│   │   ├── job_service.go
│   │   ├── execution_service.go
│   │   ├── ssh_service.go
│   │   └── schedule_service.go
│   │
│   ├── middleware/
│   │   ├── auth.go
│   │   ├── logger.go
│   │   └── cors.go
│   │
│   └── router/
│       └── router.go
│
└── web/                     # ← EVERYTHING HERE IS EMBEDDED in the binary
    ├── templates/           # Go html/template (embedded via embed.go)
    │   ├── layout/
    │   │   ├── base.html
    │   │   ├── sidebar.html
    │   │   └── topbar.html
    │   ├── auth/
    │   │   └── login.html
    │   ├── dashboard/
    │   │   └── index.html
    │   ├── projects/
    │   │   ├── list.html
    │   │   └── detail.html
    │   ├── nodes/
    │   │   ├── list.html
    │   │   └── detail.html
    │   ├── jobs/
    │   │   ├── list.html
    │   │   ├── create.html
    │   │   └── detail.html
    │   ├── executions/
    │   │   ├── list.html
    │   │   └── detail.html
    │   └── settings/
    │       └── index.html
    │
    └── static/              # Static assets (embedded via embed.go)
        ├── css/
        │   ├── input.css    # Tailwind 4.2.0 entry font (@import "tailwindcss")
        │   └── app.css      # Compiled CSS — build generated, embedded
        ├── js/
        │   ├── jquery-4.0.0.min.js
        │   ├── jquery.toast.min.js
        │   ├── app.js       # Global init + jQuery helpers
        │   └── execution.js # Live log SSE streaming
        └── img/
            ├── logo.svg
            └── favicon.ico
```

### 3.2 `embed.go` File (project root)

```go
package main

import "embed"

// TemplatesFS contains all HTML templates embedded in the binary.
//
//go:embed web/templates
var TemplatesFS embed.FS

// StaticFS contains all static assets embedded in the binary.
//
//go:embed web/static
var StaticFS embed.FS
```

### 3.3 `main.go` File (project root)

```go
package main

import (
    "go-rundeck/cmd"
)

// Build info injected via ldflags at build time
var (
    Version   = "dev"
    BuildTime = "unknown"
    GitCommit = "unknown"
)

func main() {
    cmd.Execute(Version, BuildTime, GitCommit)
}
```

### 3.4 Usage of Embeds in the Router

```go
// internal/router/router.go
package router

import (
    "html/template"
    "io/fs"
    "net/http"

    "github.com/labstack/echo/v5"
    main "go-rundeck" // accesses TemplatesFS and StaticFS
)

func Setup(db *gorm.DB, cfg *config.Config) *echo.Echo {
    e := echo.New()

    // Embedded templates
    tmplFS, _ := fs.Sub(main.TemplatesFS, "web/templates")
    tmpl := template.Must(template.ParseFS(tmplFS, "**/*.html", "layout/*.html"))
    e.Renderer = &TemplateRenderer{templates: tmpl}

    // Embedded static files — served at /static/
    staticFS, _ := fs.Sub(main.StaticFS, "web/static")
    e.StaticFS("/static", http.FS(staticFS))

    // ... register handlers
    return e
}
```

### 3.5 Tailwind CSS 4.2.0 — Configuration

Tailwind 4.2.0 uses the new CSS-first format (no `tailwind.config.js`):

```css
/* web/static/css/input.css */
@import "tailwindcss";

/* Neo-Brutalist Theme — square borders, directional shadows */
@theme {
  --color-accent:    #FF5C00;
  --color-accent-2:  #FFD600;
  --color-surface:   #F5F0E8;
  --color-surface-dark: #0C0C0C;
  --color-border:    #0C0C0C;

  --font-display: "IBM Plex Mono", "JetBrains Mono", monospace;
  --font-body:    "Inter", "DM Sans", sans-serif;

  --shadow-brutal-sm: 2px 2px 0 var(--color-border);
  --shadow-brutal-md: 4px 4px 0 var(--color-border);
  --shadow-brutal-lg: 6px 6px 0 var(--color-border);

  --radius: 0;           /* global rounded-none */
  --radius-sm: 0;
  --radius-md: 0;
  --radius-lg: 0;
  --radius-xl: 0;
}

/* Custom Utilities */
@utility shadow-brutal {
  box-shadow: var(--shadow-brutal-md);
}
@utility shadow-brutal-sm {
  box-shadow: var(--shadow-brutal-sm);
}
@utility shadow-brutal-lg {
  box-shadow: var(--shadow-brutal-lg);
}
@utility border-brutal {
  border: 2px solid var(--color-border);
}
```

Compilation via Tailwind 4.2.0 CLI (standalone, without Node/npm):

```bash
# Install standalone CLI (without npm)
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v4.2.0/tailwindcss-linux-x64
chmod +x tailwindcss-linux-x64
mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Compile
tailwindcss -i web/static/css/input.css -o web/static/css/app.css --minify
```

### 3.6 `go.mod`

```
module go-rundeck

go 1.26
```

All internal imports use the `go-rundeck/...` prefix:

```go
import (
    "go-rundeck/config"
    "go-rundeck/internal/handler"
    "go-rundeck/internal/model"
    "go-rundeck/internal/service"
    "go-rundeck/internal/router"
)
```

### 3.7 Configuration — `config.toml`

```toml
[app]
name    = "Go-Rundeck"
env     = "development"       # development | production
port    = 8080
secret  = "CHANGE_ME_SECRET"
debug   = true

[database]
host     = "localhost"
port     = 3306
user     = "rundeck"
password = "rundeck_pass"
name     = "gorundeck"
charset  = "utf8mb4"

[ssh]
default_user       = "deploy"
default_port       = 22
connect_timeout    = 10      # seconds
key_storage_path   = "./keys"

[scheduler]
enabled            = true
check_interval     = 30      # seconds

[log]
level  = "info"
format = "json"
output = "stdout"
```

---

## 4. Features — MVP (v1.0)

### 4.1 Authentication and Users

| ID | Feature | Priority |
|---|---|---|
| AUTH-01 | Login with user/password | 🔴 High |
| AUTH-02 | Logout and session invalidation | 🔴 High |
| AUTH-03 | JWT with refresh token | 🟡 Medium |
| AUTH-04 | User management (CRUD) | 🟡 Medium |
| AUTH-05 | Role control: Admin, Operator, Viewer | 🟡 Medium |

### 4.2 Projects

| ID | Feature | Priority |
|---|---|---|
| PROJ-01 | List projects | 🔴 High |
| PROJ-02 | Create project (name, description, tags) | 🔴 High |
| PROJ-03 | Edit project | 🔴 High |
| PROJ-04 | Archive/delete project | 🟡 Medium |
| PROJ-05 | Project dashboard (jobs, nodes, executions) | 🔴 High |

### 4.3 Nodes (Servers)

| ID | Feature | Priority |
|---|---|---|
| NODE-01 | List nodes per project | 🔴 High |
| NODE-02 | Add node (host, IP, SSH port, user, auth) | 🔴 High |
| NODE-03 | Edit node | 🔴 High |
| NODE-04 | Remove node | 🔴 High |
| NODE-05 | Tags/labels on nodes for job selection | 🔴 High |
| NODE-06 | SSH connectivity test (ping) | 🔴 High |
| NODE-07 | Import nodes via JSON/YAML | 🟡 Medium |

### 4.4 Key Storage (SSH Keys Management)

| ID | Feature | Priority |
|---|---|---|
| KEY-01 | Upload private SSH key | 🔴 High |
| KEY-02 | Store encrypted password/passphrase | 🔴 High |
| KEY-03 | Associate key to a node or project | 🔴 High |
| KEY-04 | List and remove keys | 🔴 High |

### 4.5 Jobs

| ID | Feature | Priority |
|---|---|---|
| JOB-01 | List jobs per project | 🔴 High |
| JOB-02 | Create job (name, description, steps, target nodes) | 🔴 High |
| JOB-03 | Edit job | 🔴 High |
| JOB-04 | Delete job | 🔴 High |
| JOB-05 | Manually execute job (Run Now) | 🔴 High |
| JOB-06 | Job steps: Command (shell), Inline script | 🔴 High |
| JOB-07 | Target nodes selection by tag or name | 🔴 High |
| JOB-08 | Input parameters (options/arguments) | 🟡 Medium |
| JOB-09 | Execution strategy: sequential / parallel | 🟡 Medium |
| JOB-10 | Failure control: stop on error / continue | 🟡 Medium |
| JOB-11 | Email notification on completion/failure | 🟢 Low |

### 4.6 Executions

| ID | Feature | Priority |
|---|---|---|
| EXEC-01 | Execution history per job/project | 🔴 High |
| EXEC-02 | Execution detail with real-time log (SSE) | 🔴 High |
| EXEC-03 | Status: Running, Succeeded, Failed, Aborted | 🔴 High |
| EXEC-04 | Abort running execution | 🟡 Medium |
| EXEC-05 | Re-execute (retry) an execution | 🟡 Medium |
| EXEC-06 | Filter and search in history | 🟡 Medium |

### 4.7 Schedules

| ID | Feature | Priority |
|---|---|---|
| SCHED-01 | Associate cron schedule to a job | 🟡 Medium |
| SCHED-02 | Enable/disable schedule | 🟡 Medium |
| SCHED-03 | List upcoming scheduled executions | 🟡 Medium |

### 4.8 Activity (Global Dashboard)

| ID | Feature | Priority |
|---|---|---|
| ACT-01 | Recent executions feed (all projects) | 🔴 High |
| ACT-02 | Filters by project, job, status, period | 🟡 Medium |

---

## 5. Data Models

### 5.1 `users`

```sql
id            BIGINT PK AUTO_INCREMENT
username      VARCHAR(100) UNIQUE NOT NULL
password_hash VARCHAR(255) NOT NULL
email         VARCHAR(255)
role          ENUM('admin','operator','viewer') DEFAULT 'operator'
active        BOOLEAN DEFAULT TRUE
created_at    DATETIME
updated_at    DATETIME
deleted_at    DATETIME NULL
```

### 5.2 `projects`

```sql
id          BIGINT PK AUTO_INCREMENT
name        VARCHAR(100) UNIQUE NOT NULL
description TEXT
tags        VARCHAR(255)
active      BOOLEAN DEFAULT TRUE
created_by  BIGINT FK users.id
created_at  DATETIME
updated_at  DATETIME
deleted_at  DATETIME NULL
```

### 5.3 `nodes`

```sql
id           BIGINT PK AUTO_INCREMENT
project_id   BIGINT FK projects.id
name         VARCHAR(100) NOT NULL
hostname     VARCHAR(255) NOT NULL
ssh_port     INT DEFAULT 22
ssh_user     VARCHAR(100)
auth_type    ENUM('password','key') DEFAULT 'key'
key_id       BIGINT FK key_storage.id NULL
tags         VARCHAR(255)
description  TEXT
os_family    VARCHAR(50)
active       BOOLEAN DEFAULT TRUE
created_at   DATETIME
updated_at   DATETIME
deleted_at   DATETIME NULL
```

### 5.4 `key_storage`

```sql
id            BIGINT PK AUTO_INCREMENT
project_id    BIGINT FK projects.id NULL
name          VARCHAR(100) NOT NULL
key_type      ENUM('private_key','password')
content_enc   TEXT NOT NULL        -- encrypted content (AES-256)
description   TEXT
created_by    BIGINT FK users.id
created_at    DATETIME
updated_at    DATETIME
```

### 5.5 `jobs`

```sql
id             BIGINT PK AUTO_INCREMENT
project_id     BIGINT FK projects.id
name           VARCHAR(200) NOT NULL
description    TEXT
node_filter    VARCHAR(255)         -- e.g., tags=web,prod
exec_strategy  ENUM('sequential','parallel') DEFAULT 'sequential'
on_error       ENUM('stop','continue') DEFAULT 'stop'
timeout_sec    INT DEFAULT 0
created_by     BIGINT FK users.id
created_at     DATETIME
updated_at     DATETIME
deleted_at     DATETIME NULL
```

### 5.6 `job_steps`

```sql
id          BIGINT PK AUTO_INCREMENT
job_id      BIGINT FK jobs.id
step_order  INT NOT NULL
type        ENUM('command','script') NOT NULL
label       VARCHAR(200)
content     TEXT NOT NULL        -- inline script or command
interpreter VARCHAR(100)         -- e.g., /bin/bash
args        VARCHAR(500)
created_at  DATETIME
updated_at  DATETIME
```

### 5.7 `executions`

```sql
id           BIGINT PK AUTO_INCREMENT
job_id       BIGINT FK jobs.id
project_id   BIGINT FK projects.id
status       ENUM('running','succeeded','failed','aborted') DEFAULT 'running'
triggered_by BIGINT FK users.id NULL
trigger_type ENUM('manual','schedule')
started_at   DATETIME
ended_at     DATETIME NULL
duration_sec INT NULL
created_at   DATETIME
```

### 5.8 `execution_logs`

```sql
id           BIGINT PK AUTO_INCREMENT
execution_id BIGINT FK executions.id
node_name    VARCHAR(100)
step_order   INT
log_level    ENUM('info','warn','error','debug') DEFAULT 'info'
message      TEXT
logged_at    DATETIME
```

### 5.9 `schedules`

```sql
id         BIGINT PK AUTO_INCREMENT
job_id     BIGINT FK jobs.id
cron_expr  VARCHAR(100) NOT NULL
enabled    BOOLEAN DEFAULT TRUE
next_run   DATETIME
last_run   DATETIME NULL
created_at DATETIME
updated_at DATETIME
```

---

## 6. API / HTTP Routes

### 6.1 Authentication

| Method | Route | Description |
|---|---|---|
| GET | `/login` | Login page |
| POST | `/login` | Authenticate user |
| POST | `/logout` | End session |

### 6.2 Dashboard

| Method | Route | Description |
|---|---|---|
| GET | `/` | Global dashboard (activity feed) |

### 6.3 Projects

| Method | Route | Description |
|---|---|---|
| GET | `/projects` | List projects |
| GET | `/projects/new` | New project form |
| POST | `/projects` | Create project |
| GET | `/projects/:id` | Project dashboard |
| GET | `/projects/:id/edit` | Edit project |
| PUT | `/projects/:id` | Save edit |
| DELETE | `/projects/:id` | Delete project |

### 6.4 Nodes

| Method | Route | Description |
|---|---|---|
| GET | `/projects/:id/nodes` | List nodes |
| GET | `/projects/:id/nodes/new` | New node form |
| POST | `/projects/:id/nodes` | Create node |
| GET | `/projects/:id/nodes/:nid` | Node details |
| PUT | `/projects/:id/nodes/:nid` | Edit node |
| DELETE | `/projects/:id/nodes/:nid` | Remove node |
| POST | `/projects/:id/nodes/:nid/ping` | Test SSH |

### 6.5 Jobs

| Method | Route | Description |
|---|---|---|
| GET | `/projects/:id/jobs` | List jobs |
| GET | `/projects/:id/jobs/new` | New job form |
| POST | `/projects/:id/jobs` | Create job |
| GET | `/projects/:id/jobs/:jid` | Job details |
| GET | `/projects/:id/jobs/:jid/edit` | Edit job |
| PUT | `/projects/:id/jobs/:jid` | Save edit |
| DELETE | `/projects/:id/jobs/:jid` | Delete job |
| POST | `/projects/:id/jobs/:jid/run` | Execute job now |

### 6.6 Executions

| Method | Route | Description |
|---|---|---|
| GET | `/projects/:id/executions` | Execution history |
| GET | `/executions/:eid` | Execution details |
| GET | `/executions/:eid/log` | Log SSE stream (real-time) |
| POST | `/executions/:eid/abort` | Abort execution |

### 6.7 Key Storage

| Method | Route | Description |
|---|---|---|
| GET | `/keys` | List keys |
| POST | `/keys` | Upload key |
| DELETE | `/keys/:kid` | Remove key |

### 6.8 JSON API (`/api/v1` prefix)

All the above routes have REST JSON equivalents under `/api/v1/...` for external integrations and a future CLI `rd`.

---

## 7. Design and UI/UX

### 7.1 Applied Neo-Brutalist Philosophy

- **Dominant Typography:** Mono or condensed font for titles and labels (IBM Plex Mono / JetBrains Mono). Clean sans-serif font for body text (Inter / DM Sans).
- **Colors:** Off-white background `#F5F0E8` (light) or `#0C0C0C` (dark) with 1–2 saturated accent colors (orange `#FF5C00`, yellow `#FFD600`).
- **Borders:** **Square** — `border-2 border-black` (or `border-white` in dark mode). Zero rounding on any element. Tailwind Class: `rounded-none` is the global default defined in `@theme`.
- **Shadows:** `shadow-brutal-sm` (`2px 2px 0 #000`) or `shadow-brutal-md` (`4px 4px 0 #000`) — solid shadow without blur, characteristic of the Neo-Brutalist style. No diffused `box-shadow`.
- **Density:** Dense interface similar to Rundeck, with intentional breathing room through consistent padding (`p-4`, `p-6`).
- **Feedback:** Toasts positioned in the top-right corner via jQuery Toast Plugin, with solid borders and without rounding.

### 7.2 Main Layout

```
┌─────────────────────────────────────────────────────────────┐
│  TOPBAR: Logo | Global Search | User Menu | Notifications   │
├──────────────┬──────────────────────────────────────────────┤
│              │                                               │
│   SIDEBAR    │           CONTENT AREA                        │
│              │                                               │
│  • Dashboard │  ┌─────────────────────────────────────────┐ │
│  • Projects  │  │  Page Header (breadcrumb + actions)     │ │
│  • Nodes     │  ├─────────────────────────────────────────┤ │
│  • Jobs      │  │                                         │ │
│  • Activity  │  │  Main Content (tables, forms, logs)     │ │
│  • Keys      │  │                                         │ │
│  • Settings  │  └─────────────────────────────────────────┘ │
│              │                                               │
└──────────────┴──────────────────────────────────────────────┘
```

### 7.3 Main Screens

| Screen | Description |
|---|---|
| **Login** | Full-screen layout with central logo, username/password fields, login button. Striking aesthetic. |
| **Global Dashboard** | Stat cards (jobs today, executions, failures), recent activity feed per project. |
| **Projects List** | Grid of project cards with status, jobs/nodes counters. |
| **Project Dashboard** | Project KPIs + latest executions + active jobs. |
| **Nodes List** | Table with SSH connectivity status, tags, quick actions. |
| **Jobs List** | Table with name, last status, next scheduled run, "Run" button. |
| **Create/Edit Job** | Multi-step form: Basic Info → Steps → Nodes → Schedule → Notifications. |
| **Execution Detail** | Real-time log via SSE, node status, progress bar, abort button. |
| **Key Storage** | Key list with type, associated project, management actions. |

### 7.4 jQuery Components

```javascript
// Success toast example
$.toast({
    heading: 'Job started',
    text: 'Execution #1234 in progress...',
    position: 'top-right',
    loaderBg: '#F97316',
    icon: 'success',
    hideAfter: 4000
});

// Error toast example
$.toast({
    heading: 'SSH connection failed',
    text: 'Could not connect to the web-01 node',
    position: 'top-right',
    loaderBg: '#EF4444',
    icon: 'error',
    hideAfter: 6000
});
```

---

## 8. SSH Module

### 8.1 SSH Execution Flow

```
JobService.Run(jobID)
    │
    ├─► Load Job + Steps + NodeFilter
    │
    ├─► Resolve target Nodes (by tag or name)
    │
    ├─► For each Node:
    │       │
    │       ├─► SSHService.Connect(node)
    │       │       └─► Use private key OR password from KeyStorage
    │       │
    │       ├─► For each Step (sequential or parallel):
    │       │       ├─► Run command/script via SSH Session
    │       │       ├─► Capture stdout/stderr in real-time
    │       │       └─► Save ExecutionLog to the database
    │       │
    │       └─► Close SSH session
    │
    └─► Update Execution status (succeeded/failed)
```

### 8.2 SSHService Interface (Go)

```go
type SSHService interface {
    Connect(node *model.Node, key *model.KeyStorage) (*SSHClient, error)
    RunCommand(client *SSHClient, cmd string, logChan chan<- LogEntry) error
    RunScript(client *SSHClient, script string, interpreter string, logChan chan<- LogEntry) error
    TestConnectivity(node *model.Node, key *model.KeyStorage) error
    Close(client *SSHClient) error
}

type LogEntry struct {
    NodeName  string
    StepOrder int
    Level     string
    Message   string
    Timestamp time.Time
}
```

### 8.3 Real-Time Log (SSE)

- The `/executions/:eid/log` handler opens a **Server-Sent Events** connection.
- `ExecutionService` publishes log entries to a Go channel.
- The handler consumes the channel and sends SSE events to the browser.
- The frontend (`execution.js`) consumes the EventSource and updates the log terminal via jQuery.

```javascript
// execution.js
const evtSource = new EventSource(`/executions/${execId}/log`);
evtSource.onmessage = function(event) {
    const entry = JSON.parse(event.data);
    $('#log-terminal').append(
        `<div class="log-line log-${entry.level}">
            <span class="log-time">${entry.timestamp}</span>
            <span class="log-node">[${entry.node}]</span>
            <span class="log-msg">${entry.message}</span>
        </div>`
    );
    $('#log-terminal').scrollTop($('#log-terminal')[0].scrollHeight);
};
evtSource.addEventListener('done', () => evtSource.close());
```

---

## 9. CLI — Cobra

### 9.1 Commands

```
gorundeck                      # Shows help
gorundeck serve                # Starts web server (reads config.toml)
gorundeck serve --port 9090    # Port override
gorundeck migrate              # Runs GORM migrations
gorundeck migrate --rollback   # Reverts latest migration
gorundeck user create          # Creates admin user interactively
gorundeck version              # Shows version and build info
gorundeck config check         # Validates config.toml and tests DB
```

### 9.2 `cmd/serve.go` Example

```go
package cmd

import (
    "fmt"

    "github.com/spf13/cobra"
    "go-rundeck/config"
    "go-rundeck/internal/database"
    "go-rundeck/internal/router"
)

var serveCmd = &cobra.Command{
    Use:   "serve",
    Short: "Starts the Go-Rundeck web server",
    RunE: func(cmd *cobra.Command, args []string) error {
        cfg := config.Load()
        db := database.Connect(cfg.Database)
        // TemplatesFS and StaticFS are passed from main (package main, root)
        e := router.Setup(db, cfg)
        return e.Start(fmt.Sprintf(":%d", cfg.App.Port))
    },
}

func init() {
    serveCmd.Flags().Int("port", 0, "Port override (overrides config.toml)")
    rootCmd.AddCommand(serveCmd)
}
```

---

## 10. Security

| Aspect | Implementation |
|---|---|
| Authentication | Secure cookie session + optional JWT for API |
| Passwords | bcrypt with cost ≥ 12 |
| SSH Keys | Encrypted at rest with AES-256-GCM; key derived from `app.secret` via PBKDF2 |
| CSRF | Echo CSRF middleware on all POST/PUT/DELETE forms |
| XSS | html/template with automatic escape |
| SQL Injection | GORM with prepared statements; no dynamic raw queries |
| Rate Limiting | Echo rate limiter middleware on `/login` |
| Headers | Helmet-equivalent: HSTS, X-Frame-Options, basic CSP |

---

## 11. Roadmap — Phases

### Phase 1 — MVP (v1.0) ✅
- Basic authentication and users
- Projects, Nodes, Jobs CRUD
- Key Storage (SSH key + password)
- Job execution via SSH (command + inline script)
- Real-time log via SSE
- Cron scheduling
- Basic activity feed

### Phase 2 — Consolidation (v1.5)
- Import nodes via YAML/JSON
- Input parameters for jobs (options)
- Parallel execution across nodes
- Email notifications
- Roles and ACL per project
- Export/import jobs in YAML

### Phase 3 — Extension (v2.0)
- Support for **Ansible** connection (in addition to direct SSH)
- Plugin system for custom steps
- Inbound and outbound webhooks
- Complete REST API `/api/v1`
- `grd` CLI (client CLI similar to `rd`)
- Metrics dashboard (duration, success rate)
- Basic multi-tenancy

---

## 12. MVP Acceptance Criteria

- [ ] User can securely log in and log out.
- [ ] User can create a project, add nodes, and upload an SSH key.
- [ ] SSH connectivity test returns success/failure with a clear message.
- [ ] User can create a job with at least 1 `command` type step.
- [ ] Job can be executed manually with tag-based node selection.
- [ ] Execution log is displayed in real-time on the browser.
- [ ] Execution status is updated correctly (running → succeeded/failed).
- [ ] Cron schedule triggers the job at the configured time.
- [ ] jQuery Toasts show confirmations and errors non-blockingly.
- [ ] Interface uses square borders (`rounded-none`) across all elements.
- [ ] Brutalist shadows (`2px 2px 0` or `4px 4px 0` solid) applied to cards and buttons.
- [ ] Interface is responsive and functional on screens ≥ 1280px.
- [ ] `make all` compiles CSS and then the binary without errors.
- [ ] `go build .` at root generates the binary with all assets embedded.
- [ ] The resulting binary works without any external files (self-contained).
- [ ] `gorundeck serve` launches without needing the `web/` directory in the filesystem.
- [ ] `gorundeck migrate` creates all tables without errors.

---

## 13. Development Conventions

### 13.1 Go

- Follow `golangci-lint` with the `default` profile.
- Echo handlers return errors using `echo.NewHTTPError()`.
- All business logic goes in `service/`; handlers just orchestrate request/response.
- Unit tests in `*_test.go` next to the main file.
- `main.go` is **always in the root** of the project (`package main`).
- `embed.go` is in the **root** of the project, declaring `TemplatesFS` and `StaticFS`.

### 13.2 Frontend

- Named Go templates: `{{ define "content" }}` based on `layout/base.html`.
- jQuery 4.0.0 and jQuery Toast Plugin served via `//go:embed` (files in `web/static/js/`). No external CDN in production.
- **Tailwind CSS 4.2.0** — CSS-first configuration via `@theme` in `web/static/css/input.css`. **No** `tailwind.config.js`. **No** npm dependencies. Use Tailwind 4.2.0 standalone CLI.
- Compiled CSS (`app.css`) is generated before the Go build and embedded in the binary — the generated file must be present at `web/static/css/app.css` before running `go build`.
- No external assets are needed in production — everything lives inside the binary.

### 13.3 Mandatory Build Order

```bash
1. tailwindcss -i web/static/css/input.css -o web/static/css/app.css --minify
2. go build -ldflags "..." -o bin/gorundeck .
```

Step 1 must always precede step 2 so that the compiled CSS is included in the embed.

### 13.4 Git

- Default branch: `main`
- Feature branches: `feat/feature-name`
- Commits: Conventional Commits (`feat:`, `fix:`, `chore:`, `docs:`)
- `web/static/css/app.css` (compiled CSS) is **committed** to the repository to ease builds without the Tailwind CLI dependency in basic CI environments.

---

## 14. Makefile

```makefile
# ─── Variables ────────────────────────────────────────────────
APP        := gorundeck
BIN        := bin/$(APP)
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS    := -ldflags "-s -w \
               -X main.Version=$(VERSION) \
               -X main.BuildTime=$(BUILD_TIME) \
               -X main.GitCommit=$(GIT_COMMIT)"

# ─── Targets ──────────────────────────────────────────────────
.PHONY: all dev build css css-watch migrate test lint clean

## all: compiles CSS and then the binary
all: css build

## css: compiles Tailwind 4.2.0 (minified) — MUST run before build
css:
	tailwindcss -i web/static/css/input.css \
	            -o web/static/css/app.css \
	            --minify

## css-watch: recompiles Tailwind in watch mode (development)
css-watch:
	tailwindcss -i web/static/css/input.css \
	            -o web/static/css/app.css \
	            --watch

## build: compiles the binary (main.go at root — web assets must be generated beforehand)
build:
	go build $(LDFLAGS) -o $(BIN) .

## dev: live reload with air (css-watch in parallel)
dev:
	make css-watch & air -c .air.toml

## migrate: runs GORM migrations
migrate:
	./$(BIN) migrate

## test: runs unit tests
test:
	go test ./...

## lint: runs golangci-lint
lint:
	golangci-lint run

## clean: removes binaries and compiled CSS
clean:
	rm -rf bin/
	rm -f web/static/css/app.css
```

> **Note:** The `tailwindcss` CLI must be the **standalone v4.2.0** downloaded from:
> `https://github.com/tailwindlabs/tailwindcss/releases/tag/v4.2.0`
> It does not require Node.js or npm.

---

*Document updated on 2026-03-17 — Rev 2. Main changes: `go-rundeck` module, Tailwind CSS 4.2.0 (CSS-first, standalone CLI), square borders (`rounded-none`) as global default, `main.go` at the root of the project, `embed.go` at the root embedding `web/templates` and `web/static` into the binary.*
