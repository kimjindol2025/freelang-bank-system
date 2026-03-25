package handlers

import (
	"net/http"

	"freelang-bank-system/src/db"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateAccountRequest struct {
	AccountType string `json:"account_type" binding:"required,oneof=checking savings credit"`
	Currency    string `json:"currency"`
}

type AccountResponse struct {
	ID                string  `json:"id"`
	AccountType       string  `json:"account_type"`
	Currency          string  `json:"currency"`
	Balance           float64 `json:"balance"`
	Status            string  `json:"status"`
	AnnualRate        float64 `json:"annual_rate"`
	TransactionCount  int64   `json:"transaction_count"`
	InterestAccrued   float64 `json:"interest_accrued"`
}

// CreateAccount - 계좌 생성
func CreateAccount(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	var req CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "입력값 검증 실패: " + err.Error()})
		return
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	// 기본값 설정
	if req.Currency == "" {
		req.Currency = "USD"
	}

	// 이자율 및 당좌차월한 한도 설정
	var annualRate, overdraftLimit float64
	switch req.AccountType {
	case "checking":
		annualRate = 0.0
		overdraftLimit = 500.0
	case "savings":
		annualRate = 2.0
		overdraftLimit = 0.0
	case "credit":
		annualRate = 0.0
		overdraftLimit = 1000.0
	}

	account := &db.Account{
		ID:              uuid.New().String(),
		UserID:          userID,
		AccountType:     req.AccountType,
		Currency:        req.Currency,
		Balance:         0.0,
		Status:          "active",
		KYCStatus:       "pending",
		AnnualRate:      annualRate,
		OverdraftLimit:  overdraftLimit,
		TransactionCount: 0,
		InterestAccrued: 0.0,
	}

	if err := database.CreateAccount(account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "계좌 생성 실패: " + err.Error()})
		return
	}

	// 감시 로그
	database.LogAudit(&db.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       "CREATE_ACCOUNT",
		ResourceType: "account",
		ResourceID:   account.ID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusCreated, AccountResponse{
		ID:              account.ID,
		AccountType:     account.AccountType,
		Currency:        account.Currency,
		Balance:         account.Balance,
		Status:          account.Status,
		AnnualRate:      account.AnnualRate,
		TransactionCount: account.TransactionCount,
		InterestAccrued: account.InterestAccrued,
	})
}

// GetAccount - 계좌 조회
func GetAccount(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	accountID := c.Param("id")

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	account, err := database.GetAccount(accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "조회 실패: " + err.Error()})
		return
	}

	if account == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "계좌를 찾을 수 없습니다"})
		return
	}

	// 권한 확인
	if account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	c.JSON(http.StatusOK, AccountResponse{
		ID:              account.ID,
		AccountType:     account.AccountType,
		Currency:        account.Currency,
		Balance:         account.Balance,
		Status:          account.Status,
		AnnualRate:      account.AnnualRate,
		TransactionCount: account.TransactionCount,
		InterestAccrued: account.InterestAccrued,
	})
}

// ListAccounts - 내 모든 계좌 조회
func ListAccounts(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	accounts, err := database.GetUserAccounts(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "조회 실패: " + err.Error()})
		return
	}

	var response []AccountResponse
	for _, acc := range accounts {
		response = append(response, AccountResponse{
			ID:              acc.ID,
			AccountType:     acc.AccountType,
			Currency:        acc.Currency,
			Balance:         acc.Balance,
			Status:          acc.Status,
			AnnualRate:      acc.AnnualRate,
			TransactionCount: acc.TransactionCount,
			InterestAccrued: acc.InterestAccrued,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"count":    len(response),
		"accounts": response,
	})
}

