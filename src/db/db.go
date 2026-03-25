package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	conn *sql.DB
}

var dbInstance *DB

// InitDB - 데이터베이스 초기화
func InitDB(dbPath string) error {
	if dbPath == "" {
		dbPath = "bank.db"
	}

	// 데이터베이스 연결
	conn, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("데이터베이스 열기 실패: %w", err)
	}

	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	if err := conn.Ping(); err != nil {
		return fmt.Errorf("데이터베이스 연결 실패: %w", err)
	}

	dbInstance = &DB{conn: conn}

	// 스키마 생성
	if err := createSchema(); err != nil {
		return fmt.Errorf("스키마 생성 실패: %w", err)
	}

	fmt.Println("✅ 데이터베이스 초기화 완료:", dbPath)
	return nil
}

// GetDB - 데이터베이스 인스턴스 반환
func GetDB() *DB {
	return dbInstance
}

// Close - 데이터베이스 연결 종료
func (db *DB) Close() error {
	return db.conn.Close()
}

// createSchema - 테이블 생성
func createSchema() error {
	schema, err := os.ReadFile("src/db/schema.sql")
	if err != nil {
		return fmt.Errorf("스키마 파일 읽기 실패: %w", err)
	}

	_, err = dbInstance.conn.Exec(string(schema))
	if err != nil {
		return fmt.Errorf("스키마 실행 실패: %w", err)
	}

	return nil
}

// ========================================
// User CRUD
// ========================================

type User struct {
	ID            string    `json:"id"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone,omitempty"`
	PasswordHash  string    `json:"-"`
	FirstName     string    `json:"first_name,omitempty"`
	LastName      string    `json:"last_name,omitempty"`
	DateOfBirth   string    `json:"date_of_birth,omitempty"`
	KYCStatus     string    `json:"kyc_status"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (db *DB) CreateUser(user *User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := db.conn.Exec(`
		INSERT INTO users (id, email, phone, password_hash, first_name, last_name, date_of_birth, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.Email, user.Phone, user.PasswordHash, user.FirstName, user.LastName, user.DateOfBirth, user.Status, user.CreatedAt, user.UpdatedAt)

	return err
}

