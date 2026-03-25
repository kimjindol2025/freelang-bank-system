package handlers

import (
	"fmt"
	"net/http"
	"time"

	"freelang-bank-system/src/db"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const (
	JWTSecret = "freelang-bank-secret-key-2026"
	TokenExp  = 24 * time.Hour
)

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Phone     string `json:"phone"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

// Register - 회원가입
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "입력값 검증 실패: " + err.Error()})
		return
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	// 중복 사용자 확인
	existingUser, _ := database.GetUserByEmail(req.Email)
	if existingUser != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "이미 존재하는 이메일입니다"})
		return
	}

	// 비밀번호 해싱
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "비밀번호 해싱 실패"})
		return
	}

	// 사용자 생성
	user := &db.User{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		Phone:        req.Phone,
		Status:       "active",
	}

	if err := database.CreateUser(user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "사용자 생성 실패: " + err.Error()})
		return
	}

	// JWT 토큰 생성
	expiresAt := time.Now().Add(TokenExp)
	token, err := generateToken(user.ID, user.Email, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 생성 실패"})
		return
	}

	// 감시 로그
	database.LogAudit(&db.AuditLog{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Action:    "REGISTER",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, AuthResponse{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	})
}

// Login - 로그인
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "입력값 검증 실패"})
		return
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	// 사용자 조회
	user, err := database.GetUserByEmail(req.Email)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 올바르지 않습니다"})
		return
	}

	// 비밀번호 검증
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "이메일 또는 비밀번호가 올바르지 않습니다"})
		return
	}

	// JWT 토큰 생성
	expiresAt := time.Now().Add(TokenExp)
	token, err := generateToken(user.ID, user.Email, expiresAt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "토큰 생성 실패"})
		return
	}

	// 감시 로그
	database.LogAudit(&db.AuditLog{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Action:    "LOGIN",
		IPAddress: c.ClientIP(),
		UserAgent: c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, AuthResponse{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: expiresAt.Unix(),
	})
}

// generateToken - JWT 토큰 생성
func generateToken(userID, email string, expiresAt time.Time) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(JWTSecret))
}

// verifyToken - JWT 토큰 검증
func VerifyToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("토큰 검증 실패")
	}

	return claims, nil
}

// AuthMiddleware - JWT 인증 미들웨어
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization 헤더가 필요합니다"})
			c.Abort()
			return
		}

		// "Bearer {token}" 형식 파싱
		var token string
		_, err := fmt.Sscanf(authHeader, "Bearer %s", &token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format"})
			c.Abort()
			return
		}

		claims, err := VerifyToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			c.Abort()
			return
		}

		// Context에 UserID 저장
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Next()
	}
}

// GetUserID - Context에서 UserID 가져오기
func GetUserID(c *gin.Context) (string, error) {
	userID, exists := c.Get("userID")
	if !exists {
		return "", fmt.Errorf("userID not found in context")
	}
	return userID.(string), nil
}
