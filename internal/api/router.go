package api

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	"zhizhang-server/internal/api/handler"
	"zhizhang-server/internal/api/middleware"
	"zhizhang-server/internal/pkg/response"
)

// SetupRouter 配置路由
func SetupRouter(db *gorm.DB, jwtCfg *middleware.JWTConfig, wecomCfg *handler.WeComConfig) *gin.Engine {
	// 初始化 JWT
	middleware.InitJWT(jwtCfg)

	// 初始化处理器
	authHandler := handler.NewAuthHandler(db, wecomCfg)
	dictHandler := handler.NewDictHandler(db)
	deptHandler := handler.NewDepartmentHandler(db)
	empHandler := handler.NewEmployeeHandler(db)
	roleHandler := handler.NewRoleHandler(db)
	permHandler := handler.NewPermissionHandler(db)
	productHandler := handler.NewProductHandler(db)
	productSettingHandler := handler.NewProductSettingHandler(db)
	customerHandler := handler.NewCustomerHandler(db)
	warehouseHandler := handler.NewWarehouseHandler(db)
	priceHandler := handler.NewPriceHandler(db)
	purchaseHandler := handler.NewPurchaseHandler(db)
	saleHandler := handler.NewSaleHandler(db)
	inventoryHandler := handler.NewInventoryHandler(db)
	mallHandler := handler.NewMallHandler(db)
	crmHandler := handler.NewCRMHandler(db)
	approvalHandler := handler.NewApprovalHandler(db)
	statsHandler := handler.NewStatsHandler(db)

	router := gin.New()
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.LoggerMiddleware())
	router.Use(gin.Recovery())

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

		// 需要认证的路由组
		authorized := v1.Group("")
		authorized.Use(middleware.JWTMiddleware())
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

			// 客户/供应商
			customerHandler.RegisterRoutes(authorized)

			// 仓库/资金
			warehouseHandler.RegisterRoutes(authorized)

			// 价格体系
			priceHandler.RegisterRoutes(authorized)

			// 库存模块
			inventoryHandler.RegisterRoutes(authorized)

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
