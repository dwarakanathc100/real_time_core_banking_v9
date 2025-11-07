# Real-Time Core Banking Application — Complete Setup and Deployment Guide

This document provides a detailed step-by-step guide for setting up, testing, containerizing, deploying, and automating the deployment of the Real-Time Core Banking application using Docker, Docker Compose, Render.com, and GitHub Actions.

---

## 1. Local Development and Testing

### Step 1: Run the Go application locally

If you are running the application without Docker, ensure Go is installed and then execute the following command:

```bash
go run cmd/api/main.go
```

### Step 2: Verify the application is running

Open your browser or use curl to test the default endpoint:

```bash
curl http://localhost:8080/api/v1/customers
```

If authentication is enabled in your project, include a valid JWT token in the header while testing.

---

## 2. Dockerizing the Application

Docker allows you to package your Go application with all its dependencies so that it can run consistently across different environments.

### Step 1: Build the Docker image

```bash
docker build -t rtcb-app:latest .
```

This command reads the Dockerfile in your project directory and builds the image named `rtcb-app` with the tag `latest`.

### Step 2: Run the Docker image locally

```bash
docker run -d -p 8081:8080 rtcb-app:latest
```

This command starts a container from the image and maps the container port 8080 to port 8081 on your local system.

### Step 3: Verify the running container

```bash
docker ps
```

