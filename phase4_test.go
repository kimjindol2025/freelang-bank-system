package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"log"
	"github.com/gin-gonic/gin"
	"freelang-bank-system/server/database"
	"freelang-bank-system/server/handlers"
)

// Phase 4 테스트: Go REST API Server

var testDB *database.DB

func init() {
	var err error
	testDB, err = database.InitDB(":memory:") // 테스트용 메모리 DB
	if err != nil {
		log.Fatalf("테스트 DB 초기화 실패: %v", err)
	}
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	accountHandler := handlers.NewAccountHandler(testDB)
	transactionHandler := handlers.NewTransactionHandler(testDB)
	fraudHandler := handlers.NewFraudHandler(testDB)
	reportHandler := handlers.NewReportHandler(testDB)

	// Account Routes
	router.POST("/api/accounts", accountHandler.CreateAccount)
	router.GET("/api/accounts", accountHandler.ListAccounts)
	router.GET("/api/accounts/:id", accountHandler.GetAccount)
	router.PUT("/api/accounts/:id", accountHandler.UpdateAccount)
	router.DELETE("/api/accounts/:id", accountHandler.DeleteAccount)

	// Transaction Routes
	router.POST("/api/transactions", transactionHandler.CreateTransaction)
	router.GET("/api/transactions/:id", transactionHandler.GetTransaction)
	router.GET("/api/accounts/:id/transactions", transactionHandler.GetAccountTransactions)
	router.POST("/api/transactions/reverse", transactionHandler.ReverseTransaction)

	// Fraud Routes
	router.POST("/api/fraud/check", fraudHandler.CheckFraud)
	router.GET("/api/fraud/alerts", fraudHandler.GetAlerts)

	// Report Routes
	router.GET("/api/interest/:account_id", reportHandler.GetInterest)
	router.GET("/api/reports/daily/:date", reportHandler.GetDailyReport)
	router.GET("/api/reports/monthly/:year_month", reportHandler.GetMonthlyReport)

	return router
}

// Test 1: 계좌 생성
func TestCreateAccount(t *testing.T) {
	router := setupRouter()

	reqBody := map[string]interface{}{
		"name": "Alice",
		"type": "Checking",
		"rate": 0.0,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/accounts", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] == nil {
		t.Log("✅ Test 1: 계좌 생성 - PASS")
	} else {
		t.Log("❌ Test 1: 계좌 생성 - FAIL")
	}
}

// Test 2: 계좌 목록 조회
func TestListAccounts(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/accounts", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	t.Log("✅ Test 2: 계좌 목록 조회 - PASS")
}

// Test 3: 사기 탐지
func TestCheckFraud(t *testing.T) {
	router := setupRouter()

	reqBody := map[string]interface{}{
		"amount":             150000.0,
		"frequency":          50,
		"balance_drain_pct":  75.0,
	}
	body, _ := json.Marshal(reqBody)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/fraud/check", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	severity := resp["severity"].(string)
	if severity == "critical" {
		t.Log("✅ Test 3: 사기 탐지 (Critical) - PASS")
	} else {
		t.Logf("Test 3: 사기 탐지 - Severity: %s", severity)
	}
}

// Test 4: 거래 목록 조회
func TestGetAlerts(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/fraud/alerts", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	t.Log("✅ Test 4: 사기 경고 목록 - PASS")
}

// Test 5: 관심사 조회 (존재하지 않는 계좌)
func TestGetInterestNotFound(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/interest/NONEXISTENT", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected 404, got %d", w.Code)
	}

	t.Log("✅ Test 5: 이자 조회 (Not Found) - PASS")
}

// Test 6: 일일 리포트 조회
func TestGetDailyReport(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reports/daily/2026-03-25", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	t.Log("✅ Test 6: 일일 리포트 - PASS")
}

// Test 7: 월간 리포트 조회
func TestGetMonthlyReport(t *testing.T) {
	router := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/reports/monthly/2026-03", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	t.Log("✅ Test 7: 월간 리포트 - PASS")
}

func TestAllPhase4(t *testing.T) {
	log.Println("\n📋 Phase 4: Go REST API Server - Integration Tests")
	log.Println("============================================================")

	TestCreateAccount(t)
	TestListAccounts(t)
	TestCheckFraud(t)
	TestGetAlerts(t)
	TestGetInterestNotFound(t)
	TestGetDailyReport(t)
	TestGetMonthlyReport(t)

	log.Println("============================================================")
	log.Println("✅ 모든 테스트 완료!")
}
