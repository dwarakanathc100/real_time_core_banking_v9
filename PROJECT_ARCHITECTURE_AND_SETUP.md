#  Real-Time Core Banking (v9) – System Architecture and Setup Guide

---

##  1 Overview

This document explains the **architecture**, **module design**, and **step-by-step setup instructions** to run the Real-Time Core Banking application locally with PostgreSQL and Docker.

---

## 2 High-Level Architecture

```
                   ┌───────────────────────────┐
                   │        Frontend (UI)      │
                   │  (Swagger UI / Postman)   │
                   └────────────┬──────────────┘
                                │ REST API Calls
                                ▼
                   ┌───────────────────────────┐
                   │     Go HTTP Server        │
                   │       (main.go)           │
                   │  Handles routes via mux   │
                   └────────────┬──────────────┘
                                │
         ┌──────────────────────┼──────────────────────┐
         ▼                      ▼                      ▼
 ┌────────────────┐     ┌────────────────┐     ┌────────────────┐
 │   auth module  │     │ customer module│     │  account module │
 │  JWT + Login   │     │  CRUD ops      │     │  Balance Mgmt  │
 └────────────────┘     └────────────────┘     └────────────────┘
                                │
                                ▼
                      ┌────────────────────┐
                      │ PostgreSQL Database│
                      │ (Docker Container) │
                      └────────────────────┘
                                │
                                ▼
                        ┌──────────────┐
                        │   Redis (optional)
                        │   For caching / async tasks
                        └──────────────┘
```

---

##  3 Module Breakdown

### 🔹 **1. main.go**
- Entry point of the application.
- Initializes database and Redis connections.
- Configures routes for:
  - Authentication
  - Customers
  - Accounts
  - Transactions
- Starts the HTTP server on port `8080`.

###  **2. internal/auth**
Handles **user registration** and **JWT-based authentication**.

| Component | Responsibility |
|------------|----------------|
| `auth.go` | User registration, password hashing (bcrypt), JWT token generation. |
| `middleware.go` | Verifies JWT tokens for protected endpoints. |

###  **3. internal/customer**
Handles **customer creation and listing**.

| File | Description |
|------|--------------|
| `repository.go` | Database queries to insert and list customers. |
| `handler.go` | HTTP handlers to create and list customers. |

###  **4. internal/account**
Handles **account operations** – create, deposit, withdraw, transfer, and check balance.

| File | Description |
|------|--------------|
| `repository.go` | SQL logic for account and transaction tables. |
| `handler.go` | HTTP endpoints for deposits, withdrawals, and transfers. |
| `repository_test.go` | Unit test for the package. |

###  **5. internal/transaction**
Handles **transaction history** and **notifications**.

| File | Description |
|------|--------------|
| `repository.go` | Fetches list of transactions. |
| `handler.go` | Serves transaction list endpoints. |
| `worker.go` | Background job for email/SMS notifications via Redis. |

---

##  4 Database Design (PostgreSQL)

### Database: `rtcb`
Tables:
1. `users` – stores user credentials for login
2. `customers` – stores customer personal details
3. `accounts` – stores customer account info
4. `transactions` – stores all transactions

### Example Schema (Simplified)
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL
);

