package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
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

func getConfigPath() string {
	return ".mysqlctl_config"
}

func saveConfigLocked() error {
	if config == nil {
		return fmt.Errorf("no config to save")
	}

	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	return os.WriteFile(getConfigPath(), data, 0600)
}

func SaveConfig() error {
	mu.RLock()
	defer mu.RUnlock()
	return saveConfigLocked()
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

func ClearConfig() error {
	return os.Remove(getConfigPath())
}

func LoadAndConnect() error {
	cfg, err := LoadConfig()
	if err != nil {
		return err
	}

	return Connect(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)
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

	if err := saveConfigLocked(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
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
		_ = ClearConfig()
		return err
	}
	_ = ClearConfig()
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

func UpdateDatabase(dbName string) error {
	mu.Lock()
	defer mu.Unlock()

	if db == nil {
		return fmt.Errorf("not connected to MySQL")
	}

	if _, err := db.Exec(fmt.Sprintf("USE `%s`", dbName)); err != nil {
		return fmt.Errorf("failed to use database: %w", err)
	}

	config.Database = dbName
	if err := saveConfigLocked(); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}
	return nil
}
