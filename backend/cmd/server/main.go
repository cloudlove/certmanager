package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"certmanager-backend/internal/config"
	"certmanager-backend/internal/handler"
	"certmanager-backend/internal/middleware"
	"certmanager-backend/internal/model"
	"certmanager-backend/internal/repository"
	"certmanager-backend/internal/scheduler"
	"certmanager-backend/internal/service"
)

var (
	db         *gorm.DB
	rdb        *redis.Client
	configPath = "config/config.yaml"
)

func main() {
	// 读取配置
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("Config loaded: server port=%d\n", cfg.Server.Port)

	// 初始化 MySQL
	if err := initMySQL(cfg); err != nil {
		log.Fatalf("Failed to init MySQL: %v", err)
	}
	fmt.Println("MySQL initialized successfully")

	// 执行数据库种子数据初始化
	if err := seedDatabase(); err != nil {
		log.Printf("Warning: Seed failed: %v", err)
	} else {
		fmt.Println("Database seed completed")
	}

	// 初始化 Redis
	if err := initRedis(cfg); err != nil {
		log.Fatalf("Failed to init Redis: %v", err)
	}
	fmt.Println("Redis initialized successfully")

	// 创建 Gin Engine
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	// 注册中间件
	r.Use(middleware.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.CORS())
	r.Use(middleware.AuditMiddleware(db))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API 路由组
	api := r.Group("/api/v1")

	// ========== 初始化认证相关 Repository 和 Service ==========
	userRepo := repository.NewUserRepository(db)
	roleRepo := repository.NewRoleRepository(db)
	permissionRepo := repository.NewPermissionRepository(db)

	authSvc := service.NewAuthService(userRepo, roleRepo)
	userSvc := service.NewUserService(userRepo, roleRepo)
	roleSvc := service.NewRoleService(roleRepo, permissionRepo)

	// 设置权限检查器
	middleware.SetPermissionChecker(roleSvc)

	authHandler := handler.NewAuthHandler(authSvc)
	userHandler := handler.NewUserHandler(userSvc)
	roleHandler := handler.NewRoleHandler(roleSvc)

	// ========== 公开路由（无需认证）==========
	{
		// 健康检查
		api.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})

		// 认证路由
		authHandler.RegisterPublicRoutes(api)
	}

	// ========== 受保护路由（需要认证）==========
	protected := api.Group("")
	protected.Use(middleware.AuthMiddleware())

	// 预先声明 certificateSvc 以便在 scheduler 中使用
	var certificateSvc *service.CertificateService

	{
		// 认证相关的受保护路由
		authHandler.RegisterProtectedRoutes(protected)

		// ========== 管理路由（需要特定权限）==========
		// 用户管理
		users := protected.Group("/users")
		users.Use(middleware.PermissionMiddleware("users", "manage"))
		{
			users.GET("", userHandler.List)
			users.POST("", userHandler.Create)
			users.GET("/:id", userHandler.Get)
			users.PUT("/:id", userHandler.Update)
			users.DELETE("/:id", userHandler.Delete)
			users.PUT("/:id/role", userHandler.AssignRole)
			users.PUT("/:id/password", userHandler.ResetPassword)
		}

		// 角色管理
		roles := protected.Group("/roles")
		roles.Use(middleware.PermissionMiddleware("roles", "manage"))
		{
			roles.GET("", roleHandler.List)
			roles.POST("", roleHandler.Create)
			roles.GET("/all", roleHandler.ListAll)
			roles.GET("/permissions", roleHandler.ListPermissions)
			roles.GET("/:id", roleHandler.Get)
			roles.PUT("/:id", roleHandler.Update)
			roles.DELETE("/:id", roleHandler.Delete)
			roles.PUT("/:id/permissions", roleHandler.AssignPermissions)
		}

		// 云凭证管理路由
		credentialRepo := repository.NewCredentialRepository(db)
		credentialSvc := service.NewCredentialService(credentialRepo, cfg.Security.AESKey)
		credentialHandler := handler.NewCredentialHandler(credentialSvc)
		credentials := protected.Group("/credentials")
		credentials.Use(middleware.PermissionMiddleware("credentials", "manage"))
		{
			credentialHandler.RegisterRoutes(credentials)
		}

		// CSR 管理路由
		csrRepo := repository.NewCSRRepository(db)
		csrSvc := service.NewCSRService(csrRepo, cfg.Security.AESKey)
		csrHandler := handler.NewCSRHandler(csrSvc)
		csrs := protected.Group("/csrs")
		csrs.Use(middleware.PermissionMiddleware("csrs", "manage"))
		{
			csrHandler.RegisterRoutes(csrs)
		}

		// 域名管理路由
		domainRepo := repository.NewDomainRepository(db)
		certRepo := repository.NewCertRepository(db)
		domainSvc := service.NewDomainService(domainRepo, certRepo)
		domainHandler := handler.NewDomainHandler(domainSvc)
		domains := protected.Group("/domains")
		domains.Use(middleware.PermissionMiddleware("domains", "manage"))
		{
			domainHandler.RegisterRoutes(domains)
		}

		// 证书管理路由
		certificateRepo := repository.NewCertificateRepository(db)
		certCSRRepo := repository.NewCSRRepository(db)
		certCredentialRepo := repository.NewCredentialRepository(db)
		certificateSvc = service.NewCertificateService(certificateRepo, certCSRRepo, certCredentialRepo, cfg.Security.AESKey)
		certificateHandler := handler.NewCertificateHandler(certificateSvc)
		certificates := protected.Group("/certificates")
		certificates.Use(middleware.PermissionMiddleware("certificates", "manage"))
		{
			certificateHandler.RegisterRoutes(certificates)
		}

		// 部署任务管理路由
		deployRepo := repository.NewDeployRepository(db)
		deployCredentialRepo := repository.NewCredentialRepository(db)
		deployCertRepo := repository.NewCertificateRepository(db)
		deploySvc := service.NewDeployService(deployRepo, deployCredentialRepo, deployCertRepo, cfg.Security.AESKey)
		deployHandler := handler.NewDeployHandler(deploySvc)
		deploys := protected.Group("/deploys")
		deploys.Use(middleware.PermissionMiddleware("deploys", "manage"))
		{
			deployHandler.RegisterRoutes(deploys)
		}

		// 大盘统计路由
		dashboardSvc := service.NewDashboardService(db)
		dashboardHandler := handler.NewDashboardHandler(dashboardSvc)
		dashboardHandler.RegisterRoutes(protected)

		// 通知管理路由
		notificationRepo := repository.NewNotificationRepository(db)
		notificationSvc := service.NewNotificationService(notificationRepo)
		notificationHandler := handler.NewNotificationHandler(notificationSvc)
		notifications := protected.Group("/notifications")
		notifications.Use(middleware.PermissionMiddleware("notifications", "manage"))
		{
			notificationHandler.RegisterRoutes(notifications)
		}

		// Nginx 集群管理路由
		nginxRepo := repository.NewNginxRepository(db)
		nginxCertRepo := repository.NewCertificateRepository(db)
		nginxSvc := service.NewNginxService(nginxRepo, nginxCertRepo, cfg.Security.AESKey)
		nginxHandler := handler.NewNginxHandler(nginxSvc)
		nginx := protected.Group("/nginx")
		nginx.Use(middleware.PermissionMiddleware("nginx", "manage"))
		{
			nginxHandler.RegisterRoutes(nginx)
		}

		// 审计日志路由
		auditHandler := handler.NewAuditHandler(db)
		auditHandler.RegisterRoutes(protected)
	}

	// WebSocket 路由（在 API 组外）
	{
		deployRepo := repository.NewDeployRepository(db)
		deployCredentialRepo := repository.NewCredentialRepository(db)
		deployCertRepo := repository.NewCertificateRepository(db)
		deploySvc := service.NewDeployService(deployRepo, deployCredentialRepo, deployCertRepo, cfg.Security.AESKey)
		wsHandler := handler.NewDeployWebSocketHandler(deploySvc)
		wsHandler.RegisterWebSocketRoutes(r)
	}

	// 初始化并启动定时任务调度器
	notificationRepo := repository.NewNotificationRepository(db)
	notificationSvc := service.NewNotificationService(notificationRepo)
	sched := scheduler.NewScheduler(db, notificationSvc, certificateSvc)
	sched.Start()
	defer sched.Stop()
	fmt.Println("Scheduler started successfully")

	// 启动 HTTP 服务
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// 优雅关闭
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
	fmt.Printf("Server started on port %d\n", cfg.Server.Port)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	// 优雅关闭，设置超时时间
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// 关闭数据库连接
	sqlDB, err := db.DB()
	if err == nil {
		sqlDB.Close()
	}
	if rdb != nil {
		rdb.Close()
	}

	fmt.Println("Server exited")
}

