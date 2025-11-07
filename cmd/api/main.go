package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"github.com/example/real_time_core_banking_v9/internal/account"
	"github.com/example/real_time_core_banking_v9/internal/auth"
	"github.com/example/real_time_core_banking_v9/internal/customer"
	"github.com/example/real_time_core_banking_v9/internal/db"
	"github.com/example/real_time_core_banking_v9/internal/transaction"
	"github.com/go-redis/redis/v8"
)

func main() {
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true})
	logrus.Info("Starting real-time core banking app...")

	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found — using system environment variables")
	}

	// --- Get Database URL
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@localhost:5432/rtcb?sslmode=disable"
		logrus.Warn("DATABASE_URL not set, using default local URL")
	}
	dbConn, err := sql.Open("postgres", databaseURL)
	if err != nil {
		logrus.Fatal(err)
	}
	if err := db.WaitForDB(dbConn, 60*time.Second); err != nil {
		logrus.Fatal("db failed to be ready:", err)
	}
	if err := db.ExecMigrations(dbConn, "internal/db/migrations/001_init.sql"); err != nil {
		logrus.Fatal("migration failed:", err)
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: redisAddr})

	repoCustomer := customer.NewRepo(dbConn)
	handlerCustomer := customer.NewHandler(repoCustomer)

	repoAccount := account.NewRepo(dbConn)
	handlerAccount := account.NewHandler(repoAccount)

	repoTxn := transaction.NewRepo(dbConn)
	handlerTxn := transaction.NewHandler(repoTxn, repoAccount, rdb)

	// auth
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "devsecret"
	}
	authSvc := auth.NewAuthService(dbConn, jwtSecret)

	mux := http.NewServeMux()

	// basic endpoints
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	// docs (static swagger yaml/json)
	mux.HandleFunc("/docs", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/swagger.yaml")
	})

	// auth endpoints
	mux.HandleFunc("/v1/register", authSvc.RegisterHandler)
	mux.HandleFunc("/v1/login", authSvc.LoginHandler)

	// customer endpoints
	mux.HandleFunc("/v1/customers", auth.WithAuth(handlerCustomer.CreateCustomer, jwtSecret))
	mux.HandleFunc("/v1/customers/list", auth.WithAuth(handlerCustomer.ListCustomers, jwtSecret))
	mux.HandleFunc("/v2/customers/list", handlerCustomer.ListCustomers)
	mux.HandleFunc("/v2/balance", handlerAccount.GetBalance)
	mux.HandleFunc("/v2/transactions/list", handlerTxn.ListTransactions)

	// accounts
	mux.HandleFunc("/v1/accounts", auth.WithAuth(handlerAccount.CreateAccount, jwtSecret))
	mux.HandleFunc("/v1/accounts/balance", auth.WithAuth(handlerAccount.GetBalance, jwtSecret))
	mux.HandleFunc("/v1/accounts/deposit", auth.WithAuth(handlerAccount.Deposit, jwtSecret))
	mux.HandleFunc("/v1/accounts/withdraw", auth.WithAuth(handlerAccount.Withdraw, jwtSecret))
	mux.HandleFunc("/v1/accounts/transfer", auth.WithAuth(handlerAccount.Transfer, jwtSecret))

	// transactions
	mux.HandleFunc("/v1/transactions/list", auth.WithAuth(handlerTxn.ListTransactions, jwtSecret))

	// start background workers
	go transaction.StartNotificationWorker(rdb, dbConn)

	// start scheduler for statements
	go startScheduler(rdb)

	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + p
	}
	logrus.Infof("listening on %s", addr)
	srv := &http.Server{
		Addr:         addr,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	if err := srv.ListenAndServe(); err != nil {
		logrus.Fatal(err)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logrus.Infof("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func startScheduler(rdb *redis.Client) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	// for demo purposes run once at startup
	ctx := context.Background()
	payload := map[string]interface{}{"type": "statement_email", "info": "daily statement"}
	b, _ := json.Marshal(payload)
	rdb.LPush(ctx, "notifications", b)
	for range ticker.C {
		ctx := context.Background()
		payload := map[string]interface{}{"type": "statement_email", "info": "daily statement"}
		b, _ := json.Marshal(payload)
		rdb.LPush(ctx, "notifications", b)
	}
}