func (db *DB) GetUserByEmail(email string) (*User, error) {
	user := &User{}
	err := db.conn.QueryRow(`
		SELECT id, email, phone, password_hash, first_name, last_name, date_of_birth, status, created_at, updated_at
		FROM users WHERE email = ?
	`, email).Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.FirstName, &user.LastName, &user.DateOfBirth, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

func (db *DB) GetUserByID(id string) (*User, error) {
	user := &User{}
	err := db.conn.QueryRow(`
		SELECT id, email, phone, password_hash, first_name, last_name, date_of_birth, status, created_at, updated_at
		FROM users WHERE id = ?
	`, id).Scan(&user.ID, &user.Email, &user.Phone, &user.PasswordHash, &user.FirstName, &user.LastName, &user.DateOfBirth, &user.Status, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return user, err
}

// ========================================
// Account CRUD
// ========================================

type Account struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id"`
	AccountType       string    `json:"account_type"`
	Currency          string    `json:"currency"`
	Balance           float64   `json:"balance"`
	Status            string    `json:"status"`
	KYCStatus         string    `json:"kyc_status"`
	AnnualRate        float64   `json:"annual_rate"`
	OverdraftLimit    float64   `json:"overdraft_limit"`
	TransactionCount  int64     `json:"transaction_count"`
	InterestAccrued   float64   `json:"interest_accrued"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func (db *DB) CreateAccount(account *Account) error {
	account.CreatedAt = time.Now()
	account.UpdatedAt = time.Now()

	_, err := db.conn.Exec(`
		INSERT INTO accounts (id, user_id, account_type, currency, balance, status, kyc_status, annual_rate, overdraft_limit, transaction_count, interest_accrued, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, account.ID, account.UserID, account.AccountType, account.Currency, account.Balance, account.Status, account.KYCStatus, account.AnnualRate, account.OverdraftLimit, account.TransactionCount, account.InterestAccrued, account.CreatedAt, account.UpdatedAt)

	return err
}

func (db *DB) GetAccount(id string) (*Account, error) {
	account := &Account{}
	err := db.conn.QueryRow(`
		SELECT id, user_id, account_type, currency, balance, status, kyc_status, annual_rate, overdraft_limit, transaction_count, interest_accrued, created_at, updated_at
		FROM accounts WHERE id = ?
	`, id).Scan(&account.ID, &account.UserID, &account.AccountType, &account.Currency, &account.Balance, &account.Status, &account.KYCStatus, &account.AnnualRate, &account.OverdraftLimit, &account.TransactionCount, &account.InterestAccrued, &account.CreatedAt, &account.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return account, err
}

func (db *DB) GetUserAccounts(userID string) ([]*Account, error) {
	rows, err := db.conn.Query(`
		SELECT id, user_id, account_type, currency, balance, status, kyc_status, annual_rate, overdraft_limit, transaction_count, interest_accrued, created_at, updated_at
		FROM accounts WHERE user_id = ? ORDER BY created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var accounts []*Account
	for rows.Next() {
		account := &Account{}
		err := rows.Scan(&account.ID, &account.UserID, &account.AccountType, &account.Currency, &account.Balance, &account.Status, &account.KYCStatus, &account.AnnualRate, &account.OverdraftLimit, &account.TransactionCount, &account.InterestAccrued, &account.CreatedAt, &account.UpdatedAt)
		if err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}

	return accounts, rows.Err()
}

func (db *DB) UpdateAccount(account *Account) error {
	account.UpdatedAt = time.Now()

	_, err := db.conn.Exec(`
		UPDATE accounts SET balance = ?, status = ?, kyc_status = ?, transaction_count = ?, interest_accrued = ?, updated_at = ?
		WHERE id = ?
	`, account.Balance, account.Status, account.KYCStatus, account.TransactionCount, account.InterestAccrued, account.UpdatedAt, account.ID)

	return err
}

func (db *DB) CloseAccount(id string) error {
	_, err := db.conn.Exec(`UPDATE accounts SET status = 'closed', updated_at = ? WHERE id = ?`, time.Now(), id)
	return err
}

// ========================================
// Transaction CRUD
// ========================================

type Transaction struct {
	ID                  string    `json:"id"`
	FromAccountID       string    `json:"from_account_id"`
	ToAccountID         string    `json:"to_account_id"`
	Amount              float64   `json:"amount"`
	Currency            string    `json:"currency"`
	Type                string    `json:"type"`
	Description         string    `json:"description"`
	Status              string    `json:"status"`
	Fee                 float64   `json:"fee"`
	BlockchainHash      string    `json:"blockchain_hash,omitempty"`
	BlockchainBlockNum  int64     `json:"blockchain_block_number,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	ConfirmedAt         *time.Time `json:"confirmed_at,omitempty"`
}

func (db *DB) CreateTransaction(tx *Transaction) error {
	tx.CreatedAt = time.Now()

	_, err := db.conn.Exec(`
		INSERT INTO transactions (id, from_account_id, to_account_id, amount, currency, type, description, status, fee, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tx.ID, tx.FromAccountID, tx.ToAccountID, tx.Amount, tx.Currency, tx.Type, tx.Description, tx.Status, tx.Fee, tx.CreatedAt)

	return err
}

func (db *DB) GetTransaction(id string) (*Transaction, error) {
	tx := &Transaction{}
	err := db.conn.QueryRow(`
		SELECT id, from_account_id, to_account_id, amount, currency, type, description, status, fee, blockchain_hash, blockchain_block_number, created_at, confirmed_at
		FROM transactions WHERE id = ?
	`, id).Scan(&tx.ID, &tx.FromAccountID, &tx.ToAccountID, &tx.Amount, &tx.Currency, &tx.Type, &tx.Description, &tx.Status, &tx.Fee, &tx.BlockchainHash, &tx.BlockchainBlockNum, &tx.CreatedAt, &tx.ConfirmedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	return tx, err
}

func (db *DB) GetAccountTransactions(accountID string, limit, offset int) ([]*Transaction, error) {
	rows, err := db.conn.Query(`
		SELECT id, from_account_id, to_account_id, amount, currency, type, description, status, fee, blockchain_hash, blockchain_block_number, created_at, confirmed_at
		FROM transactions
		WHERE from_account_id = ? OR to_account_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, accountID, accountID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		tx := &Transaction{}
		err := rows.Scan(&tx.ID, &tx.FromAccountID, &tx.ToAccountID, &tx.Amount, &tx.Currency, &tx.Type, &tx.Description, &tx.Status, &tx.Fee, &tx.BlockchainHash, &tx.BlockchainBlockNum, &tx.CreatedAt, &tx.ConfirmedAt)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}

	return transactions, rows.Err()
}

func (db *DB) UpdateTransaction(tx *Transaction) error {
	now := time.Now()
	_, err := db.conn.Exec(`
		UPDATE transactions SET status = ?, fee = ?, blockchain_hash = ?, blockchain_block_number = ?, confirmed_at = ?
		WHERE id = ?
	`, tx.Status, tx.Fee, tx.BlockchainHash, tx.BlockchainBlockNum, now, tx.ID)

	return err
}

// ========================================
// Audit Log
// ========================================

type AuditLog struct {
	ID           string                 `json:"id"`
	UserID       string                 `json:"user_id,omitempty"`
	Action       string                 `json:"action"`
	ResourceType string                 `json:"resource_type,omitempty"`
	ResourceID   string                 `json:"resource_id,omitempty"`
	Details      map[string]interface{} `json:"details,omitempty"`
	IPAddress    string                 `json:"ip_address,omitempty"`
	UserAgent    string                 `json:"user_agent,omitempty"`
	Timestamp    time.Time              `json:"timestamp"`
}

func (db *DB) LogAudit(log *AuditLog) error {
	log.Timestamp = time.Now()

	detailsJSON, _ := json.Marshal(log.Details)

	_, err := db.conn.Exec(`
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, timestamp)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, log.ID, log.UserID, log.Action, log.ResourceType, log.ResourceID, string(detailsJSON), log.IPAddress, log.UserAgent, log.Timestamp)

	return err
}

func (db *DB) GetAuditLogs(limit, offset int) ([]*AuditLog, error) {
	rows, err := db.conn.Query(`
		SELECT id, user_id, action, resource_type, resource_id, details, ip_address, user_agent, timestamp
		FROM audit_logs
		ORDER BY timestamp DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*AuditLog
	for rows.Next() {
		log := &AuditLog{}
		var detailsJSON string
		err := rows.Scan(&log.ID, &log.UserID, &log.Action, &log.ResourceType, &log.ResourceID, &detailsJSON, &log.IPAddress, &log.UserAgent, &log.Timestamp)
		if err != nil {
			return nil, err
		}
		json.Unmarshal([]byte(detailsJSON), &log.Details)
		logs = append(logs, log)
	}

	return logs, rows.Err()
}
