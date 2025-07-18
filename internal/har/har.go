package har

import (
	"encoding/json"
	"os"
)

type HAR struct {
	Log Log `json:"log"`
}

type Log struct {
	Entries []Entry `json:"entries"`
}

type Entry struct {
	Request  Request  `json:"request"`
	Response Response `json:"response"`
}

type Request struct {
	Method string `json:"method"`
	URL    string `json:"url"`
}

type Response struct {
	Status  int      `json:"status"`
	Headers []Header `json:"headers"`
	Content Content  `json:"content"`
	Error   string   `json:"_error"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Content struct {
	Text     string `json:"text"`
	MimeType string `json:"mimeType"`
}

// LoadAndParse reads a HAR file from the given path and unmarshals its JSON content.
func LoadAndParse(filePath string) (*HAR, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var harData HAR
	if err := json.Unmarshal(fileBytes, &harData); err != nil {
		return nil, err
	}

	return &harData, nil
}
