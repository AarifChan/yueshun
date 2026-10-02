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
