package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"web-tester/internal/config"

	_ "github.com/lib/pq"

	"github.com/chromedp/cdproto/network"
	"github.com/google/uuid"
)

func InsertIntoDB(logger *slog.Logger, db *sql.DB, testID uuid.UUID, event struct {
	RequestID network.RequestID
	Type      string
	URL       string
	Content   interface{}
	Body      []byte
}) error {
	eventJSON, err := json.Marshal(event.Content)
	if err != nil {
		logger.Error("failed to marshal event content: ", "error: ", err)
		eventJSON = []byte{}
	}

	parsedURL, err := url.Parse(event.URL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}
	host := parsedURL.Host
	host = strings.Split(host, ":")[0]

	logger.Debug("Inserting into events table: ", "testID: ", testID.String(), "type: ", event.Type, "domain: ", host)
	_, err = db.Exec("INSERT INTO events (test_id, type, domain, payload, body) VALUES ($1, $2, $3, $4, $5)", testID, event.Type, host, string(eventJSON), event.Body)
	if err != nil {
		return fmt.Errorf("failed to insert into events table: %v", err)
	}
	return nil
}

// InitiateDB creates a new database connection to a postgres database
func Init(logger *slog.Logger, dbcfg config.DBConfig) (*sql.DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbcfg.Host, dbcfg.Port, dbcfg.User, dbcfg.Password, dbcfg.DBName)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	logger.Info("Successfully connected to the database")
	return db, nil
}
