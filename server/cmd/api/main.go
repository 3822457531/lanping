package main

import (
	"database/sql"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lanping/lp-server/internal/auth"
	"github.com/lanping/lp-server/internal/config"
	"github.com/lanping/lp-server/internal/handler"
	"github.com/lanping/lp-server/internal/middleware"
	"github.com/lanping/lp-server/internal/repo"
	"github.com/lanping/lp-server/internal/service"
	"github.com/lanping/lp-server/internal/sms"
	"github.com/pressly/goose/v3"
)

func main() {
	cfgPath := os.Getenv("LP_CONFIG")
	if cfgPath == "" {
		cfgPath = "configs/config.yaml"
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatal(err)
	}

	db, err := repo.Open(cfg.DB.DSN)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer db.Close()

	if cfg.Migrate.Auto {
		dir := cfg.Migrate.Dir
		if !filepath.IsAbs(dir) {
			dir = filepath.Clean(dir)
		}
		if err := runMigrate(db.DB, dir); err != nil {
			log.Fatalf("migrate: %v", err)
		}
	}

	jwtMgr := auth.NewManager(cfg.JWT.Secret, cfg.JWT.ExpireHours)
	rstore := repo.New(db)
	svc := service.New(rstore, jwtMgr)
	smsClient, err := sms.New(sms.Config{
		Enabled:         cfg.SMS.Enabled,
		AccessKeyID:     cfg.SMS.AccessKeyID,
		AccessKeySecret: cfg.SMS.AccessKeySecret,
		Endpoint:        cfg.SMS.Endpoint,
		SignName:        cfg.SMS.SignName,
		TplBirthday:     cfg.SMS.TplBirthday,
		TplConsume:      cfg.SMS.TplConsume,
		TplAppoint:      cfg.SMS.TplAppoint,
	})
	if err != nil {
		log.Printf("sms init warning: %v", err)
		smsClient, _ = sms.New(sms.Config{SignName: "兰瓶"})
	}
	svc.SetSMS(smsClient)
	if err := svc.ReloadSMSFromSettings(); err != nil {
		log.Printf("sms reload from settings: %v", err)
	} else if svc.SMS != nil && svc.SMS.Enabled() {
		log.Printf("aliyun sms ready, sign=%s", svc.SMS.SignName())
	} else {
		log.Printf("aliyun sms idle (可在管理后台「系统设置」配置 AccessKey)")
	}
	api := &handler.API{Svc: svc}

	gin.SetMode(cfg.HTTP.Mode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())

	r.GET("/health", api.Health)
	r.GET("/api/health", api.Health)

	r.POST("/api/auth/login", api.Login)
	r.POST("/api/auth/register", api.Register)
	r.POST("/api/customer/login", api.CustomerLogin)
	r.GET("/api/public/stores", api.ListStores)
	r.GET("/api/public/verification/:token", api.PublicVerification)
	r.POST("/api/public/verification/:token/confirm", api.ConfirmPublicVerification)

	authz := r.Group("/api")
	authz.Use(middleware.JWT(jwtMgr))
	{
		authz.GET("/auth/me", api.Me)
		authz.GET("/customer/me", api.CustomerMe)
		authz.GET("/stores", api.ListStores)
		authz.POST("/stores", middleware.Roles("admin"), api.CreateStore)
		authz.PUT("/stores/:id", middleware.Roles("admin"), api.UpdateStore)
		authz.GET("/geo/search", middleware.Roles("admin", "manager"), api.GeoSearch)
		authz.GET("/dashboard", api.Dashboard)
		authz.GET("/workbench", api.Workbench)

		authz.GET("/customers", api.ListCustomers)
		authz.GET("/customers/birthday/upcoming", api.BirthdayUpcoming)
		authz.POST("/customers/birthday/notify", middleware.Roles("admin", "manager"), api.BirthdayNotify)
		authz.GET("/customers/:id", api.GetCustomer)
		authz.POST("/customers", api.CreateCustomer)
		authz.PUT("/customers/:id", api.UpdateCustomer)
		authz.DELETE("/customers/:id", api.DeleteCustomer)

		authz.GET("/users", api.ListUsers)
		authz.POST("/users", middleware.Roles("admin", "manager"), api.CreateUser)
		authz.PUT("/users/:id", middleware.Roles("admin", "manager"), api.UpdateUser)

		authz.GET("/projects", api.ListProjects)
		authz.POST("/projects", middleware.Roles("admin"), api.CreateProject)
		authz.PUT("/projects/:id", middleware.Roles("admin"), api.UpdateProject)
		authz.DELETE("/projects/:id", middleware.Roles("admin"), api.DeleteProject)

		authz.GET("/cards", api.ListCards)
		authz.POST("/cards", middleware.Roles("admin", "manager", "staff"), api.CreateCard)

		authz.GET("/verifications", api.ListVerifies)
		authz.POST("/verifications/draft", middleware.Roles("admin", "manager", "staff"), api.DraftVerify)
		authz.GET("/verifications/:id", api.GetVerify)
		authz.POST("/verifications/:id/confirm", middleware.Roles("admin", "manager", "staff"), api.ConfirmVerify)
		authz.POST("/verifications/:id/cancel", middleware.Roles("admin", "manager", "staff"), api.CancelVerify)
		authz.POST("/verifications/:id/refund", middleware.Roles("admin", "manager"), api.RefundVerify)

		authz.GET("/appointments", api.ListAppointments)
		authz.POST("/appointments", api.CreateAppointment)
		authz.PUT("/appointments/:id", api.UpdateAppointment)
		authz.POST("/appointments/:id/status", api.UpdateAppointment)
		authz.DELETE("/appointments/:id", api.DeleteAppointment)
		authz.GET("/staff/availability", api.ListStaffAvailability)

		authz.GET("/inventory/stock", api.ListInventory)
		authz.GET("/inventory/items", api.ListStockItems)
		authz.POST("/inventory/items", middleware.Roles("admin", "manager"), api.CreateStockItem)
		authz.PUT("/inventory/items/:id", middleware.Roles("admin", "manager"), api.UpdateStockItem)
		authz.DELETE("/inventory/items/:id", middleware.Roles("admin", "manager"), api.DeleteStockItem)
		authz.GET("/inventory/orders", api.ListInventoryOrders)
		authz.GET("/inventory/transfers/pending", api.ListPendingInventoryTransfers)
		authz.POST("/inventory/orders", middleware.Roles("admin", "manager"), api.CreateInventoryOrder)
		authz.POST("/inventory/orders/:id/confirm", middleware.Roles("admin", "manager"), api.ConfirmInventoryTransfer)

		authz.POST("/attendance/check", middleware.Roles("admin", "manager", "staff"), api.CheckIn)
		authz.GET("/attendance/today", api.AttendanceToday)
		authz.GET("/attendance", middleware.Roles("admin", "manager"), api.ListAttendance)

		authz.GET("/audits", api.ListAudits)
		authz.POST("/audits/:id/approve", middleware.Roles("admin", "manager"), api.ApproveAudit)
		authz.POST("/audits/:id/reject", middleware.Roles("admin", "manager"), api.RejectAudit)

		authz.GET("/settings", api.GetSettings)
		authz.PUT("/settings", middleware.Roles("admin", "manager"), api.SaveSettings)

		authz.GET("/assessments", api.ListAssessments)
		authz.GET("/assessments/result", api.ListAssessments)
		authz.POST("/assessments", middleware.Roles("admin", "manager"), api.CreateAssessment)
		authz.PUT("/assessments/:id", middleware.Roles("admin", "manager"), api.UpdateAssessment)
		authz.DELETE("/assessments/:id", middleware.Roles("admin"), api.DeleteAssessment)

		authz.GET("/salary/rules", api.GetSalaryRules)
		authz.PUT("/salary/rules", middleware.Roles("admin"), api.SaveSalaryRules)
		authz.POST("/salary/compute", middleware.Roles("admin"), api.ComputeSalary)
		authz.GET("/salary/records", api.ListSalary)
		authz.POST("/salary/records/:id/status", middleware.Roles("admin"), api.SalaryStatus)

		authz.GET("/welfare", api.ListWelfare)
		authz.GET("/welfare/issued", api.ListWelfareIssued)
		authz.GET("/welfare/available", api.ListAvailableWelfare)
		authz.POST("/welfare", middleware.Roles("admin"), api.CreateWelfare)
		authz.PUT("/welfare/:id", middleware.Roles("admin"), api.UpdateWelfare)
		authz.DELETE("/welfare/:id", middleware.Roles("admin"), api.DeleteWelfare)
		authz.POST("/welfare/:id/issue", middleware.Roles("admin"), api.IssueWelfare)

		authz.GET("/messages", api.ListMessages)
		authz.GET("/sms/logs", middleware.Roles("admin"), api.ListSMS)
		authz.GET("/operate-logs", middleware.Roles("admin"), api.ListOperateLogs)
	}

	// 定时将超时未处理预约置为「已失效」
	go func() {
		t := time.NewTicker(2 * time.Minute)
		defer t.Stop()
		for range t.C {
			if n, err := svc.ExpireStaleAppointments(); err != nil {
				log.Printf("appoint expire: %v", err)
			} else if n > 0 {
				log.Printf("appoint expire: marked %d as 已失效", n)
			}
		}
	}()

	log.Printf("lanping api listening on %s", cfg.HTTP.Addr)
	if err := r.Run(cfg.HTTP.Addr); err != nil {
		log.Fatal(err)
	}
}

func runMigrate(db *sql.DB, dir string) error {
	if err := goose.SetDialect("mysql"); err != nil {
		return err
	}
	return goose.Up(db, dir)
}
