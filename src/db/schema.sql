-- 🏦 FreeLang Bank System - SQLite Schema
-- Phase 3: Database Integration

-- 1️⃣ Users (사용자)
CREATE TABLE IF NOT EXISTS users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  phone TEXT,
  password_hash TEXT NOT NULL,
  first_name TEXT,
  last_name TEXT,
  date_of_birth TEXT,
  kyc_document TEXT,
  kyc_verified_at TEXT,
  status TEXT DEFAULT 'active' CHECK(status IN ('active', 'suspended', 'deleted')),
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- 2️⃣ Accounts (계좌)
CREATE TABLE IF NOT EXISTS accounts (
  id TEXT PRIMARY KEY,
  user_id TEXT NOT NULL,
  account_type TEXT NOT NULL CHECK(account_type IN ('checking', 'savings', 'credit')),
  currency TEXT DEFAULT 'USD',
  balance REAL NOT NULL DEFAULT 0.0,
  status TEXT DEFAULT 'active' CHECK(status IN ('active', 'suspended', 'closed')),
  kyc_status TEXT DEFAULT 'pending' CHECK(kyc_status IN ('pending', 'verified', 'rejected')),
  annual_rate REAL DEFAULT 0.0,
  overdraft_limit REAL DEFAULT 0.0,
  transaction_count INTEGER DEFAULT 0,
  interest_accrued REAL DEFAULT 0.0,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);
CREATE INDEX IF NOT EXISTS idx_accounts_status ON accounts(status);
CREATE INDEX IF NOT EXISTS idx_accounts_kyc_status ON accounts(kyc_status);

-- 3️⃣ Transactions (거래)
CREATE TABLE IF NOT EXISTS transactions (
  id TEXT PRIMARY KEY,
  from_account_id TEXT NOT NULL,
  to_account_id TEXT NOT NULL,
  amount REAL NOT NULL,
  currency TEXT DEFAULT 'USD',
  type TEXT NOT NULL CHECK(type IN ('transfer', 'deposit', 'withdrawal', 'fee')),
  description TEXT,
  status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'confirmed', 'failed', 'reversed')),
  fee REAL DEFAULT 0.0,
  blockchain_hash TEXT,
  blockchain_block_number INTEGER,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP,
  confirmed_at TEXT,
  FOREIGN KEY (from_account_id) REFERENCES accounts(id),
  FOREIGN KEY (to_account_id) REFERENCES accounts(id)
);

CREATE INDEX IF NOT EXISTS idx_transactions_from_account ON transactions(from_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_to_account ON transactions(to_account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);
CREATE INDEX IF NOT EXISTS idx_transactions_blockchain_hash ON transactions(blockchain_hash);
CREATE INDEX IF NOT EXISTS idx_transactions_created_at ON transactions(created_at);

-- 4️⃣ Blockchain Blocks (블록체인)
CREATE TABLE IF NOT EXISTS blockchain_blocks (
  block_number INTEGER PRIMARY KEY,
  timestamp TEXT NOT NULL,
  previous_hash TEXT,
  transactions_count INTEGER DEFAULT 0,
  merkle_root TEXT,
  nonce INTEGER,
  miner TEXT,
  difficulty INTEGER,
  total_fees REAL DEFAULT 0.0,
  block_hash TEXT UNIQUE,
  created_at TEXT DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_blocks_timestamp ON blockchain_blocks(timestamp);
CREATE INDEX IF NOT EXISTS idx_blocks_block_hash ON blockchain_blocks(block_hash);

-- 5️⃣ Audit Logs (감시 로그)
CREATE TABLE IF NOT EXISTS audit_logs (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  action TEXT NOT NULL,
  resource_type TEXT,
  resource_id TEXT,
  details TEXT,
  ip_address TEXT,
  user_agent TEXT,
  timestamp TEXT DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX IF NOT EXISTS idx_audit_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp);
