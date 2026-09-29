package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/response"
)

// SettingHandler 公司级键值设置处理器
type SettingHandler struct {
	db *gorm.DB
}

func NewSettingHandler(db *gorm.DB) *SettingHandler {
	return &SettingHandler{db: db}
}

func (h *SettingHandler) RegisterRoutes(r *gin.RouterGroup) {
	s := r.Group("/settings")
	{
		s.GET("", h.ListSettings)
		s.PUT("", h.SaveSettings)
	}
}

func (h *SettingHandler) companyID(c *gin.Context) uint {
	return middleware.GetCompanyID(c)
}

// ListSettings 获取本公司全部设置
func (h *SettingHandler) ListSettings(c *gin.Context) {
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	var list []model.CompanySetting
	if err := h.db.Where("company_id = ?", companyID).Find(&list).Error; err != nil {
		log.Error().Err(err).Msg("list company settings failed")
		response.ServerError(c, "查询失败")
		return
	}
	data := make(map[string]string, len(list))
	for _, item := range list {
		data[item.Key] = item.Value
	}
	response.Ok(c, data)
}

// SaveSettings 批量 upsert 设置
func (h *SettingHandler) SaveSettings(c *gin.Context) {
	var req map[string]string
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}
	companyID := h.companyID(c)
	if companyID == 0 {
		response.Unauthorized(c, "未登录")
		return
	}
	err := h.db.Transaction(func(tx *gorm.DB) error {
		for key, value := range req {
			var setting model.CompanySetting
			err := tx.Where("company_id = ? AND key = ?", companyID, key).First(&setting).Error
			if err == nil {
				if err := tx.Model(&setting).Update("value", value).Error; err != nil {
					return err
				}
				continue
			}
			if err != gorm.ErrRecordNotFound {
				return err
			}
			setting = model.CompanySetting{
				BaseModelWithCompany: model.BaseModelWithCompany{CompanyID: companyID},
				Key:                  key,
				Value:                value,
			}
			if err := tx.Create(&setting).Error; err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		log.Error().Err(err).Msg("save company settings failed")
		response.ServerError(c, "保存失败")
		return
	}
	response.OkWithMessage(c, "保存成功", nil)
}
