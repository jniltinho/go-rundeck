# Go-Rundeck

**Go-Rundeck** is a web-based runbook automation and task orchestration platform inspired by Rundeck, built natively in **Go**. The goal is to provide a lightweight, highly performant, and modern alternative to the original Rundeck (Java), maintaining a familiar layout and workflow but with a much leaner and more efficient tech stack.

## Features

- **Zero Dependencies**: Distributed as a single, self-contained binary with all web assets (HTML, CSS, JS) embedded.
- **Low Footprint**: No JVM required. Fast startup and minimal memory usage.
- **SSH Automation**: Execute commands and scripts on remote Linux/Unix nodes via SSH.
- **Projects & Nodes**: Organize your infrastructure by projects and manage nodes with tags, SSH keys, or passwords.
- **Jobs & Executions**: Create jobs with sequential or parallel steps, schedule cron jobs, and view real-time execution logs (SSE).
- **TLS/SSL Support**: Optional HTTPS with self-signed or CA-signed certificates; `Secure` cookie flag set automatically.
- **Session Security**: Configurable inactivity timeout (sliding window) with automatic logout and `HttpOnly`/`SameSite` cookie flags.
- **Modern UI**: Neo-brutalist aesthetic with square borders, high contrast, and dynamic interactions powered by Tailwind CSS and jQuery.

## Tech Stack
- **Backend:** Go, Echo Framework v5, GORM, MariaDB 10.x+, Cobra CLI.
- **Frontend:** Tailwind CSS, HTML Templates, jQuery.

## Installation & Build

### Prerequisites
- Go 1.26 or higher
- MariaDB 10.x+
- Tailwind CSS CLI (for development)

### Build from Source
```bash
# Clone the repository
git clone https://github.com/jniltinho/go-rundeck.git
cd go-rundeck

# Build the CSS and the Go binary
make all

# The compiled binary will be available at bin/gorundeck
```

### Running the Application

1. Create a MariaDB database (e.g., `gorundeck`).
2. Copy the example configuration file:
   ```bash
   cp config.toml.example config.toml
   ```
3. Edit `config.toml` with your database credentials and server settings.
4. Run migrations:
   ```bash
   ./bin/gorundeck migrate
   ```
5. Create an initial admin user:
   ```bash
   ./bin/gorundeck admin --add-user admin@example.com:password123
   ```
6. Start the server:
   ```bash
   ./bin/gorundeck serve
   ```

## Configuration

All options are documented in `config.toml.example`. Key sections:

### `[server]`

| Key | Default | Description |
|-----|---------|-------------|
| `name` | `Go-Rundeck` | Application name shown in logs |
| `env` | `development` | Environment (`development` \| `production`) |
| `port` | `8080` | HTTP/HTTPS listen port |
| `session_secret` | — | Random hex string, **minimum 32 characters** (generate with `openssl rand -hex 32`) |
| `session_timeout` | `60` | Inactivity timeout in minutes before the user is logged out |
| `debug` | `false` | Enable verbose debug output |
| `ssl_enable` | `false` | Enable HTTPS |
| `ssl_cert` | — | Path to TLS certificate file |
| `ssl_key` | — | Path to TLS private key file |

### Enabling SSL

Generate a self-signed certificate for development:
```bash
make certs
```

Then set in `config.toml`:
```toml
[server]
ssl_enable = true
ssl_cert   = "ssl/server.crt"
ssl_key    = "ssl/server.key"
```

When `ssl_enable = true` the session cookie is automatically marked `Secure`.

## 📖 Installation and Configuration Guide

For complete step-by-step instructions on how to set up Go-Rundeck on Ubuntu with MariaDB, systemd service, and HTTPS, see our [Complete Setup Guide](DOCUMENTS/setup/README.md).

## License

MIT License (or your chosen license)