You can now visit [http://localhost:8081](http://localhost:8081) in your browser to confirm the application is working.

---

## 3. Using Environment Variables (.env)

To avoid hardcoding configuration values in your code, create a file named `.env` in the root directory of your project. Add the following configuration:

```dotenv
PORT=8080
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
POSTGRES_DB=rtcb
DATABASE_URL=postgres://postgres:postgres@db:5432/rtcb?sslmode=disable
REDIS_ADDR=redis:6379
JWT_SECRET=verysecretjwtkey
```

This file will be used by Docker Compose and your Go application to load runtime environment variables.

---

## 4. Running the Application with Docker Compose

Docker Compose allows you to run multiple containers together (for example, your application, PostgreSQL, and Redis).

### Step 1: Create a file named `docker-compose.yml`

```yaml
version: '3.8'

services:
  db:
    image: postgres:15-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: rtcb
    volumes:
      - db-data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine

  app:
    image: dwarak1262/real_time_core_banking:latest
    depends_on:
      - db
      - redis
    env_file: .env
    ports:
      - "8081:8080"

volumes:
  db-data:
```

### Step 2: Start all containers

```bash
docker compose up --build
```

This command will build the image (if not built yet) and start all services defined in the docker-compose.yml file.

### Step 3: Access the application

Once all containers are running, open your browser and visit:

```
http://localhost:8081
```

To stop and remove all containers, run:

```bash
docker compose down
```

---

## 5. Sharing docker compose file

You can distribute the following files to your students so they can run the complete setup locally:
- `docker-compose.yml`
- `.env.example` (containing sample values)

simply run the following command:

```bash
docker compose up
```

This will start the database, Redis, and the Go application automatically without any manual setup.

---

## 6. Deploying to Render.com

Render allows you to host Dockerized applications in the cloud with minimal configuration.

### Step 1: Push your Docker image to Docker Hub

First, tag and push your Docker image:

```bash
docker tag rtcb-app:latest dwarak1262/real_time_core_banking:v2
docker push dwarak1262/real_time_core_banking:v2
docker tag dwarak1262/real_time_core_banking:v2 dwarak1262/real_time_core_banking:latest
docker push dwarak1262/real_time_core_banking:latest
```

### Step 2: Create a new Render web service

1. Log in to [https://render.com](https://render.com)
2. Click **New → Web Service**
3. Select **Deploy from Docker image**
4. Use the image name:  
   `dwarak1262/real_time_core_banking:latest`
5. Set the environment variable `PORT=8080`
6. Deploy the service

---

## 7. Creating PostgreSQL and Redis Databases on Render

### Step 1: Create PostgreSQL database

1. Go to **Render → New → PostgreSQL**
2. After creation, open the **Connect** tab
3. Copy the **Internal Database URL**, for example:

```
postgres://rtcb_user:password@dpg-xxxx/rtcb
```

### Step 2: Add environment variable to your web service

Go to your web service → **Environment tab** → Add:

```
DATABASE_URL=postgres://rtcb_user:password@dpg-xxxx/rtcb
```

### Step 3: (Optional) Create a Redis instance

If your application requires Redis caching, create **New → Redis** and add:

```
REDIS_ADDR=redis-hostname:6379
```

---

## 8. Configuring Environment Variables in Render

In the Render dashboard, go to your **Web Service → Environment** tab and add the following variables:

| Key | Value |
|-----|--------|
| DATABASE_URL | The internal database URL copied from Render |
| REDIS_ADDR | Redis hostname and port |
| JWT_SECRET | verysecretjwtkey |
| PORT | 8080 |

After saving, click **Redeploy** to apply the new configuration.

---

## 9. Setting Up Continuous Integration and Deployment (CI/CD)

You can automate Docker image building and deployment to Render using GitHub Actions.

### Step 1: Create GitHub Workflow File

Create a directory `.github/workflows` and a file named `deploy.yml` inside it:

```yaml
name: Build and Deploy to Docker Hub and Render

on:
  push:
    branches:
      - main

jobs:
  build_and_deploy:
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Docker Buildx
        uses: docker/setup-buildx-action@v3

      - name: Log in to Docker Hub
        uses: docker/login-action@v3
        with:
          username: ${{ secrets.DOCKERHUB_USERNAME }}
          password: ${{ secrets.DOCKERHUB_TOKEN }}

      - name: Build and push Docker image
        uses: docker/build-push-action@v5
        with:
          context: .
          push: true
          tags: |
            ${{ secrets.DOCKERHUB_USERNAME }}/real_time_core_banking:latest
            ${{ secrets.DOCKERHUB_USERNAME }}/real_time_core_banking:${{ github.run_number }}

      - name: Trigger Render Deployment
        run: |
          curl -X POST "https://api.render.com/v1/services/${{ secrets.RENDER_SERVICE_ID }}/deploys"           -H "Authorization: Bearer ${{ secrets.RENDER_API_KEY }}"           -H "Accept: application/json"           -H "Content-Type: application/json"           -d '{}'
```

---

## 10. Configuring GitHub Secrets

Add the following secrets to your repository in GitHub:

| Secret Name | Description |
|--------------|-------------|
| `DOCKERHUB_USERNAME` | Your Docker Hub username |
| `DOCKERHUB_TOKEN` | Docker Hub access token with Read/Write/Delete permissions |
| `RENDER_SERVICE_ID` | Your Render service ID (found in the Render web service URL) |
| `RENDER_API_KEY` | Render API Key (generated from Render → Account Settings → API Keys) |

This setup ensures every push to the `main` branch automatically triggers a Docker image build, pushes it to Docker Hub, and deploys the new image to Render.

---

## 11. Commonly Used Docker Commands

```bash
# Build a Docker image
docker build -t rtcb-app:latest .

# Run a Docker container
docker run -d -p 8081:8080 rtcb-app:latest

# List Docker images
docker images

# List running containers
docker ps

# Stop a container
docker stop <container_id>

# Tag and push Docker image to Docker Hub
docker tag rtcb-app:latest dwarak1262/real_time_core_banking:v2
docker push dwarak1262/real_time_core_banking:v2

# Start all services with Docker Compose
docker compose up --build

# Stop and remove containers
docker compose down
```
```
Programming Languages: Go (Golang), Python, Shell Scripting
Frameworks & Libraries: Echo, Gin, Gorilla Mux, Cobra, Viper
Cloud Platforms: AWS (Lambda, ECS, EKS, IAM, CloudWatch), Azure Functions
DevOps & CI/CD: Docker, Helm, Terraform, GitHub Actions, GitLab CI
Databases: PostgreSQL, Redis, MongoDB, InfluxDB
Messaging & Streaming: Apache Kafka, RabbitMQ, NATS
Monitoring & Logging: Prometheus, Grafana, Jaeger, Fluentd, ELK Stack, New Relic
Protocols & Security: REST, gRPC, WebSockets, OAuth2, JWT, TLS
```
---

