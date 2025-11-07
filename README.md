# real_time_core_banking_v9

Production-ready backend scaffold (Go) for a real-time core-banking demo.
Features:
- JWT authentication (bcrypt + JWT)
- Customer onboarding (CAF)
- Account management (view, deposit, withdraw, list)
- Transfer between accounts (atomic transactional)
- Transaction posting and ledger
- Loan scaffolding
- Redis queue for notifications (email/SMS simulation)
- Scheduler for auto statements
- Swagger docs served at /docs (static)
- PostgreSQL migrations executed on startup
- Docker + docker-compose setup
- Makefile to ease development
- Simple unit tests & GitHub Actions workflow

## Quickstart (local)
- Install Docker & Docker Compose
- `make docker-up` (starts postgres, redis and app)
- docker build -t rtcb:latest .
- App listens on :8080


--docker-compose up --build

-docker compose -f /path/to/your/project/docker-compose.yml up

-docker compose up -d

-docker images --list images

--docker login  --to login
Build your Docker image
docker build -t username/image-name:tag .
docker build -t dwarak1262/real_time_core_banking:v1 .
docker images
docker push username/image-name:tag
docker push dwarak1262/real_time_core_banking:v1

docker pull username/image-name:tag

docker run -d -p 8080:8080 username/image-name:tag

docker run -d -p 8080:8080 dwarak1262/real_time_core_banking:v1

docker stop <container_id>

docker rm <container_id>



]

-Useful Docker Compose Commands
   Command	Description
   docker compose up	Build and start containers
   docker compose up -d	Start in detached (background) mode
   docker compose down	Stop and remove containers, networks, volumes
   docker compose ps	Show running containers
   docker compose logs -f	View live logs
   docker compose build	Rebuild images manually


-docker ps
---
-docker exec -it real_time_core_banking_v9-db-1 psql -U postgres -d rtcb
-\dt  -->list tables
---
-SELECT * FROM customers;


-docker volume ls

   # Remove volume

-docker volume rm project_postgres_data



##  Go Code Documentation (using `pkgsite`)

### ▶️ Install `pkgsite`

`pkgsite` is the official Go tool for generating and viewing Go documentation locally.

```bash
go install golang.org/x/pkgsite/cmd/pkgsite@latest
```

Verify installation:

```bash
pkgsite version
```

If you get `'pkgsite' is not recognized`, make sure this folder is in your PATH:
```
C:\Users\<your-username>\go\bin
```

---

### ▶ Start Local Documentation Server

From your project root:
```bash
cd C:\2025\Go\exercises\real_time_core_banking_v9
pkgsite
```

You’ll see:
```
Serving documentation at http://localhost:8080
```

---

### ▶ View in Browser

Open:
```
http://localhost:8080
```

If your `go.mod` contains:

```go
module github.com/example/real_time_core_banking_v9
```
then open:
```
http://localhost:8080/github.com/example/real_time_core_banking_v9
```

If it contains:
```go
module real_time_core_banking_v9
```
then open:
```
http://localhost:8080/real_time_core_banking_v9
```

---

