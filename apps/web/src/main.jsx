import React from "react";
import { createRoot } from "react-dom/client";
import App from "./App.jsx";
import "./styles/tokens.css";
import "./styles/base.css";
import "./styles/vendor.css";

class AppErrorBoundary extends React.Component {
  constructor(props) {
    super(props);
    this.state = { error: null };
  }

  static getDerivedStateFromError(error) {
    return { error };
  }

  componentDidCatch(error, info) {
    console.error("Robot Control UI render error", error, info);
  }

  render() {
    if (this.state.error) {
      const message = import.meta.env.DEV
        ? this.state.error.message
        : "화면을 새로고침하거나 시스템 상태를 확인하세요.";
      return (
        <main className="grid min-h-screen place-content-center gap-2 bg-command-950 p-6 text-center text-slate-50">
          <strong className="block text-lg font-black">관제 화면을 표시하지 못했습니다.</strong>
          <span className="block text-sm font-bold text-red-200">{message}</span>
        </main>
      );
    }
    return this.props.children;
  }
}

createRoot(document.getElementById("root")).render(
  <React.StrictMode>
    <AppErrorBoundary>
      <App />
    </AppErrorBoundary>
  </React.StrictMode>
);
