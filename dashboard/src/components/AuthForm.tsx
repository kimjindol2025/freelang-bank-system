import React, { useState } from "react";
import api from "../services/api";

interface AuthFormProps {
  onAuthSuccess: (token: string) => void;
}

type AuthMode = "login" | "register";

const AuthForm: React.FC<AuthFormProps> = ({ onAuthSuccess }) => {
  const [mode, setMode] = useState<AuthMode>("login");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [username, setUsername] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (!email || !password) {
        setError("이메일과 비밀번호를 입력하세요");
        setLoading(false);
        return;
      }

      const response = await api.login(email, password);
      if (response.token) {
        localStorage.setItem("token", response.token);
        localStorage.setItem("user_id", response.user_id);
        localStorage.setItem("username", response.username);
        onAuthSuccess(response.token);
      }
    } catch (err: any) {
      setError(err.message || "로그인 실패");
    } finally {
      setLoading(false);
    }
  };

  const handleRegister = async (e: React.FormEvent) => {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      if (!username || !email || !password) {
        setError("모든 필드를 입력하세요");
        setLoading(false);
        return;
      }

      if (password.length < 8) {
        setError("비밀번호는 최소 8자 이상이어야 합니다");
        setLoading(false);
        return;
      }

      const response = await api.register(username, email, password);
      if (response.token) {
        localStorage.setItem("token", response.token);
        localStorage.setItem("user_id", response.user_id);
        localStorage.setItem("username", response.username);
        onAuthSuccess(response.token);
      }
    } catch (err: any) {
      setError(err.message || "회원가입 실패");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: "100vh",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        backgroundColor: "#ecf0f1",
        fontFamily: "Arial, sans-serif",
      }}
    >
      <div
        style={{
          backgroundColor: "white",
          padding: "40px",
          borderRadius: "8px",
          boxShadow: "0 4px 6px rgba(0,0,0,0.1)",
          width: "100%",
          maxWidth: "400px",
        }}
      >
        {/* 헤더 */}
        <div style={{ textAlign: "center", marginBottom: "30px" }}>
          <h1 style={{ fontSize: "32px", margin: "0 0 10px 0" }}>🏦</h1>
          <h2 style={{ margin: "0 0 5px 0", color: "#2c3e50" }}>FreeLang Bank</h2>
          <p style={{ margin: "0", color: "#7f8c8d", fontSize: "14px" }}>
            안전한 금융 서비스
          </p>
        </div>

        {/* 에러 메시지 */}
        {error && (
          <div
            style={{
              backgroundColor: "#ffe6e6",
              border: "1px solid #ff6b6b",
              color: "#e74c3c",
              padding: "12px",
              borderRadius: "4px",
              marginBottom: "20px",
              fontSize: "14px",
              textAlign: "center",
            }}
          >
            {error}
          </div>
        )}

        {/* 탭 */}
        <div style={{ display: "flex", gap: "10px", marginBottom: "30px" }}>
          <button
            onClick={() => {
              setMode("login");
              setError("");
            }}
            style={{
              flex: 1,
              padding: "10px",
              backgroundColor: mode === "login" ? "#3498db" : "#ecf0f1",
              color: mode === "login" ? "white" : "#2c3e50",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
              fontWeight: "bold",
              transition: "all 0.3s",
            }}
          >
            로그인
          </button>
          <button
            onClick={() => {
              setMode("register");
              setError("");
            }}
            style={{
              flex: 1,
              padding: "10px",
              backgroundColor: mode === "register" ? "#3498db" : "#ecf0f1",
              color: mode === "register" ? "white" : "#2c3e50",
              border: "none",
              borderRadius: "4px",
              cursor: "pointer",
              fontWeight: "bold",
              transition: "all 0.3s",
            }}
          >
            회원가입
          </button>
        </div>

        {/* 로그인 폼 */}
        {mode === "login" && (
          <form onSubmit={handleLogin}>
            <div style={{ marginBottom: "15px" }}>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                  color: "#2c3e50",
                }}
              >
                이메일
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="your@email.com"
                style={{
                  width: "100%",
                  padding: "10px",
                  border: "1px solid #bdc3c7",
                  borderRadius: "4px",
                  fontSize: "14px",
                  boxSizing: "border-box",
                }}
              />
            </div>

            <div style={{ marginBottom: "20px" }}>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                  color: "#2c3e50",
                }}
              >
                비밀번호
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="••••••••"
                style={{
                  width: "100%",
                  padding: "10px",
                  border: "1px solid #bdc3c7",
                  borderRadius: "4px",
                  fontSize: "14px",
                  boxSizing: "border-box",
                }}
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              style={{
                width: "100%",
                padding: "12px",
                backgroundColor: loading ? "#bdc3c7" : "#3498db",
                color: "white",
                border: "none",
                borderRadius: "4px",
                fontWeight: "bold",
                cursor: loading ? "not-allowed" : "pointer",
                fontSize: "14px",
                transition: "all 0.3s",
              }}
            >
              {loading ? "로그인 중..." : "로그인"}
            </button>
          </form>
        )}

        {/* 회원가입 폼 */}
        {mode === "register" && (
          <form onSubmit={handleRegister}>
            <div style={{ marginBottom: "15px" }}>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                  color: "#2c3e50",
                }}
              >
                사용자명
              </label>
              <input
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="3자 이상"
                style={{
                  width: "100%",
                  padding: "10px",
                  border: "1px solid #bdc3c7",
                  borderRadius: "4px",
                  fontSize: "14px",
                  boxSizing: "border-box",
                }}
              />
            </div>

            <div style={{ marginBottom: "15px" }}>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                  color: "#2c3e50",
                }}
              >
                이메일
              </label>
              <input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="your@email.com"
                style={{
                  width: "100%",
                  padding: "10px",
                  border: "1px solid #bdc3c7",
                  borderRadius: "4px",
                  fontSize: "14px",
                  boxSizing: "border-box",
                }}
              />
            </div>

            <div style={{ marginBottom: "20px" }}>
              <label
                style={{
                  display: "block",
                  marginBottom: "5px",
                  fontWeight: "bold",
                  color: "#2c3e50",
                }}
              >
                비밀번호
              </label>
              <input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="8자 이상"
                style={{
                  width: "100%",
                  padding: "10px",
                  border: "1px solid #bdc3c7",
                  borderRadius: "4px",
                  fontSize: "14px",
                  boxSizing: "border-box",
                }}
              />
            </div>

            <button
              type="submit"
              disabled={loading}
              style={{
                width: "100%",
                padding: "12px",
                backgroundColor: loading ? "#bdc3c7" : "#27ae60",
                color: "white",
                border: "none",
                borderRadius: "4px",
                fontWeight: "bold",
                cursor: loading ? "not-allowed" : "pointer",
                fontSize: "14px",
                transition: "all 0.3s",
              }}
            >
              {loading ? "회원가입 중..." : "회원가입"}
            </button>
          </form>
        )}

        {/* 정보 */}
        <div
          style={{
            marginTop: "20px",
            padding: "15px",
            backgroundColor: "#ecf0f1",
            borderRadius: "4px",
            fontSize: "12px",
            color: "#7f8c8d",
            textAlign: "center",
          }}
        >
          <p style={{ margin: "0" }}>
            🔒 모든 거래는 SSL 암호화로 보호됩니다
          </p>
          <p style={{ margin: "5px 0 0 0" }}>
            안전한 JWT 인증 시스템을 사용합니다
          </p>
        </div>
      </div>
    </div>
  );
};

export default AuthForm;
