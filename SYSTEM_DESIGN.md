# 🏦 FreeLang Bank System - 시스템 설계

**상태**: 🟡 설계 진행 중 | **목표**: 6월 말 프로토타입 완성

---

## 📋 개요

FreeLang으로 구현된 **제로 의존성 블록체인 기반 은행 시스템**입니다.

### 핵심 목표
- ✅ 자가호스팅 가능한 은행 시스템
- ✅ 블록체인 기반 거래 검증
- ✅ 규제 준수 (AML, KYC)
- ✅ 고성능 (1,000+ TPS)
- ✅ 확장성 (샤딩, 스케일링)

---

## 🏗️ 아키텍처

### 3계층 구조

```
┌─────────────────────────────────┐
│   API & Frontend Layer          │  (REST API, Web UI)
├─────────────────────────────────┤
│   Business Logic Layer          │  (Account, Transaction, Validation)
├─────────────────────────────────┤
│   Blockchain & Storage Layer    │  (Ledger, Chain, Database)
└─────────────────────────────────┘
```

### 시스템 다이어그램

```
User Interface
    ↓
    ├─ Web UI (React)
    ├─ Mobile App (Flutter)
    └─ CLI (FreeLang)
    ↓
API Gateway (Express/Go)
    ├─ /auth/* (인증)
    ├─ /accounts/* (계좌)
    ├─ /transactions/* (거래)
    ├─ /blockchain/* (블록체인)
    └─ /admin/* (관리)
    ↓
Business Logic (FreeLang/Go)
    ├─ Account Manager
    ├─ Transaction Processor
    ├─ Validation Engine
    ├─ Security Manager
    └─ Blockchain Handler
    ↓
Storage Layer
    ├─ SQLite (계좌, 거래 메타데이터)
    ├─ RocksDB (블록체인 상태)
    └─ Redis (캐싱)
    ↓
Blockchain Network
    ├─ Local Node
    └─ P2P Network (선택사항)
```

---

## 📊 데이터 모델

### 1. Account (계좌)

```sql
CREATE TABLE accounts (
  id TEXT PRIMARY KEY,                    -- UUID
  user_id TEXT NOT NULL,                  -- 사용자 ID
  account_type ENUM('checking', 'savings', 'credit'),
  currency TEXT DEFAULT 'USD',
  balance DECIMAL(15,2) NOT NULL,
  status ENUM('active', 'suspended', 'closed'),
  kyc_status ENUM('pending', 'verified', 'rejected'),
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_user_accounts ON accounts(user_id);
```

**주요 필드**:
- `balance`: 계좌 잔액
- `status`: 계좌 상태 (활성, 정지, 종료)
- `kyc_status`: KYC(Know Your Customer) 검증 상태

### 2. User (사용자)

```sql
CREATE TABLE users (
  id TEXT PRIMARY KEY,
  email TEXT UNIQUE NOT NULL,
  phone TEXT UNIQUE,
  password_hash TEXT NOT NULL,
  first_name TEXT,
  last_name TEXT,
  date_of_birth DATE,
  kyc_document TEXT,
  kyc_verified_at TIMESTAMP,
  status ENUM('active', 'suspended', 'deleted'),
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);
```

### 3. Transaction (거래)

```sql
CREATE TABLE transactions (
  id TEXT PRIMARY KEY,                    -- UUID
  from_account_id TEXT NOT NULL,
  to_account_id TEXT NOT NULL,
  amount DECIMAL(15,2) NOT NULL,
  currency TEXT DEFAULT 'USD',
  type ENUM('transfer', 'deposit', 'withdrawal', 'fee'),
  description TEXT,
  status ENUM('pending', 'confirmed', 'failed', 'reversed'),
  blockchain_hash TEXT,                   -- 블록체인 해시
  blockchain_block_number INTEGER,
  created_at TIMESTAMP,
  confirmed_at TIMESTAMP,
  FOREIGN KEY (from_account_id) REFERENCES accounts(id),
  FOREIGN KEY (to_account_id) REFERENCES accounts(id)
);

CREATE INDEX idx_from_account ON transactions(from_account_id);
CREATE INDEX idx_to_account ON transactions(to_account_id);
CREATE INDEX idx_status ON transactions(status);
CREATE INDEX idx_blockchain_hash ON transactions(blockchain_hash);
```

### 4. Blockchain Block (블록)

```sql
CREATE TABLE blockchain_blocks (
  block_number INTEGER PRIMARY KEY,
  timestamp TIMESTAMP NOT NULL,
  previous_hash TEXT,
  transactions_count INTEGER,
  merkle_root TEXT,
  nonce INTEGER,
  miner TEXT,
  difficulty INTEGER,
  total_fees DECIMAL(15,2),
  block_hash TEXT UNIQUE,
  created_at TIMESTAMP
);
```

### 5. Audit Log (감시)

```sql
CREATE TABLE audit_logs (
  id TEXT PRIMARY KEY,
  user_id TEXT,
  action TEXT,
  resource_type TEXT,
  resource_id TEXT,
  details JSON,
  ip_address TEXT,
  user_agent TEXT,
  timestamp TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE INDEX idx_user_audit ON audit_logs(user_id);
CREATE INDEX idx_action ON audit_logs(action);
```

