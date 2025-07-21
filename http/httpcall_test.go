package http

import (
	"net/url"
	"testing"
)

func TestHttpCall_GET(t *testing.T) {
	resp, status, _, err := HttpCall("GET", "https://httpbin.org/get", nil, nil, nil)
	if err != nil || status != 200 {
		t.Errorf("GET failed: %v, status=%d", err, status)
	}
	if len(resp) == 0 {
		t.Error("empty response")
	}
}

func TestHttpCall_POST_JSON(t *testing.T) {
	data := map[string]any{"foo": "bar"}
	resp, status, _, err := HttpCall("POST", "https://httpbin.org/post", data, nil, nil)
	if err != nil || status != 200 {
		t.Errorf("POST JSON failed: %v, status=%d", err, status)
	}
	if len(resp) == 0 {
		t.Error("empty response")
	}
}

func TestHttpCall_POST_FORM(t *testing.T) {
	form := url.Values{"a": {"1"}, "b": {"2"}}
	resp, status, _, err := HttpCall("POST", "https://httpbin.org/post", form, nil, nil)
	if err != nil || status != 200 {
		t.Errorf("POST form failed: %v, status=%d", err, status)
	}
	if len(resp) == 0 {
		t.Error("empty response")
	}
}
