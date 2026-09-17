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
