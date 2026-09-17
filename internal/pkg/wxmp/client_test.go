package wxmp

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type memoryCache struct{ m map[string]string }

func newMemoryCache() *memoryCache { return &memoryCache{m: map[string]string{}} }
func (c *memoryCache) get(key string) (string, error) {
	v, ok := c.m[key]
	if !ok {
		return "", errors.New("miss")
	}
	return v, nil
}
func (c *memoryCache) set(key, value string, _ time.Duration) error {
	c.m[key] = value
	return nil
}

func newTestServer(tokenCalls *int) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/cgi-bin/token":
			*tokenCalls++
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "mp-token", "expires_in": 7200,
			})
		case r.URL.Path == "/wxa/business/getuserphonenumber":
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["code"] == "valid" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"errcode": 0, "errmsg": "ok",
					"phone_info": map[string]interface{}{
						"phoneNumber": "+8613800000001", "purePhoneNumber": "13800000001", "countryCode": "86",
					},
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 40029, "errmsg": "invalid code"})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func TestGetPhoneNumber(t *testing.T) {
	tokenCalls := 0
	srv := newTestServer(&tokenCalls)
	defer srv.Close()
	c := &httpClient{
		cfg:     &Config{AppID: "wx-test", Secret: "s"},
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: srv.URL,
		cache:   newMemoryCache(),
	}

	phone, err := c.GetPhoneNumber("valid")
	if err != nil || phone != "13800000001" {
		t.Fatalf("expected purePhoneNumber, got %q err=%v", phone, err)
	}
	if _, err = c.GetPhoneNumber("expired"); err != ErrInvalidCode {
		t.Fatalf("expected ErrInvalidCode, got %v", err)
	}

	// token 缓存：前两次调用只应请求一次 token
	if tokenCalls != 1 {
		t.Fatalf("expected 1 token call, got %d", tokenCalls)
	}
}

func TestNotConfigured(t *testing.T) {
	c := &httpClient{cfg: &Config{}, http: &http.Client{}, cache: newMemoryCache()}
	if _, err := c.GetPhoneNumber("valid"); err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}
