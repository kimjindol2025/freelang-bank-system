package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
	"freelang-bank-system/server/database"
	"math"
)

type ReportHandler struct {
	db *database.DB
}

func NewReportHandler(db *database.DB) *ReportHandler {
	return &ReportHandler{db: db}
}

// GetInterest GET /api/interest/:account_id
func (h *ReportHandler) GetInterest(c *gin.Context) {
	accountID := c.Param("account_id")

	// 계좌 조회
	var account database.Account
	err := h.db.QueryRow(
		`SELECT id, name, type, balance, rate, status, created_at, updated_at FROM accounts WHERE id = ?`,
		accountID,
	).Scan(&account.ID, &account.Name, &account.Type, &account.Balance, &account.Rate,
		&account.Status, &account.CreatedAt, &account.UpdatedAt)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Not Found",
			"message": "계좌를 찾을 수 없습니다: " + accountID,
		})
		return
	}

	// 이자 계산
	dailyInterest := account.Balance * (account.Rate / 100) / 365
	monthlyInterest := account.Balance * (account.Rate / 100) / 12
	annualInterest := account.Balance * (account.Rate / 100)

	// 세금 처리 (24% 연방세)
	taxRate := 0.24
	annualInterestAfterTax := annualInterest * (1 - taxRate)

	c.JSON(http.StatusOK, gin.H{
		"account_id":              accountID,
		"balance":                 account.Balance,
		"rate":                    account.Rate,
		"daily_interest":          math.Round(dailyInterest*100) / 100,
		"monthly_interest":        math.Round(monthlyInterest*100) / 100,
		"annual_interest":         math.Round(annualInterest*100) / 100,
		"annual_interest_after_tax": math.Round(annualInterestAfterTax*100) / 100,
		"tax_rate":                taxRate * 100,
	})
}

// GetDailyReport GET /api/reports/daily/:date
func (h *ReportHandler) GetDailyReport(c *gin.Context) {
	date := c.Param("date") // YYYY-MM-DD 형식

	// 일일 거래 통계
	var totalTxns int
	var totalVolume, totalFees float64

	h.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(amount), 0), COALESCE(SUM(fee), 0)
		 FROM transactions WHERE date(datetime(created_at, 'unixepoch')) = ?`,
		date,
	).Scan(&totalTxns, &totalVolume, &totalFees)

	// 사기 경고 건수
	var fraudAlerts int
	h.db.QueryRow(
		`SELECT COUNT(*) FROM fraud_alerts WHERE date(datetime(timestamp, 'unixepoch')) = ?`,
		date,
	).Scan(&fraudAlerts)

	c.JSON(http.StatusOK, gin.H{
		"date":                date,
		"total_transactions":  totalTxns,
		"total_volume":        math.Round(totalVolume*100) / 100,
		"total_fees":          math.Round(totalFees*100) / 100,
		"fraud_alerts":        fraudAlerts,
	})
}

// GetMonthlyReport GET /api/reports/monthly/:year_month
func (h *ReportHandler) GetMonthlyReport(c *gin.Context) {
	yearMonth := c.Param("year_month") // YYYY-MM 형식

	// 월간 거래 통계
	var totalTxns int
	var totalVolume, totalFees, totalInterest float64

	h.db.QueryRow(
		`SELECT COUNT(*), COALESCE(SUM(amount), 0), COALESCE(SUM(fee), 0)
		 FROM transactions WHERE substr(datetime(created_at, 'unixepoch'), 1, 7) = ?`,
		yearMonth,
	).Scan(&totalTxns, &totalVolume, &totalFees)

	// 월간 이자 통계
	h.db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM interest_records WHERE substr(datetime(timestamp, 'unixepoch'), 1, 7) = ?`,
		yearMonth,
	).Scan(&totalInterest)

	// 평균 거래액
	var avgVolume float64
	if totalTxns > 0 {
		avgVolume = totalVolume / float64(totalTxns)
	}

	c.JSON(http.StatusOK, gin.H{
		"month":                 yearMonth,
		"total_transactions":    totalTxns,
		"total_volume":          math.Round(totalVolume*100) / 100,
		"average_transaction":   math.Round(avgVolume*100) / 100,
		"total_fees":            math.Round(totalFees*100) / 100,
		"total_interest":        math.Round(totalInterest*100) / 100,
	})
}

// ApplyDailyInterest 일일 이자 적용 (관리자 전용)
func (h *ReportHandler) ApplyDailyInterest(c *gin.Context) {
	// 모든 활성 계좌 조회
	rows, err := h.db.Query(
		`SELECT id, balance, rate FROM accounts WHERE status = 'active'`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "계좌 조회 실패: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	now := time.Now().Unix()
	totalInterest := 0.0
	count := 0

	for rows.Next() {
		var accountID string
		var balance, rate float64

		if err := rows.Scan(&accountID, &balance, &rate); err != nil {
			continue
		}

		// 일일 이자 계산
		dailyInterest := balance * (rate / 100) / 365

		if dailyInterest > 0 {
			// 계좌 잔액 업데이트
			newBalance := balance + dailyInterest
			h.db.Exec(
				`UPDATE accounts SET balance = ?, updated_at = ? WHERE id = ?`,
				newBalance, now, accountID,
			)

			// 이자 기록 저장
			interestID := "INT-" + uuid.New().String()[:8]
			h.db.Exec(
				`INSERT INTO interest_records (id, account_id, amount, rate, period, timestamp)
				 VALUES (?, ?, ?, ?, ?, ?)`,
				interestID, accountID, dailyInterest, rate, "daily", now,
			)

			totalInterest += dailyInterest
			count++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":          "일일 이자 적용 완료",
		"accounts_updated": count,
		"total_interest":   math.Round(totalInterest*100) / 100,
	})
}

// GetStatistics GET /api/statistics
func (h *ReportHandler) GetStatistics(c *gin.Context) {
	// 전체 통계
	var totalAccounts int
	var totalTransactions int
	var totalBalance float64

	h.db.QueryRow(`SELECT COUNT(*) FROM accounts`).Scan(&totalAccounts)
	h.db.QueryRow(`SELECT COUNT(*) FROM transactions`).Scan(&totalTransactions)
	h.db.QueryRow(`SELECT COALESCE(SUM(balance), 0) FROM accounts`).Scan(&totalBalance)

	c.JSON(http.StatusOK, gin.H{
		"total_accounts":     totalAccounts,
		"total_transactions": totalTransactions,
		"total_balance":      math.Round(totalBalance*100) / 100,
	})
}
