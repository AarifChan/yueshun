package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// WeComConfig 企业微信登录配置
type WeComConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	CorpID      string `mapstructure:"corp_id"`
	AgentID     string `mapstructure:"agent_id"`
	Secret      string `mapstructure:"secret"`
	RedirectURL string `mapstructure:"redirect_url"` // 授权回调地址，如 https://域名/api/v1/auth/wecom/callback
}

// Configured 是否已完成企业微信配置
func (c *WeComConfig) Configured() bool {
	return c.Enabled && c.CorpID != "" && c.AgentID != "" && c.Secret != "" && c.RedirectURL != ""
}

const weComAPIBase = "https://qyapi.weixin.qq.com/cgi-bin"

// weComAccessToken 企业微信 access_token 内存缓存
type weComAccessToken struct {
	token     string
	expiresAt time.Time
	mu        sync.Mutex
}

var weComTokenCache weComAccessToken

// getWeComAccessToken 获取企业微信 access_token（带缓存）
func getWeComAccessToken(cfg *WeComConfig) (string, error) {
	weComTokenCache.mu.Lock()
	defer weComTokenCache.mu.Unlock()

	if weComTokenCache.token != "" && time.Now().Before(weComTokenCache.expiresAt) {
		return weComTokenCache.token, nil
	}

	reqURL := fmt.Sprintf("%s/gettoken?corpid=%s&corpsecret=%s",
		weComAPIBase, url.QueryEscape(cfg.CorpID), url.QueryEscape(cfg.Secret))
	var result struct {
		ErrCode     int    `json:"errcode"`
		ErrMsg      string `json:"errmsg"`
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := weComGetJSON(reqURL, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("企业微信接口错误(%d): %s", result.ErrCode, result.ErrMsg)
	}

	weComTokenCache.token = result.AccessToken
	// 提前 5 分钟过期
	weComTokenCache.expiresAt = time.Now().Add(time.Duration(result.ExpiresIn-300) * time.Second)
	return result.AccessToken, nil
}

// getWeComUserID 用网页授权 code 换取企业微信 UserID
func getWeComUserID(cfg *WeComConfig, code string) (string, error) {
	token, err := getWeComAccessToken(cfg)
	if err != nil {
		return "", err
	}
	reqURL := fmt.Sprintf("%s/auth/getuserinfo?access_token=%s&code=%s",
		weComAPIBase, url.QueryEscape(token), url.QueryEscape(code))
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
		UserID  string `json:"UserId"`
	}
	if err := weComGetJSON(reqURL, &result); err != nil {
		return "", err
	}
	if result.ErrCode != 0 {
		return "", fmt.Errorf("企业微信接口错误(%d): %s", result.ErrCode, result.ErrMsg)
	}
	if result.UserID == "" {
		return "", fmt.Errorf("未获取到企业微信用户身份")
	}
	return result.UserID, nil
}

func weComGetJSON(reqURL string, out interface{}) error {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		return fmt.Errorf("请求企业微信接口失败: %w", err)
	}
	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("解析企业微信响应失败: %w", err)
	}
	return nil
}

// weComLoginStates 二维码登录的 state 校验（内存，10 分钟有效）
var weComLoginStates sync.Map

func newWeComState() string {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		log.Error().Err(err).Msg("generate wecom state failed")
	}
	state := hex.EncodeToString(buf)
	weComLoginStates.Store(state, time.Now().Add(10*time.Minute))
	return state
}

func verifyWeComState(state string) bool {
	v, ok := weComLoginStates.LoadAndDelete(state)
	if !ok {
		return false
	}
	expiry, ok := v.(time.Time)
	return ok && time.Now().Before(expiry)
}

// WeComStatus 查询企业微信登录是否可用
// @Summary 企业微信登录状态
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response "enabled 表示是否已配置"
// @Router /api/v1/auth/wecom/status [get]
func (h *AuthHandler) WeComStatus(c *gin.Context) {
	response.Ok(c, gin.H{"enabled": h.wecom.Configured()})
}

// WeComQRCodeURL 生成企业微信扫码登录地址
// @Summary 企业微信扫码登录地址
// @Tags 认证
// @Produce json
// @Success 200 {object} response.Response "url 为扫码页面地址"
// @Router /api/v1/auth/wecom/qrcode-url [get]
func (h *AuthHandler) WeComQRCodeURL(c *gin.Context) {
	if !h.wecom.Configured() {
		response.Fail(c, response.CodeBizError, "企业微信登录未配置，请联系管理员")
		return
	}
	state := newWeComState()
	qrURL := fmt.Sprintf("https://open.work.weixin.qq.com/wwopen/sso/qrConnect?appid=%s&agentid=%s&redirect_uri=%s&state=%s",
		url.QueryEscape(h.wecom.CorpID),
		url.QueryEscape(h.wecom.AgentID),
		url.QueryEscape(h.wecom.RedirectURL),
		state)
	response.Ok(c, gin.H{"url": qrURL})
}

// WeComCallback 企业微信授权回调
// 成功/失败都渲染 HTML：在 iframe 中通过 postMessage 通知父页面；
// 若被企业微信顶跳到顶层窗口，则展示结果页并提供返回链接。
// @Summary 企业微信授权回调
// @Tags 认证
// @Produce html
// @Param code query string true "授权码"
// @Param state query string true "状态码"
// @Router /api/v1/auth/wecom/callback [get]
func (h *AuthHandler) WeComCallback(c *gin.Context) {
	render := func(payload string, message string) {
		c.Data(200, "text/html; charset=utf-8", []byte(fmt.Sprintf(`<!DOCTYPE html>
<html><head><meta charset="utf-8"><title>企业微信登录</title></head>
<body style="font-family:sans-serif;text-align:center;padding-top:60px;">
<script>
if (window.parent !== window) {
  window.parent.postMessage(%s, window.location.origin);
} else {
  document.body.innerHTML = '<p style="font-size:18px;">%s</p><p><a href="/">返回系统登录页</a></p>';
}
</script>
</body></html>`, payload, message)))
	}

	code := c.Query("code")
	state := c.Query("state")
	if code == "" || !verifyWeComState(state) {
		render(`{"type":"wecom-login","ok":false,"message":"登录状态无效或已过期，请重试"}`, "登录状态无效或已过期，请重试")
		return
	}

	weComUserID, err := getWeComUserID(h.wecom, code)
	if err != nil {
		log.Error().Err(err).Msg("wecom get user info failed")
		render(`{"type":"wecom-login","ok":false,"message":"企业微信身份获取失败"}`, "企业微信身份获取失败")
		return
	}

	// 匹配绑定了该企业微信账号的职员
	var emp model.Employee
	if err := h.db.Where("we_com_user_id = ? AND status = 1", weComUserID).First(&emp).Error; err != nil {
		render(`{"type":"wecom-login","ok":false,"message":"当前企业微信账号未绑定系统用户"}`, "当前企业微信账号未绑定系统用户，请联系管理员绑定")
		return
	}

	resp, err := h.issueTokens(c, &emp)
	if err != nil {
		render(`{"type":"wecom-login","ok":false,"message":"令牌生成失败"}`, "令牌生成失败")
		return
	}
	render(fmt.Sprintf(`{"type":"wecom-login","ok":true,"accessToken":%q,"refreshToken":%q,"user":%q}`,
		resp.AccessToken, resp.RefreshToken, resp.User.Name), "企业微信登录成功")
}
