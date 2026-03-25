# 🏦 FreeLang Bank System - Phase 5: React 웹 대시보드

**작성일**: 2026-03-25 | **상태**: ✅ 구현 완료 | **완성도**: 85%

---

## 📋 개요

Phase 5는 **React 웹 대시보드**를 구현하여 Go REST API 서버를 시각적 인터페이스로 제어할 수 있게 합니다.

### 핵심 기술 스택
- **프레임워크**: React 18.2 (TypeScript)
- **HTTP 클라이언트**: Axios
- **UI/스타일**: CSS-in-JS (인라인 스타일)
- **포트**: 3000

---

## 📁 프로젝트 구조

```
dashboard/
├── src/
│   ├── types/
│   │   └── index.ts              # TypeScript 타입 정의
│   ├── services/
│   │   └── api.ts                # API 클라이언트
│   ├── components/
│   │   ├── Dashboard.tsx          # 메인 대시보드
│   │   ├── AccountList.tsx        # 계좌 목록
│   │   ├── TransactionForm.tsx    # 거래 생성 폼
│   │   └── FraudDetection.tsx     # 사기 탐지
│   ├── App.tsx                   # 메인 앱 컴포넌트
│   ├── App.css                   # 앱 스타일
│   ├── index.tsx                 # 진입점
│   └── index.css                 # 글로벌 스타일
├── public/
│   └── index.html                # HTML 템플릿
├── package.json                  # 의존성
└── .env.example                  # 환경변수 예제
```

---

## 🎨 컴포넌트 구조

### 1️⃣ TypeScript 타입 (src/types/index.ts - 87줄)

✅ **데이터 모델**
```typescript
// Account
interface Account {
  id, name, type, balance, rate, status
  created_at, updated_at
}

// Transaction
interface Transaction {
  id, from_account_id, to_account_id
  amount, fee, type, status
  description, created_at, completed_at
}

// FraudAlert
interface FraudAlert {
  id, transaction_id, severity, score
  reason, timestamp
}

// InterestInfo
interface InterestInfo {
  account_id, balance, rate
  daily_interest, monthly_interest
  annual_interest, annual_interest_after_tax
}
```

### 2️⃣ API 서비스 (src/services/api.ts - 165줄)

✅ **API 클라이언트 클래스**
```typescript
class BankAPI {
  // 📋 계좌 API
  getAccounts()           // GET /api/accounts
  getAccount(id)          // GET /api/accounts/:id
  createAccount(...)      // POST /api/accounts
  updateAccount(id, ...) // PUT /api/accounts/:id
  deleteAccount(id)       // DELETE /api/accounts/:id

  // 💳 거래 API
  createTransaction(...) // POST /api/transactions
  getTransaction(id)      // GET /api/transactions/:id
  getAccountTransactions(id) // GET /api/accounts/:id/transactions
  reverseTransaction(id)  // POST /api/transactions/reverse

  // 🔍 사기 탐지 API
  checkFraud(...)         // POST /api/fraud/check
  getFraudAlerts()        // GET /api/fraud/alerts

  // 💰 이자/리포트 API
  getInterest(id)         // GET /api/interest/:id
  getDailyReport(date)    // GET /api/reports/daily/:date
  getMonthlyReport(month) // GET /api/reports/monthly/:month

  // 🏥 헬스 체크
  healthCheck()           // GET /health
}
```

✅ **인터셉터**
- 요청: 자동 JWT 토큰 추가
- 응답: 401 에러 시 로그인 페이지 리다이렉트

### 3️⃣ Dashboard 컴포넌트 (src/components/Dashboard.tsx - 236줄)

✅ **메인 대시보드**

**레이아웃**:
```
┌─────────────────────────────────────────┐
│  헤더: 🏦 FreeLang Bank Dashboard      │
│  서버 상태: ✅ 연결됨                    │
├─────────────────────────────────────────┤
│  현재 계좌 정보                         │
│  ┌──────────┬──────────┬──────────┐   │
│  │ 계좌명   │ 타입     │ 잔액     │   │
│  │ 이율     │ 상태     │          │   │
│  └──────────┴──────────┴──────────┘   │
├─────────────────────────────────────────┤
│  탭 네비게이션: [📋계좌] [💳거래] [🔍사기탐지] │
├─────────────────────────────────────────┤
│  탭 콘텐츠 (동적)                       │
├─────────────────────────────────────────┤
│  푸터: FreeLang Bank System v1.0.0     │
└─────────────────────────────────────────┘
```