---

## 💼 비즈니스 로직

### Account Manager

```freelang
// Account 인터페이스
struct Account {
  id: string,
  userId: string,
  accountType: string,  // "checking", "savings", "credit"
  currency: string,
  balance: f64,
  status: string,
  kycStatus: string,
  createdAt: i64
}

// 계좌 생성
fn create_account(user_id: string, account_type: string) -> Result<Account, Error> {
  // 1. 사용자 존재 확인
  // 2. 계좌 유형 검증
  // 3. 계좌 ID 생성
  // 4. DB에 저장
  // 5. 감시 로그 기록
  // 반환: 새 계좌
}

// 계좌 조회
fn get_account(account_id: string) -> Result<Account, Error> {
  // 1. 권한 확인
  // 2. DB 조회
  // 3. 캐시 업데이트
  // 반환: 계좌 정보
}

// 계좌 비활성화
fn close_account(account_id: string) -> Result<void, Error> {
  // 1. 잔액 확인 (0이어야 함)
  // 2. 상태 업데이트
  // 3. 감시 로그 기록
}
```

### Transaction Processor

```freelang
// Transaction 인터페이스
struct Transaction {
  id: string,
  fromAccountId: string,
  toAccountId: string,
  amount: f64,
  currency: string,
  type: string,  // "transfer", "deposit", "withdrawal"
  status: string,
  blockchainHash: string,
  createdAt: i64
}

// 거래 실행
fn execute_transaction(
  from_account_id: string,
  to_account_id: string,
  amount: f64
) -> Result<Transaction, Error> {
  // 1. 입력 검증 (amount > 0)
  // 2. 계좌 존재 확인
  // 3. 잔액 확인
  // 4. 거래 생성 (상태: pending)
  // 5. 블록체인에 추가
  // 6. 잔액 업데이트
  // 7. 거래 상태 변경 (confirmed)
  // 8. 감시 로그 기록
  // 반환: 거래 정보
}

// 거래 취소 (Reversal)
fn reverse_transaction(transaction_id: string) -> Result<void, Error> {
  // 1. 거래 조회
  // 2. 상태 확인 (confirmed만 가능)
  // 3. 역거래 생성
  // 4. 잔액 복구
  // 5. 원본 거래 상태 변경 (reversed)
}

// 거래 히스토리
fn get_transaction_history(
  account_id: string,
  limit: i64,
  offset: i64
) -> Result<Transaction[], Error> {
  // 1. 권한 확인
  // 2. DB 조회 (페이지네이션)
  // 3. 캐시 업데이트
  // 반환: 거래 목록
}
```

### Validation Engine

```freelang
// 거래 검증
fn validate_transaction(tx: Transaction) -> Result<void, Error> {
  // 1. 금액 검증
  validate_amount(tx.amount)?

  // 2. 계좌 검증
  validate_accounts(tx.from_account_id, tx.to_account_id)?

  // 3. 잔액 검증
  validate_balance(tx.from_account_id, tx.amount)?

  // 4. KYC 검증 (금액에 따라)
  if tx.amount > 10000.0 {
    validate_kyc(tx.from_account_id)?
  }

  // 5. AML 검증 (의심 거래 탐지)
  validate_aml(tx)?

  // 6. 일일 한도 검증
  validate_daily_limit(tx.from_account_id, tx.amount)?

  return Ok(())
}
```

---

## 🔗 블록체인 통합

### Blockchain Structure

```freelang
struct Block {
  blockNumber: i64,
  timestamp: i64,
  previousHash: string,
  transactions: Transaction[],
  transactionsCount: i64,
  merkleRoot: string,
  nonce: i64,
  miner: string,
  difficulty: i64,
  totalFees: f64,
  blockHash: string
}

struct Blockchain {
  chain: Block[],
  difficulty: i64,
  miningReward: f64,
  pendingTransactions: Transaction[]
}
```

### 합의 메커니즘

#### Proof of Work (PoW)

```freelang
fn mine_block(blockchain: &mut Blockchain) -> Block {
  let pending = blockchain.pendingTransactions

  // 1. Merkle Tree 생성
  let merkle_root = calculate_merkle_root(pending)

  // 2. Block 생성
  let mut block = Block {
    blockNumber: len(blockchain.chain),
    timestamp: current_time(),
    previousHash: blockchain.chain[len-1].blockHash,
    transactions: pending,
    transactionsCount: len(pending),
    merkleRoot: merkle_root,
    nonce: 0,
    miner: "system",
    difficulty: blockchain.difficulty,
    totalFees: calculate_total_fees(pending),
    blockHash: ""
  }

  // 3. Proof of Work 계산
  while !is_valid_pow(calculate_hash(&block), blockchain.difficulty) {
    block.nonce = block.nonce + 1
  }

  block.blockHash = calculate_hash(&block)
  return block
}

fn calculate_hash(block: Block) -> string {
  let data = str(block.blockNumber) +
             str(block.timestamp) +
             block.previousHash +
             block.merkleRoot +
             str(block.nonce)
  return sha256(data)
}
```

