# 企业微信登录接入实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为智账系统三端接入企业微信登录——除超级管理员外，内部用户必须通过企业微信身份（Web 扫码 / 小程序手机号校验）登录。

**Architecture:** 后端新增 `internal/pkg/wecom` 与 `internal/pkg/wxmp` 两个接口化 API 客户端包（access_token 走 Redis 缓存，Redis 不可用时跳过缓存直调 API），认证层新增三个端点（config / mp / web），员工模型增加 `wecom_user_id` 绑定字段；Web 管理端登录页改双 Tab + WWLogin 扫码回调页；小程序端重写登录页（getPhoneNumber 一键登录 + 超管密码通道）。

**Tech Stack:** Go 1.26 + Gin + GORM + go-redis；Vue 3 + Element Plus + Vitest（frontend/）；uni-best v4.4.1（uni-app + Vue3 + wot-ui + Vitest，app/）。

**Spec:** `docs/superpowers/specs/2026-09-17-wecom-login-design.md`（实现者必须先读）

## Global Constraints

- 统一响应信封 `{code, message, data}`；成功码 200；业务错误码段 4000-4999，新增 4101/4102/4103。
- 所有第三方凭证只走 `ZHIZHANG_` 前缀环境变量注入；`config.yaml` 仅占位空值；**禁止把真实 secret 提交进仓库**。
- 所有登录路径（含企业微信）统一要求员工 `status = 1`。
- 身份锚点是 `wecom_user_id`，手机号仅作首次绑定桥梁。
- Go 测试用 `github.com/glebarez/sqlite`（纯 Go，无 CGO）内存库 + `httptest`；不引入 testify，用标准库 `t.Fatal/t.Errorf`。
- 前端测试用 Vitest（frontend/ 与 app/ 均已配置）。
- 提交信息格式 `<type>: <description>`，结尾带 `Co-Authored-By: Claude Code <noreply@anthropic.com>`。
- 每个任务的代码必须遵循 TDD：先写失败测试，再实现。

---

### Task 1: wecom 包 —— 企业微信 API 客户端

**Files:**
- Create: `internal/pkg/wecom/types.go`
- Create: `internal/pkg/wecom/token.go`
- Create: `internal/pkg/wecom/client.go`
- Test: `internal/pkg/wecom/client_test.go`

**Interfaces:**
- Consumes: `internal/pkg/redis` 的 `Set/Get` 与全局 `Client`（可为 nil）。
- Produces:
  - `wecom.Config{CorpID string; AgentID int; Secret string; InviteQRURL string; RedirectHost string}`
  - `wecom.UserDetail{UserID string; Name string; Mobile string}`
  - 哨兵错误 `wecom.ErrNotConfigured / ErrNotMember / ErrInvalidCode`；`wecom.APIError{Code int; Msg string}`
  - `wecom.Client` 接口：`GetUserIDByMobile(mobile string) (string, error)`、`GetUserInfoByCode(code string) (string, error)`、`GetUserDetail(userID string) (*UserDetail, error)`
  - `wecom.NewClient(cfg *Config) Client`
- 后续 Task 6 的 handler 依赖以上全部签名。

- [ ] **Step 1: 添加依赖**

```bash
cd /Users/fuqiang/Project/zhizhang-server
go get golang.org/x/sync
```

- [ ] **Step 2: 写失败测试 `internal/pkg/wecom/client_test.go`**

```go
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
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/pkg/wecom/ -v 2>&1 | head -5`
Expected: 编译失败，package 不存在。

- [ ] **Step 4: 实现 `internal/pkg/wecom/types.go`**

```go
package wecom

import (
	"errors"
	"fmt"
)

// Config 企业微信自建应用配置（secret 仅服务端持有）
type Config struct {
	CorpID       string
	AgentID      int
	Secret       string
	InviteQRURL  string // 企业邀请二维码图片地址，用于「未加入企业」引导页
	RedirectHost string // Web 扫码回调域名（须为应用可信域名），为空时前端用 location.origin
}

// UserDetail 企业微信成员详情（本系统关心的子集）
type UserDetail struct {
	UserID string
	Name   string
	Mobile string
}

var (
	// ErrNotConfigured corpid/secret 未配置
	ErrNotConfigured = errors.New("企业微信未配置")
	// ErrNotMember 手机号/授权码对应的用户不是企业成员
	ErrNotMember = errors.New("非企业成员")
	// ErrInvalidCode 授权码无效或已过期
	ErrInvalidCode = errors.New("授权码无效或已过期")
)

// APIError 企业微信 API 返回的业务错误（errcode != 0 且无法归类时）
type APIError struct {
	Code int
	Msg  string
}

func (e *APIError) Error() string { return fmt.Sprintf("wecom api error %d: %s", e.Code, e.Msg) }
```

- [ ] **Step 5: 实现 `internal/pkg/wecom/token.go`**

```go
package wecom

import (
	"errors"
	"time"

	"zhizhang-server/internal/pkg/redis"
)

const tokenCacheKey = "wecom:access_token"

var errCacheMiss = errors.New("cache miss")

// tokenCache access_token 缓存抽象（测试用内存实现注入）
type tokenCache interface {
	get(key string) (string, error)
	set(key, value string, ttl time.Duration) error
}

// redisTokenCache 基于 pkg/redis；redis.Client 为 nil（Redis 可选非阻塞）时退化为 noop
type redisTokenCache struct{}

func (redisTokenCache) get(key string) (string, error) {
	if redis.Client == nil {
		return "", errCacheMiss
	}
	v, err := redis.Get(key)
	if err != nil {
		return "", errCacheMiss
	}
	return v, nil
}

func (redisTokenCache) set(key, value string, ttl time.Duration) error {
	if redis.Client == nil {
		return nil // 跳过缓存，不影响登录链路
	}
	return redis.Set(key, value, ttl)
}
```

- [ ] **Step 6: 实现 `internal/pkg/wecom/client.go`**

```go
package wecom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/sync/singleflight"
)

const defaultBaseURL = "https://qyapi.weixin.qq.com"

// Client 企业微信 API 客户端（接口化便于 handler 测试 mock）
type Client interface {
	// GetUserIDByMobile 按手机号查企业成员 userid；非成员返回 ErrNotMember
	GetUserIDByMobile(mobile string) (string, error)
	// GetUserInfoByCode WWLogin 授权码换成员 userid；外部联系人返回 ErrNotMember，码无效返回 ErrInvalidCode
	GetUserInfoByCode(code string) (string, error)
	// GetUserDetail 按 userid 取成员详情；不存在返回 ErrNotMember
	GetUserDetail(userID string) (*UserDetail, error)
}

type httpClient struct {
	cfg     *Config
	http    *http.Client
	baseURL string
	cache   tokenCache
	sf      singleflight.Group
}

// NewClient 生产构造：真实 API 地址 + Redis 缓存
func NewClient(cfg *Config) Client {
	return &httpClient{
		cfg:     cfg,
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: defaultBaseURL,
		cache:   redisTokenCache{},
	}
}

type errcodeBody struct {
	Errcode int    `json:"errcode"`
	Errmsg  string `json:"errmsg"`
}

func (c *httpClient) accessToken() (string, error) {
	if c.cfg.CorpID == "" || c.cfg.Secret == "" {
		return "", ErrNotConfigured
	}
	if token, err := c.cache.get(tokenCacheKey); err == nil && token != "" {
		return token, nil
	}
	// singleflight：并发刷新只发一次请求
	v, err, _ := c.sf.Do("refresh", func() (interface{}, error) {
		if token, err := c.cache.get(tokenCacheKey); err == nil && token != "" {
			return token, nil
		}
		url := fmt.Sprintf("%s/cgi-bin/gettoken?corpid=%s&corpsecret=%s", c.baseURL, c.cfg.CorpID, c.cfg.Secret)
		resp, err := c.http.Get(url)
		if err != nil {
			return "", fmt.Errorf("gettoken request failed: %w", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("gettoken read body failed: %w", err)
		}
		var r struct {
			errcodeBody
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
		}
		if err := json.Unmarshal(body, &r); err != nil {
			return "", fmt.Errorf("gettoken decode failed: %w", err)
		}
		if r.Errcode != 0 {
			return "", &APIError{Code: r.Errcode, Msg: r.Errmsg}
		}
		// 提前 300s 过期，给在途请求留余量
		ttl := time.Duration(r.ExpiresIn-300) * time.Second
		_ = c.cache.set(tokenCacheKey, r.AccessToken, ttl)
		return r.AccessToken, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (c *httpClient) GetUserIDByMobile(mobile string) (string, error) {
	token, err := c.accessToken()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/cgi-bin/user/getuserid?access_token=%s", c.baseURL, token)
	payload, _ := json.Marshal(map[string]string{"mobile": mobile})
	resp, err := c.http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("getuserid request failed: %w", err)
	}
	defer resp.Body.Close()
	var r struct {
		errcodeBody
		UserID string `json:"userid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("getuserid decode failed: %w", err)
	}
	// 手机号不在企业内时 errcode 非 0 且无 userid，统一归为非成员
	if r.UserID == "" {
		return "", ErrNotMember
	}
	return r.UserID, nil
}

func (c *httpClient) GetUserInfoByCode(code string) (string, error) {
	token, err := c.accessToken()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/cgi-bin/auth/getuserinfo?access_token=%s&code=%s", c.baseURL, token, code)
	resp, err := c.http.Get(url)
	if err != nil {
		return "", fmt.Errorf("getuserinfo request failed: %w", err)
	}
	defer resp.Body.Close()
	var r struct {
		errcodeBody
		UserID string `json:"userid"`
		OpenID string `json:"openid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("getuserinfo decode failed: %w", err)
	}
	if r.Errcode == 40029 || r.Errcode == 40001 {
		return "", ErrInvalidCode
	}
	// 外部联系人 errcode=0 但只返回 openid/external_userid
	if r.UserID == "" {
		return "", ErrNotMember
	}
	return r.UserID, nil
}

func (c *httpClient) GetUserDetail(userID string) (*UserDetail, error) {
	token, err := c.accessToken()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/cgi-bin/user/get?access_token=%s&userid=%s", c.baseURL, token, userID)
	resp, err := c.http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("user/get request failed: %w", err)
	}
	defer resp.Body.Close()
	var r struct {
		errcodeBody
		UserID string `json:"userid"`
		Name   string `json:"name"`
		Mobile string `json:"mobile"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, fmt.Errorf("user/get decode failed: %w", err)
	}
	if r.Errcode != 0 || r.UserID == "" {
		return nil, ErrNotMember
	}
	return &UserDetail{UserID: r.UserID, Name: r.Name, Mobile: r.Mobile}, nil
}
```

- [ ] **Step 7: 运行测试确认通过**

Run: `go test ./internal/pkg/wecom/ -v -race`
Expected: 5 个测试全部 PASS。

- [ ] **Step 8: Commit**

```bash
git add internal/pkg/wecom/ go.mod go.sum
git commit -m "feat: add wecom API client package

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 2: wxmp 包 —— 微信小程序 API 客户端

**Files:**
- Create: `internal/pkg/wxmp/types.go`
- Create: `internal/pkg/wxmp/token.go`
- Create: `internal/pkg/wxmp/client.go`
- Test: `internal/pkg/wxmp/client_test.go`

**Interfaces:**
- Consumes: `internal/pkg/redis`。
- Produces:
  - `wxmp.Config{AppID string; Secret string}`
  - 哨兵错误 `wxmp.ErrNotConfigured / ErrInvalidCode`；`wxmp.APIError`
  - `wxmp.Client` 接口：`GetPhoneNumber(code string) (purePhone string, err error)`
  - `wxmp.NewClient(cfg *Config) Client`
- Task 6 依赖 `wxmp.Client` 与错误哨兵。

- [ ] **Step 1: 写失败测试 `internal/pkg/wxmp/client_test.go`**

```go
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

func newTestClient(baseURL string, tokenCalls *int) *httpClient {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	// 注意：测试服务器生命周期交给调用方 Close 的场景更常见，这里直接挂在 client 上不优雅；
	// 为简单起见让调用方通过返回值自己管理——本文件只有一个测试服务器场景。
	return &httpClient{
		cfg:     &Config{AppID: "wx-test", Secret: "secret"},
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: baseURL,
		cache:   newMemoryCache(),
	}
}
```

