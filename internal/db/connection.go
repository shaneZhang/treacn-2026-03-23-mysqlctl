package db

import (
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db     *sql.DB
	mu     sync.RWMutex
	config *Config
)

type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	Database string
}

func IsConnected() bool {
	mu.RLock()
	defer mu.RUnlock()
	return db != nil
}

func GetDB() *sql.DB {
	mu.RLock()
	defer mu.RUnlock()
	return db
}

func GetConfig() *Config {
	mu.RLock()
	defer mu.RUnlock()
	return config
}

func Connect(host string, port int, user, password, database string) error {
	mu.Lock()
	defer mu.Unlock()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/?parseTime=true&multiStatements=true",
		user, password, host, port)

	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %w", err)
	}

	if database != "" {
		if _, err := conn.Exec(fmt.Sprintf("USE `%s`", database)); err != nil {
			conn.Close()
			return fmt.Errorf("failed to use database: %w", err)
		}
	}

	db = conn
	config = &Config{
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		Database: database,
	}

	return nil
}

func Disconnect() error {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		err := db.Close()
		db = nil
		config = nil
		return err
	}
	return nil
}

func GetStatus() (connected bool, dbName string) {
	mu.RLock()
	defer mu.RUnlock()
	if db != nil {
		return true, config.Database
	}
	return false, ""
}
