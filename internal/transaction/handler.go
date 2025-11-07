package transaction

import (
    "context"
    "database/sql"
    "encoding/json"
    "net/http"
    "strconv"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/sirupsen/logrus"
)

type Handler struct { repo *Repo; acctRepo interface{}; rdb *redis.Client }
func NewHandler(r *Repo, acctRepo interface{}, rdb *redis.Client) *Handler { return &Handler{repo:r, acctRepo:acctRepo, rdb:rdb} }

func (h *Handler) ListTransactions(w http.ResponseWriter, r *http.Request) {
    q := r.URL.Query()
    from := q.Get("from")
    to := q.Get("to")
    if from=="" { from = time.Now().AddDate(0, -1, 0).Format(time.RFC3339) }
    if to=="" { to = time.Now().Format(time.RFC3339) }
    aid := q.Get("account_id")
    if aid=="" { http.Error(w, "missing", http.StatusBadRequest); return }
    id, _ := strconv.Atoi(aid)
    list, err := h.repo.ListForAccount(id, from, to)
    if err!=nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
    json.NewEncoder(w).Encode(list)
}

func StartNotificationWorker(rdb *redis.Client, dbConn *sql.DB) {
    ctx := context.Background()
    for {
        res, err := rdb.BRPop(ctx, 0, "notifications").Result()
        if err!=nil { time.Sleep(1*time.Second); continue }
        if len(res) < 2 { continue }
        payload := res[1]
        logrus.Infof("processing notification: %s", payload)
    }
}
