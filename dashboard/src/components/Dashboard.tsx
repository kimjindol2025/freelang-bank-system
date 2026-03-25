// 메인 대시보드 컴포넌트

import React, { useState, useEffect } from "react";
import { Account, Transaction } from "../types";
import api from "../services/api";
import AccountList from "./AccountList";
import TransactionForm from "./TransactionForm";
import FraudDetection from "./FraudDetection";

interface DashboardProps {
  onLogout?: () => void;
}

const Dashboard: React.FC<DashboardProps> = ({ onLogout }) => {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [selectedAccount, setSelectedAccount] = useState<Account | null>(null);
  const [transactions, setTransactions] = useState<Transaction[]>([]);
  const [activeTab, setActiveTab] = useState<"accounts" | "transactions" | "fraud">(
    "accounts"
  );
  const [serverStatus, setServerStatus] = useState<"connected" | "disconnected">(
    "disconnected"
  );
  const [refreshTrigger, setRefreshTrigger] = useState(0);
  const [username, setUsername] = useState("");

  // 초기 로드
  useEffect(() => {
    loadAccounts();
    checkServerStatus();
    const storedUsername = localStorage.getItem("username") || "User";
    setUsername(storedUsername);
  }, []);

  // 선택된 계좌의 거래 내역 로드
  useEffect(() => {
    if (selectedAccount) {
      loadTransactions(selectedAccount.id);
    }
  }, [selectedAccount, refreshTrigger]);

  const checkServerStatus = async () => {
    try {
      const status = await api.healthCheck();
      setServerStatus(status ? "connected" : "disconnected");
    } catch {
      setServerStatus("disconnected");
    }
  };

  const loadAccounts = async () => {
    try {
      const data = await api.getAccounts();
      setAccounts(data);
      if (data.length > 0 && !selectedAccount) {
        setSelectedAccount(data[0]);
      }
    } catch (error) {
      console.error("Failed to load accounts:", error);
    }
  };

  const loadTransactions = async (accountId: string) => {
    try {
      const data = await api.getAccountTransactions(accountId);
      setTransactions(data);
    } catch (error) {
      console.error("Failed to load transactions:", error);
    }
  };

  const handleTransactionCreated = () => {
    loadAccounts();
    setRefreshTrigger((prev) => prev + 1);
  };

  const formatBalance = (balance: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(balance);
  };

  return (
    <div style={{ minHeight: "100vh", backgroundColor: "#ecf0f1" }}>
      {/* 헤더 */}
      <header
        style={{
          backgroundColor: "#2c3e50",
          color: "white",
          padding: "20px",
          boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
        }}
      >
        <div style={{ maxWidth: "1200px", margin: "0 auto" }}>
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
            <h1 style={{ margin: 0 }}>🏦 FreeLang Bank Dashboard</h1>
            <div style={{ display: "flex", gap: "20px", alignItems: "center" }}>
              <div>
                {serverStatus === "connected" ? (
                  <span style={{ color: "#2ecc71" }}>✅ 서버 연결됨</span>
                ) : (
                  <span style={{ color: "#e74c3c" }}>❌ 서버 연결 끊김</span>
                )}
              </div>
              <div style={{ display: "flex", gap: "10px", alignItems: "center" }}>
                <span>👤 {username}</span>
                <button
                  onClick={onLogout}
                  style={{
                    padding: "8px 16px",
                    backgroundColor: "#e74c3c",
                    color: "white",
                    border: "none",
                    borderRadius: "4px",
                    cursor: "pointer",
                    fontSize: "12px",
                    fontWeight: "bold",
                  }}
                >
                  로그아웃
                </button>
              </div>
            </div>
          </div>
        </div>
      </header>

      {/* 메인 콘텐츠 */}
      <div style={{ maxWidth: "1200px", margin: "0 auto", padding: "20px" }}>
        {/* 현재 계좌 정보 */}
        {selectedAccount && (
          <div
            style={{
              backgroundColor: "white",
              padding: "20px",
              borderRadius: "8px",
              marginBottom: "20px",
              boxShadow: "0 2px 4px rgba(0,0,0,0.1)",
            }}
          >
            <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr 1fr", gap: "20px" }}>
              <div>
                <div style={{ fontSize: "12px", color: "#666" }}>계좌명</div>
                <div style={{ fontSize: "18px", fontWeight: "bold" }}>
                  {selectedAccount.name}
                </div>
              </div>
              <div>
                <div style={{ fontSize: "12px", color: "#666" }}>계좌 타입</div>
                <div style={{ fontSize: "18px", fontWeight: "bold" }}>
                  {selectedAccount.type}
                </div>
              </div>
              <div>
                <div style={{ fontSize: "12px", color: "#666" }}>잔액</div>
                <div
                  style={{
                    fontSize: "24px",
                    fontWeight: "bold",
                    color: "#2ecc71",
                  }}
                >
                  {formatBalance(selectedAccount.balance)}
                </div>
              </div>
              <div>
                <div style={{ fontSize: "12px", color: "#666" }}>이율</div>
                <div style={{ fontSize: "18px", fontWeight: "bold" }}>
                  {selectedAccount.rate}% APY
                </div>
              </div>
            </div>
          </div>
        )}

        {/* 탭 네비게이션 */}
        <div
          style={{
            display: "flex",
            gap: "10px",
            marginBottom: "20px",
            borderBottom: "2px solid #3498db",
          }}
        >
          {["accounts", "transactions", "fraud"].map((tab) => (
            <button
              key={tab}
              onClick={() => setActiveTab(tab as any)}
              style={{
                padding: "10px 20px",
                backgroundColor: activeTab === tab ? "#3498db" : "transparent",
                color: activeTab === tab ? "white" : "#2c3e50",
                border: "none",
                cursor: "pointer",
                fontWeight: activeTab === tab ? "bold" : "normal",
                borderBottom: activeTab === tab ? "3px solid white" : "none",
              }}
            >
              {tab === "accounts" && "📋 계좌"}
              {tab === "transactions" && "💳 거래"}
              {tab === "fraud" && "🔍 사기탐지"}
            </button>
          ))}
        </div>

        {/* 탭 콘텐츠 */}
        <div style={{ backgroundColor: "white", borderRadius: "8px", boxShadow: "0 2px 4px rgba(0,0,0,0.1)" }}>
          {activeTab === "accounts" && (
            <div>
              <AccountList
                onSelectAccount={setSelectedAccount}
                refreshTrigger={refreshTrigger}
              />
              {selectedAccount && (
                <div style={{ padding: "20px", borderTop: "1px solid #eee" }}>
                  <TransactionForm
                    accounts={accounts}
                    onTransactionCreated={handleTransactionCreated}
                    selectedAccountId={selectedAccount.id}
                  />
                </div>
              )}
            </div>
          )}

          {activeTab === "transactions" && (
            <div style={{ padding: "20px" }}>
              <h2>💳 거래 내역</h2>
              {selectedAccount ? (
                transactions.length === 0 ? (
                  <p>거래 내역이 없습니다</p>
                ) : (
                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "repeat(auto-fill, minmax(300px, 1fr))",
                      gap: "15px",
                    }}
                  >
                    {transactions.map((txn) => (
                      <div
                        key={txn.id}
                        style={{
                          border: "1px solid #ddd",
                          borderRadius: "8px",
                          padding: "15px",
                          backgroundColor: "#f8f9fa",
                        }}
                      >
                        <div style={{ fontWeight: "bold", marginBottom: "10px" }}>
                          {txn.type}
                        </div>
                        <div style={{ fontSize: "12px", marginBottom: "5px" }}>
                          <strong>ID:</strong> {txn.id}
                        </div>
                        <div
                          style={{
                            fontSize: "20px",
                            fontWeight: "bold",
                            color: txn.type === "Deposit" ? "#2ecc71" : "#e74c3c",
                            marginBottom: "10px",
                          }}
                        >
                          {txn.type === "Deposit" ? "+" : "-"}${txn.amount.toFixed(2)}
                        </div>
                        <div style={{ fontSize: "12px" }}>
                          수수료: ${txn.fee.toFixed(2)}
                        </div>
                        <div
                          style={{
                            fontSize: "12px",
                            marginTop: "10px",
                            padding: "5px",
                            backgroundColor: txn.status === "completed" ? "#d4edda" : "#fff3cd",
                            borderRadius: "4px",
                            textAlign: "center",
                          }}
                        >
                          {txn.status === "completed" ? "✅" : "⏳"} {txn.status}
                        </div>
                      </div>
                    ))}
                  </div>
                )
              ) : (
                <p>계좌를 선택하세요</p>
              )}
            </div>
          )}

          {activeTab === "fraud" && <FraudDetection />}
        </div>
      </div>

      {/* 푸터 */}
      <footer
        style={{
          backgroundColor: "#2c3e50",
          color: "white",
          padding: "20px",
          textAlign: "center",
          marginTop: "40px",
        }}
      >
        <p style={{ margin: 0 }}>
          🏦 FreeLang Bank System - Phase 5 React Dashboard | v1.0.0
        </p>
      </footer>
    </div>
  );
};

export default Dashboard;
