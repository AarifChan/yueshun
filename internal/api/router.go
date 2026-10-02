package api

import (
	"os"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/handler"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/pkg/response"
	"zhizhang-server/internal/pkg/wecom"
	"zhizhang-server/internal/pkg/wxmp"
)

// SetupRouter 配置路由
func SetupRouter(db *gorm.DB, jwtCfg *middleware.JWTConfig, wecomCfg *wecom.Config, wxmpCfg *wxmp.Config, uploadDir string) *gin.Engine {
	// 初始化 JWT
	middleware.InitJWT(jwtCfg)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(db)
	wecomAuthHandler := handler.NewWecomAuthHandler(db, wecom.NewClient(wecomCfg), wxmp.NewClient(wxmpCfg), wecomCfg)
	dictHandler := handler.NewDictHandler(db)
	deptHandler := handler.NewDepartmentHandler(db)
	empHandler := handler.NewEmployeeHandler(db)
	roleHandler := handler.NewRoleHandler(db)
	permHandler := handler.NewPermissionHandler(db)
	productHandler := handler.NewProductHandler(db)
	productSettingHandler := handler.NewProductSettingHandler(db)
	saleScopeHandler := handler.NewSaleScopeHandler(db)
	settingHandler := handler.NewSettingHandler(db)
	customerHandler := handler.NewCustomerHandler(db)
	supplierHandler := handler.NewSupplierHandler(db)
	warehouseHandler := handler.NewWarehouseHandler(db)
	priceHandler := handler.NewPriceHandler(db)
	priceManageHandler := handler.NewPriceManageHandler(db)
	purchaseHandler := handler.NewPurchaseHandler(db)
	saleHandler := handler.NewSaleHandler(db)
	inventoryHandler := handler.NewInventoryHandler(db)
	otherInStockHandler := handler.NewOtherInStockHandler(db)
	otherOutStockHandler := handler.NewOtherOutStockHandler(db)
	mallHandler := handler.NewMallHandler(db)
	crmHandler := handler.NewCRMHandler(db)
	approvalHandler := handler.NewApprovalHandler(db)
	statsHandler := handler.NewStatsHandler(db)
	uploadHandler := handler.NewUploadHandler(uploadDir)
	baseDataHandler := handler.NewBaseDataHandler(db)
	inventoryExtHandler := handler.NewInventoryExtHandler(db)
	transferExtHandler := handler.NewTransferExtHandler(db)
	inventoryReportHandler := handler.NewInventoryReportHandler(db)
	billCenterHandler := handler.NewBillCenterHandler(db)
	saleReportHandler := handler.NewSaleReportHandler(db)
	purchaseReportHandler := handler.NewPurchaseReportHandler(db)
	financeHandler := handler.NewFinanceHandler(db)
	crmExtHandler := handler.NewCRMExtHandler(db)
	productExtHandler := handler.NewProductExtHandler(db)
	marketingHandler := handler.NewMarketingHandler(db)
	biHandler := handler.NewBIHandler(db)
	ecosystemHandler := handler.NewEcosystemHandler(db)
	settingExtHandler := handler.NewSettingExtHandler(db)

	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(gin.Recovery())

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		log.Warn().Err(err).Str("dir", uploadDir).Msg("create upload dir failed, static serving disabled")
	} else {
		router.Static("/uploads", uploadDir)
	}

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		response.Ok(c, gin.H{"status": "ok", "service": "zhizhang-server"})
	})

	// Swagger 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1
	v1 := router.Group("/api/v1")
	{
		// 认证（公开）
		authHandler.RegisterRoutes(v1)
		wecomAuthHandler.RegisterRoutes(v1)

		// 需要认证的路由组
		authorized := v1.Group("")
		authorized.Use(middleware.JWTMiddleware())
		authorized.Use(middleware.OperationLogger(db))
		{
			// 基础资料
			dictHandler.RegisterRoutes(authorized)
			deptHandler.RegisterRoutes(authorized)
			empHandler.RegisterRoutes(authorized)
			roleHandler.RegisterRoutes(authorized)
			permHandler.RegisterRoutes(authorized)

			// 商品资料
			productHandler.RegisterRoutes(authorized)
			productSettingHandler.RegisterRoutes(authorized)
			saleScopeHandler.RegisterRoutes(authorized)
			settingHandler.RegisterRoutes(authorized)

			// 客户/供应商
			customerHandler.RegisterRoutes(authorized)
			supplierHandler.RegisterRoutes(authorized)

			// 仓库/资金
			warehouseHandler.RegisterRoutes(authorized)

			// 价格体系
			priceHandler.RegisterRoutes(authorized)
			priceManageHandler.RegisterRoutes(authorized)

			// 库存模块
			inventoryHandler.RegisterRoutes(authorized)
			otherInStockHandler.RegisterRoutes(authorized)
			otherOutStockHandler.RegisterRoutes(authorized)

			// 商城模块
			mallHandler.RegisterRoutes(authorized)

			// CRM模块
			crmHandler.RegisterRoutes(authorized)

			// 审批模块
			approvalHandler.RegisterRoutes(authorized)

			// 采购模块
			purchaseHandler.RegisterRoutes(authorized)

			// 销售模块
			saleHandler.RegisterRoutes(authorized)

			// 数据统计
			statsHandler.RegisterRoutes(authorized)

			// 资料模块
			baseDataHandler.RegisterRoutes(authorized)

			// 库存扩展模块
			inventoryExtHandler.RegisterRoutes(authorized)
			transferExtHandler.RegisterRoutes(authorized)
			inventoryReportHandler.RegisterRoutes(authorized)
			billCenterHandler.RegisterRoutes(authorized)
			saleReportHandler.RegisterRoutes(authorized)
			purchaseReportHandler.RegisterRoutes(authorized)
			financeHandler.RegisterRoutes(authorized)
			crmExtHandler.RegisterRoutes(authorized)
			productExtHandler.RegisterRoutes(authorized)
			marketingHandler.RegisterRoutes(authorized)
			biHandler.RegisterRoutes(authorized)
			ecosystemHandler.RegisterRoutes(authorized)
			settingExtHandler.RegisterRoutes(authorized)

			// 文件上传
			uploadHandler.RegisterRoutes(authorized)

			// 占位
			authorized.GET("/ping", func(c *gin.Context) {
				userID := middleware.GetUserID(c)
				response.Ok(c, gin.H{
					"message": "pong",
					"userID":  userID,
				})
			})
		}
	}

	// 404
	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "接口不存在")
	})

	return router
}
