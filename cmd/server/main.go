package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"gorm.io/gorm"

	_ "zhizhang-server/docs" // swaggo 生成的文档

	"zhizhang-server/internal/api"
	"zhizhang-server/internal/api/handler"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/model"
	"zhizhang-server/internal/pkg/database"
	"zhizhang-server/internal/pkg/redis"
	"zhizhang-server/internal/pkg/wecom"
	"zhizhang-server/internal/pkg/wxmp"
)

// Config 应用配置
type Config struct {
	App struct {
		Name    string `mapstructure:"name"`
		Version string `mapstructure:"version"`
		Env     string `mapstructure:"env"`
		Port    int    `mapstructure:"port"`
	} `mapstructure:"app"`
	JWT struct {
		Secret     string `mapstructure:"secret"`
		AccessTTL  int    `mapstructure:"access_ttl"`
		RefreshTTL int    `mapstructure:"refresh_ttl"`
	} `mapstructure:"jwt"`
	Database database.Config `mapstructure:"database"`
	Redis    redis.Config  `mapstructure:"redis"`
	Wecom    struct {
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
	Log struct {
		Level  string `mapstructure:"level"`
		Format string `mapstructure:"format"`
		Output string `mapstructure:"output"`
	} `mapstructure:"log"`
}

func main() {
	// 加载配置
	cfg, err := loadConfig()
	if err != nil {
		fmt.Printf("failed to load config: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	initLogger(cfg)

	log.Info().
		Str("app", cfg.App.Name).
		Str("version", cfg.App.Version).
		Str("env", cfg.App.Env).
		Msg("starting server")

	// 初始化数据库
	db, err := database.Init(&cfg.Database)
	if err != nil {
		log.Fatal().Err(err).Msg("database init failed")
	}

	// 自动迁移
	if err := autoMigrate(db); err != nil {
		log.Fatal().Err(err).Msg("auto migrate failed")
	}

	// 初始化种子数据
	if err := handler.InitSeedData(db); err != nil {
		log.Fatal().Err(err).Msg("seed data init failed")
	}

	// 初始化 Redis（可选，非阻塞）
	if _, err := redis.Init(&cfg.Redis); err != nil {
		log.Warn().Err(err).Msg("redis init failed, continuing without cache")
	}

	// 设置运行模式
	if cfg.App.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}

	// JWT 配置
	jwtCfg := &middleware.JWTConfig{
		Secret:     cfg.JWT.Secret,
		AccessTTL:  cfg.JWT.AccessTTL,
		RefreshTTL: cfg.JWT.RefreshTTL,
	}

	// 企业微信/微信小程序未配置时给出告警（端点会返回 4001，不影响其余功能）
	if cfg.Wecom.CorpID == "" || cfg.Wecom.Secret == "" {
		log.Warn().Msg("wecom login not configured (ZHIZHANG_WECOM_* empty); wecom endpoints will respond 4001")
	}
	if cfg.WechatMP.AppID == "" || cfg.WechatMP.Secret == "" {
		log.Warn().Msg("wechat mp not configured (ZHIZHANG_WECHAT_MP_* empty); mp phone login unavailable")
	}

	// 配置路由
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

	// 启动服务
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Info().Str("addr", addr).Msg("server listening")
	if err := router.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("server failed")
	}
}

func loadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")

	// 环境变量覆盖
	viper.SetEnvPrefix("ZHIZHANG")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_")) // ZHIZHANG_WECOM_CORPID → wecom.corpid
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func initLogger(cfg *Config) {
	level, _ := zerolog.ParseLevel(cfg.Log.Level)
	zerolog.SetGlobalLevel(level)

	if cfg.Log.Format == "console" {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}).With().Timestamp().Logger()
	} else {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}

func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&model.Company{},
		&model.Department{},
		&model.Role{},
		&model.Permission{},
		&model.RolePermission{},
		&model.Employee{},
		&model.DictType{},
		&model.DictItem{},
		// 商品资料
		&model.ProductCategory{},
		&model.Brand{},
		&model.ProductSeries{},
		&model.Product{},
		&model.ProductUnit{},
		&model.ProductBarcode{},
		// 客户/供应商
		&model.CustomerCategory{},
		&model.Region{},
		&model.CustomerLevel{},
		&model.Customer{},
		// 仓库/资金
		&model.Warehouse{},
		&model.WarehousePosition{},
		&model.Account{},
		&model.IncomeExpenseItem{},
		// 价格体系
		&model.PriceLevel{},
		&model.ProductPrice{},
		// 采购模块
		&model.PurchaseOrder{},
		&model.PurchaseOrderItem{},
		&model.PurchaseInStock{},
		&model.PurchaseInStockItem{},
		&model.PurchaseReturn{},
		&model.PurchaseReturnItem{},
		&model.PurchasePayment{},
		&model.PurchasePaymentItem{},
		// 销售模块
		&model.SalesOrder{},
		&model.SalesOrderItem{},
		&model.SalesOutStock{},
		&model.SalesOutStockItem{},
		&model.SalesReturn{},
		&model.SalesReturnItem{},
		&model.SalesReceipt{},
		&model.SalesReceiptItem{},
		// 库存模块
		&model.InventoryCheck{},
		&model.InventoryCheckItem{},
		&model.InventoryTransfer{},
		&model.InventoryTransferItem{},
		&model.InventoryWarning{},
		// 商城模块
		&model.MallProduct{},
		&model.MallCart{},
		&model.MallOrder{},
		&model.MallOrderItem{},
		// CRM模块
		&model.FollowUp{},
		&model.Opportunity{},
		&model.Contract{},
		// 审批模块
		&model.ApprovalProcess{},
		&model.ApprovalRecord{},
	); err != nil {
		return err
	}
	// 企业微信绑定：(公司, 手机号) 复合唯一索引
	// 不用 GORM tag 的原因：CompanyID 在共享的 BaseModelWithCompany 中，打 tag 会污染所有租户模型
	return db.Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_emp_phone_company ON employees (company_id, phone)").Error
}
