package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	_ "github.com/go-sql-driver/mysql"
)

var (
	db     *sql.DB
	mu     sync.RWMutex
	config *Config
)

type Config struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func getConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}
	return filepath.Join(homeDir, ".mysqlctl_config.json")
}

func SaveConfig() error {
	mu.RLock()
	defer mu.RUnlock()
	if config == nil {
		return nil
	}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getConfigPath(), data, 0600)
}

func LoadConfig() (*Config, error) {
	data, err := os.ReadFile(getConfigPath())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
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
		if _, err := conn.Exec(fmt.Sprintf("USE %s", escapeIdentifier(database))); err != nil {
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

func SetCurrentDatabase(database string) {
	mu.Lock()
	defer mu.Unlock()
	if config != nil {
		config.Database = database
	}
}

func escapeIdentifier(name string) string {
	return "`" + strings.ReplaceAll(name, "`", "``") + "`"
}
