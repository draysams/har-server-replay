package server

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/draysams/har-server-replay/internal/har"
)

var isVerbose bool

// ReplayServer manages HAR replay state and request matching.
type ReplayServer struct {
	harData      *har.HAR
	requestIndex map[string]int
	mu           sync.Mutex
}

// SetVerbose enables or disables verbose logging.
func SetVerbose(v bool) {
	isVerbose = v
}

// NewReplayServer initializes a new ReplayServer instance.
func NewReplayServer(harData *har.HAR) *ReplayServer {
	return &ReplayServer{
		harData:      harData,
		requestIndex: make(map[string]int),
	}
}

// handler matches incoming HTTP requests to HAR entries and serves the recorded response.
func (s *ReplayServer) handler(w http.ResponseWriter, r *http.Request) {
	requestSignature := fmt.Sprintf("%s %s", r.Method, r.URL.Path)

	if isVerbose {
		log.Printf("Received request: %s", requestSignature)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	currentIndex := s.requestIndex[requestSignature]
	var foundEntry *har.Entry

	for i := currentIndex; i < len(s.harData.Log.Entries); i++ {
		entry := s.harData.Log.Entries[i]
		harURL, err := url.Parse(entry.Request.URL)
		if err != nil {
			if isVerbose {
				log.Printf("Could not parse URL in HAR entry #%d: %s", i, entry.Request.URL)
			}
			continue
		}
		if entry.Request.Method == r.Method && harURL.Path == r.URL.Path {
			foundEntry = &s.harData.Log.Entries[i]
			s.requestIndex[requestSignature] = i + 1
			break
		}
	}

	if foundEntry == nil {
		http.Error(w, fmt.Sprintf("No more HAR entries for request: %s", requestSignature), http.StatusNotFound)
		return
	}

	if isVerbose {
		log.Printf("Found match at index %d. Replaying response (Status: %d, Error: '%s')", s.requestIndex[requestSignature]-1, foundEntry.Response.Status, foundEntry.Response.Error)
	}

	if foundEntry.Response.Error != "" {
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			conn.Close()
		}
		return
	}

	for _, header := range foundEntry.Response.Headers {
		if !strings.EqualFold(header.Name, "Content-Length") {
			w.Header().Set(header.Name, header.Value)
		}
	}

	if foundEntry.Response.Content.MimeType != "" {
		w.Header().Set("Content-Type", foundEntry.Response.Content.MimeType)
	}

	w.WriteHeader(foundEntry.Response.Status)
	_, _ = w.Write([]byte(foundEntry.Response.Content.Text))
}

// Start launches the HTTP server on the specified port with the replay handler.
func Start(port int, harData *har.HAR) error {
	replayServer := NewReplayServer(harData)
	mux := http.NewServeMux()
	mux.HandleFunc("/", replayServer.handler)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting HAR replay server on http://localhost%s", addr)
	log.Printf("Loaded %d total entries from the HAR file.", len(harData.Log.Entries))

	return http.ListenAndServe(addr, mux)
}
