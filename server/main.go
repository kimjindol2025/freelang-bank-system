package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"freelang-bank-system/server/handlers"
	"freelang-bank-system/server/database"
)

func main() {
	// 데이터베이스 초기화
	db, err := database.InitDB("freelang_bank.db")
	if err != nil {
		log.Fatalf("데이터베이스 초기화 실패: %v", err)
	}
	defer db.Close()

	// Gin 라우터 설정
	router := gin.Default()

	// 미들웨어
	router.Use(corsMiddleware())
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 헬스 체크
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
			"message": "FreeLang Bank Server is running",
		})
	})

	// Account 핸들러
	accountHandler := handlers.NewAccountHandler(db)
	router.POST("/api/accounts", accountHandler.CreateAccount)
	router.GET("/api/accounts", accountHandler.ListAccounts)
	router.GET("/api/accounts/:id", accountHandler.GetAccount)
	router.PUT("/api/accounts/:id", accountHandler.UpdateAccount)
	router.DELETE("/api/accounts/:id", accountHandler.DeleteAccount)

	// Transaction 핸들러
	transactionHandler := handlers.NewTransactionHandler(db)
	router.POST("/api/transactions", transactionHandler.CreateTransaction)
	router.GET("/api/transactions/:id", transactionHandler.GetTransaction)
	router.GET("/api/accounts/:id/transactions", transactionHandler.GetAccountTransactions)
	router.POST("/api/transactions/reverse", transactionHandler.ReverseTransaction)

	// Fraud Detection 핸들러
	fraudHandler := handlers.NewFraudHandler(db)
	router.POST("/api/fraud/check", fraudHandler.CheckFraud)
	router.GET("/api/fraud/alerts", fraudHandler.GetAlerts)

	// Interest & Reports 핸들러
	reportHandler := handlers.NewReportHandler(db)
	router.GET("/api/interest/:account_id", reportHandler.GetInterest)
	router.GET("/api/reports/daily/:date", reportHandler.GetDailyReport)
	router.GET("/api/reports/monthly/:year_month", reportHandler.GetMonthlyReport)

	// Auth 핸들러
	jwtSecret := "freelang-bank-secret-key-2026" // 프로덕션에서는 환경변수에서 로드
	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.POST("/api/auth/refresh", authHandler.RefreshToken)
	router.GET("/api/auth/profile", handlers.AuthMiddleware(jwtSecret), authHandler.GetProfile)

	// Protected routes (JWT 미들웨어 적용)
	protected := router.Group("/api")
	protected.Use(handlers.AuthMiddleware(jwtSecret))
	{
		// 이후 추가될 보호된 엔드포인트들을 여기에 추가
	}

	// 서버 시작
	log.Println("🚀 FreeLang Bank Server 시작...")
	log.Println("📍 http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("서버 시작 실패: %v", err)
	}
}

// CORS 미들웨어
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