**기능**:
- 실시간 서버 상태 확인
- 탭 기반 네비게이션
- 계좌 자동 선택
- 거래 내역 자동 로드
- 새로고침 트리거

### 4️⃣ AccountList 컴포넌트 (src/components/AccountList.tsx - 90줄)

✅ **계좌 카드 목록**

**기능**:
- 그리드 레이아웃 (자동 반응형)
- 계좌 타입별 색상 구분
  - Checking (파란색): #3498db
  - Savings (초록색): #2ecc71
  - MoneyMarket (주황색): #f39c12
  - CD (보라색): #9b59b6

**상호작용**:
- 마우스 호버: 카드 확대 (scale 1.05)
- 클릭: 계좌 선택
- 로딩 상태: 스피너 표시
- 에러 처리: 에러 메시지 표시

**정보**:
- 계좌명
- 계좌 타입
- 현재 잔액 (통화 포맷)
- 이율 (%)
- 상태 (활성/동결)

### 5️⃣ TransactionForm 컴포넌트 (src/components/TransactionForm.tsx - 147줄)

✅ **거래 생성 폼**

**입력 필드**:
1. 송금 계좌 (선택)
2. 수취 계좌 (선택)
3. 금액 (숫자)
4. 설명 (선택)

**검증**:
- 모든 필드 필수 확인
- 송금/수취 계좌 중복 확인
- 금액 > 0 확인
- 숫자 형식 확인

**피드백**:
- 성공: 초록색 메시지 + 자동 닫기 (2초)
- 에러: 빨간색 메시지
- 로딩: 버튼 비활성화

**기능**:
- 선택된 계좌 자동 선택
- 거래 후 폼 초기화
- 계좌 목록 자동 새로고침

### 6️⃣ FraudDetection 컴포넌트 (src/components/FraudDetection.tsx - 177줄)

✅ **사기 탐지 시스템**

**좌측: 사기 점수 계산**
- 거래 금액 입력
- 시간당 거래 건수 입력
- 잔액 감소율 (%) 입력
- 검사 버튼

**결과 표시**:
```
🚨 점수: 90/100
🚨 Critical (차단)
• Large transaction (>$100K)
• Unusual frequency (>100/hour)
• Balance drain (>80%)
```

**색상 코드**:
- 🚨 Critical: #e74c3c (빨강)
- 🔴 High: #f39c12 (주황)
- 🟡 Medium: #f1c40f (노랑)
- ✅ Low: #2ecc71 (초록)

**우측: 경고 목록**
- 최근 10개 경고 표시
- 스크롤 가능 (최대 500px)
- 타임스탬프 표시
- 심각도별 색상

---

## 🚀 실행 방법

### 개발 환경

```bash
# 1. 의존성 설치
cd dashboard
npm install

# 2. 환경변수 설정
cp .env.example .env
# REACT_APP_API_URL=http://localhost:8080

# 3. 개발 서버 시작
npm start

# 브라우저에서 http://localhost:3000 접속
```

### 프로덕션 빌드

```bash
npm run build
# build/ 디렉토리에 최적화된 파일 생성
```

---

## 🎨 UI/UX 특징

### 색상 스키마
```
Primary:   #3498db (파란색)
Success:   #2ecc71 (초록색)
Warning:   #f1c40f (노랑색)
Error:     #e74c3c (빨강색)
Dark:      #2c3e50 (짙은 회색)
Light:     #ecf0f1 (밝은 회색)
```

### 반응형 디자인
```
그리드 레이아웃: auto-fill, minmax(250px, 1fr)
모바일 지원: 디바이스 너비에 자동 조정
최대 너비: 1200px (컨테이너)
```

### 인터랙션
- 호버: 카드 확대, 버튼 색상 변화
- 클릭: 계좌 선택, 거래 생성
- 포커스: 입력 필드 테두리 변화
- 로딩: 버튼 비활성화, 텍스트 변화

---

## 🔄 API 통신 흐름

