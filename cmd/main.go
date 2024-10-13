package main

import (
	"log/slog"
	"os"
	"time"
	"web-tester/internal/browser"
	"web-tester/internal/config"
	"web-tester/internal/database"

	"github.com/chromedp/cdproto/network"
)

// main is the entry point of the web-tester application. It performs the following tasks:
// 1. Initializes a logger with JSON output and info level logging.
// 2. Creates a new browser client for the specified URL and ensures it is properly canceled on exit.
// 3. Loads the database configuration and initializes the database connection.
// 4. Sets up channels and structures to handle browser events, requests, and responses.
// 5. Listens to browser events and runs the browser for a specified duration.
// 6. Watches for event finishers and logs the successful run of the browser.
// 7. Iterates over the captured requests and responses, inserting them into the database.
//
// If any errors occur during database initialization, browser execution, or database insertion,
// they are logged appropriately.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))

	client := browser.New("https://google.com")
	defer client.Cancel()

	dbConfig := &config.DBConfig{}
	db, err := database.Init(logger, dbConfig.Load())
	if err != nil {
		logger.Error("failed to initialize database", "error: ", err)
	}

	var finisherChan = client.NewFinisherChannel()
	var responses = browser.Responses{}
	var requests = browser.Requests{}

	client.ListenToEvents(logger, &responses, &requests, &finisherChan)

	err = client.Run(5 * time.Second)
	if err != nil {
		logger.Error("failed to run browser:", "error: ", err)
		panic(err)
	}

	client.WatchEventFinishers(logger, &finisherChan, &responses)

	logger.Info("browser ran successfully, starting database input")

	for _, r := range requests {
		r.SetBody(client.GetCtx())
		err = database.InsertIntoDB(logger, db, client.TestID(), struct {
			RequestID network.RequestID
			Type      string
			URL       string
			Content   interface{}
			Body      []byte
		}{RequestID: r.RequestID, Type: r.Type, URL: r.URL, Content: r.Content, Body: r.Body})
		if err != nil {
			logger.Error("failed to insert into database", "error: ", err)
		}
	}

	for _, r := range responses.ResponseMap {
		err = database.InsertIntoDB(logger, db, client.TestID(), struct {
			RequestID network.RequestID
			Type      string
			URL       string
			Content   interface{}
			Body      []byte
		}{RequestID: r.RequestID, Type: r.Type, URL: r.URL, Content: r.Content, Body: r.Body})
		if err != nil {
			logger.Error("failed to insert into database: ", "error: ", err)
		}
	}
}
