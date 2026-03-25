package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
	"freelang-bank-system/server/database"
)

type AccountHandler struct {
	db *database.DB
}

func NewAccountHandler(db *database.DB) *AccountHandler {
	return &AccountHandler{db: db}
}

type CreateAccountRequest struct {
	Name string  `json:"name" binding:"required"`
	Type string  `json:"type" binding:"required"`
	Rate float64 `json:"rate"`
}

type CreateAccountResponse struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Balance   float64 `json:"balance"`
	Status    string  `json:"status"`
	Message   string  `json:"message"`
}

// CreateAccount POST /api/accounts
func (h *AccountHandler) CreateAccount(c *gin.Context) {
	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad Request",
			"message": "잘못된 요청: " + err.Error(),
		})
		return
	}

	// 계좌 생성
	account := database.Account{
		ID:        "ACC" + uuid.New().String()[:8],
		Name:      req.Name,
		Type:      req.Type,
		Balance:   0.0,
		Rate:      req.Rate,
		Status:    "active",
		CreatedAt: time.Now().Unix(),
		UpdatedAt: time.Now().Unix(),
	}

	// 데이터베이스에 저장
	_, err := h.db.Exec(
		`INSERT INTO accounts (id, name, type, balance, rate, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		account.ID, account.Name, account.Type, account.Balance, account.Rate,
		account.Status, account.CreatedAt, account.UpdatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "계좌 생성 실패: " + err.Error(),
		})
		return
	}

	// 감시 로그 기록
	h.logAudit("ACCOUNT_CREATED", account.ID, "새 계좌 생성", c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusCreated, CreateAccountResponse{
		ID:      account.ID,
		Name:    account.Name,
		Type:    account.Type,
		Balance: account.Balance,
		Status:  account.Status,
		Message: "계좌 생성 완료",
	})
}

// ListAccounts GET /api/accounts
func (h *AccountHandler) ListAccounts(c *gin.Context) {
	rows, err := h.db.Query(`SELECT id, name, type, balance, rate, status, created_at, updated_at FROM accounts`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "계좌 목록 조회 실패: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var accounts []database.Account
	for rows.Next() {
		var acc database.Account
		if err := rows.Scan(&acc.ID, &acc.Name, &acc.Type, &acc.Balance, &acc.Rate,
			&acc.Status, &acc.CreatedAt, &acc.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal Server Error",
				"message": "행 스캔 실패: " + err.Error(),
			})
			return
		}
		accounts = append(accounts, acc)
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
		"count":    len(accounts),
	})
}

// GetAccount GET /api/accounts/:id
func (h *AccountHandler) GetAccount(c *gin.Context) {
	id := c.Param("id")

	var account database.Account
	err := h.db.QueryRow(
		`SELECT id, name, type, balance, rate, status, created_at, updated_at FROM accounts WHERE id = ?`,
		id,
	).Scan(&account.ID, &account.Name, &account.Type, &account.Balance, &account.Rate,
		&account.Status, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"message": "계좌를 찾을 수 없습니다: " + id,
		})
		return
	}

	c.JSON(http.StatusOK, account)
}

type UpdateAccountRequest struct {
	Name   string  `json:"name"`
	Status string  `json:"status"`
	Rate   float64 `json:"rate"`
}

// UpdateAccount PUT /api/accounts/:id
func (h *AccountHandler) UpdateAccount(c *gin.Context) {
	id := c.Param("id")
	var req UpdateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad Request",
			"message": "잘못된 요청: " + err.Error(),
		})
		return
	}

	// 기존 계좌 확인
	var account database.Account
	err := h.db.QueryRow(
		`SELECT id, name, type, balance, rate, status, created_at, updated_at FROM accounts WHERE id = ?`,
		id,
	).Scan(&account.ID, &account.Name, &account.Type, &account.Balance, &account.Rate,
		&account.Status, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"message": "계좌를 찾을 수 없습니다: " + id,
		})
		return
	}

	// 업데이트
	if req.Name != "" {
		account.Name = req.Name
	}
	if req.Status != "" {
		account.Status = req.Status
	}
	if req.Rate > 0 {
		account.Rate = req.Rate
	}
	account.UpdatedAt = time.Now().Unix()

	_, err = h.db.Exec(
		`UPDATE accounts SET name = ?, status = ?, rate = ?, updated_at = ? WHERE id = ?`,
		account.Name, account.Status, account.Rate, account.UpdatedAt, id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "계좌 업데이트 실패: " + err.Error(),
		})
		return
	}

	// 감시 로그 기록
	h.logAudit("ACCOUNT_UPDATED", id, "계좌 업데이트", c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, gin.H{
		"id":     id,
		"status": "updated",
		"message": "계좌 업데이트 완료",
	})
}

// DeleteAccount DELETE /api/accounts/:id
func (h *AccountHandler) DeleteAccount(c *gin.Context) {
	id := c.Param("id")

	// 계좌 존재 확인
	var exists int
	err := h.db.QueryRow(`SELECT COUNT(*) FROM accounts WHERE id = ?`, id).Scan(&exists)
	if err != nil || exists == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"message": "계좌를 찾을 수 없습니다: " + id,
		})
		return
	}

	_, err = h.db.Exec(`DELETE FROM accounts WHERE id = ?`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "계좌 삭제 실패: " + err.Error(),
		})
		return
	}

	// 감시 로그 기록
	h.logAudit("ACCOUNT_DELETED", id, "계좌 삭제", c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusNoContent, nil)
}

// logAudit 감시 로그 기록
func (h *AccountHandler) logAudit(action, accountID, description, ipAddress, userAgent string) {
	auditLog := database.AuditLog{
		ID:          "AUDIT-" + uuid.New().String()[:8],
		Action:      action,
		AccountID:   accountID,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Timestamp:   time.Now().Unix(),
	}

	_, err := h.db.Exec(
		`INSERT INTO audit_logs (id, action, account_id, description, ip_address, user_agent, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		auditLog.ID, auditLog.Action, auditLog.AccountID, auditLog.Description,
		auditLog.IPAddress, auditLog.UserAgent, auditLog.Timestamp,
	)

	if err != nil {
		// 로그 기록 실패는 무시
	}
}
