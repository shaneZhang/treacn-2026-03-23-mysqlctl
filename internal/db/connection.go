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
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
}

func getConfigPath() string {
	// Use current directory for config file
	return ".mysqlctl.json"
}

func LoadConfig() (*Config, error) {
	configPath := getConfigPath()
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func SaveConfig(cfg *Config) error {
	configPath := getConfigPath()
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, data, 0600)
}

func ClearConfig() error {
	configPath := getConfigPath()
	return os.Remove(configPath)
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

	// Save config to file for persistence
	if err := SaveConfig(config); err != nil {
		// Non-fatal error, just log it
		fmt.Fprintf(os.Stderr, "Warning: failed to save config: %v\n", err)
	}

	return nil
}

func ConnectWithSavedConfig() error {
	cfg, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg == nil {
		return fmt.Errorf("no saved configuration found. Use 'connect' command first")
	}
	return Connect(cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Database)
}

func Disconnect() error {
	mu.Lock()
	defer mu.Unlock()

	if db != nil {
		err := db.Close()
		db = nil
		config = nil
		// Clear saved config
		ClearConfig()
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

func UpdateDatabase(dbName string) {
	mu.Lock()
	defer mu.Unlock()
	if config != nil {
		config.Database = dbName
		// Save updated config
		SaveConfig(config)
	}
}
