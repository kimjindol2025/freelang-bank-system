package main

import (
	"fmt"
	"net/http"
	"os"

	"freelang-bank-system/handlers"
	"freelang-bank-system/src/db"
	"github.com/gin-gonic/gin"
)

func main() {
	// 데이터베이스 초기화
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "bank.db"
	}

	if err := db.InitDB(dbPath); err != nil {
		fmt.Println("❌ 데이터베이스 초기화 실패:", err)
		os.Exit(1)
	}
	defer db.GetDB().Close()

	// Gin 라우터 초기화
	router := gin.Default()

	// CORS 설정
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// ========================================
	// 헬스 체크
	// ========================================
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "FreeLang Bank System",
			"version": "1.0.0",
		})
	})

	// ========================================
	// 인증 엔드포인트
	// ========================================
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", handlers.Register)
		authGroup.POST("/login", handlers.Login)
	}

	// ========================================
	// 계좌 엔드포인트 (인증 필요)
	// ========================================
	accountGroup := router.Group("/api/accounts")
	accountGroup.Use(handlers.AuthMiddleware())
	{
		accountGroup.POST("", handlers.CreateAccount)
		accountGroup.GET("", handlers.ListAccounts)
		accountGroup.GET("/:id", handlers.GetAccount)
		accountGroup.GET("/:id/balance", handlers.GetBalance)
		accountGroup.POST("/:id/deposit", handlers.DepositToAccount)
		accountGroup.POST("/:id/withdraw", handlers.WithdrawFromAccount)
		accountGroup.DELETE("/:id", handlers.CloseAccount)
		accountGroup.GET("/:id/transactions", handlers.GetAccountTransactions)
	}

	// ========================================
	// 거래 엔드포인트 (인증 필요)
	// ========================================
	transactionGroup := router.Group("/api/transactions")
	transactionGroup.Use(handlers.AuthMiddleware())
	{
		transactionGroup.POST("", handlers.Transfer)
		transactionGroup.GET("/:id", handlers.GetTransaction)
		transactionGroup.POST("/:id/reverse", handlers.ReverseTransaction)
	}

	// ========================================
	// 서버 시작
	// ========================================
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║   FreeLang Bank System - Server       ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("✅ Database:", dbPath)
	fmt.Println("📍 Server:  http://localhost:" + port)
	fmt.Println()
	fmt.Println("🔐 API Endpoints:")
	fmt.Println("  - Authentication:")
	fmt.Println("    POST   /api/auth/register")
	fmt.Println("    POST   /api/auth/login")
	fmt.Println()
	fmt.Println("  - Accounts:")
	fmt.Println("    POST   /api/accounts")
	fmt.Println("    GET    /api/accounts")
	fmt.Println("    GET    /api/accounts/:id")
	fmt.Println("    GET    /api/accounts/:id/balance")
	fmt.Println("    POST   /api/accounts/:id/deposit")
	fmt.Println("    POST   /api/accounts/:id/withdraw")
	fmt.Println("    DELETE /api/accounts/:id")
	fmt.Println()
	fmt.Println("  - Transactions:")
	fmt.Println("    POST   /api/transactions")
	fmt.Println("    GET    /api/transactions/:id")
	fmt.Println("    GET    /api/accounts/:id/transactions")
	fmt.Println("    POST   /api/transactions/:id/reverse")
	fmt.Println()

	if err := router.Run(":" + port); err != nil {
		fmt.Println("❌ 서버 시작 실패:", err)
		os.Exit(1)
	}
}
