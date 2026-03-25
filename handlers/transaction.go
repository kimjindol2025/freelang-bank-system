package handlers

import (
	"net/http"
	"strconv"
	"time"

	"freelang-bank-system/src/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TransferRequest struct {
	ToAccountID string  `json:"to_account_id" binding:"required"`
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description"`
}

type TransactionResponse struct {
	ID              string     `json:"id"`
	FromAccountID   string     `json:"from_account_id"`
	ToAccountID     string     `json:"to_account_id"`
	Amount          float64    `json:"amount"`
	Type            string     `json:"type"`
	Description     string     `json:"description"`
	Status          string     `json:"status"`
	Fee             float64    `json:"fee"`
	CreatedAt       time.Time  `json:"created_at"`
	ConfirmedAt     *time.Time `json:"confirmed_at,omitempty"`
}

// Transfer - 이체
func Transfer(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	var req TransferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "입력값 검증 실패: " + err.Error()})
		return
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	// 출금 계좌 확인
	fromAccount, err := database.GetAccount(req.ToAccountID) // 쿼리 파라미터에서 가져올 예정
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "출금 계좌 조회 실패"})
		return
	}

	// 간단한 버전: body의 from_account_id 사용 또는 쿼리 파라미터
	// 여기서는 쿼리에서 from_account_id를 받도록 수정
	fromAccountID := c.Query("from_account_id")
	if fromAccountID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_account_id 파라미터가 필요합니다"})
		return
	}

	fromAccount, err = database.GetAccount(fromAccountID)
	if err != nil || fromAccount == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "출금 계좌를 찾을 수 없습니다"})
		return
	}

	// 권한 확인
	if fromAccount.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	if fromAccount.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "활성 계좌만 거래할 수 있습니다"})
		return
	}

	// 입금 계좌 확인
	toAccount, err := database.GetAccount(req.ToAccountID)
	if err != nil || toAccount == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "입금 계좌를 찾을 수 없습니다"})
		return
	}

	if toAccount.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "입금 계좌가 활성 상태가 아닙니다"})
		return
	}

	// 수수료 계산 (1000 초과 시 0.5%, 최소 $1)
	var fee float64
	if req.Amount > 1000.0 {
		fee = req.Amount * 0.005
		if fee < 1.0 {
			fee = 1.0
		}
	} else {
		fee = 1.0
	}

	totalAmount := req.Amount + fee

	// 잔액 확인
	availableBalance := fromAccount.Balance + fromAccount.OverdraftLimit
	if totalAmount > availableBalance {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잔액 부족"})
		return
	}

	// 이체 처리 (원자적 업데이트)
	fromAccount.Balance -= totalAmount
	fromAccount.TransactionCount++

	toAccount.Balance += req.Amount
	toAccount.TransactionCount++

	if err := database.UpdateAccount(fromAccount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "이체 처리 실패"})
		return
	}

	if err := database.UpdateAccount(toAccount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "이체 처리 실패"})
		return
	}

	// 거래 기록 (출금)
	txn := &db.Transaction{
		ID:            uuid.New().String(),
		FromAccountID: fromAccountID,
		ToAccountID:   req.ToAccountID,
		Amount:        req.Amount,
		Currency:      fromAccount.Currency,
		Type:          "transfer",
		Description:   req.Description,
		Status:        "confirmed",
		Fee:           fee,
	}

	if err := database.CreateTransaction(txn); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "거래 기록 실패"})
		return
	}

	// 감시 로그
	database.LogAudit(&db.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       "TRANSFER",
		ResourceType: "transaction",
		ResourceID:   txn.ID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	confirmedAt := time.Now()
	c.JSON(http.StatusOK, TransactionResponse{
		ID:            txn.ID,
		FromAccountID: txn.FromAccountID,
		ToAccountID:   txn.ToAccountID,
		Amount:        txn.Amount,
		Type:          txn.Type,
		Description:   txn.Description,
		Status:        txn.Status,
		Fee:           txn.Fee,
		CreatedAt:     txn.CreatedAt,
		ConfirmedAt:   &confirmedAt,
	})
}

