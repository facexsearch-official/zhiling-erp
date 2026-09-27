package main

import (
	"fmt"
	"log"
	"pisa_server/internal/config"
	"pisa_server/internal/db"
	"pisa_server/internal/handler"
	"pisa_server/internal/middleware"
	jwtpkg "pisa_server/internal/pkg/jwt"
	"pisa_server/internal/pkg/snowflake"
	"pisa_server/internal/repository"
	"pisa_server/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func main() {
	if err := config.Load("config.yaml"); err != nil {
		log.Fatalf("load config: %v", err)
	}
	cfg := config.GlobalConfig

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	if err := snowflake.Init(1); err != nil {
		log.Printf("snowflake init warning: %v", err)
	}

	tokenManager := jwtpkg.NewTokenManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)

	var database *gorm.DB
	var limitChecker *middleware.LimitChecker
	database, err := db.Connect(cfg.Database)
	if err != nil {
		log.Printf("db init warning (continuing without DB): %v", err)
	} else {
		db.AutoCreateTables(database)
		db.Seed(database)
		limitChecker = middleware.NewLimitChecker(database)
	}

	r := gin.Default()

	r.Use(middleware.CORS())
	r.Use(middleware.NoCacheHTML())

	r.Static("/css", "./web/css")
	r.Static("/js", "./web/js")
	r.Static("/lib", "./web/lib")
	r.Static("/uploads", "./web/uploads")
	r.StaticFile("/login.html", "./web/login.html")
	r.StaticFile("/", "./web/index.html")
	r.StaticFile("/index.html", "./web/index.html")
	r.StaticFile("/supplier.html", "./web/supplier.html")
	r.StaticFile("/customer.html", "./web/customer.html")
	r.StaticFile("/warehouse.html", "./web/warehouse.html")
	r.StaticFile("/goods.html", "./web/goods.html")
	r.StaticFile("/spec.html", "./web/spec.html")
	r.StaticFile("/unit.html", "./web/unit.html")
	r.StaticFile("/attr.html", "./web/attr.html")
	r.StaticFile("/price.html", "./web/price.html")
	r.StaticFile("/account.html", "./web/account.html")

	api := r.Group("/api")

	authHandler := handler.NewAuthHandler(database, tokenManager)

	auth := api.Group("/auth")
	auth.POST("/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(middleware.Auth(tokenManager))
	{
		protected.POST("/auth/switch-tenant", authHandler.SwitchTenant)

		tenantGroup := protected.Group("/tenant")
		if database != nil {
			tenantGroup.Use(middleware.TenantMiddleware(database))
		}

		shopGroup := protected.Group("/shop")
		if database != nil {
			shopGroup.Use(middleware.TenantMiddleware(database))
		}

		if database != nil {
			base := repository.NewBaseRepository(database)

			supplierRepo := repository.NewSupplierRepository(base)
			supplierCategoryRepo := repository.NewSupplierCategoryRepository(base)
			staffRepo := repository.NewStaffRepository(base)
			customerRepo := repository.NewCustomerRepository(base)
			warehouseRepo := repository.NewWarehouseRepository(base)
			goodsRepo := repository.NewGoodsRepository(base)
			accountRepo := repository.NewAccountRepository(base)
			categoryRepo := repository.NewCategoryRepository(base)
			unitRepo := repository.NewUnitRepository(base)

			supplierHandler := handler.NewSupplierHandler(supplierRepo)
			supplierCategoryHandler := handler.NewSupplierCategoryHandler(supplierCategoryRepo)
			staffHandler := handler.NewStaffHandler(staffRepo)
			customerHandler := handler.NewCustomerHandler(customerRepo)
			warehouseHandler := handler.NewWarehouseHandler(warehouseRepo)
			goodsHandler := handler.NewGoodsHandler(goodsRepo)
			accountHandler := handler.NewAccountHandler(accountRepo)
			categoryHandler := handler.NewCategoryHandler(categoryRepo)
			unitHandler := handler.NewUnitHandler(unitRepo)

			attributeRepo := repository.NewAttributeRepository(base)
			attributeHandler := handler.NewAttributeHandler(attributeRepo)

			propertyRepo := repository.NewPropertyRepository(base)
			propertyHandler := handler.NewPropertyHandler(propertyRepo)

			priceRepo := repository.NewPriceRepository(base)
			priceHandler := handler.NewPriceHandler(priceRepo)

			shopGroup.GET("/supplier/list", supplierHandler.List)
			shopGroup.GET("/supplier/all", supplierHandler.ListAll)
			shopGroup.GET("/supplier/:id", supplierHandler.GetByID)
			shopGroup.POST("/supplier", supplierHandler.Create)
			shopGroup.PUT("/supplier/:id", supplierHandler.Update)
			shopGroup.DELETE("/supplier/:id", supplierHandler.Delete)

			shopGroup.GET("/supplier-category/all", supplierCategoryHandler.ListAll)
			shopGroup.POST("/supplier-category", supplierCategoryHandler.Create)
			shopGroup.PUT("/supplier-category/:id", supplierCategoryHandler.Update)
			shopGroup.DELETE("/supplier-category/:id", supplierCategoryHandler.Delete)

			shopGroup.GET("/staff/users", staffHandler.ListUsers)
			shopGroup.POST("/staff/users", middleware.PlanLimitMiddleware(limitChecker, "staff"), staffHandler.CreateUser)
			shopGroup.PUT("/staff/users/:id", staffHandler.UpdateUser)
			shopGroup.DELETE("/staff/users/:id", staffHandler.DeleteUser)
			shopGroup.GET("/salesmen", staffHandler.ListSalesmen)
			shopGroup.POST("/salesmen", staffHandler.CreateSalesman)
			shopGroup.PUT("/salesmen/:id", staffHandler.UpdateSalesman)
			shopGroup.DELETE("/salesmen/:id", staffHandler.DeleteSalesman)
			shopGroup.GET("/shops", staffHandler.ListShops)
			shopGroup.POST("/shops", middleware.PlanLimitMiddleware(limitChecker, "shop"), staffHandler.CreateShop)
			shopGroup.PUT("/shops/:id", staffHandler.UpdateShop)
			shopGroup.DELETE("/shops/:id", staffHandler.DeleteShop)

			shopGroup.GET("/customer/list", customerHandler.List)
			shopGroup.GET("/customer/all", customerHandler.ListAll)
			shopGroup.GET("/customer/:id", customerHandler.GetByID)
			shopGroup.POST("/customer", customerHandler.Create)
			shopGroup.PUT("/customer/:id", customerHandler.Update)
			shopGroup.DELETE("/customer/:id", customerHandler.Delete)

			shopGroup.GET("/warehouse/list", warehouseHandler.List)
			shopGroup.GET("/warehouse/all", warehouseHandler.ListAll)
			shopGroup.GET("/warehouse/:id", warehouseHandler.GetByID)
			shopGroup.POST("/warehouse", warehouseHandler.Create)
			shopGroup.PUT("/warehouse/:id", warehouseHandler.Update)
			shopGroup.DELETE("/warehouse/:id", warehouseHandler.Delete)

			shopGroup.GET("/goods/list", goodsHandler.List)
			shopGroup.GET("/goods/all", goodsHandler.ListAll)
			shopGroup.GET("/goods/next-code", goodsHandler.NextCode)
			shopGroup.GET("/goods/brands", goodsHandler.Brands)
			shopGroup.GET("/goods/origins", goodsHandler.Origins)
			shopGroup.GET("/goods/:id", goodsHandler.GetByID)
			shopGroup.POST("/goods", middleware.PlanLimitMiddleware(limitChecker, "goods"), goodsHandler.Create)
			shopGroup.PUT("/goods/:id", goodsHandler.Update)
			shopGroup.DELETE("/goods/:id", goodsHandler.Delete)

			shopGroup.GET("/category/all", categoryHandler.ListAll)
			shopGroup.POST("/category", categoryHandler.Create)
			shopGroup.PUT("/category/:id", categoryHandler.Update)
			shopGroup.DELETE("/category/:id", categoryHandler.Delete)
			shopGroup.GET("/unit/all", unitHandler.ListAll)
			shopGroup.POST("/unit", unitHandler.Create)
			shopGroup.PUT("/unit/:id", unitHandler.Update)
			shopGroup.DELETE("/unit/:id", unitHandler.Delete)
			shopGroup.GET("/attribute/all", attributeHandler.ListAll)
			shopGroup.GET("/attribute/list", attributeHandler.List)
			shopGroup.POST("/attribute", attributeHandler.Create)
			shopGroup.PUT("/attribute/:id", attributeHandler.Update)
			shopGroup.DELETE("/attribute/:id", attributeHandler.Delete)
			shopGroup.POST("/attribute/:id/value", attributeHandler.AddValue)
			shopGroup.GET("/property/all", propertyHandler.ListAll)
			shopGroup.POST("/property", propertyHandler.Create)
			shopGroup.PUT("/property/:id", propertyHandler.Update)
			shopGroup.DELETE("/property/:id", propertyHandler.Delete)
			shopGroup.GET("/price/list", priceHandler.List)
			shopGroup.PUT("/price/batch", priceHandler.Batch)
			shopGroup.PUT("/price/rows", priceHandler.SaveRows)
			shopGroup.POST("/upload", handler.Upload)

			shopGroup.GET("/account/list", accountHandler.List)
			shopGroup.GET("/account/all", accountHandler.ListAll)
			shopGroup.GET("/account/:id", accountHandler.GetByID)
			shopGroup.POST("/account", accountHandler.Create)
			shopGroup.PUT("/account/:id", accountHandler.Update)
			shopGroup.DELETE("/account/:id", accountHandler.Delete)

			purchaseRepo := repository.NewPurchaseRepository(base)
			purchaseItemRepo := repository.NewPurchaseItemRepository(base)
			purchaseReturnRepo := repository.NewPurchaseReturnRepository(base)
			purchaseReturnItemRepo := repository.NewPurchaseReturnItemRepository(base)

			purchaseSvc := service.NewPurchaseService(database, purchaseRepo, purchaseItemRepo, goodsRepo)
			purchaseHandler := handler.NewPurchaseHandler(purchaseSvc)
			purchaseReturnHandler := handler.NewPurchaseReturnHandler(purchaseReturnRepo, purchaseReturnItemRepo, database)
			purchaseOrderHandler := handler.NewPurchaseOrderHandler(database)
			salesHandler := handler.NewSalesHandler(database)
			commissionHandler := handler.NewCommissionHandler(database)
			comboHandler := handler.NewComboHandler(database)
			stockCountHandler := handler.NewStockCountHandler(database)
			assemblyHandler := handler.NewAssemblyHandler(database)
			recipeHandler := handler.NewRecipeHandler(database)
			stockQueryHandler := handler.NewStockQueryHandler(database)
			batchHandler := handler.NewBatchHandler(database)
			transferHandler := handler.NewTransferHandler(database)
			receiptHandler := handler.NewReceiptHandler(database)
			paymentHandler := handler.NewPaymentHandler(database)
			incomeHandler := handler.NewIncomeHandler(database)
			incomeTypeHandler := handler.NewIncomeTypeHandler(database)

			shopGroup.GET("/purchase/list", purchaseHandler.List)
			shopGroup.GET("/purchase/:id", purchaseHandler.GetByID)
			shopGroup.POST("/purchase", purchaseHandler.Create)
			shopGroup.PUT("/purchase/:id", purchaseHandler.Update)
			shopGroup.DELETE("/purchase/:id", purchaseHandler.Delete)
			shopGroup.POST("/purchase/:id/audit", purchaseHandler.Audit)
			shopGroup.POST("/purchase/:id/unaudit", purchaseHandler.UnAudit)

			shopGroup.GET("/purchase-return/list", purchaseReturnHandler.List)
			shopGroup.GET("/purchase-return/:id", purchaseReturnHandler.GetByID)
			shopGroup.POST("/purchase-return", purchaseReturnHandler.Create)
			shopGroup.DELETE("/purchase-return/:id", purchaseReturnHandler.Delete)
			shopGroup.POST("/purchase-return/:id/audit", purchaseReturnHandler.Audit)

			shopGroup.GET("/purchase-order/list", purchaseOrderHandler.List)
			shopGroup.GET("/purchase-order/:id", purchaseOrderHandler.GetByID)
			shopGroup.POST("/purchase-order", purchaseOrderHandler.Create)
			shopGroup.DELETE("/purchase-order/:id", purchaseOrderHandler.Delete)
			shopGroup.POST("/purchase-order/:id/audit", purchaseOrderHandler.Audit)

			shopGroup.GET("/sale/list", salesHandler.ListSales)
			shopGroup.GET("/sale/:id", salesHandler.GetSale)
			shopGroup.POST("/sale", salesHandler.CreateSale)
			shopGroup.DELETE("/sale/:id", salesHandler.DeleteSale)
			shopGroup.GET("/sale-order/list", salesHandler.ListSaleOrders)
			shopGroup.GET("/sale-order/:id", salesHandler.GetSaleOrder)
			shopGroup.POST("/sale-order", salesHandler.CreateSaleOrder)
			shopGroup.DELETE("/sale-order/:id", salesHandler.DeleteSaleOrder)
			shopGroup.GET("/sale-return/list", salesHandler.ListSalesReturns)
			shopGroup.GET("/sale-return/:id", salesHandler.GetSalesReturn)
			shopGroup.POST("/sale-return", salesHandler.CreateSalesReturn)
			shopGroup.DELETE("/sale-return/:id", salesHandler.DeleteSalesReturn)
			shopGroup.GET("/quote/list", salesHandler.ListQuotes)
			shopGroup.GET("/quote/:id", salesHandler.GetQuote)
			shopGroup.POST("/quote", salesHandler.CreateQuote)
			shopGroup.DELETE("/quote/:id", salesHandler.DeleteQuote)

			shopGroup.GET("/commission-rule/list", commissionHandler.ListRules)
			shopGroup.POST("/commission-rule", commissionHandler.CreateRule)
			shopGroup.PUT("/commission-rule/:id", commissionHandler.UpdateRule)
			shopGroup.DELETE("/commission-rule/:id", commissionHandler.DeleteRule)
			shopGroup.GET("/commission/report", commissionHandler.Report)

			shopGroup.GET("/combo/list", comboHandler.List)
			shopGroup.GET("/combo/price-levels", comboHandler.PriceLevels)
			shopGroup.GET("/combo/:id", comboHandler.GetByID)
			shopGroup.POST("/combo", comboHandler.Create)
			shopGroup.DELETE("/combo/:id", comboHandler.Delete)

			shopGroup.GET("/stock-count/list", stockCountHandler.List)
			shopGroup.GET("/stock-count/:id", stockCountHandler.GetByID)
			shopGroup.POST("/stock-count", stockCountHandler.Create)
			shopGroup.POST("/stock-count/:id/void", stockCountHandler.Void)

			shopGroup.GET("/assembly/list", assemblyHandler.List)
			shopGroup.GET("/assembly/:id", assemblyHandler.GetByID)
			shopGroup.POST("/assembly", assemblyHandler.Create)
			shopGroup.PUT("/assembly/:id", assemblyHandler.Update)
			shopGroup.POST("/assembly/:id/void", assemblyHandler.Void)

			shopGroup.GET("/split/list", assemblyHandler.ListSplit)
			shopGroup.GET("/split/:id", assemblyHandler.GetByID)
			shopGroup.POST("/split", assemblyHandler.CreateSplit)
			shopGroup.PUT("/split/:id", assemblyHandler.Update)
			shopGroup.POST("/split/:id/void", assemblyHandler.Void)

			shopGroup.GET("/recipe/list", recipeHandler.List)
			shopGroup.GET("/recipe/:id", recipeHandler.GetByID)
			shopGroup.POST("/recipe", recipeHandler.Create)
			shopGroup.PUT("/recipe/:id", recipeHandler.Update)
			shopGroup.DELETE("/recipe/:id", recipeHandler.Delete)

			shopGroup.GET("/stock-query/list", stockQueryHandler.List)
			shopGroup.GET("/stock-query/alert", stockQueryHandler.Alert)
			shopGroup.GET("/stock-query/flow/:id", stockQueryHandler.Flow)
			shopGroup.GET("/stock-query/cost/:id", stockQueryHandler.Cost)

			shopGroup.GET("/batch/list", batchHandler.List)
			shopGroup.GET("/batch/expiry", batchHandler.Expiry)

			shopGroup.GET("/transfer/list", transferHandler.List)
			shopGroup.POST("/transfer", transferHandler.Create)

			shopGroup.GET("/receipt/list", receiptHandler.List)
			shopGroup.GET("/receipt/:id", receiptHandler.GetByID)
			shopGroup.POST("/receipt", receiptHandler.Create)
			shopGroup.PUT("/receipt/:id", receiptHandler.Update)
			shopGroup.POST("/receipt/:id/void", receiptHandler.Void)

			shopGroup.GET("/payment/list", paymentHandler.List)
			shopGroup.GET("/payment/:id", paymentHandler.GetByID)
			shopGroup.POST("/payment", paymentHandler.Create)
			shopGroup.PUT("/payment/:id", paymentHandler.Update)
			shopGroup.POST("/payment/:id/void", paymentHandler.Void)

			shopGroup.GET("/income-type/list", incomeTypeHandler.List)
			shopGroup.POST("/income-type", incomeTypeHandler.Create)
			shopGroup.PUT("/income-type/:id", incomeTypeHandler.Update)
			shopGroup.DELETE("/income-type/:id", incomeTypeHandler.Delete)

			shopGroup.GET("/income/list", incomeHandler.List)
			shopGroup.GET("/income/:id", incomeHandler.GetByID)
			shopGroup.POST("/income", incomeHandler.Create)
			shopGroup.PUT("/income/:id", incomeHandler.Update)
			shopGroup.POST("/income/:id/void", incomeHandler.Void)

			shopGroup.GET("/expense/list", incomeHandler.ListExpense)
			shopGroup.GET("/expense/:id", incomeHandler.GetByID)
			shopGroup.POST("/expense", incomeHandler.CreateExpense)
			shopGroup.PUT("/expense/:id", incomeHandler.Update)
			shopGroup.POST("/expense/:id/void", incomeHandler.Void)
		}

		if limitChecker != nil {
			shopGroup.POST("/sales", middleware.PlanLimitMiddleware(limitChecker, "order"))
		}
	}

	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	fmt.Printf("PISA Server starting on :%d\n", cfg.Server.Port)
	r.Run(fmt.Sprintf(":%d", cfg.Server.Port))
}
