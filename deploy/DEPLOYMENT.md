# VinzHub REST API - Deployment Guide

## 📁 Files to Upload
Upload entire `vinzhub-rest-api` folder to server.

## 🚀 Quick Start (Docker)

### 1. Build & Run
```bash
cd /path/to/vinzhub-rest-api
docker build -t vinzhub-api .
docker run -d --name vinzhub-api -p 8080:8080 --env-file .env vinzhub-api
```

### 2. Check Status
```bash
docker ps | grep vinzhub-api
docker logs vinzhub-api
```

---

## 🔧 Management Commands

### Start
```bash
docker start vinzhub-api
```

### Stop
```bash
docker stop vinzhub-api
```

### Restart
```bash
docker restart vinzhub-api
```

### View Logs
```bash
docker logs -f vinzhub-api          # Follow logs
docker logs --tail 100 vinzhub-api  # Last 100 lines
```

### Clear/Remove
```bash
docker stop vinzhub-api
docker rm vinzhub-api
docker rmi vinzhub-api  # Remove image
```

---

## 🔄 Update Deployment

```bash
# 1. Stop old container
docker stop vinzhub-api
docker rm vinzhub-api

# 2. Rebuild with new code
docker build -t vinzhub-api .

# 3. Start new container
docker run -d --name vinzhub-api -p 8080:8080 --env-file .env vinzhub-api
```

---

## 📋 Environment Variables (.env)

Make sure `.env` file exists with:
```env
PORT=8080
DB_HOST=your-db-host
DB_PORT=3306
DB_USER=your-user
DB_PASS=your-password
DB_NAME=game_log_db
API_KEY=vinzhub_sk_live_xxx
```

---

## 🏥 Health Check

```bash
curl http://localhost:8080/api/v1/health
```

Expected: `{"status":"ok"}`

---

## 🐧 Without Docker (Direct Binary)

### Build on Server
```bash
cd /path/to/vinzhub-rest-api
go build -o api ./cmd/api
./api
```

### Run as Service (systemd)
```bash
# Create service file
sudo nano /etc/systemd/system/vinzhub-api.service
```

Content:
```ini
[Unit]
Description=VinzHub REST API
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/path/to/vinzhub-rest-api
ExecStart=/path/to/vinzhub-rest-api/api
Restart=always
RestartSec=5
EnvironmentFile=/path/to/vinzhub-rest-api/.env

[Install]
WantedBy=multi-user.target
```

Commands:
```bash
sudo systemctl enable vinzhub-api
sudo systemctl start vinzhub-api
sudo systemctl status vinzhub-api
sudo systemctl restart vinzhub-api
sudo journalctl -u vinzhub-api -f  # View logs
```
