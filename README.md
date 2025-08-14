# Go Microservices System with Grafana Loki Logging

This project contains two Golang microservices with comprehensive logging and monitoring setup:

- **user-service**: Handles authentication and user management
- **collab-service**: Manages collaboration features and communicates with `user-service`

## 🛠️ Prerequisites

Make sure you have the following installed:

- [Docker](https://www.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## 📂 Project Structure

```
.
├── user-service/
│   ├── Dockerfile
│   └── logs/
├── collab-service/
│   ├── Dockerfile
│   └── logs/
├── docker-compose.yml
├── promtail-config.yml
└── README.md
```

## 🚀 Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/your-repo.git
cd your-repo
```

### 2. Build and Run All Services

```bash
docker-compose up --build
```

Wait until all services (PostgreSQL, user-service, collab-service, Loki, Promtail, Grafana) are up and running.

## ⚙️ Service Details

### 🔐 Auth Service

- **URL**: http://localhost:4000/query
- **Port**: 4000
- **Database**: `user_service`

### 🤝 Collab Service

- **URL**: http://localhost:8080
- **Port**: 8080
- **Database**: `collab_service`

### 🐘 PostgreSQL

- **Host**: localhost
- **Port**: 5432
- **User**: postgres
- **Password**: truonghoang2004
- **Databases**:
  - `user_service`
  - `collab_service`

## 📊 Logging & Monitoring

### 🔎 Grafana Setup

- **URL**: http://localhost:3000
- **Login**: admin / admin

#### Setup Instructions:

1. Navigate to **Settings** → **Data Sources**
2. Add a Loki data source with the URL: `http://loki:3100`
3. Go to **Explore** tab and choose:
   - `{job="user-service"}` for auth logs
   - `{job="collab-service"}` for collab logs

### 📜 Promtail Configuration

Promtail reads logs from the following mounted files:

- `user-service/logs/auth_service.log`
- `collab-service/logs/server.log`

**Important**: Ensure your services write logs to these paths in structured format (preferably JSON).

## 🔧 Individual Service Management

### Running a Single Service

To build and run only the user-service:

```bash
docker-compose build user-service
docker-compose up -d user-service
```

To build and run only the collab-service:

```bash
docker-compose build collab-service
docker-compose up -d collab-service
```

### Debugging Container Environment

Check environment variables inside a running container:

```bash
docker exec -it <container_name> env
```

Access container shell:

```bash
docker exec -it <container_name> sh
```

## 🧹 Cleanup

Stop all containers and remove volumes (including PostgreSQL data):

```bash
docker-compose down -v
```

## ⚙️ Configuration

All service-specific settings (e.g., JWTAccessSecret, DB_HOST, etc.) are configured inside `docker-compose.yml`.

Environment variables can be customized by modifying the docker-compose file or using `.env` files.

## 📌 Important Notes

- Use **JSON or structured logs** for better parsing in Grafana
- The mounted `logs/` directory is required for Promtail to function properly
- Ensure environment values are properly synced between services
- Log files must be created in the specified paths for monitoring to work

## 🐛 Troubleshooting

### Common Issues:

1. **Services not connecting**: Check if all containers are running with `docker-compose ps`
2. **Logs not appearing in Grafana**: Verify log files exist in the mounted directories
3. **Database connection errors**: Ensure PostgreSQL container is healthy before starting other services

### Viewing Logs:

```bash
# View user-service logs
docker-compose logs user-service

# View collab-service logs
docker-compose logs collab-service

# View all logs
docker-compose logs
```

## 📚 Additional Resources

- [Go Documentation](https://golang.org/doc/)
- [Docker Compose Reference](https://docs.docker.com/compose/compose-file/)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/)
- [Promtail Configuration](https://grafana.com/docs/loki/latest/clients/promtail/configuration/)