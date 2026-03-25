package handlers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"net/http"
	"time"
	"freelang-bank-system/server/database"
)

type FraudHandler struct {
	db *database.DB
}

func NewFraudHandler(db *database.DB) *FraudHandler {
	return &FraudHandler{db: db}
}

type CheckFraudRequest struct {
	Amount          float64 `json:"amount" binding:"required,gt=0"`
	Frequency       int     `json:"frequency"` // 시간당 거래 건수
	BalanceDrainPct float64 `json:"balance_drain_pct"` // 잔액 감소율 (%)
}

// CheckFraud POST /api/fraud/check
func (h *FraudHandler) CheckFraud(c *gin.Context) {
	var req CheckFraudRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bad Request",
			"message": "잘못된 요청: " + err.Error(),
		})
		return
	}

	score := 0
	var reasons []string

	// 1️⃣ 거래액 체크
	if req.Amount > 100000 {
		score += 30
		reasons = append(reasons, "Large transaction (>$100K)")
	} else if req.Amount > 50000 {
		score += 20
		reasons = append(reasons, "Large transaction (>$50K)")
	} else if req.Amount > 10000 {
		score += 10
		reasons = append(reasons, "Large transaction (>$10K)")
	}

	// 2️⃣ 거래 빈도 체크
	if req.Frequency > 100 {
		score += 25
		reasons = append(reasons, "Unusual frequency (>100/hour)")
	} else if req.Frequency > 50 {
		score += 15
		reasons = append(reasons, "Unusual frequency (>50/hour)")
	} else if req.Frequency > 20 {
		score += 10
		reasons = append(reasons, "Unusual frequency (>20/hour)")
	}

	// 3️⃣ 잔액 급감 체크
	if req.BalanceDrainPct > 80 {
		score += 25
		reasons = append(reasons, "Balance drain (>80%)")
	} else if req.BalanceDrainPct > 50 {
		score += 15
		reasons = append(reasons, "Balance drain (>50%)")
	} else if req.BalanceDrainPct > 30 {
		score += 10
		reasons = append(reasons, "Balance drain (>30%)")
	}

	// 4️⃣ 시간대 체크 (야간)
	now := time.Now()
	hour := now.Hour()
	if hour >= 0 && hour < 6 {
		score += 10
		reasons = append(reasons, "Unusual time (00:00-06:00)")
	}

	// 심각도 판정
	var severity string
	if score >= 80 {
		severity = "critical"
	} else if score >= 60 {
		severity = "high"
	} else if score >= 40 {
		severity = "medium"
	} else {
		severity = "low"
	}

	// 사기 경고 저장 (score >= 40인 경우만)
	if score >= 40 {
		alertID := "ALERT-" + uuid.New().String()[:8]
		reasonStr := ""
		for _, r := range reasons {
			reasonStr += r + "; "
		}

		h.db.Exec(
			`INSERT INTO fraud_alerts (id, transaction_id, severity, score, reason, timestamp)
			 VALUES (?, ?, ?, ?, ?, ?)`,
			alertID, "TXN-pending", severity, score, reasonStr, time.Now().Unix(),
		)
	}

	c.JSON(http.StatusOK, gin.H{
		"score":      score,
		"severity":   severity,
		"reasons":    reasons,
		"risk_level": map[string]string{
			"critical": "🚨 Critical (차단)",
			"high":     "🔴 High (경고)",
			"medium":   "🟡 Medium (모니터링)",
			"low":      "✅ Low (안전)",
		}[severity],
	})
}

// GetAlerts GET /api/fraud/alerts
func (h *FraudHandler) GetAlerts(c *gin.Context) {
	rows, err := h.db.Query(
		`SELECT id, transaction_id, severity, score, reason, timestamp
		 FROM fraud_alerts ORDER BY timestamp DESC LIMIT 100`,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Internal Server Error",
			"message": "경고 조회 실패: " + err.Error(),
		})
		return
	}
	defer rows.Close()

	var alerts []database.FraudAlert
	for rows.Next() {
		var alert database.FraudAlert
		if err := rows.Scan(&alert.ID, &alert.TransactionID, &alert.Severity,
			&alert.Score, &alert.Reason, &alert.Timestamp); err != nil {
			continue
		}
		alerts = append(alerts, alert)
	}

	c.JSON(http.StatusOK, gin.H{
		"alerts": alerts,
		"count":  len(alerts),
	})
}