CREATE TABLE customers (
    id SERIAL PRIMARY KEY,
    user_id INT,
    first_name TEXT,
    last_name TEXT,
    email TEXT,
    mobile TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE accounts (
    id SERIAL PRIMARY KEY,
    customer_id INT REFERENCES customers(id),
    account_number TEXT UNIQUE,
    currency TEXT DEFAULT 'INR',
    balance NUMERIC DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE transactions (
    id SERIAL PRIMARY KEY,
    account_id INT REFERENCES accounts(id),
    related_account_id INT,
    amount NUMERIC,
    type TEXT,
    narration TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
```

---

##  5 Step-by-Step Setup

### 🧱 **1. Prerequisites**
- Go (v1.21+)
- PostgreSQL (local or Docker)
- Redis (optional)
- PgAdmin (for UI)
- Docker Desktop (for container setup)

---

###  **2. Create Database Manually (Optional)**

#### a. Start PostgreSQL locally
```bash
psql -U postgres
```

#### b. Create database and user
```sql
CREATE DATABASE rtcb;
CREATE USER rtcb_user WITH ENCRYPTED PASSWORD 'password';
GRANT ALL PRIVILEGES ON DATABASE rtcb TO rtcb_user;
```

#### c. Connect using PgAdmin
1. Open **PgAdmin**
2. Add new server connection:
   - Host: `localhost`
   - Port: `5432`
   - Username: `postgres`
   - Password: `postgres`
3. Select the `rtcb` database.

---

###  **3. Run Database in Docker**

Create a file `docker-compose.yml` in project root:

```yaml
version: '3.8'
services:
  postgres:
    image: postgres:15
    container_name: rtcb_postgres
    restart: always
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: rtcb
    ports:
      - "5433:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data

  pgadmin:
    image: dpage/pgadmin4
    container_name: rtcb_pgadmin
    environment:
      PGADMIN_DEFAULT_EMAIL: admin@local.com
      PGADMIN_DEFAULT_PASSWORD: admin
    ports:
      - "5050:80"

volumes:
  pgdata:
```

#### Start containers
```bash
docker-compose up -d
```

PgAdmin will be available at:  
👉 http://localhost:5050  
Login with `admin@local.com` / `admin`  
Then connect to PostgreSQL:  
- Host: `postgres`
- Port: `5432`
- Username: `postgres`
- Password: `postgres`
- Database: `rtcb`

---

###  **4. Configure Environment Variables**

Create `.env` file (optional but recommended):

```bash
DATABASE_URL=postgres://postgres:postgres@localhost:5433/rtcb?sslmode=disable
REDIS_ADDR=localhost:6379
JWT_SECRET=devsecret
PORT=8080
```

---

###  **5. Run the Application**

```bash
go run main.go
```

Then visit:
- Health check: [http://localhost:8080/health](http://localhost:8080/health)
- Swagger docs: [http://localhost:8080/docs](http://localhost:8080/docs)

---

##  6 Dockerize the Entire App

Create a `Dockerfile` in project root:

```dockerfile
FROM golang:1.21-alpine

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o app .

EXPOSE 8080

CMD ["./app"]
```

Build and run container:
```bash
docker build -t rtcb_app .
docker run -p 8080:8080 --env-file .env rtcb_app
```

Now your app runs inside Docker and connects to PostgreSQL + PgAdmin containers.

---

##  7 Verification Steps

1. Open Swagger UI (via Docker): [http://localhost:8081](http://localhost:8081)
2. Register user `/v1/register`
3. Login `/v1/login` → copy JWT token
4. Authorize in Swagger UI → add `Bearer <token>`
5. Create customer `/v1/customers`
6. Create account `/v1/accounts`
7. Deposit / Withdraw / Transfer funds
8. Check transactions `/v1/transactions/list`

---

##  8 Summary of Components

| Component | Description |
|------------|--------------|
| Go Server | Main backend logic for REST APIs |
| PostgreSQL | Persistent storage for accounts, users, transactions |
| Redis | Queuing for async notifications (optional) |
| PgAdmin | Database UI |
| Swagger | API testing and documentation UI |

---

##  9 End-to-End Flow

1. User registers → stored in `users` table  
2. User logs in → gets JWT token  
3. Creates customer → entry in `customers` table  
4. Opens an account → `accounts` table  
5. Performs deposit/withdraw/transfer → records in `transactions`  
6. View balances & history → secure via JWT middleware  

---

## 10 Final URLs

| Service | URL |
|----------|------|
| Go Application | http://localhost:8080 |
| PgAdmin | http://localhost:5050 |
| Swagger Docs | http://localhost:8080/docs |
| Swagger UI (Docker) | http://localhost:8081 |

---
