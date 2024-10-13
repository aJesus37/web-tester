// Package browser provides functionality to interact with a web browser using the chromedp package.
package browser

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/google/uuid"
)

// Browser represents a browser instance with a target URL, context, and cancel function.
type Browser struct {
	target string
	ctx    context.Context
	cancel context.CancelFunc
	testID uuid.UUID
}

// New creates a new Browser instance with the specified target URL.
// It initializes a chromedp context with logging and sets a timeout of 60 seconds to prevent infinite wait loops.
func New(target string) *Browser {
	// create context
	ctx, _ := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)

	id, err := uuid.NewV7()
	if err != nil {
		log.Fatalf("failed to create test ID: %v", err)
	}

	// create a timeout as a safety net to prevent any infinite wait loops
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	return &Browser{target: target, ctx: ctx, cancel: cancel, testID: id}
}

// TestID returns the browser's test ID.
func (b *Browser) TestID() uuid.UUID {
	return b.testID
}

// Cancel cancels the browser's context, stopping any ongoing operations.
func (b *Browser) Cancel() {
	b.cancel()
}

// GetCtx returns the browser's context.
func (b *Browser) GetCtx() context.Context {
	return b.ctx
}

// ListenToEvents sets up listeners for various browser events and processes them accordingly.
// It listens for network request, response, and loading finished events, and logs the events
// using the provided logger. The events are also added to the respective Requests and Responses
// collections, and the loading finished events are sent to the finisher channel.
//
// Parameters:
//   - logger: A pointer to an slog.Logger used for logging event information.
//   - responses: A pointer to a Responses collection where response events are added.
//   - requests: A pointer to a Requests collection where request events are added.
//   - finisherChan: A pointer to a channel where loading finished events are sent.
func (b *Browser) ListenToEvents(logger *slog.Logger, responses *Responses, requests *Requests, finisherChan *chan network.EventLoadingFinished) {
	// listen for events
	chromedp.ListenTarget(b.ctx, func(ev interface{}) {
		switch ev := ev.(type) {
		// case *page.EventFrameNavigated:
		// 	fmt.Printf("frame navigated: %s\n", ev.Frame.URL)
		case *network.EventRequestWillBeSent:
			go func() {
				logger.Info("EventRequestWillBeSent: ", "requestID: ", ev.RequestID)
				requests.Add(Request{RequestID: ev.RequestID, Type: "request", URL: ev.Request.URL, Content: ev})
			}()

		case *network.EventResponseReceived:
			go func() {
				logger.Info("EventResponseReceived:", "requestID: ", ev.RequestID)
				responses.Add(Response{RequestID: ev.RequestID, Type: "response", URL: ev.Response.URL, Content: ev})
			}()

		case *network.EventLoadingFinished:
			go func() {
				logger.Info("EventLoadingFinished:", "requestID: ", ev.RequestID)
				*finisherChan <- *ev
			}()

		}
	})
}

// Run navigates the browser to the target URL specified in the Browser struct.
// It uses the chromedp package to perform the navigation.
// Returns an error if the navigation fails.
func (b *Browser) Run(waitTime time.Duration) error {
	// navigate to the target URL
	if err := chromedp.Run(b.ctx, chromedp.Navigate(b.target)); err != nil {
		return err
	}

	// wait for the specified duration
	chromedp.Sleep(waitTime)

	return nil
}

// GetResponseBody retrieves the response body for a given request and updates the response map.
// It logs the initial and final lengths of the response body at various stages of the process.
//
// Parameters:
// - logger: A structured logger for logging information and errors.
// - r: A pointer to the Response struct containing the request ID and body.
// - responses: A pointer to the Responses struct containing a map of responses and a mutex for synchronization.
//
// Returns:
// - error: An error if the response body could not be retrieved or updated.
func (b *Browser) GetResponseBody(logger *slog.Logger, r *Response, responses *Responses) error {
	logger.Info("initial response body length: ", "len: ", len(r.Body))

	err := chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		body, err := network.GetResponseBody(r.RequestID).Do(ctx)
		if err != nil {
			logger.Error("failed to get response body: ", "error: ", err)
			return fmt.Errorf("failed to get response body: %v", err)
		}
		r.Body = body
		return nil
	}))

	if err != nil {
		return fmt.Errorf("could not get response body: %v", err)
	}

	// Lock the mutex before updating the map
	responses.mu.Lock()
	defer responses.mu.Unlock()
	responses.ResponseMap[r.RequestID] = *r

	return nil
}

// NewFinisherChannel creates and returns a new channel for network.EventLoadingFinished events.
// This channel can be used to receive notifications when a network loading event has finished.
func (b *Browser) NewFinisherChannel() chan network.EventLoadingFinished {
	return make(chan network.EventLoadingFinished)
}

// WatchEventFinishers listens for network loading finished events and processes the responses.
// It logs the event details and retrieves the response body for each event.
//
// Parameters:
//   - logger: A pointer to an slog.Logger instance for logging event details.
//   - f: A pointer to a channel of network.EventLoadingFinished events to watch.
//   - responses: A pointer to a Responses struct containing the response map and mutex.
func (b *Browser) WatchEventFinishers(logger *slog.Logger, f *chan network.EventLoadingFinished, responses *Responses) {
	log.Printf("Watching for event finishers")
	go func(responses *Responses) {
		for event := range *f {
			logger.Info("EventLoadingFinished, getting body:", "requestID: ", event.RequestID)

			// Lock the mutex before reading from the map
			responses.mu.Lock()
			resp := responses.ResponseMap[event.RequestID]
			responses.mu.Unlock()

			b.GetResponseBody(logger, &resp, responses)
		}
	}(responses)
}