---

## 🔐 보안

### Authentication

```freelang
// 로그인
fn login(email: string, password: string) -> Result<Token, Error> {
  // 1. 사용자 조회
  let user = get_user_by_email(email)?

  // 2. 비밀번호 확인 (bcrypt)
  verify_password(password, user.password_hash)?

  // 3. JWT 토큰 생성
  let token = create_jwt(user.id, expiration: 24h)

  // 4. 로그인 기록
  log_audit("LOGIN", "user", user.id)

  return Ok(token)
}

// 토큰 검증
fn verify_token(token: string) -> Result<Claims, Error> {
  let claims = decode_jwt(token)?
  return Ok(claims)
}
```

### Authorization

```freelang
// 역할 기반 접근 제어 (RBAC)
enum Role {
  Admin,
  User,
  Support,
  Auditor
}

fn check_permission(user_id: string, action: string) -> Result<void, Error> {
  let user = get_user(user_id)?
  let role = get_user_role(user_id)?

  match action {
    "transfer" => {
      if role != Role::User && role != Role::Admin {
        return Err("Permission denied")
      }
    },
    "close_account" => {
      if role != Role::Admin {
        return Err("Permission denied")
      }
    },
    _ => return Err("Unknown action")
  }

  return Ok(())
}
```

---

## 📡 API 엔드포인트

### Authentication
```
POST   /api/auth/register         (회원가입)
POST   /api/auth/login            (로그인)
POST   /api/auth/logout           (로그아웃)
POST   /api/auth/refresh          (토큰 갱신)
```

### Accounts
```
POST   /api/accounts              (계좌 생성)
GET    /api/accounts/:id          (계좌 조회)
GET    /api/accounts              (사용자 모든 계좌)
PUT    /api/accounts/:id          (계좌 정보 수정)
DELETE /api/accounts/:id          (계좌 종료)
```

### Transactions
```
POST   /api/transactions          (거래 생성)
GET    /api/transactions/:id      (거래 조회)
GET    /api/accounts/:id/transactions  (계좌 거래 히스토리)
POST   /api/transactions/:id/reverse    (거래 취소)
```

### Blockchain
```
GET    /api/blockchain/blocks     (블록 목록)
GET    /api/blockchain/blocks/:number (특정 블록)
GET    /api/blockchain/status     (체인 상태)
POST   /api/blockchain/mine       (채굴)
```

### Admin
```
GET    /api/admin/users           (사용자 목록)
GET    /api/admin/logs            (감시 로그)
PUT    /api/admin/kyc/:id         (KYC 승인)
```

---

## 📈 성능 목표

| 메트릭 | 목표 | 달성 여부 |
|--------|------|----------|
| 거래 처리량 (TPS) | 1,000+ | ⏳ |
| 거래 확인 시간 | < 2초 | ⏳ |
| API 응답 시간 | < 100ms | ⏳ |
| 가용성 | 99.9% | ⏳ |
| 동시 사용자 | 10,000+ | ⏳ |

---

## 🧪 테스트 전략

### Unit Tests
- Account 생성/조회/종료
- Transaction 검증/실행/취소
- Blockchain 채굴/검증

### Integration Tests
- 계좌 생성 → 거래 → 블록체인 기록
- 권한 검증 → 거래 실행
- KYC 검증 → 대액 거래

### Load Tests
- 1,000 TPS 거래 처리
- 10,000 동시 사용자
- 네트워크 지연 시뮬레이션

---

## 🗺️ 구현 로드맵

### Phase 1: 핵심 모듈 (2주)
- [ ] Account Manager 완성
- [ ] Transaction Processor 완성
- [ ] Validation Engine 완성

### Phase 2: 블록체인 (1주)
- [ ] Blockchain 구현
- [ ] Mining 구현
- [ ] Chain Validation 구현

### Phase 3: API & 배포 (1주)
- [ ] REST API 완성
- [ ] Docker 설정
- [ ] 라이브 배포

### Phase 4: 테스트 & 최적화 (1주)
- [ ] 성능 테스트
- [ ] 보안 감시
- [ ] 모니터링

---

## 📚 기술 스택

| 계층 | 기술 |
|------|------|
| 언어 | FreeLang v4, Go 1.18+ |
| API | Express.js / Go Gin |
| DB | SQLite (프로토타입), PostgreSQL (프로덕션) |
| 캐싱 | Redis |
| 블록체인 | 자체 구현 (PoW) |
| 보안 | JWT, bcrypt, SHA-256 |
| 배포 | Docker, Docker Compose |
| 모니터링 | Prometheus, Grafana |

---

## 📝 다음 단계

1. **데이터 모델 최종 확정** - DBA 리뷰
2. **Account Manager 구현 시작** - Go/FreeLang
3. **Database 스키마 생성** - SQLite
4. **API 스켈레톤 작성** - Express/Go

---

**설계 완료**: 2026-06-01
**상태**: 🟡 **구현 준비 완료**

다음: [Account Manager 구현](/docs/bank-system/implementation)
