package database

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type DB struct {
	*sql.DB
}

// InitDB 데이터베이스 초기화
func InitDB(filepath string) (*DB, error) {
	sqlDB, err := sql.Open("sqlite3", filepath)
	if err != nil {
		return nil, fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, fmt.Errorf("데이터베이스 핑 실패: %w", err)
	}

	db := &DB{sqlDB}

	// 테이블 생성
	if err := db.createTables(); err != nil {
		return nil, err
	}

	log.Println("✅ 데이터베이스 초기화 완료")
	return db, nil
}

// createTables 필요한 모든 테이블 생성
func (db *DB) createTables() error {
	queries := []string{
		// 사용자 테이블
		`CREATE TABLE IF NOT EXISTS users (
			id TEXT PRIMARY KEY,
			username TEXT NOT NULL UNIQUE,
			email TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL,
			role TEXT NOT NULL DEFAULT 'user',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,

		// 계좌 테이블
		`CREATE TABLE IF NOT EXISTS accounts (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			type TEXT NOT NULL,
			balance REAL NOT NULL,
			rate REAL NOT NULL DEFAULT 0.0,
			status TEXT NOT NULL DEFAULT 'active',
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)`,

		// 거래 테이블
		`CREATE TABLE IF NOT EXISTS transactions (
			id TEXT PRIMARY KEY,
			from_account_id TEXT,
			to_account_id TEXT,
			amount REAL NOT NULL,
			fee REAL NOT NULL DEFAULT 0.0,
			type TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending',
			description TEXT,
			created_at INTEGER NOT NULL,
			completed_at INTEGER
		)`,

		// 감시 로그 테이블
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id TEXT PRIMARY KEY,
			action TEXT NOT NULL,
			account_id TEXT,
			description TEXT,
			ip_address TEXT,
			user_agent TEXT,
			timestamp INTEGER NOT NULL
		)`,

		// 사기 경고 테이블
		`CREATE TABLE IF NOT EXISTS fraud_alerts (
			id TEXT PRIMARY KEY,
			transaction_id TEXT,
			severity TEXT NOT NULL,
			score INTEGER NOT NULL,
			reason TEXT,
			timestamp INTEGER NOT NULL
		)`,

		// 이자 기록 테이블
		`CREATE TABLE IF NOT EXISTS interest_records (
			id TEXT PRIMARY KEY,
			account_id TEXT NOT NULL,
			amount REAL NOT NULL,
			rate REAL NOT NULL,
			period TEXT NOT NULL,
			timestamp INTEGER NOT NULL
		)`,
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("테이블 생성 실패: %w", err)
		}
	}

	return nil
}

// Account 레코드
type Account struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Type      string  `json:"type"`
	Balance   float64 `json:"balance"`
	Rate      float64 `json:"rate"`
	Status    string  `json:"status"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// Transaction 레코드
type Transaction struct {
	ID               string  `json:"id"`
	FromAccountID    string  `json:"from_account_id"`
	ToAccountID      string  `json:"to_account_id"`
	Amount           float64 `json:"amount"`
	Fee              float64 `json:"fee"`
	Type             string  `json:"type"`
	Status           string  `json:"status"`
	Description      string  `json:"description"`
	CreatedAt        int64  `json:"created_at"`
	CompletedAt      *int64 `json:"completed_at,omitempty"`
}

// AuditLog 레코드
type AuditLog struct {
	ID        string `json:"id"`
	Action    string `json:"action"`
	AccountID string `json:"account_id"`
	Description string `json:"description"`
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Timestamp int64  `json:"timestamp"`
}

// FraudAlert 레코드
type FraudAlert struct {
	ID            string `json:"id"`
	TransactionID string `json:"transaction_id"`
	Severity      string `json:"severity"`
	Score         int    `json:"score"`
	Reason        string `json:"reason"`
	Timestamp     int64  `json:"timestamp"`
}

// InterestRecord 레코드
type InterestRecord struct {
	ID        string  `json:"id"`
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
	Rate      float64 `json:"rate"`
	Period    string  `json:"period"`
	Timestamp int64  `json:"timestamp"`
}
