// 🏦 FreeLang Bank Dashboard - TypeScript Types

// 계좌 타입
export interface Account {
  id: string;
  name: string;
  type: "Checking" | "Savings" | "MoneyMarket" | "CD";
  balance: number;
  rate: number;
  status: "active" | "frozen" | "closed";
  created_at: number;
  updated_at: number;
}

// 거래 타입
export interface Transaction {
  id: string;
  from_account_id: string;
  to_account_id: string;
  amount: number;
  fee: number;
  type: "Deposit" | "Withdraw" | "Transfer" | "Interest" | "reverse";
  status: "pending" | "completed" | "failed" | "reversed";
  description?: string;
  created_at: number;
  completed_at?: number;
}

// 사기 경고 타입
export interface FraudAlert {
  id: string;
  transaction_id: string;
  severity: "critical" | "high" | "medium" | "low";
  score: number;
  reason: string;
  timestamp: number;
}

// 이자 정보 타입
export interface InterestInfo {
  account_id: string;
  balance: number;
  rate: number;
  daily_interest: number;
  monthly_interest: number;
  annual_interest: number;
  annual_interest_after_tax: number;
  tax_rate: number;
}

// 일일 리포트 타입
export interface DailyReport {
  date: string;
  total_transactions: number;
  total_volume: number;
  total_fees: number;
  fraud_alerts: number;
}

// 월간 리포트 타입
export interface MonthlyReport {
  month: string;
  total_transactions: number;
  total_volume: number;
  average_transaction: number;
  total_fees: number;
  total_interest: number;
}

// 감시 로그 타입
export interface AuditLog {
  id: string;
  action: string;
  account_id: string;
  description: string;
  ip_address: string;
  user_agent: string;
  timestamp: number;
}

// API 응답 타입
export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

// 대시보드 통계 타입
export interface DashboardStats {
  total_accounts: number;
  total_balance: number;
  total_transactions: number;
  total_volume: number;
  fraud_alerts_count: number;
}