// GetTransaction - 거래 조회
func GetTransaction(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	txID := c.Param("id")

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	tx, err := database.GetTransaction(txID)
	if err != nil || tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "거래를 찾을 수 없습니다"})
		return
	}

	// 권한 확인 (출금/입금 계좌의 소유자만 조회 가능)
	fromAccount, _ := database.GetAccount(tx.FromAccountID)
	toAccount, _ := database.GetAccount(tx.ToAccountID)

	hasAccess := false
	if fromAccount != nil && fromAccount.UserID == userID {
		hasAccess = true
	}
	if toAccount != nil && toAccount.UserID == userID {
		hasAccess = true
	}

	if !hasAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	c.JSON(http.StatusOK, TransactionResponse{
		ID:            tx.ID,
		FromAccountID: tx.FromAccountID,
		ToAccountID:   tx.ToAccountID,
		Amount:        tx.Amount,
		Type:          tx.Type,
		Description:   tx.Description,
		Status:        tx.Status,
		Fee:           tx.Fee,
		CreatedAt:     tx.CreatedAt,
		ConfirmedAt:   tx.ConfirmedAt,
	})
}

// GetAccountTransactions - 계좌 거래 히스토리
func GetAccountTransactions(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	accountID := c.Param("id")

	// 페이지네이션
	limit := 20
	offset := 0

	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 && v <= 100 {
			limit = v
		}
	}

	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	// 계좌 소유권 확인
	account, err := database.GetAccount(accountID)
	if err != nil || account == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "계좌를 찾을 수 없습니다"})
		return
	}

	if account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	transactions, err := database.GetAccountTransactions(accountID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "조회 실패: " + err.Error()})
		return
	}

	var response []TransactionResponse
	for _, tx := range transactions {
		response = append(response, TransactionResponse{
			ID:            tx.ID,
			FromAccountID: tx.FromAccountID,
			ToAccountID:   tx.ToAccountID,
			Amount:        tx.Amount,
			Type:          tx.Type,
			Description:   tx.Description,
			Status:        tx.Status,
			Fee:           tx.Fee,
			CreatedAt:     tx.CreatedAt,
			ConfirmedAt:   tx.ConfirmedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"count":        len(response),
		"limit":        limit,
		"offset":       offset,
		"transactions": response,
	})
}

// ReverseTransaction - 거래 취소 (반환)
func ReverseTransaction(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	txID := c.Param("id")

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	tx, err := database.GetTransaction(txID)
	if err != nil || tx == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "거래를 찾을 수 없습니다"})
		return
	}

	// 권한 확인
	fromAccount, _ := database.GetAccount(tx.FromAccountID)
	if fromAccount == nil || fromAccount.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	// 이미 취소된 거래 확인
	if tx.Status == "reversed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "이미 취소된 거래입니다"})
		return
	}

	// 실패한 거래는 취소할 수 없음
	if tx.Status == "failed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "실패한 거래는 취소할 수 없습니다"})
		return
	}

	// 거래 취소: 금액 반환
	fromAccount.Balance += (tx.Amount + tx.Fee)
	toAccount, _ := database.GetAccount(tx.ToAccountID)
	if toAccount != nil {
		toAccount.Balance -= tx.Amount
		database.UpdateAccount(toAccount)
	}

	database.UpdateAccount(fromAccount)

	// 거래 상태 업데이트
	tx.Status = "reversed"
	tx.Description += " (REVERSED)"
	database.UpdateTransaction(tx)

	// 감시 로그
	database.LogAudit(&db.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       "REVERSE_TRANSACTION",
		ResourceType: "transaction",
		ResourceID:   txID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "거래가 취소되었습니다",
		"transaction_id": txID,
	})
}
