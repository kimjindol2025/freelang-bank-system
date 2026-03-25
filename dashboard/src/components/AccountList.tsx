// 계좌 목록 컴포넌트

import React, { useState, useEffect } from "react";
import { Account } from "../types";
import api from "../services/api";

interface AccountListProps {
  onSelectAccount: (account: Account) => void;
  refreshTrigger?: number;
}

const AccountList: React.FC<AccountListProps> = ({
  onSelectAccount,
  refreshTrigger = 0,
}) => {
  const [accounts, setAccounts] = useState<Account[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadAccounts();
  }, [refreshTrigger]);

  const loadAccounts = async () => {
    setLoading(true);
    setError(null);
    try {
      const data = await api.getAccounts();
      setAccounts(data);
    } catch (err) {
      setError("계좌 목록을 불러올 수 없습니다");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const formatBalance = (balance: number) => {
    return new Intl.NumberFormat("en-US", {
      style: "currency",
      currency: "USD",
    }).format(balance);
  };

  const getAccountTypeColor = (type: string) => {
    switch (type) {
      case "Checking":
        return "#3498db";
      case "Savings":
        return "#2ecc71";
      case "MoneyMarket":
        return "#f39c12";
      case "CD":
        return "#9b59b6";
      default:
        return "#95a5a6";
    }
  };

  if (loading) {
    return <div style={{ padding: "20px" }}>📋 계좌 목록 로드 중...</div>;
  }

  if (error) {
    return <div style={{ padding: "20px", color: "red" }}>❌ {error}</div>;
  }

  return (
    <div style={{ padding: "20px" }}>
      <h2>📋 계좌 목록</h2>
      {accounts.length === 0 ? (
        <p>등록된 계좌가 없습니다</p>
      ) : (
        <div
          style={{
            display: "grid",
            gridTemplateColumns: "repeat(auto-fill, minmax(250px, 1fr))",
            gap: "20px",
          }}
        >
          {accounts.map((account) => (
            <div
              key={account.id}
              onClick={() => onSelectAccount(account)}
              style={{
                border: `3px solid ${getAccountTypeColor(account.type)}`,
                borderRadius: "8px",
                padding: "15px",
                cursor: "pointer",
                transition: "transform 0.2s",
                backgroundColor: "#f8f9fa",
              }}
              onMouseOver={(e) => {
                (e.currentTarget as HTMLDivElement).style.transform =
                  "scale(1.05)";
              }}
              onMouseOut={(e) => {
                (e.currentTarget as HTMLDivElement).style.transform =
                  "scale(1)";
              }}
            >
              <div style={{ fontSize: "18px", fontWeight: "bold" }}>
                {account.name}
              </div>
              <div style={{ fontSize: "12px", color: "#666", margin: "5px 0" }}>
                📝 {account.type}
              </div>
              <div style={{ fontSize: "24px", fontWeight: "bold", margin: "10px 0" }}>
                {formatBalance(account.balance)}
              </div>
              <div style={{ fontSize: "12px", color: "#666" }}>
                📊 이율: {account.rate}%
              </div>
              <div
                style={{
                  fontSize: "12px",
                  marginTop: "10px",
                  padding: "5px",
                  backgroundColor: account.status === "active" ? "#d4edda" : "#f8d7da",
                  borderRadius: "4px",
                  textAlign: "center",
                }}
              >
                {account.status === "active" ? "✅ 활성" : "🔒 동결"}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default AccountList;
