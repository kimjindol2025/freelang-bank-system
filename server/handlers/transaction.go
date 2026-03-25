package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
	"freelang-bank-system/server/database"
)

type TransactionHandler struct {
	db *database.DB
}

func NewTransactionHandler(db *database.DB) *TransactionHandler {
	return &TransactionHandler{db: db}
}

type CreateTransactionRequest struct {
	FromAccountID string  `json:"from_account_id" binding:"required"`
	ToAccountID   string  `json:"to_account_id" binding:"required"`
	Amount        float64 `json:"amount" binding:"required,gt=0"`
	Type          string  `json:"type" binding:"required"`
	Description   string  `json:"description"`
}

// CreateTransaction POST /api/transactions
func (h *TransactionHandler) CreateTransaction(c *gin.Context) {
	var req CreateTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad Request",
			"message": "잘못된 요청: " + err.Error(),
		})
		return
	}

	// 송금 계좌 존재 확인
	var fromBalance float64
	err := h.db.QueryRow(`SELECT balance FROM accounts WHERE id = ?`, req.FromAccountID).Scan(&fromBalance)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"message": "송금 계좌를 찾을 수 없습니다",
		})
		return
	}

	// 수취 계좌 존재 확인
	var toBalance float64
	err = h.db.QueryRow(`SELECT balance FROM accounts WHERE id = ?`, req.ToAccountID).Scan(&toBalance)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"message": "수취 계좌를 찾을 수 없습니다",
		})
		return
	}

	// 잔액 확인
	fee := req.Amount * 0.002 // 0.2% 수수료
	totalAmount := req.Amount + fee

	if fromBalance < totalAmount {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Insufficient Balance",
			"message": "잔액 부족",
			"required": totalAmount,
			"balance": fromBalance,
		})
		return
	}

	// 거래 생성
	now := time.Now().Unix()
	transaction := database.Transaction{
		ID:            "TXN-" + uuid.New().String()[:8],
		FromAccountID: req.FromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Fee:           fee,
		Type:          req.Type,
		Status:        "pending",
		Description:   req.Description,
		CreatedAt:     now,
	}

	// 거래 저장
	_, err = h.db.Exec(
		`INSERT INTO transactions (id, from_account_id, to_account_id, amount, fee, type, status, description, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		transaction.ID, transaction.FromAccountID, transaction.ToAccountID,
		transaction.Amount, transaction.Fee, transaction.Type, transaction.Status,
		transaction.Description, transaction.CreatedAt,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "거래 저장 실패: " + err.Error(),
		})
		return
	}

	// 계좌 업데이트 (출금/입금)
	newFromBalance := fromBalance - totalAmount
	newToBalance := toBalance + req.Amount

	h.db.Exec(`UPDATE accounts SET balance = ?, updated_at = ? WHERE id = ?`,
		newFromBalance, now, req.FromAccountID)
	h.db.Exec(`UPDATE accounts SET balance = ?, updated_at = ? WHERE id = ?`,
		newToBalance, now, req.ToAccountID)

	// 거래 완료 처리
	h.db.Exec(`UPDATE transactions SET status = ?, completed_at = ? WHERE id = ?`,
		"completed", now, transaction.ID)

	// 감시 로그 기록
	h.logAudit("TRANSACTION_COMPLETED", req.FromAccountID, "거래 완료: "+transaction.ID, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusCreated, gin.H{
		"id":                transaction.ID,
		"from_account_id":   transaction.FromAccountID,
		"to_account_id":     transaction.ToAccountID,
		"amount":            transaction.Amount,
		"fee":               transaction.Fee,
		"status":            "completed",
		"message":           "거래 완료",
	})
}

// GetTransaction GET /api/transactions/:id
func (h *TransactionHandler) GetTransaction(c *gin.Context) {
	id := c.Param("id")

	var txn database.Transaction
	err := h.db.QueryRow(
		`SELECT id, from_account_id, to_account_id, amount, fee, type, status, description, created_at, completed_at
		 FROM transactions WHERE id = ?`,
		id,
	).Scan(&txn.ID, &txn.FromAccountID, &txn.ToAccountID, &txn.Amount, &txn.Fee,
		&txn.Type, &txn.Status, &txn.Description, &txn.CreatedAt, &txn.CompletedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"message": "거래를 찾을 수 없습니다: " + id,
		})
		return
	}

	c.JSON(http.StatusOK, txn)
}

// GetAccountTransactions GET /api/accounts/:id/transactions
func (h *TransactionHandler) GetAccountTransactions(c *gin.Context) {
	accountID := c.Param("id")

	rows, err := h.db.Query(
		`SELECT id, from_account_id, to_account_id, amount, fee, type, status, description, created_at, completed_at
		 FROM transactions WHERE from_account_id = ? OR to_account_id = ? ORDER BY created_at DESC LIMIT 100`,
		accountID, accountID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "거래 조회 실패: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var transactions []database.Transaction
	for rows.Next() {
		var txn database.Transaction
		if err := rows.Scan(&txn.ID, &txn.FromAccountID, &txn.ToAccountID,
			&txn.Amount, &txn.Fee, &txn.Type, &txn.Status,
			&txn.Description, &txn.CreatedAt, &txn.CompletedAt); err != nil {
			continue
		}
		transactions = append(transactions, txn)
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id":    accountID,
		"transactions":  transactions,
		"count":         len(transactions),
	})
}

type ReverseTransactionRequest struct {
	TransactionID string `json:"transaction_id" binding:"required"`
}

// ReverseTransaction POST /api/transactions/reverse
func (h *TransactionHandler) ReverseTransaction(c *gin.Context) {
	var req ReverseTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad Request",
			"message": "잘못된 요청: " + err.Error(),
		})
		return
	}

	// 원본 거래 조회
	var txn database.Transaction
	var fromBalance, toBalance float64

	err := h.db.QueryRow(
		`SELECT id, from_account_id, to_account_id, amount, fee, type, status, description, created_at, completed_at
		 FROM transactions WHERE id = ?`,
		req.TransactionID,
	).Scan(&txn.ID, &txn.FromAccountID, &txn.ToAccountID, &txn.Amount, &txn.Fee,
		&txn.Type, &txn.Status, &txn.Description, &txn.CreatedAt, &txn.CompletedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"message": "거래를 찾을 수 없습니다",
		})
		return
	}

	// 계좌 잔액 조회
	h.db.QueryRow(`SELECT balance FROM accounts WHERE id = ?`, txn.FromAccountID).Scan(&fromBalance)
	h.db.QueryRow(`SELECT balance FROM accounts WHERE id = ?`, txn.ToAccountID).Scan(&toBalance)

	now := time.Now().Unix()

	// 원본 거래 취소 처리
	h.db.Exec(`UPDATE transactions SET status = ? WHERE id = ?`, "reversed", req.TransactionID)

	// 역거래 생성
	reverseID := "TXN-" + uuid.New().String()[:8]
	h.db.Exec(
		`INSERT INTO transactions (id, from_account_id, to_account_id, amount, fee, type, status, description, created_at, completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		reverseID, txn.ToAccountID, txn.FromAccountID, txn.Amount, txn.Fee,
		"reverse", "completed", "거래 취소: "+req.TransactionID, now, now,
	)

	// 계좌 잔액 복구
	newFromBalance := fromBalance + txn.Amount + txn.Fee
	newToBalance := toBalance - txn.Amount

	h.db.Exec(`UPDATE accounts SET balance = ?, updated_at = ? WHERE id = ?`,
		newFromBalance, now, txn.FromAccountID)
	h.db.Exec(`UPDATE accounts SET balance = ?, updated_at = ? WHERE id = ?`,
		newToBalance, now, txn.ToAccountID)

	// 감시 로그 기록
	h.logAudit("TRANSACTION_REVERSED", txn.FromAccountID, "거래 취소: "+req.TransactionID, c.ClientIP(), c.Request.UserAgent())

	c.JSON(http.StatusOK, gin.H{
		"original_transaction_id": req.TransactionID,
		"reverse_transaction_id":  reverseID,
		"status":                  "reversed",
		"message":                 "거래 취소 완료",
	})
}

// logAudit 감시 로그 기록
func (h *TransactionHandler) logAudit(action, accountID, description, ipAddress, userAgent string) {
	auditLog := database.AuditLog{
		ID:          "AUDIT-" + uuid.New().String()[:8],
		Action:      action,
		AccountID:   accountID,
		Description: description,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Timestamp:   time.Now().Unix(),
	}

	h.db.Exec(
		`INSERT INTO audit_logs (id, action, account_id, description, ip_address, user_agent, timestamp)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		auditLog.ID, auditLog.Action, auditLog.AccountID, auditLog.Description,
		auditLog.IPAddress, auditLog.UserAgent, auditLog.Timestamp,
	)
}