等一下——上面的写法把 server 创建放进了 client 构造里，server 无法 Close。**改用显式拆分**（实现者按此修正版写）：

```go
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
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/pkg/wxmp/ -v 2>&1 | head -5`
Expected: 编译失败。

- [ ] **Step 3: 实现 `internal/pkg/wxmp/types.go`**

```go
package wxmp

import (
	"errors"
	"fmt"
)

// Config 微信小程序配置（secret 仅服务端持有）
type Config struct {
	AppID  string
	Secret string
}

var (
	ErrNotConfigured = errors.New("微信小程序未配置")
	ErrInvalidCode   = errors.New("授权码无效或已过期")
)

// APIError 微信 API 返回的业务错误
type APIError struct {
	Code int
	Msg  string
}

func (e *APIError) Error() string { return fmt.Sprintf("wxmp api error %d: %s", e.Code, e.Msg) }
```

- [ ] **Step 4: 实现 `internal/pkg/wxmp/token.go`**

```go
package wxmp

import (
	"errors"
	"time"

	"zhizhang-server/internal/pkg/redis"
)

const tokenCacheKey = "wxmp:access_token"

var errCacheMiss = errors.New("cache miss")

type tokenCache interface {
	get(key string) (string, error)
	set(key, value string, ttl time.Duration) error
}

// redisTokenCache 基于 pkg/redis；redis.Client 为 nil 时退化为不缓存
type redisTokenCache struct{}

func (redisTokenCache) get(key string) (string, error) {
	if redis.Client == nil {
		return "", errCacheMiss
	}
	v, err := redis.Get(key)
	if err != nil {
		return "", errCacheMiss
	}
	return v, nil
}

func (redisTokenCache) set(key, value string, ttl time.Duration) error {
	if redis.Client == nil {
		return nil
	}
	return redis.Set(key, value, ttl)
}
```

- [ ] **Step 5: 实现 `internal/pkg/wxmp/client.go`**

```go
package wxmp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/sync/singleflight"
)

const defaultBaseURL = "https://api.weixin.qq.com"

// Client 微信小程序 API 客户端
type Client interface {
	// GetPhoneNumber 用 getPhoneNumber 的 code 换手机号，返回不带区号的 purePhoneNumber
	GetPhoneNumber(code string) (string, error)
}

type httpClient struct {
	cfg     *Config
	http    *http.Client
	baseURL string
	cache   tokenCache
	sf      singleflight.Group
}

func NewClient(cfg *Config) Client {
	return &httpClient{
		cfg:     cfg,
		http:    &http.Client{Timeout: 10 * time.Second},
		baseURL: defaultBaseURL,
		cache:   redisTokenCache{},
	}
}

func (c *httpClient) accessToken() (string, error) {
	if c.cfg.AppID == "" || c.cfg.Secret == "" {
		return "", ErrNotConfigured
	}
	if token, err := c.cache.get(tokenCacheKey); err == nil && token != "" {
		return token, nil
	}
	v, err, _ := c.sf.Do("refresh", func() (interface{}, error) {
		if token, err := c.cache.get(tokenCacheKey); err == nil && token != "" {
			return token, nil
		}
		url := fmt.Sprintf("%s/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", c.baseURL, c.cfg.AppID, c.cfg.Secret)
		resp, err := c.http.Get(url)
		if err != nil {
			return "", fmt.Errorf("token request failed: %w", err)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("token read body failed: %w", err)
		}
		var r struct {
			AccessToken string `json:"access_token"`
			ExpiresIn   int    `json:"expires_in"`
			Errcode     int    `json:"errcode"`
			Errmsg      string `json:"errmsg"`
		}
		if err := json.Unmarshal(body, &r); err != nil {
			return "", fmt.Errorf("token decode failed: %w", err)
		}
		if r.Errcode != 0 || r.AccessToken == "" {
			return "", &APIError{Code: r.Errcode, Msg: r.Errmsg}
		}
		ttl := time.Duration(r.ExpiresIn-300) * time.Second
		_ = c.cache.set(tokenCacheKey, r.AccessToken, ttl)
		return r.AccessToken, nil
	})
	if err != nil {
		return "", err
	}
	return v.(string), nil
}

func (c *httpClient) GetPhoneNumber(code string) (string, error) {
	token, err := c.accessToken()
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/wxa/business/getuserphonenumber?access_token=%s", c.baseURL, token)
	payload, _ := json.Marshal(map[string]string{"code": code})
	resp, err := c.http.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("getuserphonenumber request failed: %w", err)
	}
	defer resp.Body.Close()
	var r struct {
		Errcode   int    `json:"errcode"`
		Errmsg    string `json:"errmsg"`
		PhoneInfo struct {
			PhoneNumber     string `json:"phoneNumber"`
			PurePhoneNumber string `json:"purePhoneNumber"`
			CountryCode     string `json:"countryCode"`
		} `json:"phone_info"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", fmt.Errorf("getuserphonenumber decode failed: %w", err)
	}
	if r.Errcode != 0 {
		// 40001/40029：code 无效或已过期（一次性、5 分钟有效）
		return "", ErrInvalidCode
	}
	if r.PhoneInfo.PurePhoneNumber == "" {
		return "", &APIError{Code: r.Errcode, Msg: "empty purePhoneNumber"}
	}
	return r.PhoneInfo.PurePhoneNumber, nil
}
```

- [ ] **Step 6: 运行测试确认通过**

Run: `go test ./internal/pkg/wxmp/ -v -race`
Expected: 2 个测试 PASS。

- [ ] **Step 7: Commit**

```bash
git add internal/pkg/wxmp/
git commit -m "feat: add wxmp API client package

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 3: response 业务码 + Employee 模型 + 迁移索引

**Files:**
- Modify: `internal/pkg/response/response.go`（codes 常量区，约 line 26-30）
- Modify: `internal/model/base.go`（Employee 结构，line 24-36）
- Modify: `cmd/server/main.go`（`autoMigrate` 函数尾部，约 line 143 起）
- Test: `cmd/server/main_test.go`

**Interfaces:**
- Produces: `response.CodeNeedWecom = 4101`、`response.CodeNotWecomMember = 4102`、`response.CodeEmpNotRegistered = 4103`（Task 5/6/7 依赖）；`model.Employee.WecomUserID string`；`autoMigrate` 行为变化（新增复合唯一索引）。
- Consumes: 现有 `model.Employee`、`model.BaseModelWithCompany`。

- [ ] **Step 1: 添加 sqlite 测试依赖**

```bash
cd /Users/fuqiang/Project/zhizhang-server
go get github.com/glebarez/sqlite
```

- [ ] **Step 2: 写失败测试 `cmd/server/main_test.go`**

```go
package main

import (
	"testing"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
)

// 验证 autoMigrate 后：employees 有 wecom_user_id 列，且 (company_id, phone) 复合唯一索引生效
func TestAutoMigrateEmployee(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:wecom-migrate-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := autoMigrate(db); err != nil {
		t.Fatal(err)
	}

	// 密码可空：不传密码应能建员工
	emp := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: 1},
		DeptID: 1, RoleID: 1, Username: "u1", Name: "甲", Phone: "13800000001", Status: 1,
	}
	if err := db.Create(&emp).Error; err != nil {
		t.Fatalf("create without password should succeed: %v", err)
	}

	// 同公司同手机号 → 唯一索引冲突
	dup := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: 1},
		DeptID: 1, RoleID: 1, Username: "u2", Name: "乙", Phone: "13800000001", Status: 1,
	}
	if err := db.Create(&dup).Error; err == nil {
		t.Fatal("expected unique index violation on (company_id, phone), got nil")
	}

	// 不同公司同手机号 → 允许
	other := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: 2},
		DeptID: 1, RoleID: 1, Username: "u3", Name: "丙", Phone: "13800000001", Status: 1,
	}
	if err := db.Create(&other).Error; err != nil {
		t.Fatalf("same phone in another company should be allowed: %v", err)
	}

	// wecom_user_id 可写入
	if err := db.Model(&emp).Update("wecom_user_id", "zhangsan").Error; err != nil {
		t.Fatal(err)
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./cmd/server/ -run TestAutoMigrateEmployee -v 2>&1 | head -15`
Expected: FAIL —— 同公司同手机号未触发唯一冲突（索引还不存在）。

- [ ] **Step 4: 修改 `internal/pkg/response/response.go`**

在 codes 常量区追加：

```go
	CodeValidationError = 4000
	CodeBizError        = 4001
	CodeDuplicate       = 4002
	CodeNotExist        = 4003
	CodeNoPermission    = 4004

	// 企业微信登录 41xx
	CodeNeedWecom        = 4101 // 非超管尝试密码登录
	CodeNotWecomMember   = 4102 // 非企业成员
	CodeEmpNotRegistered = 4103 // 企业成员但系统未建档
```

- [ ] **Step 5: 修改 `internal/model/base.go` Employee**

```go
type Employee struct {
	BaseModelWithCompany
	DeptID       uint       `json:"deptId" gorm:"index;not null"`
	RoleID       uint       `json:"roleId" gorm:"index;not null"`
	Username     string     `json:"username" gorm:"size:64;uniqueIndex:idx_emp_username_company"`
	Password     string     `json:"-" gorm:"size:128"` // 企业微信登录员工无密码
	Name         string     `json:"name" gorm:"size:64;not null"`
	Phone        string     `json:"phone" gorm:"size:20;index"`
	Email        string     `json:"email" gorm:"size:128"`
	WecomUserID  string     `json:"wecomUserId" gorm:"size:64;index"` // 企业微信成员 userid，首次登录绑定
	Status       int8       `json:"status" gorm:"default:1;comment:1启用 0禁用"`
	LastLoginAt  *time.Time `json:"lastLoginAt"`
	LastLoginIP  string     `json:"lastLoginIp" gorm:"size:64"`
}
```

改动点：`Password` 去掉 `not null`；新增 `WecomUserID`；其余不动。（Phone 的唯一约束由迁移 SQL 建复合索引，不打 GORM 唯一 tag —— 避免污染共享的 BaseModelWithCompany。）

- [ ] **Step 6: 修改 `cmd/server/main.go` 的 `autoMigrate`**

在 `autoMigrate` 函数中 `return db.AutoMigrate(...)` 改为先捕获再补索引：

```go
func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Company{},
		// ……现有全部模型保持不动……
	); err != nil {
		return err
	}
	// 企业微信绑定：(公司, 手机号) 复合唯一索引
	// 不用 GORM tag 的原因：CompanyID 在共享的 BaseModelWithCompany 中，打 tag 会污染所有租户模型
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_emp_phone_company ON employees (company_id, phone)").Error
}
```

- [ ] **Step 7: 运行测试确认通过**

Run: `go test ./cmd/server/ -run TestAutoMigrateEmployee -v -race`
Expected: PASS。

- [ ] **Step 8: Commit**

```bash
git add internal/pkg/response/response.go internal/model/base.go cmd/server/ go.mod go.sum
git commit -m "feat: add wecom login business codes and employee binding field

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 4: viper 环境变量修复 + 配置节 + .env.example

**Files:**
- Modify: `cmd/server/main.go`（Config 结构 line 23-45 区域；loadConfig 函数 line 105-127）
- Modify: `configs/config.yaml`
- Modify: `.env.example`
- Test: `cmd/server/config_test.go`

**Interfaces:**
- Produces: `Config.Wecom{CorpID, AgentID, Secret, InviteQRURL, RedirectHost}`、`Config.WechatMP{AppID, Secret}`；env 变量 `ZHIZHANG_WECOM_*` / `ZHIZHANG_WECHAT_MP_*` 真正生效。
- Consumes: 现有 viper 加载流程。Task 6 的 main 接线依赖这些字段。

- [ ] **Step 1: 写失败测试 `cmd/server/config_test.go`**

```go
package main

import (
	"os"
	"testing"
)

