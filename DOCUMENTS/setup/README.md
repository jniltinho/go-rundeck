# Installation Guide: Go-Rundeck on Ubuntu

This guide walks you through installing **Go-Rundeck** on an Ubuntu server using a **MariaDB** database, running the application as a **systemd service**, and optionally enabling **HTTPS** with a self-signed or Let's Encrypt certificate.

---

## 1. Update the System and Install Dependencies

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install mariadb-server curl git openssl -y
```

---

## 2. Install Go (if building from source)

```bash
sudo apt install golang -y
```

Or download the latest release from https://go.dev/dl and install manually:

```bash
wget https://go.dev/dl/go1.24.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.24.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

---

## 3. Configure the MariaDB Database

```bash
sudo mysql_secure_installation
```

Then create the database and user:

```sql
sudo mysql -u root -p

CREATE DATABASE rundeck CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
CREATE USER 'rundeck'@'localhost' IDENTIFIED BY 'strong_password_here';
GRANT ALL PRIVILEGES ON rundeck.* TO 'rundeck'@'localhost';
FLUSH PRIVILEGES;
EXIT;
```

---

## 4. Build and Install Go-Rundeck

```bash
# Clone the repository
git clone https://github.com/jniltinho/go-rundeck.git
cd go-rundeck

# Install Tailwind CSS CLI (required for CSS build)
curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/latest/download/tailwindcss-linux-x64
chmod +x tailwindcss-linux-x64
sudo mv tailwindcss-linux-x64 /usr/local/bin/tailwindcss

# Build the binary (CSS + Go)
make all

# Create the installation directory
sudo mkdir -p /opt/go-rundeck

# Copy the binary and config
sudo cp bin/gorundeck /opt/go-rundeck/
sudo cp config.toml.example /opt/go-rundeck/config.toml
```

---

## 5. Configure the Application

Edit `/opt/go-rundeck/config.toml`:

```bash
sudo nano /opt/go-rundeck/config.toml
```

Minimum required settings:

```toml
[server]
name           = "Go-Rundeck"
env            = "production"
port           = 8080
session_secret = "replace_with_64_char_hex_openssl_rand_hex_32"
session_timeout = 60
ssl_enable     = false

[database]
host     = "localhost"
port     = 3306
user     = "rundeck"
password = "strong_password_here"
name     = "rundeck"
charset  = "utf8mb4"
```

Generate a secure session secret:

```bash
openssl rand -hex 32
```

---

## 6. Run Database Migrations

```bash
cd /opt/go-rundeck
./gorundeck migrate
```

---

## 7. Create the Initial Admin User

```bash
cd /opt/go-rundeck
./gorundeck admin --add-user admin@example.com:password123
```

---

## 8. Install and Start the systemd Service

```bash
# Copy the service file
sudo cp /path/to/go-rundeck/DOCUMENTS/setup/gorundeck.service /etc/systemd/system/

# Reload systemd and enable the service
sudo systemctl daemon-reload
sudo systemctl enable gorundeck
sudo systemctl start gorundeck

# Check status
sudo systemctl status gorundeck
```

View logs:

```bash
sudo tail -f /opt/go-rundeck/gorundeck.log
```

---

## 9. Enable HTTPS (Optional)

### Option A — Self-signed certificate (development)

```bash
cd /opt/go-rundeck
mkdir -p ssl
openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout ssl/server.key -out ssl/server.crt \
    -subj "/C=BR/ST=SP/L=Sao Paulo/O=MyOrg/CN=yourdomain.com"
```

### Option B — Let's Encrypt (production)

```bash
sudo apt install certbot -y
sudo certbot certonly --standalone -d yourdomain.com
```

Certificate files will be at:
- `/etc/letsencrypt/live/yourdomain.com/fullchain.pem`
- `/etc/letsencrypt/live/yourdomain.com/privkey.pem`

### Enable SSL in config.toml

```toml
[server]
ssl_enable = true
ssl_cert   = "ssl/server.crt"
ssl_key    = "ssl/server.key"
```

Restart the service:

```bash
sudo systemctl restart gorundeck
```

When `ssl_enable = true`, the session cookie is automatically marked `Secure` and sent only over HTTPS.

---

## 10. Firewall (UFW)

Allow the application port:

```bash
sudo ufw allow 8080/tcp   # HTTP
# or
sudo ufw allow 443/tcp    # HTTPS (if using port 443)
sudo ufw reload
```

---

## Directory Structure on the Server

```
/opt/go-rundeck/
├── gorundeck          # compiled binary
├── config.toml        # configuration file
├── gorundeck.log      # application log
├── keys/              # SSH key storage
└── ssl/               # TLS certificates (if using SSL)
```

---

## Useful Commands

| Command | Description |
|---------|-------------|
| `systemctl start gorundeck` | Start the service |
| `systemctl stop gorundeck` | Stop the service |
| `systemctl restart gorundeck` | Restart the service |
| `systemctl status gorundeck` | Check service status |
| `tail -f /opt/go-rundeck/gorundeck.log` | Follow logs |
| `./gorundeck migrate` | Run database migrations |
| `./gorundeck admin --add-user user@x.com:pass` | Add admin user |
| `./gorundeck config-check` | Validate config and test DB connection |
