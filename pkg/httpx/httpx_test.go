package httpx

import (
	"encoding/base64"
	"io"
	"testing"
)

func TestClientDo(t *testing.T) {
	client := NewDefaultClient("http://192.168.12.34:8084")
	resp, err := client.Do(NewRequestOption(
		WithMethodGet(),
		WithPath("/api/v2/1/diagnosis-reporter/4372/content"),
		WithHeaders(map[string]string{
			"Content-Type":  "application/json",
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte("admin:27557cf63164184a32957d4d87ae6954")),
		}),
		WithPrintLog(true),
		WithSensitive(true),
	))
	if err != nil {
		t.Fatal(err)
	}
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(bytes))
}