// 验证 ZHIZHANG_ 前缀环境变量能覆盖嵌套配置（需要 SetEnvKeyReplacer）
func TestLoadConfigFromEnv(t *testing.T) {
	// config.yaml 在仓库根目录 configs/ 下，loadConfig 的相对路径基于仓库根
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	t.Setenv("ZHIZHANG_WECOM_CORPID", "ww_from_env")
	t.Setenv("ZHIZHANG_WECOM_AGENTID", "1000002")
	t.Setenv("ZHIZHANG_WECOM_SECRET", "secret_from_env")
	t.Setenv("ZHIZHANG_WECOM_INVITE_QR_URL", "https://example.com/qr.png")
	t.Setenv("ZHIZHANG_WECHAT_MP_APPID", "wx_from_env")

	cfg, err := loadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Wecom.CorpID != "ww_from_env" {
		t.Fatalf("env override failed, got corpid=%q", cfg.Wecom.CorpID)
	}
	if cfg.Wecom.AgentID != 1000002 {
		t.Fatalf("env override failed, got agentid=%d", cfg.Wecom.AgentID)
	}
	if cfg.Wecom.InviteQRURL != "https://example.com/qr.png" {
		t.Fatalf("invite qr url mismatch: %q", cfg.Wecom.InviteQRURL)
	}
	if cfg.WechatMP.AppID != "wx_from_env" {
		t.Fatalf("wxmp appid mismatch: %q", cfg.WechatMP.AppID)
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./cmd/server/ -run TestLoadConfigFromEnv -v 2>&1 | head -10`
Expected: FAIL —— `Wecom` 字段不存在（编译错误）或环境变量不生效。

- [ ] **Step 3: 修改 `cmd/server/main.go`**

Config 结构追加（紧跟 `Redis` 字段之后）：

```go
	Wecom struct {
		CorpID       string `mapstructure:"corpid"`
		AgentID      int    `mapstructure:"agentid"`
		Secret       string `mapstructure:"secret"`
		InviteQRURL  string `mapstructure:"invite_qr_url"`
		RedirectHost string `mapstructure:"redirect_host"`
	} `mapstructure:"wecom"`
	WechatMP struct {
		AppID  string `mapstructure:"appid"`
		Secret string `mapstructure:"secret"`
	} `mapstructure:"wechat_mp"`
```

`loadConfig` 中 `viper.AutomaticEnv()` 之前加 key replacer（文件 import 需加 `"strings"`）：

```go
	viper.SetEnvPrefix("ZHIZHANG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // ZHIZHANG_WECOM_CORPID → wecom.corpid
	viper.AutomaticEnv()
```

- [ ] **Step 4: `configs/config.yaml` 追加占位节**

```yaml
# 企业微信（自建应用）—— 真实值一律用环境变量注入，勿写进仓库
wecom:
  corpid: ""
  agentid: 0
  secret: ""
  invite_qr_url: ""
  redirect_host: ""

# 微信小程序 —— 真实值一律用环境变量注入
wechat_mp:
  appid: ""
  secret: ""
```

- [ ] **Step 5: `.env.example` 追加**

```bash
# 企业微信（自建应用，见 docs/wecom-setup.md）
ZHIZHANG_WECOM_CORPID=
ZHIZHANG_WECOM_AGENTID=
ZHIZHANG_WECOM_SECRET=
ZHIZHANG_WECOM_INVITE_QR_URL=
ZHIZHANG_WECOM_REDIRECT_HOST=

# 微信小程序（getPhoneNumber 换手机号）
ZHIZHANG_WECHAT_MP_APPID=
ZHIZHANG_WECHAT_MP_SECRET=
```

- [ ] **Step 6: 运行测试确认通过**

Run: `go test ./cmd/server/ -v -race`
Expected: 本测试与 Task 3 测试都 PASS。

- [ ] **Step 7: Commit**

```bash
git add cmd/server/ configs/config.yaml .env.example
git commit -m "feat: add wecom/wxmp config sections and fix viper env key replacer

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 5: auth.go —— issueTokenPair 提取 + 超管密码登录门

**Files:**
- Modify: `internal/api/handler/auth.go`（Login、Refresh）
- Test: `internal/api/handler/testhelper_test.go`（新建）、`internal/api/handler/auth_test.go`（新建）

**Interfaces:**
- Produces: 包级函数 `issueTokenPair(emp *model.Employee) (*LoginResp, error)`（Task 6 复用）；`setupTestDB(t) *gorm.DB`、`seedTenant(t, db)`、`performRequest(...)` 测试辅助（Task 6/7 复用）。
- Consumes: Task 3 的 `response.CodeNeedWecom`；现有 `middleware.GenerateToken/GenerateRefreshToken/GetAccessTTL/GetRefreshTTL`。

- [ ] **Step 1: 写测试辅助 `internal/api/handler/testhelper_test.go`**

```go
package handler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
)

func init() {
	gin.SetMode(gin.TestMode)
	middleware.InitJWT(&middleware.JWTConfig{Secret: "test-secret", AccessTTL: 7200, RefreshTTL: 604800})
}

// setupTestDB 每个测试用独立的内存库名，避免 cache=shared 串数据
func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:testdb-%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(
		&model.Company{}, &model.Department{}, &model.Role{}, &model.Employee{},
	); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_emp_phone_company ON employees (company_id, phone)").Error; err != nil {
		t.Fatal(err)
	}
	return db
}

// seedTenant 建 公司/部门/两个角色，返回（公司, 部门, 超管角色, 普通角色）
func seedTenant(t *testing.T, db *gorm.DB) (model.Company, model.Department, model.Role, model.Role) {
	t.Helper()
	company := model.Company{Name: "测试公司", Code: "T"}
	if err := db.Create(&company).Error; err != nil {
		t.Fatal(err)
	}
	dept := model.Department{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		Name:                 "总部", Code: "HQ",
	}
	if err := db.Create(&dept).Error; err != nil {
		t.Fatal(err)
	}
	superRole := model.Role{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		Name:                 "超级管理员", Code: "super_admin",
	}
	staffRole := model.Role{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		Name:                 "销售员", Code: "staff",
	}
	if err := db.Create(&superRole).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&staffRole).Error; err != nil {
		t.Fatal(err)
	}
	return company, dept, superRole, staffRole
}

// performRequest 发 JSON 请求并返回 recorder
func performRequest(r http.Handler, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	var reader *bytes.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req, _ := http.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	r.ServeHTTP(w, req)
	return w
}

// parseBody 解析统一响应信封
func parseBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var m map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &m); err != nil {
		t.Fatalf("decode response failed: %v, body=%s", err, w.Body.String())
	}
	return m
}
```

- [ ] **Step 2: 写失败测试 `internal/api/handler/auth_test.go`**

```go
package handler

import (
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

func setupAuthRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	v1 := r.Group("/api/v1")
	NewAuthHandler(db).RegisterRoutes(v1)
	return r
}

func TestLoginPasswordGate(t *testing.T) {
	db := setupTestDB(t)
	company, dept, superRole, staffRole := seedTenant(t, db)

	hashed, _ := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.DefaultCost)
	// 超管
	admin := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		DeptID: dept.ID, RoleID: superRole.ID,
		Username: "admin", Password: string(hashed), Name: "管理员", Phone: "13800000001", Status: 1,
	}
	// 普通员工
	staff := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		DeptID: dept.ID, RoleID: staffRole.ID,
		Username: "staff", Password: string(hashed), Name: "小张", Phone: "13800000002", Status: 1,
	}
	db.Create(&admin)
	db.Create(&staff)

	r := setupAuthRouter(db)

	// 超管密码登录 → 放行
	w := performRequest(r, "POST", "/api/v1/auth/login", map[string]string{"username": "admin", "password": "123456"}, nil)
	resp := parseBody(t, w)
	if resp["code"].(float64) != 200 {
		t.Fatalf("super admin login should succeed, got %v", resp)
	}
	data := resp["data"].(map[string]interface{})
	if data["accessToken"] == "" || data["refreshToken"] == "" {
		t.Fatal("expected token pair")
	}
	if data["accessExpiresIn"].(float64) != 7200 {
		t.Fatalf("expected accessExpiresIn 7200, got %v", data["accessExpiresIn"])
	}

	// 普通员工密码登录 → 4101
	w = performRequest(r, "POST", "/api/v1/auth/login", map[string]string{"username": "staff", "password": "123456"}, nil)
	resp = parseBody(t, w)
	if resp["code"].(float64) != response.CodeNeedWecom {
		t.Fatalf("expected 4101, got %v", resp)
	}
}
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/api/handler/ -run TestLoginPasswordGate -v 2>&1 | head -10`
Expected: FAIL —— staff 登录返回 200 而非 4101。

- [ ] **Step 4: 修改 `internal/api/handler/auth.go`**

提取包级签发函数（放在 DTO 定义之后）：

```go
// issueTokenPair 为员工签发 access/refresh 令牌对（密码登录、刷新、企业微信登录共用）
func issueTokenPair(emp *model.Employee) (*LoginResp, error) {
	accessToken, err := middleware.GenerateToken(emp.ID, emp.Username, emp.RoleID, emp.DeptID, emp.CompanyID)
	if err != nil {
		return nil, err
	}
	refreshToken, err := middleware.GenerateRefreshToken(emp.ID)
	if err != nil {
		return nil, err
	}
	return &LoginResp{
		AccessToken:      accessToken,
		RefreshToken:     refreshToken,
		AccessExpiresIn:  middleware.GetAccessTTL(),
		RefreshExpiresIn: middleware.GetRefreshTTL(),
		User: UserInfo{
			ID: emp.ID, Username: emp.Username, Name: emp.Name, Phone: emp.Phone,
			RoleID: emp.RoleID, DeptID: emp.DeptID,
		},
	}, nil
}
```

`Login` 中：密码校验通过后、签发令牌前插入角色门；签发段替换为 `issueTokenPair`（保留 LastLoginAt/IP 更新）：

```go
	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(emp.Password), []byte(req.Password)); err != nil {
		response.Fail(c, response.CodeUnauthorized, "用户名或密码错误")
		return
	}

	// 密码登录仅超级管理员可用，其余角色必须走企业微信登录
	var role model.Role
	if err := h.db.First(&role, emp.RoleID).Error; err != nil || role.Code != "super_admin" {
		response.Fail(c, response.CodeNeedWecom, "当前账号请使用企业微信登录")
		return
	}

	// 更新登录时间
	now := time.Now()
	emp.LastLoginAt = &now
	emp.LastLoginIP = c.ClientIP()
	h.db.Save(&emp)

	resp, err := issueTokenPair(&emp)
	if err != nil {
		log.Error().Err(err).Msg("issue token pair failed")
		response.ServerError(c, "令牌生成失败")
		return
	}
	response.Ok(c, resp)
```

`Refresh` 中签发段同样替换为 `resp, err := issueTokenPair(&emp)` + 错误处理，删掉原有的两段 Generate 调用和内联 LoginResp 构造。

- [ ] **Step 5: 运行测试确认通过**

Run: `go test ./internal/api/handler/ -v -race`
Expected: PASS；`go build ./...` 通过。

- [ ] **Step 6: Commit**

```bash
git add internal/api/handler/
git commit -m "feat: restrict password login to super admin and extract issueTokenPair

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 6: auth_wecom.go —— 企业微信登录三端点

**Files:**
- Create: `internal/api/handler/auth_wecom.go`
- Test: `internal/api/handler/auth_wecom_test.go`
- Modify: `internal/pkg/response/response.go`（加 `FailWithData`）
- Modify: `internal/api/router.go`（SetupRouter 签名 + 挂载）
- Modify: `cmd/server/main.go`（调用点传配置）

**Interfaces:**
- Consumes: `wecom.Client`/`wxmp.Client`（Task 1/2）；`issueTokenPair`、`LoginResp`（Task 5）；`response.CodeNotWecomMember/CodeEmpNotRegistered`（Task 3）；`Config.Wecom/WechatMP`（Task 4）；测试辅助（Task 5）。
- Produces:
  - `NewWecomAuthHandler(db *gorm.DB, wc wecom.Client, wx wxmp.Client, cfg *wecom.Config) *WecomAuthHandler`
  - `SetupRouter(db *gorm.DB, jwtCfg *middleware.JWTConfig, wecomCfg *wecom.Config, wxmpCfg *wxmp.Config) *gin.Engine`
  - `response.FailWithData(c *gin.Context, code int, message string, data interface{})`

- [ ] **Step 1: `internal/pkg/response/response.go` 加 `FailWithData`**

在现有 `Fail` 函数旁追加：

```go
// FailWithData 失败响应（带附加数据，如引导二维码地址）
func FailWithData(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{Code: code, Message: message, Data: data})
}
```

（若现有 `Fail` 实现不是 `http.StatusOK`，以现有 `Fail` 的状态码为准，保持一致。）

- [ ] **Step 2: 写失败测试 `internal/api/handler/auth_wecom_test.go`**

```go
package handler

import (
	"errors"
	"testing"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
	"zhizhang-server/internal/pkg/wecom"
	"zhizhang-server/internal/pkg/wxmp"
)

// --- mock 客户端 ---

type fakeWecom struct {
	byMobile   map[string]string // mobile → userid（不存在则 ErrNotMember）
	byCode     map[string]string // auth code → userid
	details    map[string]*wecom.UserDetail
}

func (f *fakeWecom) GetUserIDByMobile(mobile string) (string, error) {
	if id, ok := f.byMobile[mobile]; ok {
		return id, nil
	}
	return "", wecom.ErrNotMember
}
func (f *fakeWecom) GetUserInfoByCode(code string) (string, error) {
	if id, ok := f.byCode[code]; ok {
		return id, nil
	}
	return "", wecom.ErrNotMember
}
func (f *fakeWecom) GetUserDetail(userID string) (*wecom.UserDetail, error) {
	if d, ok := f.details[userID]; ok {
		return d, nil
	}
	return nil, wecom.ErrNotMember
}

type fakeWxmp struct {
	byCode map[string]string // phone code → 手机号
}

func (f *fakeWxmp) GetPhoneNumber(code string) (string, error) {
	if p, ok := f.byCode[code]; ok {
		return p, nil
	}
	return "", wxmp.ErrInvalidCode
}

// --- 装配 ---

func setupWecomRouter(db *gorm.DB, wc wecom.Client, wx wxmp.Client) *gin.Engine {
	r := gin.New()
	v1 := r.Group("/api/v1")
	cfg := &wecom.Config{CorpID: "ww-test", AgentID: 1000002, InviteQRURL: "https://example.com/qr.png"}
	NewWecomAuthHandler(db, wc, wx, cfg).RegisterRoutes(v1)
	return r
}

func seedStaff(t *testing.T, db *gorm.DB, phone string) model.Employee {
	t.Helper()
	company, dept, _, staffRole := seedTenant(t, db)
	emp := model.Employee{
		BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: company.ID},
		DeptID: dept.ID, RoleID: staffRole.ID,
		Username: "staff-" + phone, Name: "员工" + phone[len(phone)-4:], Phone: phone, Status: 1,
	}
	if err := db.Create(&emp).Error; err != nil {
		t.Fatal(err)
	}
	return emp
}

// --- 用例 ---

func TestWecomConfig(t *testing.T) {
	db := setupTestDB(t)
	r := setupWecomRouter(db, &fakeWecom{}, &fakeWxmp{})
	w := performRequest(r, "GET", "/api/v1/auth/wecom/config", nil, nil)
	resp := parseBody(t, w)
	data := resp["data"].(map[string]interface{})
	if data["corpid"] != "ww-test" || data["inviteQrUrl"] == "" {
		t.Fatalf("unexpected config resp: %v", resp)
	}
}

func TestMpLoginFirstBind(t *testing.T) {
	db := setupTestDB(t)
	emp := seedStaff(t, db, "13800000001")
	wc := &fakeWecom{byMobile: map[string]string{"13800000001": "wx_zhangsan"}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "13800000001"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	resp := parseBody(t, w)
	if resp["code"].(float64) != 200 {
		t.Fatalf("mp login should succeed, got %v", resp)
	}
	// 首次登录回写 wecom_user_id
	var reloaded model.Employee
	db.First(&reloaded, emp.ID)
	if reloaded.WecomUserID != "wx_zhangsan" {
		t.Fatalf("expected wecom_user_id bound, got %q", reloaded.WecomUserID)
	}
}

func TestMpLoginBoundUserSkipsPhoneMatch(t *testing.T) {
	db := setupTestDB(t)
	emp := seedStaff(t, db, "13800000001")
	db.Model(&emp).Update("wecom_user_id", "wx_zhangsan")
	// 成员换了手机号：企业微信侧手机号变了，按 wecom_user_id 直接命中并更新 phone
	wc := &fakeWecom{byMobile: map[string]string{"13900000009": "wx_zhangsan"}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "13900000009"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("bound user should login by wecom_user_id, got %v", resp)
	}
	var reloaded model.Employee
	db.First(&reloaded, emp.ID)
	if reloaded.Phone != "13900000009" {
		t.Fatalf("phone should sync from wecom, got %q", reloaded.Phone)
	}
}

func TestMpLoginNotMember(t *testing.T) {
	db := setupTestDB(t)
	wc := &fakeWecom{byMobile: map[string]string{}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "19900000000"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	resp := parseBody(t, w)
	if resp["code"].(float64) != response.CodeNotWecomMember {
		t.Fatalf("expected 4102, got %v", resp)
	}
	if resp["data"].(map[string]interface{})["inviteQrUrl"] == "" {
		t.Fatal("4102 should carry inviteQrUrl")
	}
}

func TestMpLoginNotRegistered(t *testing.T) {
	db := setupTestDB(t)
	seedTenant(t, db) // 有租户但无该手机号的员工
	wc := &fakeWecom{byMobile: map[string]string{"13800000003": "wx_lisi"}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "13800000003"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeEmpNotRegistered {
		t.Fatalf("expected 4103, got %v", resp)
	}
}

func TestMpLoginDisabled(t *testing.T) {
	db := setupTestDB(t)
	emp := seedStaff(t, db, "13800000004")
	db.Model(&emp).Update("status", 0)
	wc := &fakeWecom{byMobile: map[string]string{"13800000004": "wx_wangwu"}}
	wx := &fakeWxmp{byCode: map[string]string{"phone-code": "13800000004"}}
	r := setupWecomRouter(db, wc, wx)

	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "phone-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeForbidden {
		t.Fatalf("expected 403, got %v", resp)
	}
}

func TestMpLoginInvalidCode(t *testing.T) {
	db := setupTestDB(t)
	r := setupWecomRouter(db, &fakeWecom{}, &fakeWxmp{byCode: map[string]string{}})
	w := performRequest(r, "POST", "/api/v1/auth/wecom/mp", map[string]string{"code": "expired"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeBadRequest {
		t.Fatalf("expected 400, got %v", resp)
	}
}

func TestWebLoginByUserID(t *testing.T) {
	db := setupTestDB(t)
	emp := seedStaff(t, db, "13800000005")
	db.Model(&emp).Update("wecom_user_id", "wx_zhaoliu")
	wc := &fakeWecom{
		byCode:  map[string]string{"web-code": "wx_zhaoliu"},
		details: map[string]*wecom.UserDetail{"wx_zhaoliu": {UserID: "wx_zhaoliu", Name: "赵六", Mobile: "13800000005"}},
	}
	r := setupWecomRouter(db, wc, &fakeWxmp{})

	w := performRequest(r, "POST", "/api/v1/auth/wecom/web", map[string]string{"code": "web-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("web login should succeed, got %v", resp)
	}
}

func TestWebLoginFallbackBindByMobile(t *testing.T) {
	db := setupTestDB(t)
	emp := seedStaff(t, db, "13800000006")
	wc := &fakeWecom{
		byCode:  map[string]string{"web-code": "wx_qiqi"},
		details: map[string]*wecom.UserDetail{"wx_qiqi": {UserID: "wx_qiqi", Name: "七七", Mobile: "13800000006"}},
	}
	r := setupWecomRouter(db, wc, &fakeWxmp{})

	w := performRequest(r, "POST", "/api/v1/auth/wecom/web", map[string]string{"code": "web-code"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("web login fallback should succeed, got %v", resp)
	}
	var reloaded model.Employee
	db.First(&reloaded, emp.ID)
	if reloaded.WecomUserID != "wx_qiqi" {
		t.Fatalf("expected wecom_user_id bound, got %q", reloaded.WecomUserID)
	}
}

func TestWebLoginExternalUser(t *testing.T) {
	db := setupTestDB(t)
	wc := &fakeWecom{byCode: map[string]string{}} // 任何 code 都 ErrNotMember
	r := setupWecomRouter(db, wc, &fakeWxmp{})
	w := performRequest(r, "POST", "/api/v1/auth/wecom/web", map[string]string{"code": "ext"}, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeNotWecomMember {
		t.Fatalf("expected 4102, got %v", resp)
	}
}

// 未配置时 config 端点返回 4001
func TestWecomConfigNotConfigured(t *testing.T) {
	db := setupTestDB(t)
	r := gin.New()
	v1 := r.Group("/api/v1")
	NewWecomAuthHandler(db, &fakeWecom{}, &fakeWxmp{}, &wecom.Config{}).RegisterRoutes(v1)
	w := performRequest(r, "GET", "/api/v1/auth/wecom/config", nil, nil)
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeBizError {
		t.Fatalf("expected 4001, got %v", resp)
	}
}

var _ = errors.Is // 保留给实现侧 stdlib errors 使用说明：注意与 internal/pkg/errors 区分
```

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/api/handler/ -run TestWecom -v 2>&1 | head -5`
Expected: 编译失败，`NewWecomAuthHandler` 未定义。

- [ ] **Step 4: 实现 `internal/api/handler/auth_wecom.go`**

```go
package handler

import (
	"errors" // 注意：这里是标准库 errors（errors.Is），不是 internal/pkg/errors

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
	"zhizhang-server/internal/pkg/wecom"
	"zhizhang-server/internal/pkg/wxmp"
)

// WecomAuthHandler 企业微信登录处理器
type WecomAuthHandler struct {
	db    *gorm.DB
	wecom wecom.Client
	wxmp  wxmp.Client
	cfg   *wecom.Config
}

// NewWecomAuthHandler 创建企业微信登录处理器
func NewWecomAuthHandler(db *gorm.DB, wc wecom.Client, wx wxmp.Client, cfg *wecom.Config) *WecomAuthHandler {
	return &WecomAuthHandler{db: db, wecom: wc, wxmp: wx, cfg: cfg}
}

// RegisterRoutes 注册路由（公开端点，无 JWT）
func (h *WecomAuthHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/auth/wecom/config", h.GetConfig)
	r.POST("/auth/wecom/mp", h.MpLogin)
	r.POST("/auth/wecom/web", h.WebLogin)
}

// GetConfig 企业微信登录前端配置
// @Summary 企业微信登录配置
// @Description 返回 corpid/agentId/inviteQrUrl/redirectHost 供登录页使用（不含 secret）
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response "获取成功"
// @Router /api/v1/auth/wecom/config [get]
func (h *WecomAuthHandler) GetConfig(c *gin.Context) {
	if h.cfg.CorpID == "" {
		response.Fail(c, response.CodeBizError, "企业微信登录未配置")
		return
	}
	response.Ok(c, gin.H{
		"corpid":       h.cfg.CorpID,
		"agentId":      h.cfg.AgentID,
		"inviteQrUrl":  h.cfg.InviteQRURL,
		"redirectHost": h.cfg.RedirectHost,
	})
}

type wecomCodeReq struct {
	Code string `json:"code" binding:"required"`
}

// MpLogin 小程序企业微信登录
// @Summary 小程序企业微信登录
// @Description getPhoneNumber code → 手机号 → 校验企业成员身份 → 绑定/签发 JWT
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body wecomCodeReq true "手机号授权码"
// @Success 200 {object} response.Response{data=LoginResp} "登录成功"
// @Router /api/v1/auth/wecom/mp [post]
func (h *WecomAuthHandler) MpLogin(c *gin.Context) {
	var req wecomCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	phone, err := h.wxmp.GetPhoneNumber(req.Code)
	if err != nil {
		if errors.Is(err, wxmp.ErrInvalidCode) {
			response.BadRequest(c, "授权已过期，请重试")
			return
		}
		log.Error().Err(err).Msg("wxmp get phone number failed")
		response.ServerError(c, "登录失败，请稍后重试")
		return
	}

	wecomUserID, err := h.wecom.GetUserIDByMobile(phone)
	if err != nil {
		if errors.Is(err, wecom.ErrNotMember) {
			response.FailWithData(c, response.CodeNotWecomMember, "请先使用企业微信扫码加入企业", gin.H{"inviteQrUrl": h.cfg.InviteQRURL})
			return
		}
		log.Error().Err(err).Msg("wecom get userid by mobile failed")
		response.ServerError(c, "登录失败，请稍后重试")
		return
	}

	h.wecomLogin(c, wecomUserID, phone)
}

// WebLogin Web 管理端企业微信扫码登录
// @Summary Web 企业微信扫码登录
// @Description WWLogin 授权 code → 成员 userid → 详情(手机号) → 绑定/签发 JWT
// @Tags 认证
// @Accept json
// @Produce json
// @Param body body wecomCodeReq true "扫码授权码"
// @Success 200 {object} response.Response{data=LoginResp} "登录成功"
// @Router /api/v1/auth/wecom/web [post]
func (h *WecomAuthHandler) WebLogin(c *gin.Context) {
	var req wecomCodeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	wecomUserID, err := h.wecom.GetUserInfoByCode(req.Code)
	if err != nil {
		if errors.Is(err, wecom.ErrInvalidCode) {
			response.BadRequest(c, "授权已过期，请重试")
			return
		}
		if errors.Is(err, wecom.ErrNotMember) {
			response.FailWithData(c, response.CodeNotWecomMember, "请先使用企业微信扫码加入企业", gin.H{"inviteQrUrl": h.cfg.InviteQRURL})
			return
		}
		log.Error().Err(err).Msg("wecom get userinfo by code failed")
		response.ServerError(c, "登录失败，请稍后重试")
		return
	}

	var mobile string
	detail, err := h.wecom.GetUserDetail(wecomUserID)
	if err != nil {
		// 详情取不到不阻断：userid 已能证明成员身份，走 wecom_user_id 匹配
		log.Warn().Err(err).Str("wecomUserId", wecomUserID).Msg("wecom get user detail failed, fallback to userid-only match")
	} else {
		mobile = detail.Mobile
	}

	h.wecomLogin(c, wecomUserID, mobile)
}

// wecomLogin 企业微信登录共享绑定判定：先按 wecom_user_id，再按手机号首绑
func (h *WecomAuthHandler) wecomLogin(c *gin.Context, wecomUserID, mobile string) {
	var emp model.Employee
	err := h.db.Where("wecom_user_id = ?", wecomUserID).First(&emp).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		if mobile == "" {
			response.Fail(c, response.CodeEmpNotRegistered, "请联系管理员在后台添加员工档案")
			return
		}
		// 首次登录：手机号匹配预建档员工（同手机号多公司命中时取最早一条并告警）
		err = h.db.Where("phone = ?", mobile).Order("id asc").First(&emp).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Fail(c, response.CodeEmpNotRegistered, "请联系管理员在后台添加员工档案")
			return
		}
	}
	if err != nil {
		log.Error().Err(err).Msg("wecom login employee query failed")
		response.ServerError(c, "登录失败")
		return
	}

	if emp.Status != 1 {
		response.Fail(c, response.CodeForbidden, "账号已禁用，请联系管理员")
		return
	}

	// 绑定/同步：身份锚点是 wecom_user_id，手机号以企业微信为准
	changed := false
	if emp.WecomUserID != wecomUserID {
		emp.WecomUserID = wecomUserID
		changed = true
	}
	if mobile != "" && emp.Phone != mobile {
		emp.Phone = mobile
		changed = true
	}
	if changed {
		if err := h.db.Save(&emp).Error; err != nil {
			log.Error().Err(err).Msg("wecom login bind employee failed")
			response.ServerError(c, "登录失败")
			return
		}
	}

	resp, err := issueTokenPair(&emp)
	if err != nil {
		log.Error().Err(err).Msg("issue token pair failed")
		response.ServerError(c, "令牌生成失败")
		return
	}
	response.Ok(c, resp)
}
```

- [ ] **Step 5: 接线 `internal/api/router.go`**

```go
import (
	// 现有 import 不动，追加：
	"zhizhang-server/internal/pkg/wecom"
	"zhizhang-server/internal/pkg/wxmp"
)

// SetupRouter 配置路由
func SetupRouter(db *gorm.DB, jwtCfg *middleware.JWTConfig, wecomCfg *wecom.Config, wxmpCfg *wxmp.Config) *gin.Engine {
	middleware.InitJWT(jwtCfg)

	authHandler := handler.NewAuthHandler(db)
	wecomAuthHandler := handler.NewWecomAuthHandler(db, wecom.NewClient(wecomCfg), wxmp.NewClient(wxmpCfg), wecomCfg)
	// ……其余 handler 不动……
```

在 `authHandler.RegisterRoutes(v1)` 之后加一行 `wecomAuthHandler.RegisterRoutes(v1)`。

- [ ] **Step 6: 修改 `cmd/server/main.go` 调用点 + 启动告警**

```go
	router := api.SetupRouter(db, jwtCfg, &wecom.Config{
		CorpID:       cfg.Wecom.CorpID,
		AgentID:      cfg.Wecom.AgentID,
		Secret:       cfg.Wecom.Secret,
		InviteQRURL:  cfg.Wecom.InviteQRURL,
		RedirectHost: cfg.Wecom.RedirectHost,
	}, &wxmp.Config{
		AppID:  cfg.WechatMP.AppID,
		Secret: cfg.WechatMP.Secret,
	})
```

`SetupRouter` 调用前加告警（import 无需变）：

```go
	if cfg.Wecom.CorpID == "" || cfg.Wecom.Secret == "" {
		log.Warn().Msg("wecom login not configured (ZHIZHANG_WECOM_* empty); wecom endpoints will respond 4001")
	}
	if cfg.WechatMP.AppID == "" || cfg.WechatMP.Secret == "" {
		log.Warn().Msg("wechat mp not configured (ZHIZHANG_WECHAT_MP_* empty); mp phone login unavailable")
	}
```

main.go import 追加 `"zhizhang-server/internal/pkg/wecom"` 与 `"zhizhang-server/internal/pkg/wxmp"`。

- [ ] **Step 7: 运行全部后端测试**

Run: `go build ./... && go test ./internal/... ./cmd/... -race`
Expected: 全部 PASS。

- [ ] **Step 8: Commit**

```bash
git add internal/ cmd/
git commit -m "feat: add wecom login endpoints (config/mp/web)

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 7: employee.go —— 手机号必填 / 重复校验 / 企微解绑

**Files:**
- Modify: `internal/api/handler/employee.go`（CreateEmployeeReq line 44-56、UpdateEmployeeReq line 58-64、CreateEmployee、UpdateEmployee）
- Test: `internal/api/handler/employee_test.go`（新建）

**Interfaces:**
- Consumes: 测试辅助（Task 5）；`response.CodeDuplicate`（现有）。
- Produces: Create 请求 `phone` 必填、`password` 选填；Update 请求支持 `wecomUserId` 显式传空串解绑（`*string` 区分「未传」与「传空」）。

- [ ] **Step 1: 写失败测试 `internal/api/handler/employee_test.go`**

需要 JWT 鉴权——测试里直接用 middleware 生成 token 带 Authorization 头：

```go
package handler

import (
	"fmt"
	"testing"

	"github.com/gin-gonic/gin"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

func setupEmpRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()
	v1 := r.Group("/api/v1")
	authorized := v1.Group("")
	authorized.Use(middleware.JWTMiddleware())
	NewEmployeeHandler(db).RegisterRoutes(authorized)
	return r
}

func authHeader(t *testing.T, companyID, userID uint) map[string]string {
	t.Helper()
	token, err := middleware.GenerateToken(userID, "admin", 1, 1, companyID)
	if err != nil {
		t.Fatal(err)
	}
	return map[string]string{"Authorization": "Bearer " + token}
}

func TestCreateEmployeePhoneRequired(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	r := setupEmpRouter(db)

	// 缺手机号 → 400
	w := performRequest(r, "POST", "/api/v1/employees", map[string]interface{}{
		"deptId": dept.ID, "roleId": staffRole.ID, "username": "nophone", "name": "无号", "status": 1,
	}, authHeader(t, company.ID, 1))
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeBadRequest {
		t.Fatalf("expected 400 for missing phone, got %v", resp)
	}

	// 不带密码 + 带手机号 → 200
	w = performRequest(r, "POST", "/api/v1/employees", map[string]interface{}{
		"deptId": dept.ID, "roleId": staffRole.ID, "username": "withphone", "name": "有号",
		"phone": "13800000010", "status": 1,
	}, authHeader(t, company.ID, 1))
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}
}

func TestCreateEmployeeDuplicatePhone(t *testing.T) {
	db := setupTestDB(t)
	company, dept, _, staffRole := seedTenant(t, db)
	seedStaff(t, db, "13800000011")
	r := setupEmpRouter(db)

	w := performRequest(r, "POST", "/api/v1/employees", map[string]interface{}{
		"deptId": dept.ID, "roleId": staffRole.ID, "username": "another", "name": "重复",
		"phone": "13800000011", "status": 1,
	}, authHeader(t, company.ID, 1))
	if resp := parseBody(t, w); resp["code"].(float64) != response.CodeDuplicate {
		t.Fatalf("expected 4002, got %v", resp)
	}
}

func TestUpdateEmployeeUnbindWecom(t *testing.T) {
	db := setupTestDB(t)
	company, _, _, _ := seedTenant(t, db)
	emp := seedStaff(t, db, "13800000012")
	db.Model(&emp).Update("wecom_user_id", "wx_abc")
	r := setupEmpRouter(db)

	// 显式传空串 → 解绑
	w := performRequest(r, "PUT", fmt.Sprintf("/api/v1/employees/%d", emp.ID), map[string]interface{}{
		"wecomUserId": "", "status": 1,
	}, authHeader(t, company.ID, 1))
	if resp := parseBody(t, w); resp["code"].(float64) != 200 {
		t.Fatalf("expected 200, got %v", resp)
	}
	var reloaded model.Employee
	db.First(&reloaded, emp.ID)
	if reloaded.WecomUserID != "" {
		t.Fatalf("expected unbound, got %q", reloaded.WecomUserID)
	}
}
```

（注意：`seedStaff` 种子的 company 与 `authHeader` 的 companyID 需一致——`seedStaff` 内部调 `seedTenant` 会各自建公司；若用例混用，先调 `seedTenant` 再手工建员工。实现者可把 `seedStaff` 重构为接收 company/dept/role 参数。）

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/api/handler/ -run TestCreateEmployee -v 2>&1 | head -8`
Expected: FAIL —— 缺手机号仍创建成功；重复手机号未返回 4002。

- [ ] **Step 3: 修改 `employee.go` DTO 与逻辑**

```go
type CreateEmployeeReq struct {
	DeptID   uint   `json:"deptId" binding:"required"`
	RoleID   uint   `json:"roleId" binding:"required"`
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"omitempty,min=6,max=64"` // 企业微信员工可留空
	Name     string `json:"name" binding:"required,max=64"`
	Phone    string `json:"phone" binding:"required,max=20"`           // 企业微信绑定键，必填
	Email    string `json:"email" binding:"max=128"`
	Status   int8   `json:"status" binding:"oneof=0 1"`
}

type UpdateEmployeeReq struct {
	DeptID      uint    `json:"deptId"`
	RoleID      uint    `json:"roleId"`
	Name        string  `json:"name" binding:"max=64"`
	Phone       string  `json:"phone" binding:"max=20"`
	Email       string  `json:"email" binding:"max=128"`
	Status      int8    `json:"status" binding:"oneof=0 1"`
	WecomUserID *string `json:"wecomUserId"` // 指针：未传=不变；传空串=解绑
}
```

`CreateEmployee`：用户名重复检查后加手机号重复检查；密码改为条件哈希：

```go
	// 检查手机号是否重复（企业微信绑定键，公司内唯一）
	h.db.Model(&model.Employee{}).Where("company_id = ? AND phone = ?", companyID, req.Phone).Count(&count)
	if count > 0 {
		response.Fail(c, response.CodeDuplicate, "手机号已存在")
		return
	}

	// 密码选填：企业微信登录员工无需密码
	var hashedPwd string
	if req.Password != "" {
		b, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			response.ServerError(c, "密码加密失败")
			return
		}
		hashedPwd = string(b)
	}
	emp := model.Employee{
		// …… Password: hashedPwd, 其余不变 ……
	}
```

`UpdateEmployee`：手机号变更查重（排除自身）+ 解绑支持，加在 `emp.Status = req.Status` 之前：

```go
	if req.Phone != "" && req.Phone != emp.Phone {
		var phoneCount int64
		h.db.Model(&model.Employee{}).Where("company_id = ? AND phone = ? AND id != ?", companyID, req.Phone, emp.ID).Count(&phoneCount)
		if phoneCount > 0 {
			response.Fail(c, response.CodeDuplicate, "手机号已存在")
			return
		}
		emp.Phone = req.Phone
	}
	// 企业微信解绑：显式传空串清除绑定（未传则保持原值）
	if req.WecomUserID != nil {
		emp.WecomUserID = *req.WecomUserID
	}
```

（原有 `if req.Phone != "" { emp.Phone = req.Phone }` 段被上面的查重版本替换。）

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/api/handler/ -v -race`
Expected: 全部 PASS。

- [ ] **Step 5: Commit**

```bash
git add internal/api/handler/
git commit -m "feat: require phone for employees and support wecom unbind

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 8: 部署文档 + Swagger 重新生成

**Files:**
- Create: `docs/wecom-setup.md`
- Modify: `docs/`（`make swagger` 产物）

**Interfaces:**
- Consumes: Task 6 的 swaggo 注解。
- Produces: 运维可按文档独立完成企业微信侧配置。

- [ ] **Step 1: 写 `docs/wecom-setup.md`**

包含以下小节（每节给具体路径与操作）：

1. **创建自建应用**：企业微信管理后台 → 应用管理 → 自建 → 创建应用；记录 `AgentId` 与 `Secret`（应用详情页「查看 Secret」会推送到企业微信）；可见范围选全体成员或按部门。
2. **通讯录权限（手机号查 userid 必需）**：管理工具 → 通讯录同步 → 开启 API 接口同步，或给自建应用配置「通讯录读取」权限；`cgi-bin/user/getuserid` 需要通讯录级别权限，若自建应用 Secret 调用报 60011/48002，改用通讯录同步 Secret 或给应用授权。
3. **可信域名与 WWLogin（Web 扫码必需）**：自建应用详情 → 开发者接口 → 企业微信授权登录 → 设置授权回调域（与 `ZHIZHANG_WECOM_REDIRECT_HOST` 一致）；本地开发 localhost 无法扫码，用超管密码通道。
4. **邀请二维码**：管理后台 → 我的企业 → 邀请成员 → 下载二维码，上传至任意可访问的图床/OSS，URL 填入 `ZHIZHANG_WECOM_INVITE_QR_URL`。
5. **微信小程序 getPhoneNumber**：mp 后台（mp.weixin.qq.com）→ 开发管理 → 接口设置 → 手机号快速验证组件（需非个人主体 + 微信认证）；AppID/Secret 填入 `ZHIZHANG_WECHAT_MP_APPID/SECRET`，AppID 同步填入 `app/env/.env` 的 `VITE_WX_APPID`。
6. **环境变量清单**：对照 `.env.example` 的 `ZHIZHANG_WECOM_*` / `ZHIZHANG_WECHAT_MP_*`。
7. **验证步骤**：后端启动日志无 wecom 告警 → `GET /api/v1/auth/wecom/config` 返回 200 → Web 登录页出现二维码 → 小程序一键登录成功。

- [ ] **Step 2: 重新生成 Swagger**

Run: `make swagger`
Expected: `docs/` 更新，含 `/api/v1/auth/wecom/*` 三个端点。

- [ ] **Step 3: Commit**

```bash
git add docs/
git commit -m "docs: add wecom setup guide and regenerate swagger

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 9: Web 管理端 —— auth store wecomLogin + 回调页与路由

**Files:**
- Modify: `frontend/src/stores/auth.ts`
- Create: `frontend/src/views/WecomCallbackView.vue`
- Modify: `frontend/src/router/index.ts`（`/login` 路由附近）
- Test: `frontend/src/stores/auth.spec.ts`（新建）

**Interfaces:**
- Consumes: `POST /api/v1/auth/wecom/web`（Task 6）；现有 `api` axios 实例与响应信封。
- Produces: `useAuthStore().wecomLogin(code: string)`——成功存 token/user，失败抛错（`err.code`/`err.data` 携带业务码与 data）；Task 10 的 LoginView 依赖。

- [ ] **Step 1: 写失败测试 `frontend/src/stores/auth.spec.ts`**

```ts
import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

vi.mock('@/api/client', () => ({
  default: { post: vi.fn(), get: vi.fn() },
}))

import api from '@/api/client'
import { useAuthStore } from './auth'

describe('wecomLogin', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    localStorage.clear()
    vi.clearAllMocks()
  })

  it('扫码登录成功：存储 token 与用户信息', async () => {
    vi.mocked(api.post).mockResolvedValue({
      data: {
        code: 200,
        message: 'success',
        data: {
          accessToken: 'at',
          refreshToken: 'rt',
          accessExpiresIn: 7200,
          refreshExpiresIn: 604800,
          user: { id: 1, username: 'zhangsan', name: '张三', phone: '13800000001', roleId: 2, deptId: 1 },
        },
      },
    })
    const store = useAuthStore()
    await store.wecomLogin('auth-code')
    expect(api.post).toHaveBeenCalledWith('/api/v1/auth/wecom/web', { code: 'auth-code' })
    expect(store.token).toBe('at')
    expect(localStorage.getItem('token')).toBe('at')
    expect(store.user?.name).toBe('张三')
  })

  it('未建档：抛出带 4103 的错误', async () => {
    vi.mocked(api.post).mockResolvedValue({
      data: { code: 4103, message: '请联系管理员在后台添加员工档案' },
    })
    const store = useAuthStore()
    await expect(store.wecomLogin('auth-code')).rejects.toMatchObject({ code: 4103 })
    expect(store.token).toBe('')
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd frontend && npx vitest run src/stores/auth.spec.ts`
Expected: FAIL —— `wecomLogin` 不存在。

- [ ] **Step 3: 修改 `frontend/src/stores/auth.ts`**

在 `login` 函数后追加：

```ts
  async function wecomLogin(code: string) {
    const res = await api.post('/api/v1/auth/wecom/web', { code })
    if (res.data.code === 0 || res.data.code === 200) {
      const data = res.data.data
      token.value = data.accessToken
      localStorage.setItem('token', data.accessToken)
      localStorage.setItem('refreshToken', data.refreshToken)
      user.value = data.user
      return data
    }
    const err = new Error(res.data.message || '企业微信登录失败') as Error & { code?: number; data?: unknown }
    err.code = res.data.code
    err.data = res.data.data
    throw err
  }
```

`return` 对象中追加 `wecomLogin`。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd frontend && npx vitest run src/stores/auth.spec.ts`
Expected: 2 个用例 PASS。

- [ ] **Step 5: 新建 `frontend/src/views/WecomCallbackView.vue`**

```vue
<template>
  <div class="callback-container">
    <el-card class="callback-card">
      <el-result v-if="status === 'loading'" icon="info" title="正在登录…" sub-title="企业微信授权校验中，请稍候" />
      <el-result v-else icon="error" title="登录失败" :sub-title="errorMsg">
        <template #extra>
          <el-image v-if="inviteQrUrl" :src="inviteQrUrl" fit="contain" style="width: 200px; height: 200px" />
          <p v-if="inviteQrUrl" class="qr-tip">请使用企业微信扫码加入企业后重试</p>
          <el-button type="primary" @click="$router.replace('/login')">返回登录页</el-button>
        </template>
      </el-result>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const status = ref<'loading' | 'error'>('loading')
const errorMsg = ref('')
const inviteQrUrl = ref('')

onMounted(async () => {
  const { code, state } = route.query as Record<string, string>
  if (!code) {
    status.value = 'error'
    errorMsg.value = '缺少授权码，请重新扫码'
    return
  }
  // state 防 CSRF：与登录页生成时存入的值比对
  const savedState = sessionStorage.getItem('wecom_login_state')
  sessionStorage.removeItem('wecom_login_state')
  if (!savedState || state !== savedState) {
    status.value = 'error'
    errorMsg.value = '登录状态校验失败，请重新扫码'
    return
  }
  try {
    await authStore.wecomLogin(code)
    router.replace('/dashboard')
  } catch (e: any) {
    status.value = 'error'
    errorMsg.value = e.message || '企业微信登录失败'
    if (e.code === 4102) {
      inviteQrUrl.value = e.data?.inviteQrUrl || ''
    }
  }
})
</script>

<style scoped>
.callback-container { min-height: 100vh; display: flex; justify-content: center; align-items: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.callback-card { width: 460px; }
.qr-tip { margin: 12px 0; font-size: 14px; color: #606266; }
</style>
```

- [ ] **Step 6: 路由注册 `frontend/src/router/index.ts`**

在 `/login` 路由后追加：

```ts
    {
      path: '/login/wecom/callback',
      name: 'WecomCallback',
      component: () => import('@/views/WecomCallbackView.vue'),
      meta: { public: true },
    },
```

- [ ] **Step 7: 构建验证 + Commit**

Run: `cd frontend && npm run build && npx vitest run`
Expected: 构建成功，测试全 PASS。

```bash
git add frontend/src/
git commit -m "feat: add wecom scan login flow to web admin

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 10: Web 管理端 —— LoginView 双 Tab + WWLogin 二维码

**Files:**
- Modify: `frontend/src/views/LoginView.vue`（整体重写）

**Interfaces:**
- Consumes: `GET /api/v1/auth/wecom/config`（Task 6 返回 `{corpid, agentId, inviteQrUrl, redirectHost}`）；Task 9 的回调路由 `/login/wecom/callback`。
- Produces: 登录页默认展示企业微信二维码；密码 Tab 标注「仅超级管理员」。

- [ ] **Step 1: 重写 `frontend/src/views/LoginView.vue`**

```vue
<template>
  <div class="login-container">
    <el-card class="login-card" shadow="always">
      <template #header>
        <div class="login-header">
          <el-icon :size="40" color="#409EFF"><Box /></el-icon>
          <h2>智账系统</h2>
          <p>企业进销存管理平台</p>
        </div>
      </template>

      <el-tabs v-model="activeTab" stretch>
        <el-tab-pane label="企业微信扫码" name="wecom">
          <div v-if="wecomLoading" class="qr-loading" v-loading="true" style="height: 300px" />
          <div v-show="!wecomLoading && wecomConfigured" id="wecom-qr-container" class="qr-box" />
          <el-result v-if="!wecomLoading && !wecomConfigured" icon="warning" title="企业微信登录未配置" sub-title="请联系管理员配置，或切换账号密码登录（仅超级管理员）" />
          <p v-if="wecomConfigured" class="qr-tip">使用企业微信扫码登录，未加入企业请先联系管理员</p>
        </el-tab-pane>

        <el-tab-pane label="账号密码（仅超级管理员）" name="password">
          <el-form
            ref="formRef"
            :model="form"
            :rules="rules"
            label-position="top"
            @keyup.enter="handleLogin"
          >
            <el-form-item label="用户名" prop="username">
              <el-input v-model="form.username" placeholder="请输入用户名" prefix-icon="User" size="large" />
            </el-form-item>
            <el-form-item label="密码" prop="password">
              <el-input v-model="form.password" type="password" placeholder="请输入密码" prefix-icon="Lock" size="large" show-password />
            </el-form-item>
            <el-form-item>
              <el-button type="primary" size="large" style="width: 100%" :loading="loading" @click="handleLogin">
                登录
              </el-button>
            </el-form-item>
          </el-form>
        </el-tab-pane>
      </el-tabs>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import api from '@/api/client'
import { ElMessage } from 'element-plus'

declare global {
  interface Window { WwLogin?: new (options: Record<string, string>) => unknown }
}

const WWLOGIN_SDK_URL = 'https://wwcdn.weixin.qq.com/node/wework/wwopen/js/wwLogin-1.2.7.js'

const router = useRouter()
const authStore = useAuthStore()
const formRef = ref()
const loading = ref(false)
const activeTab = ref<'wecom' | 'password'>('wecom')
const wecomLoading = ref(true)
const wecomConfigured = ref(false)

const form = reactive({ username: '', password: '' })

const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur' }],
}

function loadScript(src: string) {
  return new Promise<void>((resolve, reject) => {
    const el = document.createElement('script')
    el.src = src
    el.onload = () => resolve()
    el.onerror = () => reject(new Error('企业微信组件加载失败'))
    document.head.appendChild(el)
  })
}

async function mountWecomQr() {
  try {
    const res = await api.get('/api/v1/auth/wecom/config')
    if (res.data.code !== 200 && res.data.code !== 0) {
      return // 未配置：保持 wecomConfigured = false
    }
    const { corpid, agentId, redirectHost } = res.data.data
    const state = crypto.randomUUID()
    sessionStorage.setItem('wecom_login_state', state)
    await loadScript(WWLOGIN_SDK_URL)
    const origin = redirectHost || window.location.origin
    new window.WwLogin!({
      id: 'wecom-qr-container',
      appid: corpid,
      agentid: String(agentId),
      redirect_uri: encodeURIComponent(`${origin}/login/wecom/callback`),
      state,
      href: '',
    })
    wecomConfigured.value = true
  } catch {
    // 网络异常等按未配置处理，用户可用密码通道
  } finally {
    wecomLoading.value = false
  }
}

onMounted(mountWecomQr)

async function handleLogin() {
  try {
    await formRef.value.validate()
    loading.value = true
    await authStore.login(form.username, form.password)
    ElMessage.success('登录成功')
    router.push('/dashboard')
  } catch (error: any) {
    // 4101：非超管密码登录被拒，引导扫码
    if (error?.response?.data?.code === 4101) {
      ElMessage.warning('请使用企业微信扫码登录')
      activeTab.value = 'wecom'
    } else {
      ElMessage.error(error.message || '登录失败')
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container { min-height: 100vh; display: flex; justify-content: center; align-items: center; background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); }
.login-card { width: 460px; }
.login-header { text-align: center; }
.login-header h2 { margin: 10px 0 5px; font-size: 24px; color: #303133; }
.login-header p { margin: 0; font-size: 14px; color: #909399; }
.qr-box { display: flex; justify-content: center; min-height: 300px; }
.qr-tip { text-align: center; font-size: 12px; color: #909399; margin: 12px 0 0; }
</style>
```

注意：密码登录失败时后端错误走 axios 拦截器还是信封判断，取决于 `client.ts` 是否对非 2xx reject——4101 以 HTTP 200 + code 返回，会落到 `authStore.login` 的 `throw new Error(res.data.message)` 分支。为让 4101 提示正确，`authStore.login` 需要像 `wecomLogin` 一样携带 code：在 `login` 的 throw 处补 `err.code = res.data.code`（与 Task 9 的写法一致），`handleLogin` 里改判 `error.code === 4101`。

- [ ] **Step 2: 修改 `frontend/src/stores/auth.ts` 的 `login`**

```ts
    throw new Error(res.data.message || '登录失败')
```
改为：
```ts
    const err = new Error(res.data.message || '登录失败') as Error & { code?: number }
    err.code = res.data.code
    throw err
```

（`auth.spec.ts` 里补一个用例：密码登录 4101 → `rejects.toMatchObject({ code: 4101 })`。）

- [ ] **Step 3: 构建 + 测试 + Commit**

Run: `cd frontend && npm run build && npx vitest run`
Expected: 成功。

```bash
git add frontend/src/
git commit -m "feat: rebuild web login page with wecom qr tab

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

- [ ] **Step 4: 手动验证（真机之外最接近的检查）**

启动后端 + `npm run dev`，打开登录页：
- 未配置 wecom 时：显示「企业微信登录未配置」，密码 Tab 可登录（admin/admin123）。
- 配置后：二维码渲染，扫码 → 回调页 → dashboard。（可信域名未配时二维码不渲染属预期，文档已说明。）

---

### Task 11: Web 管理端 —— 员工管理：手机号必填 + 企微绑定列/解绑

**Files:**
- Modify: `frontend/src/views/base/EmployeeDetailView.vue`（表单校验）
- Modify: `frontend/src/views/base/EmployeeView.vue`（列表列 + 解绑按钮）

**Interfaces:**
- Consumes: Task 7 的 employee API 行为（phone 必填、wecomUserId 空串解绑）。

- [ ] **Step 1: `EmployeeDetailView.vue` 表单改动**

手机号表单项改必填并加说明；密码项改说明（现状：新建时显示密码项）：

```vue
<el-col :span="8">
  <el-form-item label="手机号" prop="phone" :rules="[{ required: true, message: '请输入手机号', trigger: 'blur' }]">
    <el-input v-model="form.phone" placeholder="企业微信登录的手机号绑定键" />
  </el-form-item>
</el-col>
<el-col :span="8" v-if="id === 'new'">
  <el-form-item label="密码" prop="password">
    <el-input v-model="form.password" type="password" placeholder="留空则仅支持企业微信登录" />
  </el-form-item>
</el-col>
```

提交前校验（在现有 submit 函数开头）：

```ts
if (!form.phone) {
  ElMessage.warning('请填写手机号（企业微信登录绑定键）')
  return
}
```

（若页面已有 `formRef.validate()` 则把 phone 规则并入现有 rules；保持与页面现有校验风格一致。）

- [ ] **Step 2: `EmployeeView.vue` 列表改动**

`Employee` 接口加 `wecomUserId?: string`；表格在「状态」列后加：

```vue
<el-table-column label="企微绑定" width="150">
  <template #default="{ row }">
    <el-tag v-if="row.wecomUserId" type="success">{{ row.wecomUserId }}</el-tag>
    <el-tag v-else type="info">未绑定</el-tag>
  </template>
</el-table-column>
```

「操作」列加解绑按钮（仅已绑定时显示）：

```vue
<el-button v-if="row.wecomUserId" type="warning" size="small" @click="handleUnbindWecom(row)">解绑企微</el-button>
```

script 追加：

```ts
import api from '@/api/client'
import { ElMessage, ElMessageBox } from 'element-plus'

async function handleUnbindWecom(row: Employee) {
  await ElMessageBox.confirm(`确认解除「${row.name}」的企业微信绑定？解除后其下次登录将重新按手机号绑定。`, '解绑确认', { type: 'warning' })
  await api.put(`/api/v1/employees/${row.id}`, { wecomUserId: '' })
  ElMessage.success('已解绑')
  fetchList()
}
```

- [ ] **Step 3: 构建 + Commit**

Run: `cd frontend && npm run build`
Expected: 成功。

```bash
git add frontend/src/views/base/
git commit -m "feat: employee phone required and wecom bind management

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 12: 小程序端 —— API + token store 的 wecomMpLogin

**Files:**
- Modify: `app/src/api/login.ts`
- Modify: `app/src/api/types/login.ts`（业务码常量）
- Modify: `app/src/store/token.ts`
- Test: `app/src/store/token.test.ts`（新建）

**Interfaces:**
- Consumes: `POST /api/v1/auth/wecom/mp`、`GET /api/v1/auth/wecom/config`（Task 6）；app http 层 `{hideErrorToast}` 选项。
- Produces:
  - `wecomMpLogin(code: string)` API 函数（`hideErrorToast: true`，错误由页面分支处理）
  - `getWecomConfig()` → `{corpid, agentId, inviteQrUrl, redirectHost}`
  - `WECOM_CODE = { NEED_WECOM: 4101, NOT_MEMBER: 4102, NOT_REGISTERED: 4103 } as const`
  - `useTokenStore().wecomMpLogin(code)` —— Task 13 登录页依赖。

- [ ] **Step 1: 写失败测试 `app/src/store/token.test.ts`**

```ts
import { beforeEach, describe, expect, it, vi } from 'vitest'

vi.mock('@/api/login', () => ({
  login: vi.fn(),
  logout: vi.fn(),
  refreshToken: vi.fn(),
  wxLogin: vi.fn(),
  getWxCode: vi.fn(),
  getUserInfo: vi.fn(),
  wecomMpLogin: vi.fn(),
}))

import { getUserInfo, wecomMpLogin } from '@/api/login'
import { useTokenStore } from './token'

const tokenRes = {
  accessToken: 'at',
  refreshToken: 'rt',
  accessExpiresIn: 7200,
  refreshExpiresIn: 604800,
}

describe('wecomMpLogin', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('登录成功：写入 tokenInfo 并拉取用户信息', async () => {
    vi.mocked(wecomMpLogin).mockResolvedValue(tokenRes)
    vi.mocked(getUserInfo).mockResolvedValue({ userId: 1, username: 'zhangsan', nickname: '张三' } as any)

    const store = useTokenStore()
    await store.wecomMpLogin('phone-code')

    expect(wecomMpLogin).toHaveBeenCalledWith('phone-code')
    expect(store.tokenInfo).toMatchObject({ accessToken: 'at', refreshToken: 'rt' })
    expect(uni.showToast).toHaveBeenCalledWith(expect.objectContaining({ title: '登录成功' }))
  })

  it('非企业成员 4102：原样抛出由页面处理引导，不弹成功 toast', async () => {
    vi.mocked(wecomMpLogin).mockRejectedValue({ code: 4102, message: '请先使用企业微信扫码加入企业', data: { inviteQrUrl: 'https://x/qr.png' } })

    const store = useTokenStore()
    await expect(store.wecomMpLogin('phone-code')).rejects.toMatchObject({ code: 4102 })
    expect(store.tokenInfo).toMatchObject({ accessToken: '' })
    expect(uni.showToast).not.toHaveBeenCalledWith(expect.objectContaining({ title: '登录成功' }))
  })
})
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd app && npx vitest run src/store/token.test.ts`
Expected: FAIL —— `wecomMpLogin` 不在 store 暴露的方法里。

- [ ] **Step 3: 修改 `app/src/api/types/login.ts`**（文件尾部追加）

```ts
/** 企业微信登录业务码（与后端 internal/pkg/response 一致） */
export const WECOM_CODE = {
  NEED_WECOM: 4101,
  NOT_MEMBER: 4102,
  NOT_REGISTERED: 4103,
} as const
```

- [ ] **Step 4: 修改 `app/src/api/login.ts`**（尾部追加）

```ts
/**
 * 小程序企业微信登录（手机号授权码）
 * hideErrorToast: 4102/4103 等分支由登录页自行处理 UI（引导二维码等），不走全局 toast
 */
export function wecomMpLogin(code: string) {
  return http.post<IAuthLoginRes>('/auth/wecom/mp', { code }, undefined, undefined, { hideErrorToast: true })
}

export interface IWecomConfig {
  corpid: string
  agentId: number
  inviteQrUrl: string
  redirectHost: string
}

/**
 * 企业微信登录前端配置（corpid/agentId/inviteQrUrl）
 */
export function getWecomConfig() {
  return http.get<IWecomConfig>('/auth/wecom/config')
}
```

- [ ] **Step 5: 修改 `app/src/store/token.ts`**

import 区把 `wecomMpLogin` 加进 `@/api/login` 的导入（`as _wecomMpLogin`），store 内追加 action 并加入 return：

```ts
    /**
     * 小程序企业微信登录（手机号授权码换取）
     * 业务错误（4102 非成员 / 4103 未建档）原样抛出，由登录页分支处理
     */
    const wecomMpLogin = async (code: string) => {
      try {
        const res = await _wecomMpLogin(code)
        await _postLogin(res)
        uni.showToast({ title: '登录成功', icon: 'success' })
        return res
      }
      catch (error) {
        console.error('企业微信登录失败:', error)
        throw error
      }
      finally {
        updateNowTime()
      }
    }
```

return 对象的「核心API方法」区追加 `wecomMpLogin,`。

- [ ] **Step 6: 运行测试确认通过**

Run: `cd app && npx vitest run src/store/token.test.ts`
Expected: 2 个用例 PASS。

- [ ] **Step 7: Commit**

```bash
git add app/src/api/ app/src/store/
git commit -m "feat: add wecomMpLogin to miniprogram api and token store

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 13: 小程序端 —— 登录页重写

**Files:**
- Modify: `app/src/pages/auth/login.vue`（整体重写）

**Interfaces:**
- Consumes: `useTokenStore().wecomMpLogin / login`（Task 12）；`WECOM_CODE`、`getWecomConfig`（Task 12）；`HOME_PAGE`（`@/utils`）。
- Produces: 可用的真实登录页（替换 mock）。

- [ ] **Step 1: 重写 `app/src/pages/auth/login.vue`**

```vue
<script lang="ts" setup>
import { onLoad } from '@dcloudio/uni-app'
import { ref } from 'vue'
import { getWecomConfig, WECOM_CODE } from '@/api/login'
import { useTokenStore } from '@/store/token'
import { HOME_PAGE } from '@/utils'

definePage({
  style: {
    navigationBarTitleText: '登录',
  },
})

const tokenStore = useTokenStore()
const redirectUrl = ref('')

const username = ref('')
const password = ref('')
const loading = ref(false)

// 4102 引导态
const showInviteGuide = ref(false)
const inviteQrUrl = ref('')

onLoad((options) => {
  if (options?.redirect) {
    redirectUrl.value = decodeURIComponent(options.redirect)
  }
})

function navigateAfterLogin() {
  const target = redirectUrl.value || HOME_PAGE
  uni.redirectTo({ url: target })
}

async function doWecomLogin(e: any) {
  const code = e?.detail?.code
  if (!code) {
    uni.showToast({ title: '手机号授权失败，请重试', icon: 'none' })
    return
  }
  loading.value = true
  try {
    await tokenStore.wecomMpLogin(code)
    navigateAfterLogin()
  }
  catch (err: any) {
    if (err?.code === WECOM_CODE.NOT_MEMBER) {
      inviteQrUrl.value = err?.data?.inviteQrUrl || ''
      if (!inviteQrUrl.value) {
        // 后端未配邀请二维码时尝试拉一次公开配置
        try {
          inviteQrUrl.value = (await getWecomConfig()).inviteQrUrl
        }
        catch { /* 配置也没有就只显示文案 */ }
      }
      showInviteGuide.value = true
    }
    else if (err?.code === WECOM_CODE.NOT_REGISTERED) {
      uni.showToast({ title: '请联系管理员在后台添加员工档案', icon: 'none' })
    }
    else {
      uni.showToast({ title: err?.message || '登录失败，请重试', icon: 'none' })
    }
  }
  finally {
    loading.value = false
  }
}

async function doPasswordLogin() {
  if (!username.value || !password.value) {
    uni.showToast({ title: '请输入用户名和密码', icon: 'none' })
    return
  }
  loading.value = true
  try {
    await tokenStore.login({ username: username.value, password: password.value })
    navigateAfterLogin()
  }
  catch (err: any) {
    if (err?.code === WECOM_CODE.NEED_WECOM) {
      uni.showToast({ title: '请使用企业微信登录', icon: 'none' })
    }
    // 其余错误 http 层已弹 toast
  }
  finally {
    loading.value = false
  }
}
</script>

<template>
  <view class="login-page px-48rpx pt-100rpx">
    <view class="mb-80rpx text-center">
      <view class="text-44rpx font-bold">
        智账系统
      </view>
      <view class="mt-16rpx text-26rpx text-gray-400">
        企业进销存管理平台
      </view>
    </view>

    <!-- 企业微信一键登录：仅微信小程序支持 getPhoneNumber -->
    <!-- #ifdef MP-WEIXIN -->
    <button
      class="mb-32rpx h-88rpx w-full rounded-12rpx bg-green-600 text-32rpx text-white leading-88rpx"
      open-type="getPhoneNumber"
      :disabled="loading"
      @getphonenumber="doWecomLogin"
    >
      企业微信一键登录
    </button>
    <view class="mb-40rpx text-center text-24rpx text-gray-400">
      将校验您的企业微信成员身份（需已加入企业）
    </view>
    <!-- #endif -->

    <!-- 账号密码登录：仅超级管理员 -->
    <view class="mb-24rpx text-center text-24rpx text-gray-400">
      —— 账号密码登录（仅超级管理员） ——
    </view>
    <input
      v-model="username"
      class="mb-24rpx h-88rpx w-full rounded-12rpx bg-gray-100 px-24rpx text-30rpx"
      placeholder="用户名"
    >
    <input
      v-model="password"
      class="mb-32rpx h-88rpx w-full rounded-12rpx bg-gray-100 px-24rpx text-30rpx"
      type="password"
      placeholder="密码"
    >
    <button
      class="h-88rpx w-full rounded-12rpx bg-blue-600 text-32rpx text-white leading-88rpx"
      :disabled="loading"
      @click="doPasswordLogin"
    >
      登录
    </button>

    <!-- 非企业成员引导：展示企业邀请二维码 -->
    <view v-if="showInviteGuide" class="mt-60rpx flex flex-col items-center">
      <view class="mb-16rpx text-28rpx text-gray-600">
        请先使用企业微信扫码加入企业
      </view>
      <image
        v-if="inviteQrUrl"
        :src="inviteQrUrl"
        class="h-400rpx w-400rpx"
        mode="aspectFit"
        show-menu-by-longpress
      />
      <view class="mt-16rpx text-24rpx text-gray-400">
        可长按保存二维码，在企业微信中扫码加入后重试
      </view>
    </view>
  </view>
</template>

<style lang="scss" scoped>
// UnoCSS 原子类已覆盖布局，此处保留块备用
</style>
```

- [ ] **Step 2: 类型检查 + 构建**

Run: `cd app && pnpm type-check 2>&1 | grep '^src/' ; pnpm build:mp 2>&1 | tail -3`
Expected: `src/` 无类型错误（node_modules 内 wot-design-uni 的既有类型噪音忽略）；构建成功。

- [ ] **Step 3: Commit**

```bash
git add app/src/pages/auth/login.vue
git commit -m "feat: rewrite miniprogram login page with wecom one-tap login

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 14: 小程序端 —— 登录拦截配置（商城放行 / 内部强制）

**Files:**
- Modify: `app/src/router/config.ts`
- Test: `app/src/router/interceptor.test.ts`（改 1 个既有用例 + 新增 2 个）

**Interfaces:**
- Consumes: 现有黑名单策略（`DEFAULT_NO_NEED_LOGIN`：EXCLUDE 列表 = 需要登录的页面）。
- Produces: `LOGIN_PAGE_ENABLE_IN_MP = true`；`/pages/me/me` 入黑名单。后续新增内部业务页必须追加 `EXCLUDE_LOGIN_PATH_LIST`（商城页不加，默认放行）。

- [ ] **Step 1: 修改 `app/src/router/config.ts`**

```ts
// 在小程序里面是否使用H5的登录页，默认为 false
// 智账系统：小程序内部业务强制企业微信登录，启用拦截
export const LOGIN_PAGE_ENABLE_IN_MP = true
```

```ts
// 黑名单策略（DEFAULT_NO_NEED_LOGIN）：列表内页面需要登录
// 约定：内部业务页追加到本列表；商城页保持默认放行
export const EXCLUDE_LOGIN_PATH_LIST = [
  '/pages/me/me', // 「我的」展示员工身份信息，需要登录
  ...excludeLoginPathList, // 都是以 / 开头的 path
]
```

（删掉两个 `/pages/xxx/index` 示例值。）

- [ ] **Step 2: 调整既有用例 + 新增用例（`interceptor.test.ts`）**

既有用例 `未配置 roles 的 tabbar 页放行` 访问的 `/pages/me/me` 现在进了黑名单，未登录会被重定向——把该用例改为「已登录」前置：

```ts
  it('未配置 roles 的 tabbar 页放行（已登录）', async () => {
    // 让 tokenStore.hasLogin 为 true：写入双 token 并让过期时间检查通过
    vi.mocked(uni.getStorageSync).mockReturnValue(Date.now() + 3600_000)
    const { useTokenStore } = await import('@/store/token')
    useTokenStore().setTokenInfo({
      accessToken: 'a', refreshToken: 'r', accessExpiresIn: 7200, refreshExpiresIn: 604800,
    })
    const { navigateToInterceptor } = await import('./interceptor')
    const result = navigateToInterceptor.invoke({ url: '/pages/me/me' })
    expect(result).not.toBe(false)
    expect(uni.reLaunch).not.toHaveBeenCalled()
  })
```

新增：

```ts
  it('未登录访问黑名单内页面 → 重定向登录页并携带 redirect', async () => {
    const result = await invoke('/pages/me/me')
    expect(result).toBe(false)
    expect(uni.navigateTo).toHaveBeenCalledWith({
      url: `/pages/auth/login?redirect=${encodeURIComponent('/pages/me/me')}`,
    })
  })

  it('未登录访问商城/普通页（不在黑名单）→ 放行', async () => {
    const result = await invoke('/pages/index/index')
    expect(result).not.toBe(false)
    expect(uni.navigateTo).not.toHaveBeenCalled()
  })
```

（`invoke` 辅助函数沿用文件顶部现有实现。）

- [ ] **Step 3: 运行测试**

Run: `cd app && npx vitest run src/router/interceptor.test.ts`
Expected: 全部 PASS。

- [ ] **Step 4: Commit**

```bash
git add app/src/router/
git commit -m "feat: enable mp login interception with mall pages guest-accessible

Co-Authored-By: Claude Code <noreply@anthropic.com>"
```

---

### Task 15: 全量验证 + 手动验收清单

**Files:**
- 无新文件；验证既有改动。

- [ ] **Step 1: 后端全量**

Run: `cd /Users/fuqiang/Project/zhizhang-server && go build ./... && go vet ./... && go test ./... -race`
Expected: 全部 PASS。

- [ ] **Step 2: Web 管理端全量**

Run: `cd frontend && npm run build && npx vitest run`
Expected: 构建成功，测试全 PASS。

- [ ] **Step 3: 小程序端全量**

Run: `cd app && pnpm type-check 2>&1 | grep '^src/' ; pnpm test:run ; pnpm build:mp 2>&1 | tail -3`
Expected: src 无类型错误；测试全 PASS；构建成功。

- [ ] **Step 4: 启动冒烟（无真实凭证路径）**

```bash
make infra-up && go run cmd/server/main.go
```

确认日志：
- `wecom login not configured` / `wechat mp not configured` 两条 warn 出现（凭证为空时）。
- `GET http://localhost:8388/api/v1/auth/wecom/config` 返回 `{"code":4001,...}`。
- `POST /api/v1/auth/login` 用 admin/admin123 仍返回 200 且带 `accessExpiresIn/refreshExpiresIn`（web 管理端密码通道不回归）。

- [ ] **Step 5: 真机验收清单（交给用户在配置真实凭证后执行）**

- [ ] Web 端：配置 `ZHIZHANG_WECOM_*` + 可信域名后，登录页二维码渲染 → 企业微信扫码 → 回调 → dashboard
- [ ] Web 端：非企业成员扫码 → 4102 → 展示邀请二维码引导
- [ ] 小程序：微信开发者工具/真机 → 一键登录 → 已建档员工进入首页
- [ ] 小程序：未加入企业 → 4102 引导页展示邀请二维码
- [ ] 小程序：是企业成员但未建档 → 4103 提示联系管理员
- [ ] 小程序：商城/首页未登录可浏览；「我的」未登录跳登录页

## Self-Review 记录

- spec §3.1/§3.2/§3.3 → Task 1/2/5/6 覆盖；§4.1 模型与索引 → Task 3；§4.2 绑定规则 → Task 6 `wecomLogin`；§4.3 业务码 → Task 3；§5.1 Web 页面 → Task 9/10/11；§5.2 小程序 → Task 12/13/14；§6 配置 → Task 4；§7 安全 → state 校验在 Task 9、token 缓存在 Task 1/2、文案分级在 Task 6；§8 错误矩阵 → Task 6/10/13；§9 测试 → 各 Task 内嵌；§10 交付物 7（wecom-setup.md）→ Task 8。
- 类型一致性：`wecom.Config` 字段（CorpID/AgentID/Secret/InviteQRURL/RedirectHost）在 Task 1 定义、Task 4 映射、Task 6 使用，一致；`WECOM_CODE` 在 Task 12 定义、Task 13 使用，一致；`FailWithData` 在 Task 6 Step 1 定义后使用，一致。
- 已知留白：Task 7 测试中 `seedStaff` 与 `seedTenant` 的公司一致性在 Step 1 注释中说明（实现者重构 seedStaff 签名），不影响计划可执行性。
