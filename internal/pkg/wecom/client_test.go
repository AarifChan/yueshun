package wecom

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// memoryCache 测试用内存缓存
type memoryCache struct{ m map[string]string }

func newMemoryCache() *memoryCache { return &memoryCache{m: map[string]string{}} }
func (c *memoryCache) get(key string) (string, error) {
	v, ok := c.m[key]
	if !ok {
		return "", errCacheMiss
	}
	return v, nil
}
func (c *memoryCache) set(key, value string, _ time.Duration) error {
	c.m[key] = value
	return nil
}

// newTestServer 按路径分发企业微信 API 模拟响应
func newTestServer(t *testing.T, tokenCalls *int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.URL.Path == "/cgi-bin/gettoken":
			*tokenCalls++
			json.NewEncoder(w).Encode(map[string]interface{}{
				"errcode": 0, "errmsg": "ok", "access_token": "test-token", "expires_in": 7200,
			})
		case r.URL.Path == "/cgi-bin/user/getuserid":
			var body map[string]string
			json.NewDecoder(r.Body).Decode(&body)
			if body["mobile"] == "13800000001" {
				json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0, "errmsg": "ok", "userid": "zhangsan"})
			} else {
				// 手机号不在企业内：errcode 非 0 且无 userid
				json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 60111, "errmsg": "userid not found"})
			}
		case r.URL.Path == "/cgi-bin/auth/getuserinfo":
			switch r.URL.Query().Get("code") {
			case "valid-code":
				json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0, "errmsg": "ok", "userid": "zhangsan"})
			case "external-code": // 外部联系人：返回 openid 而非 userid
				json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 0, "errmsg": "ok", "openid": "oABC"})
			default:
				json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 40029, "errmsg": "invalid code"})
			}
		case r.URL.Path == "/cgi-bin/user/get":
			if r.URL.Query().Get("userid") == "zhangsan" {
				json.NewEncoder(w).Encode(map[string]interface{}{
					"errcode": 0, "errmsg": "ok", "userid": "zhangsan", "name": "张三", "mobile": "13800000001",
				})
			} else {
				json.NewEncoder(w).Encode(map[string]interface{}{"errcode": 60111, "errmsg": "userid not found"})
			}
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newTestClient(baseURL string, cache tokenCache) *httpClient {
	return &httpClient{
		cfg:     &Config{CorpID: "ww-test", AgentID: 1000002, Secret: "secret"},
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: baseURL,
		cache:   cache,
	}
}

func TestGetUserIDByMobile(t *testing.T) {
	tokenCalls := 0
	srv := newTestServer(t, &tokenCalls)
	defer srv.Close()
	c := newTestClient(srv.URL, newMemoryCache())

	id, err := c.GetUserIDByMobile("13800000001")
	if err != nil || id != "zhangsan" {
		t.Fatalf("expected zhangsan, got id=%q err=%v", id, err)
	}

	_, err = c.GetUserIDByMobile("19900000000")
	if err != ErrNotMember {
		t.Fatalf("expected ErrNotMember, got %v", err)
	}
}

func TestGetUserInfoByCode(t *testing.T) {
	tokenCalls := 0
	srv := newTestServer(t, &tokenCalls)
	defer srv.Close()
	c := newTestClient(srv.URL, newMemoryCache())

	id, err := c.GetUserInfoByCode("valid-code")
	if err != nil || id != "zhangsan" {
		t.Fatalf("expected zhangsan, got id=%q err=%v", id, err)
	}
	if _, err = c.GetUserInfoByCode("external-code"); err != ErrNotMember {
		t.Fatalf("expected ErrNotMember for external user, got %v", err)
	}
	if _, err = c.GetUserInfoByCode("bad-code"); err != ErrInvalidCode {
		t.Fatalf("expected ErrInvalidCode, got %v", err)
	}
}

func TestGetUserDetail(t *testing.T) {
	tokenCalls := 0
	srv := newTestServer(t, &tokenCalls)
	defer srv.Close()
	c := newTestClient(srv.URL, newMemoryCache())

	d, err := c.GetUserDetail("zhangsan")
	if err != nil || d.Mobile != "13800000001" || d.Name != "张三" {
		t.Fatalf("unexpected detail %+v err=%v", d, err)
	}
	if _, err = c.GetUserDetail("ghost"); err != ErrNotMember {
		t.Fatalf("expected ErrNotMember, got %v", err)
	}
}

func TestAccessTokenCached(t *testing.T) {
	tokenCalls := 0
	srv := newTestServer(t, &tokenCalls)
	defer srv.Close()
	c := newTestClient(srv.URL, newMemoryCache())

	// 连续两次调用，token 接口只能被调一次（第二次命中缓存）
	if _, err := c.GetUserIDByMobile("13800000001"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.GetUserDetail("zhangsan"); err != nil {
		t.Fatal(err)
	}
	if tokenCalls != 1 {
		t.Fatalf("expected 1 token call, got %d", tokenCalls)
	}
}

func TestNotConfigured(t *testing.T) {
	c := &httpClient{cfg: &Config{}, http: &http.Client{}, cache: newMemoryCache()}
	if _, err := c.GetUserIDByMobile("13800000001"); err != ErrNotConfigured {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}
