package db

import (
    "database/sql"
    "io/ioutil"
    "time"
)

func WaitForDB(db *sql.DB, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for {
        if time.Now().After(deadline) {
            return nil
        }
        if err := db.Ping(); err == nil {
            return nil
        }
        time.Sleep(1 * time.Second)
    }
}

func ExecMigrations(db *sql.DB, path string) error {
    content, err := ioutil.ReadFile(path)
    if err != nil {
        return err
    }
    _, err = db.Exec(string(content))
    return err
}
