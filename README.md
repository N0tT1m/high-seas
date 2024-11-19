# High Seas

A web application for searching and downloading shows/movies.

> **Note**: This application is for educational purposes only.

## Prerequisites

- Docker and Docker Compose
- NodeJS (v20.12.0 or later)
- Go (for building setup script)
- Python 3.x
- Nginx

## Quick Start

1. Clone the repository
2. Set up environment files (see Configuration section)
3. Run the installation script
4. Start the application using Docker Compose

## Installation

### Setup Script Options

Choose one of the following setup methods:

#### Go Script (Recommended)
```bash
# Windows
go build -o ./bin/setup.exe ./install-scripts/Setup.go

# Linux
go build -o ./bin/setup ./install-scripts/Setup.go
```

#### Python Script
> Note: Currently being refactored, use Go script instead.

### Frontend Setup

1. Install NodeJS:
   - Windows: Download from [NodeJS v20.12.0](https://nodejs.org/dist/v20.12.0/node-v20.12.0-x64.msi)
   - Linux: Use package manager

2. Local Development:
```bash
cd web
npm install
npm run start-local
```

3. Docker Deployment:
```bash
cd web
docker build -t high-seas-frontend .
docker run -d -p 6969:6969 high-seas-frontend
```

### Backend Setup

```bash
docker build -t high-seas-backend .
docker run -d -p 8782:8782 high-seas-backend
```

## Configuration

### 1. Frontend Environment (`./web/src/app/environments/environment.ts`)

```typescript
export const environment = {
  production: true,
  baseUrl: 'http://www.example.com:8080',
  envVar: {
    authorization: "YOUR_TMDB_API_BEARER_TOKEN",
    port: "GOLANG_API_PORT",
    ip: "GOLANG_API_IP",
    transport: "HTTPS_OR_HTTP",
  },
};
```

### 2. Backend Environment (`.env`)

Create this file in the root directory:
```env
DB_USER=DB_USER
DB_PASSWORD=DB_PASSWORD
DB_IP=DB_IP
DB_PORT=DB_PORT
DELUGE_IP=DELUGE_IP
DELUGE_PORT=DELUGE_PORT
DELUGE_USER=DELUGE_USER
DELUGE_PASSWORD=DELUGE_PASSWORD
JACKETT_IP=JACKETT_IP_HERE
JACKETT_PORT=JACKETT_PORT_HERE
JACKETT_API_KEY=YOUR_KEY_HERE
```

### 3. Plex Backend (`config.py`)
```python
HOST="192.168.1.1"
USER="root"
PASSWD="ThisIsAPassword"
DB="highseas"
IP="192.168.1.1"
PORT="32400"
```

### 4. Nginx Configuration (`./web/nginx.conf`)

```nginx
events{}

http {
    include /etc/nginx/mime.types;

    server {
        root /usr/share/nginx/html;
        index index.html;
        listen 6969;
        server_name http://goose.duocore.space http://arch.duocore.space;

        location / {
            try_files $uri $uri/ /index.html;
        }
    }
}
```

## Production Deployment

### Using Docker Compose

#### Linux
```bash
./start-dedicated.sh
```

## Project Structure

```
.
├── web/                    # Frontend application
├── install-scripts/        # Setup scripts
├── docker-compose.yml      # Docker composition
└── README.md              # This file
```

## Troubleshooting

If you encounter issues:
1. Verify all environment variables are set correctly
2. Ensure all required ports are available
3. Check Docker logs for detailed error messages

## Support

For issues and feature requests, please open an issue in the repository.
