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

// NewClient 生产构造：真实 API 地址 + Redis 缓存
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
		// 提前 300s 过期
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
