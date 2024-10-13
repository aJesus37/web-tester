package browser

import (
	"context"
	"sync"

	"github.com/chromedp/cdproto/network"
)

type Requests []Request
type Responses struct {
	mu          sync.Mutex
	ResponseMap map[network.RequestID]Response
}

type Request struct {
	RequestID network.RequestID
	Type      string
	URL       string
	Content   interface{}
	Body      []byte
}

type Response struct {
	RequestID network.RequestID
	Type      string
	URL       string
	Content   interface{}
	Body      []byte
}

func (r *Responses) Add(response Response) {
	if r.ResponseMap == nil {
		r.ResponseMap = make(map[network.RequestID]Response)
	}
	r.ResponseMap[response.RequestID] = response
}

func (rqs *Requests) Add(request Request) {
	*rqs = append(*rqs, request)
}

func (r *Request) SetBody(ctx context.Context) {
	if r.Type != "request" {
		return
	}
	postDataEntries := r.Content.(*network.EventRequestWillBeSent).Request.PostDataEntries
	var postData string
	for _, entry := range postDataEntries {
		postData += entry.Bytes
	}
	r.Body = []byte(postData)
}
