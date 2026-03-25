// 🏦 FreeLang Bank API Client

import axios, { AxiosInstance } from "axios";
import {
  Account,
  Transaction,
  FraudAlert,
  InterestInfo,
  DailyReport,
  MonthlyReport,
} from "../types";

const API_BASE_URL = process.env.REACT_APP_API_URL || "http://localhost:8080";

class BankAPI {
  private api: AxiosInstance;

  constructor() {
    this.api = axios.create({
      baseURL: API_BASE_URL,
      headers: {
        "Content-Type": "application/json",
      },
    });

    // 요청 인터셉터
    this.api.interceptors.request.use((config) => {
      const token = localStorage.getItem("token");
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }
      return config;
    });

    // 응답 인터셉터
    this.api.interceptors.response.use(
      (response) => response,
      (error) => {
        if (error.response?.status === 401) {
          localStorage.removeItem("token");
          window.location.href = "/login";
        }
        return Promise.reject(error);
      }
    );
  }

  // 📋 계좌 API

  /**
   * 모든 계좌 조회
   */
  async getAccounts(): Promise<Account[]> {
    const response = await this.api.get("/api/accounts");
    return response.data.accounts || [];
  }

  /**
   * 특정 계좌 조회
   */
  async getAccount(id: string): Promise<Account> {
    const response = await this.api.get(`/api/accounts/${id}`);
    return response.data;
  }

  /**
   * 계좌 생성
   */
  async createAccount(name: string, type: string, rate: number = 0): Promise<Account> {
    const response = await this.api.post("/api/accounts", {
      name,
      type,
      rate,
    });
    return response.data;
  }

  /**
   * 계좌 업데이트
   */
  async updateAccount(id: string, updates: Partial<Account>): Promise<Account> {
    const response = await this.api.put(`/api/accounts/${id}`, updates);
    return response.data;
  }

  /**
   * 계좌 삭제
   */
  async deleteAccount(id: string): Promise<void> {
    await this.api.delete(`/api/accounts/${id}`);
  }

  // 💳 거래 API

  /**
   * 거래 생성
   */
  async createTransaction(
    from_account_id: string,
    to_account_id: string,
    amount: number,
    type: string = "Transfer",
    description?: string
  ): Promise<Transaction> {
    const response = await this.api.post("/api/transactions", {
      from_account_id,
      to_account_id,
      amount,
      type,
      description,
    });
    return response.data;
  }

  /**
   * 거래 조회
   */
  async getTransaction(id: string): Promise<Transaction> {
    const response = await this.api.get(`/api/transactions/${id}`);
    return response.data;
  }

  /**
   * 계좌의 거래 내역 조회
   */
  async getAccountTransactions(accountId: string): Promise<Transaction[]> {
    const response = await this.api.get(`/api/accounts/${accountId}/transactions`);
    return response.data.transactions || [];
  }

  /**
   * 거래 취소
   */
  async reverseTransaction(transactionId: string): Promise<Transaction> {
    const response = await this.api.post("/api/transactions/reverse", {
      transaction_id: transactionId,
    });
    return response.data;
  }

  // 🔍 사기 탐지 API

  /**
   * 사기 검사
   */
  async checkFraud(
    amount: number,
    frequency: number = 0,
    balance_drain_pct: number = 0
  ): Promise<any> {
    const response = await this.api.post("/api/fraud/check", {
      amount,
      frequency,
      balance_drain_pct,
    });
    return response.data;
  }

  /**
   * 사기 경고 목록 조회
   */
  async getFraudAlerts(): Promise<FraudAlert[]> {
    const response = await this.api.get("/api/fraud/alerts");
    return response.data.alerts || [];
  }

  // 💰 이자 & 리포트 API

  /**
   * 이자 정보 조회
   */
  async getInterest(accountId: string): Promise<InterestInfo> {
    const response = await this.api.get(`/api/interest/${accountId}`);
    return response.data;
  }

  /**
   * 일일 리포트 조회
   */
  async getDailyReport(date: string): Promise<DailyReport> {
    const response = await this.api.get(`/api/reports/daily/${date}`);
    return response.data;
  }

  /**
   * 월간 리포트 조회
   */
  async getMonthlyReport(yearMonth: string): Promise<MonthlyReport> {
    const response = await this.api.get(`/api/reports/monthly/${yearMonth}`);
    return response.data;
  }

  // 🏥 헬스 체크

  /**
   * 서버 헬스 체크
   */
  async healthCheck(): Promise<boolean> {
    try {
      const response = await this.api.get("/health");
      return response.data.status === "OK";
    } catch {
      return false;
    }
  }
}

export default new BankAPI();
