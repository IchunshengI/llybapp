package init

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// OpenMySQLFromEnv connects to MySQL in the same docker network by default.
func OpenMySQLFromEnv() (*sql.DB, error) {
	host := getenv("MYSQL_HOST", "llyb-mysql")
	port := getenv("MYSQL_PORT", "3306")
	user := getenv("MYSQL_USER", "llyb")
	pass := getenv("MYSQL_PASSWORD", "llybpass")
	dbname := getenv("MYSQL_DATABASE", "llyb")
	params := getenv("MYSQL_PARAMS", "charset=utf8mb4&parseTime=true&loc=Local")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?%s", user, pass, host, port, dbname, params)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func EnsureAdminAccountTable(ctx context.Context, db *sql.DB) error {
	_, err := db.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS admin_account (
  id BIGINT NOT NULL AUTO_INCREMENT,
  username VARCHAR(64) NOT NULL,
  password_hash CHAR(32) NOT NULL,
  password_salt CHAR(32) NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id),
  UNIQUE KEY uk_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;`)
	return err
}

func getenv(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

