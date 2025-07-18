package server

import (
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/draysams/har-server-replay/internal/har"
)

func makeHAR(entries []har.Entry) *har.HAR {
	return &har.HAR{Log: har.Log{Entries: entries}}
}

func TestReplayServer_Handler_BasicMatch(t *testing.T) {
	harData := makeHAR([]har.Entry{
		{
			Request: har.Request{Method: "GET", URL: "http://host/foo"},
			Response: har.Response{
				Status:  200,
				Headers: []har.Header{{Name: "X-Test", Value: "yes"}},
				Content: har.Content{Text: "ok", MimeType: "text/plain"},
			},
		},
	})
	srv := NewReplayServer(harData)
	req := httptest.NewRequest("GET", "/foo", nil)
	rec := httptest.NewRecorder()
	srv.handler(rec, req)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
	if rec.Header().Get("X-Test") != "yes" {
		t.Errorf("expected X-Test header")
	}
	if rec.Header().Get("Content-Type") != "text/plain" {
		t.Errorf("expected Content-Type 'text/plain', got %q", rec.Header().Get("Content-Type"))
	}
}

func TestReplayServer_Handler_404Unmatched(t *testing.T) {
	harData := makeHAR([]har.Entry{
		{Request: har.Request{Method: "GET", URL: "http://host/foo"}, Response: har.Response{Status: 200}},
	})
	srv := NewReplayServer(harData)
	req := httptest.NewRequest("POST", "/foo", nil)
	rec := httptest.NewRecorder()
	srv.handler(rec, req)
	if rec.Code != 404 {
		t.Errorf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "No more HAR entries") {
		t.Errorf("expected error message, got %q", rec.Body.String())
	}
}

func TestReplayServer_Handler_ExhaustedEntries(t *testing.T) {
	harData := makeHAR([]har.Entry{
		{Request: har.Request{Method: "GET", URL: "http://host/foo"}, Response: har.Response{Status: 200}},
	})
	srv := NewReplayServer(harData)
	req := httptest.NewRequest("GET", "/foo", nil)
	rec := httptest.NewRecorder()
	srv.handler(rec, req)
	// Second request should 404
	rec2 := httptest.NewRecorder()
	srv.handler(rec2, req)
	if rec2.Code != 404 {
		t.Errorf("expected 404, got %d", rec2.Code)
	}
}

func TestReplayServer_Handler_ErrorSimulation(t *testing.T) {
	harData := makeHAR([]har.Entry{
		{
			Request:  har.Request{Method: "GET", URL: "http://host/foo"},
			Response: har.Response{Status: 200, Error: "ERR_CONNECTION_REFUSED"},
		},
	})
	srv := NewReplayServer(harData)
	req := httptest.NewRequest("GET", "/foo", nil)
	rec := httptest.NewRecorder()
	srv.handler(rec, req)
	// Should not write a body or status (connection closed)
	if rec.Code != 200 && rec.Code != 0 {
		// Accept 0 (no WriteHeader called) or 200 (default)
		t.Errorf("expected 0 or 200, got %d", rec.Code)
	}
}

func TestReplayServer_Handler_ConcurrentRequests(t *testing.T) {
	harData := makeHAR([]har.Entry{
		{Request: har.Request{Method: "GET", URL: "http://host/foo"}, Response: har.Response{Status: 200, Content: har.Content{Text: "1"}}},
		{Request: har.Request{Method: "GET", URL: "http://host/foo"}, Response: har.Response{Status: 200, Content: har.Content{Text: "2"}}},
	})
	srv := NewReplayServer(harData)
	var wg sync.WaitGroup
	results := make([]string, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			rec := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/foo", nil)
			srv.handler(rec, req)
			results[idx] = rec.Body.String()
		}(i)
	}
	wg.Wait()
	if results[0] == results[1] {
		t.Errorf("expected different responses for each request, got %q and %q", results[0], results[1])
	}
}

func TestReplayServer_Handler_PathMatching(t *testing.T) {
	harData := makeHAR([]har.Entry{
		{Request: har.Request{Method: "GET", URL: "http://host:1234/foo/bar?x=1"}, Response: har.Response{Status: 200, Content: har.Content{Text: "ok"}}},
	})
	srv := NewReplayServer(harData)
	req := httptest.NewRequest("GET", "/foo/bar", nil)
	rec := httptest.NewRecorder()
	srv.handler(rec, req)
	if rec.Code != 200 {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Errorf("expected body 'ok', got %q", rec.Body.String())
	}
}

func TestReplayServer_Handler_CustomHeadersAndContentType(t *testing.T) {
	harData := makeHAR([]har.Entry{
		{
			Request: har.Request{Method: "GET", URL: "http://host/foo"},
			Response: har.Response{
				Status: 200,
				Headers: []har.Header{
					{Name: "X-Foo", Value: "bar"},
					{Name: "Content-Type", Value: "application/json"},
				},
				Content: har.Content{Text: "{}", MimeType: "application/json"},
			},
		},
	})
	srv := NewReplayServer(harData)
	req := httptest.NewRequest("GET", "/foo", nil)
	rec := httptest.NewRecorder()
	srv.handler(rec, req)
	if rec.Header().Get("X-Foo") != "bar" {
		t.Errorf("expected X-Foo header")
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got %q", rec.Header().Get("Content-Type"))
	}
}

func TestReplayServer_Handler_RequestIndexIncrements(t *testing.T) {
	harData := makeHAR([]har.Entry{
		{Request: har.Request{Method: "GET", URL: "http://host/foo"}, Response: har.Response{Status: 200, Content: har.Content{Text: "1"}}},
		{Request: har.Request{Method: "GET", URL: "http://host/foo"}, Response: har.Response{Status: 200, Content: har.Content{Text: "2"}}},
	})
	srv := NewReplayServer(harData)
	req := httptest.NewRequest("GET", "/foo", nil)
	rec1 := httptest.NewRecorder()
	srv.handler(rec1, req)
	rec2 := httptest.NewRecorder()
	srv.handler(rec2, req)
	if rec1.Body.String() == rec2.Body.String() {
		t.Errorf("expected different bodies for repeated requests, got %q and %q", rec1.Body.String(), rec2.Body.String())
	}
}

func TestReplayServer_Handler_VerboseDoesNotPanic(t *testing.T) {
	SetVerbose(true)
	harData := makeHAR([]har.Entry{
		{Request: har.Request{Method: "GET", URL: "http://host/foo"}, Response: har.Response{Status: 200}},
	})
	srv := NewReplayServer(harData)
	req := httptest.NewRequest("GET", "/foo", nil)
	rec := httptest.NewRecorder()
	srv.handler(rec, req)
	SetVerbose(false)
}
