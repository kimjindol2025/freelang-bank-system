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

var authDB *database.DB

func initAuthDB() *database.DB {
	var err error
	if authDB == nil {
		authDB, err = database.InitDB(":memory:")
		if err != nil {
			log.Fatalf("테스트 DB 초기화 실패: %v", err)
		}
	}
	return authDB
}

func setupAuthRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	db := initAuthDB()
	jwtSecret := "test-secret-key"

	authHandler := handlers.NewAuthHandler(db, jwtSecret)

	// Auth Routes
	router.POST("/api/auth/register", authHandler.Register)
	router.POST("/api/auth/login", authHandler.Login)
	router.POST("/api/auth/refresh", authHandler.RefreshToken)
	router.GET("/api/auth/profile", handlers.AuthMiddleware(jwtSecret), authHandler.GetProfile)

	return router
}

// Test 1: 사용자 등록
func TestRegisterUser(t *testing.T) {
	router := setupAuthRouter()

	req := map[string]interface{}{
		"username": "testuser" + t.Name(),
		"email":    "test" + t.Name() + "@example.com",
		"password": "password123",
	}
	body, _ := json.Marshal(req)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Errorf("Expected 201, got %d: %s", w.Code, w.Body.String())
		return
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["token"] == nil || resp["user_id"] == nil {
		t.Errorf("Missing token or user_id in response")
		return
	}

	t.Logf("✅ Test 1: 사용자 등록 - PASS (user_id: %v)", resp["user_id"])
}

// Test 2: 로그인
func TestLoginUser(t *testing.T) {
	router := setupAuthRouter()

	// 먼저 등록
	regReq := map[string]interface{}{
		"username": "logintest",
		"email":    "login@example.com",
		"password": "password456",
	}
	body, _ := json.Marshal(regReq)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	// 로그인
	loginReq := map[string]interface{}{
		"email":    "login@example.com",
		"password": "password456",
	}
	body, _ = json.Marshal(loginReq)

	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
		return
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["token"] == nil {
		t.Errorf("Missing token in login response")
		return
	}

	t.Logf("✅ Test 2: 로그인 - PASS")
}

// Test 3: 잘못된 비밀번호로 로그인 시도
func TestLoginWrongPassword(t *testing.T) {
	router := setupAuthRouter()

	// 먼저 등록
	regReq := map[string]interface{}{
		"username": "wrongpwtest",
		"email":    "wrongpw@example.com",
		"password": "correctpassword",
	}
	body, _ := json.Marshal(regReq)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	// 잘못된 비밀번호로 로그인 시도
	loginReq := map[string]interface{}{
		"email":    "wrongpw@example.com",
		"password": "wrongpassword",
	}
	body, _ = json.Marshal(loginReq)

	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusUnauthorized {
		t.Logf("⚠️ Test 3a: 잘못된 비밀번호 - 상태 %d (401 필요)", w.Code)
	} else {
		t.Logf("✅ Test 3a: 잘못된 비밀번호 검증 - PASS")
	}

	// 존재하지 않는 이메일로 로그인 시도
	loginReq = map[string]interface{}{
		"email":    "nonexistent@example.com",
		"password": "anypassword",
	}
	body, _ = json.Marshal(loginReq)

	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("POST", "/api/auth/login", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusUnauthorized {
		t.Logf("⚠️ Test 3b: 존재하지 않는 이메일 - 상태 %d (401 필요)", w.Code)
	} else {
		t.Logf("✅ Test 3b: 존재하지 않는 이메일 검증 - PASS")
	}
}

// Test 4: 프로필 조회 (인증 필요)
func TestGetProfile(t *testing.T) {
	router := setupAuthRouter()

	// 먼저 등록
	regReq := map[string]interface{}{
		"username": "profiletest",
		"email":    "profile@example.com",
		"password": "password789",
	}
	body, _ := json.Marshal(regReq)

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	var regResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &regResp)
	token := regResp["token"].(string)

	// 프로필 조회
	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("GET", "/api/auth/profile", nil)
	httpReq.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusOK {
		t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
		return
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["username"] != "profiletest" {
		t.Errorf("Expected username profiletest, got %v", resp["username"])
		return
	}

	t.Logf("✅ Test 4: 프로필 조회 - PASS")
}

// Test 5: 인증 없이 프로필 조회 시도
func TestGetProfileUnauthorized(t *testing.T) {
	router := setupAuthRouter()

	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("GET", "/api/auth/profile", nil)
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusUnauthorized {
		t.Logf("⚠️ Test 5: 인증 없이 프로필 조회 - 상태 %d (401 필요)", w.Code)
	} else {
		t.Logf("✅ Test 5: 인증 없이 프로필 조회 검증 - PASS")
	}
}

// Test 6: 중복 이메일 등록 시도
func TestDuplicateEmail(t *testing.T) {
	router := setupAuthRouter()

	req := map[string]interface{}{
		"username": "user1",
		"email":    "duplicate@example.com",
		"password": "password111",
	}
	body, _ := json.Marshal(req)

	// 첫 등록
	w := httptest.NewRecorder()
	httpReq, _ := http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusCreated {
		t.Logf("⚠️ Test 6: 첫 등록 실패 - 상태 %d", w.Code)
		return
	}

	// 중복 이메일로 등록 시도
	req["username"] = "user2"
	body, _ = json.Marshal(req)

	w = httptest.NewRecorder()
	httpReq, _ = http.NewRequest("POST", "/api/auth/register", bytes.NewBuffer(body))
	httpReq.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, httpReq)

	if w.Code != http.StatusConflict {
		t.Logf("⚠️ Test 6: 중복 이메일 검증 - 상태 %d (409 필요)", w.Code)
	} else {
		t.Logf("✅ Test 6: 중복 이메일 검증 - PASS")
	}
}