```
사용자 입력
    ↓
컴포넌트 이벤트 처리
    ↓
API 클라이언트 호출
    ↓
Go REST API 서버
    ↓
SQLite 데이터베이스
    ↓
응답 반환
    ↓
상태 업데이트
    ↓
UI 리렌더링
```

### 예제: 거래 생성

```typescript
// 1. 사용자가 거래 폼 제출
const handleSubmit = async (e) => {
  const response = await api.createTransaction(
    fromAccountId,
    toAccountId,
    amount,
    "Transfer"
  );

  // 2. 성공 메시지 표시
  setMessage({ type: "success", text: "거래 완료" });

  // 3. 데이터 새로고침
  onTransactionCreated();
};
```

---

## 📊 코드 통계

### Phase 5 코드
```
src/types/index.ts:              87줄
src/services/api.ts:             165줄
src/components/Dashboard.tsx:    236줄
src/components/AccountList.tsx:  90줄
src/components/TransactionForm.tsx: 147줄
src/components/FraudDetection.tsx: 177줄
src/App.tsx:                     13줄
src/index.tsx:                   15줄
src/App.css:                     20줄
src/index.css:                   9줄
public/index.html:               20줄
package.json:                    28줄
.env.example:                    1줄

총 1,008줄 (React TypeScript)
```

### 누적 코드 (Phase 1-5)
```
Phase 1-4: 5,621줄
Phase 5:   1,008줄
========
총합:      6,629줄
```

### 완성도
```
계좌 관리:     ✅ 100%
거래 처리:     ✅ 100%
이자 계산:     ✅ 100%
사기 탐지:     ✅ 100%
데이터베이스:  ✅ 100%
API 서버:      ✅ 100%
웹 대시보드:   ✅ 100%
배포:          ⚠️ 0%

평균 완성도:   85% (B+ 등급)
```

---

## 🔐 보안 고려사항

✅ **이미 구현됨**
- CORS 지원 (백엔드)
- JWT 토큰 자동 추가
- 401 에러 시 자동 로그아웃

⚠️ **향후 추가 필요**
- [ ] HTTPS 지원
- [ ] 토큰 갱신 (refresh token)
- [ ] 입력 살균 (XSS 방지)
- [ ] CSRF 토큰
- [ ] 레이트 제한

---

## 📱 브라우저 지원

- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

---

## 🎯 다음 단계

### Phase 6: Docker/Kubernetes 배포 (2주, 95% 완성도)
- [ ] Dockerfile 작성 (React)
- [ ] Docker Compose 설정 (3개 서비스)
- [ ] Kubernetes manifests
- [ ] CI/CD 파이프라인
- [ ] Nginx 리버스 프록시

### 배포 구성도
```
┌─────────────────────────────────────────┐
│          Internet (사용자)              │
└──────────────────┬──────────────────────┘
                   │
        ┌──────────▼──────────┐
        │   Nginx (80/443)    │
        │  (리버스 프록시)     │
        └──────────┬──────────┘
                   │
        ┌──────────┴──────────┐
        │                     │
   ┌────▼────┐          ┌────▼────┐
   │ React   │          │ Go API   │
   │ :3000   │          │ :8080    │
   └────┬────┘          └────┬────┘
        │                    │
        └──────────┬─────────┘
                   │
            ┌──────▼─────┐
            │   SQLite   │
            │ :database  │
            └────────────┘
```

---

## 📚 개발 팁

### 상태 관리
```typescript
const [accounts, setAccounts] = useState<Account[]>([]);
const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
const [refreshTrigger, setRefreshTrigger] = useState(0);

// 새로고침 트리거
useEffect(() => {
  if (selectedAccount) {
    loadTransactions(selectedAccount.id);
  }
}, [selectedAccount, refreshTrigger]);
```

### 에러 처리
```typescript
try {
  const data = await api.getAccounts();
  setAccounts(data);
} catch (err) {
  setError("계좌 목록을 불러올 수 없습니다");
  console.error(err);
}
```

### 로딩 상태
```typescript
const [loading, setLoading] = useState(false);

const loadData = async () => {
  setLoading(true);
  try {
    // 데이터 로드
  } finally {
    setLoading(false);
  }
};
```

---

**상태**: ✅ Phase 5 구현 완료
**다음**: Phase 6 Docker/Kubernetes 배포
