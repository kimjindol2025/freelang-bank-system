// 사기 탐지 컴포넌트

import React, { useState, useEffect } from "react";
import api from "../services/api";
import { FraudAlert } from "../types";

const FraudDetection: React.FC = () => {
  const [amount, setAmount] = useState("");
  const [frequency, setFrequency] = useState("0");
  const [balanceDrain, setBalanceDrain] = useState("0");
  const [result, setResult] = useState<any>(null);
  const [alerts, setAlerts] = useState<FraudAlert[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadingAlerts, setLoadingAlerts] = useState(false);

  useEffect(() => {
    loadAlerts();
  }, []);

  const loadAlerts = async () => {
    setLoadingAlerts(true);
    try {
      const data = await api.getFraudAlerts();
      setAlerts(data);
    } catch (error) {
      console.error("Failed to load fraud alerts:", error);
    } finally {
      setLoadingAlerts(false);
    }
  };

  const handleCheck = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const fraudResult = await api.checkFraud(
        parseFloat(amount) || 0,
        parseInt(frequency) || 0,
        parseFloat(balanceDrain) || 0
      );
      setResult(fraudResult);
      loadAlerts();
    } catch (error) {
      console.error("Fraud check failed:", error);
    } finally {
      setLoading(false);
    }
  };

  const getSeverityColor = (severity: string) => {
    switch (severity) {
      case "critical":
        return "#e74c3c";
      case "high":
        return "#f39c12";
      case "medium":
        return "#f1c40f";
      case "low":
        return "#2ecc71";
      default:
        return "#95a5a6";
    }
  };

  const getSeverityEmoji = (severity: string) => {
    switch (severity) {
      case "critical":
        return "🚨";
      case "high":
        return "🔴";
      case "medium":
        return "🟡";
      case "low":
        return "✅";
      default:
        return "❓";
    }
  };

  return (
    <div style={{ padding: "20px" }}>
      <h2>🔍 사기 탐지 시스템</h2>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "20px" }}>
        {/* 사기 검사 폼 */}
        <div
          style={{
            backgroundColor: "#f8f9fa",
            padding: "20px",
            borderRadius: "8px",
          }}
        >
          <h3>🔎 사기 점수 계산</h3>
          <form onSubmit={handleCheck}>
            <div style={{ marginBottom: "15px" }}>
              <label style={{ display: "block", marginBottom: "5px", fontWeight: "bold" }}>
                💰 거래 금액
              </label>
              <input
                type="number"
                value={amount}
                onChange={(e) => setAmount(e.target.value)}
                placeholder="예: 150000"
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
                📊 시간당 거래 건수
              </label>
              <input
                type="number"
                value={frequency}
                onChange={(e) => setFrequency(e.target.value)}
                placeholder="예: 120"
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
                📉 잔액 감소율 (%)
              </label>
              <input
                type="number"
                value={balanceDrain}
                onChange={(e) => setBalanceDrain(e.target.value)}
                placeholder="예: 90"
                max="100"
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
              {loading ? "검사 중..." : "🔍 사기 검사"}
            </button>
          </form>

          {result && (
            <div
              style={{
                marginTop: "20px",
                padding: "15px",
                backgroundColor: "#fff",
                borderLeft: `4px solid ${getSeverityColor(result.severity)}`,
                borderRadius: "4px",
              }}
            >
              <div style={{ fontSize: "24px", fontWeight: "bold", marginBottom: "10px" }}>
                {getSeverityEmoji(result.severity)} 점수: {result.score}/100
              </div>
              <div style={{ fontSize: "16px", fontWeight: "bold", color: getSeverityColor(result.severity) }}>
                {result.risk_level}
              </div>
              <div style={{ marginTop: "10px", fontSize: "12px" }}>
                {result.reasons?.map((reason: string, idx: number) => (
                  <div key={idx}>• {reason}</div>
                ))}
              </div>
            </div>
          )}
        </div>

        {/* 경고 목록 */}
        <div
          style={{
            backgroundColor: "#f8f9fa",
            padding: "20px",
            borderRadius: "8px",
          }}
        >
          <h3>⚠️ 최근 사기 경고</h3>
          {loadingAlerts ? (
            <p>로드 중...</p>
          ) : alerts.length === 0 ? (
            <p>사기 경고가 없습니다</p>
          ) : (
            <div
              style={{
                display: "flex",
                flexDirection: "column",
                gap: "10px",
                maxHeight: "500px",
                overflowY: "auto",
              }}
            >
              {alerts.slice(0, 10).map((alert) => (
                <div
                  key={alert.id}
                  style={{
                    padding: "10px",
                    backgroundColor: "#fff",
                    borderLeft: `4px solid ${getSeverityColor(alert.severity)}`,
                    borderRadius: "4px",
                    fontSize: "12px",
                  }}
                >
                  <div style={{ fontWeight: "bold", marginBottom: "5px" }}>
                    {getSeverityEmoji(alert.severity)} {alert.severity.toUpperCase()}
                  </div>
                  <div>점수: {alert.score}</div>
                  <div style={{ color: "#666", fontSize: "11px", marginTop: "5px" }}>
                    {alert.reason}
                  </div>
                  <div style={{ color: "#999", fontSize: "10px", marginTop: "5px" }}>
                    {new Date(alert.timestamp * 1000).toLocaleString("ko-KR")}
                  </div>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default FraudDetection;
