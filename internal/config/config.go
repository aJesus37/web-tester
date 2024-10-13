package config

import "os"

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func (db *DBConfig) Load() DBConfig {
	db.Host = getEnv("DB_HOST", "localhost")
	db.Port = getEnv("DB_PORT", "5432")
	db.User = getEnv("DB_USER", "myuser")
	db.Password = getEnv("DB_PASSWORD", "mypassword")
	db.DBName = getEnv("DB_NAME", "events")

	return *db
}