// CloseAccount - 계좌 종료
func CloseAccount(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	accountID := c.Param("id")

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	account, err := database.GetAccount(accountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "조회 실패: " + err.Error()})
		return
	}

	if account == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "계좌를 찾을 수 없습니다"})
		return
	}

	// 권한 확인
	if account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	// 잔액 확인 (0이어야 함)
	if account.Balance != 0.0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잔액이 0이 아니면 계좌를 종료할 수 없습니다"})
		return
	}

	if err := database.CloseAccount(accountID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "계좌 종료 실패: " + err.Error()})
		return
	}

	// 감시 로그
	database.LogAudit(&db.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       "CLOSE_ACCOUNT",
		ResourceType: "account",
		ResourceID:   accountID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{"message": "계좌가 종료되었습니다"})
}

// DepositToAccount - 입금 (지금은 기본 데이터 베이스 업데이트만)
func DepositToAccount(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	accountID := c.Param("id")

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "입력값 검증 실패: " + err.Error()})
		return
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	account, err := database.GetAccount(accountID)
	if err != nil || account == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "계좌를 찾을 수 없습니다"})
		return
	}

	if account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	if account.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "활성 계좌만 거래할 수 있습니다"})
		return
	}

	// 입금 처리
	account.Balance += req.Amount
	account.TransactionCount++

	if err := database.UpdateAccount(account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "입금 실패: " + err.Error()})
		return
	}

	// 거래 기록
	tx := &db.Transaction{
		ID:              uuid.New().String(),
		FromAccountID:   accountID,
		ToAccountID:     accountID,
		Amount:          req.Amount,
		Currency:        account.Currency,
		Type:            "deposit",
		Description:     "입금",
		Status:          "confirmed",
		Fee:             0.0,
	}
	database.CreateTransaction(tx)

	// 감시 로그
	database.LogAudit(&db.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       "DEPOSIT",
		ResourceType: "transaction",
		ResourceID:   tx.ID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": tx.ID,
		"balance":        account.Balance,
		"message":        "입금이 완료되었습니다",
	})
}

// WithdrawFromAccount - 출금
func WithdrawFromAccount(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	accountID := c.Param("id")

	var req struct {
		Amount float64 `json:"amount" binding:"required,gt=0"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "입력값 검증 실패: " + err.Error()})
		return
	}

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	account, err := database.GetAccount(accountID)
	if err != nil || account == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "계좌를 찾을 수 없습니다"})
		return
	}

	if account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	if account.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "활성 계좌만 거래할 수 있습니다"})
		return
	}

	// 잔액 확인
	availableBalance := account.Balance + account.OverdraftLimit
	if req.Amount > availableBalance {
		c.JSON(http.StatusBadRequest, gin.H{"error": "잔액 부족"})
		return
	}

	// 출금 처리
	account.Balance -= req.Amount
	account.TransactionCount++

	if err := database.UpdateAccount(account); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "출금 실패: " + err.Error()})
		return
	}

	// 거래 기록
	tx := &db.Transaction{
		ID:              uuid.New().String(),
		FromAccountID:   accountID,
		ToAccountID:     accountID,
		Amount:          req.Amount,
		Currency:        account.Currency,
		Type:            "withdrawal",
		Description:     "출금",
		Status:          "confirmed",
		Fee:             0.0,
	}
	database.CreateTransaction(tx)

	// 감시 로그
	database.LogAudit(&db.AuditLog{
		ID:           uuid.New().String(),
		UserID:       userID,
		Action:       "WITHDRAW",
		ResourceType: "transaction",
		ResourceID:   tx.ID,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.Request.UserAgent(),
	})

	c.JSON(http.StatusOK, gin.H{
		"transaction_id": tx.ID,
		"balance":        account.Balance,
		"message":        "출금이 완료되었습니다",
	})
}

// GetBalance - 잔액 조회
func GetBalance(c *gin.Context) {
	userID, err := GetUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "인증 필요"})
		return
	}

	accountID := c.Param("id")

	database := db.GetDB()
	if database == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "데이터베이스 연결 실패"})
		return
	}

	account, err := database.GetAccount(accountID)
	if err != nil || account == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "계좌를 찾을 수 없습니다"})
		return
	}

	if account.UserID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "접근 권한이 없습니다"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account_id": account.ID,
		"balance":    account.Balance,
		"currency":   account.Currency,
	})
}
