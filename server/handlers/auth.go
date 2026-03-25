package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"net/http"
	"time"
	"freelang-bank-system/server/database"
)

type AuthHandler struct {
	db        *database.DB
	jwtSecret string
}

func NewAuthHandler(db *database.DB, jwtSecret string) *AuthHandler {
	return &AuthHandler{
		db:        db,
		jwtSecret: jwtSecret,
	}
}

// User represents an authenticated user
type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"` // admin, user
}

// JWT Claims
type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// RegisterRequest
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// LoginRequest
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse
type AuthResponse struct {
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expires_at"`
}

// Register POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "잘못된 요청: " + err.Error(),
		})
		return
	}

	// 중복 확인
	var count int
	err := h.db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, req.Email).Scan(&count)
	if err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":   "Conflict",
			"message": "이미 가입된 이메일입니다",
		})
		return
	}

	// 비밀번호 해싱
	hashedPassword := hashPassword(req.Password)

	// 사용자 생성
	userID := "USER-" + uuid.New().String()[:8]
	now := time.Now().Unix()

	_, err = h.db.Exec(
		`INSERT INTO users (id, username, email, password_hash, role, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		userID, req.Username, req.Email, hashedPassword, "user", now, now,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": "사용자 생성 실패: " + err.Error(),
		})
		return
	}

	// JWT 토큰 생성
	token, expiresAt := h.generateToken(userID, req.Username, "user")

	c.JSON(http.StatusCreated, AuthResponse{
		Token:     token,
		UserID:    userID,
		Username:  req.Username,
		Email:     req.Email,
		ExpiresAt: expiresAt,
	})
}

// Login POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Bad Request",
			"message": "잘못된 요청: " + err.Error(),
		})
		return
	}

	// 사용자 조회
	var userID, username, passwordHash, role string
	err := h.db.QueryRow(
		`SELECT id, username, password_hash, role FROM users WHERE email = ?`,
		req.Email,
	).Scan(&userID, &username, &passwordHash, &role)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "이메일 또는 비밀번호가 잘못되었습니다",
		})
		return
	}

	// 비밀번호 검증
	if !verifyPassword(req.Password, passwordHash) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "이메일 또는 비밀번호가 잘못되었습니다",
		})
		return
	}

	// JWT 토큰 생성
	token, expiresAt := h.generateToken(userID, username, role)

	c.JSON(http.StatusOK, AuthResponse{
		Token:     token,
		UserID:    userID,
		Username:  username,
		Email:     req.Email,
		ExpiresAt: expiresAt,
	})
}

// RefreshToken POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// Authorization 헤더에서 토큰 추출
	tokenString := c.GetHeader("Authorization")
	if tokenString == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "토큰이 필요합니다",
		})
		return
	}

	// Bearer 스킴 제거
	if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
		tokenString = tokenString[7:]
	}

	// 토큰 파싱
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(h.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "유효하지 않은 토큰입니다",
		})
		return
	}

	// 새 토큰 생성
	newToken, expiresAt := h.generateToken(claims.UserID, claims.Username, claims.Role)

	c.JSON(http.StatusOK, AuthResponse{
		Token:     newToken,
		UserID:    claims.UserID,
		Username:  claims.Username,
		ExpiresAt: expiresAt,
	})
}

// GetProfile GET /api/auth/profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// 미들웨어에서 추출된 user 정보
	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "인증이 필요합니다",
		})
		return
	}

	userID := user.(string)

	// 사용자 정보 조회
	var username, email, role string
	err := h.db.QueryRow(
		`SELECT username, email, role FROM users WHERE id = ?`,
		userID,
	).Scan(&username, &email, &role)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not Found",
			"message": "사용자를 찾을 수 없습니다",
		})
		return
	}

	c.JSON(http.StatusOK, User{
		ID:       userID,
		Username: username,
		Email:    email,
		Role:     role,
	})
}

// generateToken generates a JWT token
func (h *AuthHandler) generateToken(userID, username, role string) (string, int64) {
	expiresAt := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(h.jwtSecret))

	return tokenString, expiresAt.Unix()
}

// hashPassword hashes a password using SHA256
func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

// verifyPassword verifies a password against its hash
func verifyPassword(password, hash string) bool {
	return hashPassword(password) == hash
}

// AuthMiddleware validates JWT token
func AuthMiddleware(jwtSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "토큰이 필요합니다",
			})
			c.Abort()
			return
		}

		// Bearer 스킴 제거
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		// 토큰 파싱
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "유효하지 않은 토큰입니다",
			})
			c.Abort()
			return
		}

		// userID를 context에 저장
		c.Set("user", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// AdminMiddleware validates admin role
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role.(string) != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "관리자 권한이 필요합니다",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}