// initMySQL 初始化 MySQL 连接并执行 AutoMigrate
func initMySQL(cfg *config.Config) error {
	var err error
	db, err = gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	// 自动迁移所有模型
	models := []interface{}{
		&model.Certificate{},
		&model.CSRRecord{},
		&model.CloudCredential{},
		&model.Domain{},
		&model.DeployTask{},
		&model.DeployTaskItem{},
		&model.DeploySnapshot{},
		&model.NginxCluster{},
		&model.NginxNode{},
		&model.NotificationRule{},
		&model.NotificationLog{},
		&model.AuditLog{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
	}

	for _, m := range models {
		if err := db.AutoMigrate(m); err != nil {
			return fmt.Errorf("failed to migrate %T: %w", m, err)
		}
	}

	return nil
}

// initRedis 初始化 Redis 连接
func initRedis(cfg *config.Config) error {
	rdb = redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

// seedDatabase 初始化数据库种子数据
func seedDatabase() error {
	// 检查是否已存在 admin 角色
	var roleCount int64
	db.Model(&model.Role{}).Count(&roleCount)
	if roleCount > 0 {
		// 已存在数据，跳过种子
		return nil
	}

	fmt.Println("Seeding database...")

	// 创建默认权限
	permissions := []model.Permission{
		{Name: "全部权限", Resource: "*", Action: "*"},
		{Name: "用户查看", Resource: "users", Action: "view"},
		{Name: "用户管理", Resource: "users", Action: "manage"},
		{Name: "角色查看", Resource: "roles", Action: "view"},
		{Name: "角色管理", Resource: "roles", Action: "manage"},
		{Name: "证书查看", Resource: "certificates", Action: "view"},
		{Name: "证书管理", Resource: "certificates", Action: "manage"},
		{Name: "域名查看", Resource: "domains", Action: "view"},
		{Name: "域名管理", Resource: "domains", Action: "manage"},
		{Name: "凭证查看", Resource: "credentials", Action: "view"},
		{Name: "凭证管理", Resource: "credentials", Action: "manage"},
		{Name: "部署查看", Resource: "deploys", Action: "view"},
		{Name: "部署管理", Resource: "deploys", Action: "manage"},
		{Name: "Nginx查看", Resource: "nginx", Action: "view"},
		{Name: "Nginx管理", Resource: "nginx", Action: "manage"},
		{Name: "通知查看", Resource: "notifications", Action: "view"},
		{Name: "通知管理", Resource: "notifications", Action: "manage"},
		{Name: "审计日志查看", Resource: "audit_logs", Action: "view"},
	}

	for i := range permissions {
		if err := db.Create(&permissions[i]).Error; err != nil {
			return fmt.Errorf("failed to create permission: %w", err)
		}
	}

	// 创建默认角色
	// admin 角色 - 拥有所有权限
	adminRole := model.Role{
		Name:        "admin",
		Description: "管理员角色，拥有所有权限",
	}
	if err := db.Create(&adminRole).Error; err != nil {
		return fmt.Errorf("failed to create admin role: %w", err)
	}
	// 分配所有权限给 admin 角色
	var allPermissions []model.Permission
	db.Find(&allPermissions)
	if err := db.Model(&adminRole).Association("Permissions").Replace(allPermissions); err != nil {
		return fmt.Errorf("failed to assign permissions to admin role: %w", err)
	}

	// operator 角色 - 拥有操作权限
	operatorRole := model.Role{
		Name:        "operator",
		Description: "操作员角色，拥有证书、域名、部署等操作权限",
	}
	if err := db.Create(&operatorRole).Error; err != nil {
		return fmt.Errorf("failed to create operator role: %w", err)
	}
	// 分配操作权限给 operator 角色
	var operatorPermissions []model.Permission
	db.Where("resource IN ?", []string{"certificates", "domains", "deploys", "credentials", "nginx", "notifications"}).Find(&operatorPermissions)
	if err := db.Model(&operatorRole).Association("Permissions").Replace(operatorPermissions); err != nil {
		return fmt.Errorf("failed to assign permissions to operator role: %w", err)
	}

	// viewer 角色 - 只读权限
	viewerRole := model.Role{
		Name:        "viewer",
		Description: "只读角色，仅有查看权限",
	}
	if err := db.Create(&viewerRole).Error; err != nil {
		return fmt.Errorf("failed to create viewer role: %w", err)
	}
	// 分配只读权限给 viewer 角色
	var viewerPermissions []model.Permission
	db.Where("action = ?", "view").Find(&viewerPermissions)
	if err := db.Model(&viewerRole).Association("Permissions").Replace(viewerPermissions); err != nil {
		return fmt.Errorf("failed to assign permissions to viewer role: %w", err)
	}

	// 创建默认 admin 用户
	hashedPassword, err := service.HashPassword("admin123")
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	adminUser := model.User{
		Username: "admin",
		Password: hashedPassword,
		Email:    "admin@example.com",
		Nickname: "Administrator",
		RoleID:   adminRole.ID,
		Status:   "active",
	}
	if err := db.Create(&adminUser).Error; err != nil {
		return fmt.Errorf("failed to create admin user: %w", err)
	}

	fmt.Println("Database seed completed: admin user created with password 'admin123'")
	return nil
}
