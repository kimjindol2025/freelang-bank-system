// 거래 생성 폼 컴포넌트

import React, { useState } from "react";
import api from "../services/api";
import { Account } from "../types";

interface TransactionFormProps {
  accounts: Account[];
  onTransactionCreated: () => void;
  selectedAccountId?: string;
}

const TransactionForm: React.FC<TransactionFormProps> = ({
  accounts,
  onTransactionCreated,
  selectedAccountId = "",
}) => {
  const [fromAccountId, setFromAccountId] = useState(selectedAccountId);
  const [toAccountId, setToAccountId] = useState("");
  const [amount, setAmount] = useState("");
  const [description, setDescription] = useState("");
  const [loading, setLoading] = useState(false);
  const [message, setMessage] = useState<{ type: string; text: string } | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setMessage(null);

    if (!fromAccountId || !toAccountId || !amount) {
      setMessage({ type: "error", text: "모든 필드를 입력하세요" });
      return;
    }

    if (fromAccountId === toAccountId) {
      setMessage({
        type: "error",
        text: "송금 계좌와 수취 계좌가 같을 수 없습니다",
      });
      return;
    }

    const parsedAmount = parseFloat(amount);
    if (isNaN(parsedAmount) || parsedAmount <= 0) {
      setMessage({ type: "error", text: "유효한 금액을 입력하세요" });
      return;
    }

    setLoading(true);
    try {
      await api.createTransaction(
        fromAccountId,
        toAccountId,
        parsedAmount,
        "Transfer",
        description
      );

      setMessage({ type: "success", text: "거래가 완료되었습니다" });
      setFromAccountId(selectedAccountId);
      setToAccountId("");
      setAmount("");
      setDescription("");

      // 2초 후 자동으로 메시지 제거
      setTimeout(() => {
        setMessage(null);
        onTransactionCreated();
      }, 2000);
    } catch (error: any) {
      const errorMsg =
        error.response?.data?.message || "거래 생성에 실패했습니다";
      setMessage({ type: "error", text: errorMsg });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ padding: "20px", backgroundColor: "#f8f9fa", borderRadius: "8px" }}>
      <h3>💳 거래 생성</h3>

      {message && (
        <div
          style={{
            padding: "10px",
            marginBottom: "15px",
            borderRadius: "4px",
            backgroundColor: message.type === "error" ? "#f8d7da" : "#d4edda",
            color: message.type === "error" ? "#721c24" : "#155724",
            border: `1px solid ${message.type === "error" ? "#f5c6cb" : "#c3e6cb"}`,
          }}
        >
          {message.type === "error" ? "❌" : "✅"} {message.text}
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: "15px" }}>
          <label style={{ display: "block", marginBottom: "5px", fontWeight: "bold" }}>
            📤 송금 계좌
          </label>
          <select
            value={fromAccountId}
            onChange={(e) => setFromAccountId(e.target.value)}
            disabled={loading}
            style={{
              width: "100%",
              padding: "8px",
              border: "1px solid #ccc",
              borderRadius: "4px",
            }}
          >
            <option value="">계좌를 선택하세요</option>
            {accounts.map((acc) => (
              <option key={acc.id} value={acc.id}>
                {acc.name} ({acc.type}) - ${acc.balance.toFixed(2)}
              </option>
            ))}
          </select>
        </div>

        <div style={{ marginBottom: "15px" }}>
          <label style={{ display: "block", marginBottom: "5px", fontWeight: "bold" }}>
            📥 수취 계좌
          </label>
          <select
            value={toAccountId}
            onChange={(e) => setToAccountId(e.target.value)}
            disabled={loading}
            style={{
              width: "100%",
              padding: "8px",
              border: "1px solid #ccc",
              borderRadius: "4px",
            }}
          >
            <option value="">계좌를 선택하세요</option>
            {accounts.map((acc) => (
              <option key={acc.id} value={acc.id}>
                {acc.name} ({acc.type}) - ${acc.balance.toFixed(2)}
              </option>
            ))}
          </select>
        </div>

        <div style={{ marginBottom: "15px" }}>
          <label style={{ display: "block", marginBottom: "5px", fontWeight: "bold" }}>
            💰 금액
          </label>
          <input
            type="number"
            value={amount}
            onChange={(e) => setAmount(e.target.value)}
            disabled={loading}
            placeholder="0.00"
            step="0.01"
            min="0"
            style={{
              width: "100%",
              padding: "8px",
              border: "1px solid #ccc",
              borderRadius: "4px",
              boxSizing: "border-box",
            }}
          />
        </div>

        <div style={{ marginBottom: "15px" }}>
          <label style={{ display: "block", marginBottom: "5px", fontWeight: "bold" }}>
            📝 설명 (선택)
          </label>
          <input
            type="text"
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            disabled={loading}
            placeholder="거래 설명"
            style={{
              width: "100%",
              padding: "8px",
              border: "1px solid #ccc",
              borderRadius: "4px",
              boxSizing: "border-box",
            }}
          />
        </div>

        <button
          type="submit"
          disabled={loading}
          style={{
            width: "100%",
            padding: "10px",
            backgroundColor: loading ? "#ccc" : "#3498db",
            color: "white",
            border: "none",
            borderRadius: "4px",
            cursor: loading ? "not-allowed" : "pointer",
            fontWeight: "bold",
          }}
        >
          {loading ? "처리 중..." : "💳 거래 생성"}
        </button>
      </form>
    </div>
  );
};

export default TransactionForm;
