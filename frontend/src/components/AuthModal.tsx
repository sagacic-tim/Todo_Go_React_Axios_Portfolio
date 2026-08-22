// src/components/AuthModal.tsx
import React, { FormEvent, useEffect, useRef, useState } from "react";
import { login, signup } from "../services/authService";
import { useAuth } from "../contexts/AuthContext";
import { toMessage } from "../services/apiClient";
import styles from "./AuthModal.module.css";

type Tab = "login" | "signup";

interface AuthModalProps {
  open: boolean;
  /** Which tab to show when the modal opens. Defaults to "login". */
  initialTab?: Tab;
  onClose: () => void;
}

export const AuthModal: React.FC<AuthModalProps> = ({
  open,
  initialTab = "login",
  onClose,
}) => {
  const { setUser } = useAuth();
  const [tab, setTab] = useState<Tab>(initialTab);
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const emailRef = useRef<HTMLInputElement>(null);

  // Reset form whenever the modal opens or the tab changes.
  useEffect(() => {
    if (open) {
      setTab(initialTab);
      setEmail("");
      setPassword("");
      setConfirm("");
      setError("");
      setLoading(false);
      // Focus the email field on next tick.
      setTimeout(() => emailRef.current?.focus(), 50);
    }
  }, [open, initialTab]);

  // Allow Escape to close.
  useEffect(() => {
    if (!open) return;
    const handler = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open, onClose]);

  if (!open) return null;

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setError("");

    if (tab === "signup" && password !== confirm) {
      setError("Passwords do not match.");
      return;
    }
    if (tab === "signup" && password.length < 8) {
      setError("Password must be at least 8 characters.");
      return;
    }

    setLoading(true);
    try {
      const user =
        tab === "login"
          ? await login(email, password)
          : await signup(email, password);
      setUser(user);
      onClose();
    } catch (err) {
      setError(toMessage(err));
    } finally {
      setLoading(false);
    }
  };

  const switchTab = (next: Tab) => {
    setTab(next);
    setError("");
    setEmail("");
    setPassword("");
    setConfirm("");
  };

  return (
    <div
      className={styles.overlay}
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
      role="dialog"
      aria-modal="true"
      aria-label={tab === "login" ? "Sign in" : "Create account"}
    >
      <div className={styles.card}>
        <button
          className={styles.closeBtn}
          onClick={onClose}
          aria-label="Close"
          type="button"
        >
          ✕
        </button>

        {/* Tabs */}
        <div className={styles.tabs} role="tablist">
          <button
            role="tab"
            aria-selected={tab === "login"}
            className={`${styles.tab} ${tab === "login" ? styles.tabActive : ""}`}
            onClick={() => switchTab("login")}
            type="button"
          >
            Sign In
          </button>
          <button
            role="tab"
            aria-selected={tab === "signup"}
            className={`${styles.tab} ${tab === "signup" ? styles.tabActive : ""}`}
            onClick={() => switchTab("signup")}
            type="button"
          >
            Create Account
          </button>
        </div>

        {/* Form */}
        <form className={styles.form} onSubmit={handleSubmit} noValidate>
          {error && <p className={styles.errorMsg}>{error}</p>}

          <div className={styles.fieldGroup}>
            <label className={styles.label} htmlFor="auth-email">
              Email
            </label>
            <input
              id="auth-email"
              ref={emailRef}
              type="email"
              autoComplete="email"
              className={`${styles.input} ${error ? styles.inputError : ""}`}
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              required
              disabled={loading}
              maxLength={254}
              spellCheck={false}
              autoCapitalize="none"
              autoCorrect="off"
            />
          </div>

          <div className={styles.fieldGroup}>
            <label className={styles.label} htmlFor="auth-password">
              Password
            </label>
            <input
              id="auth-password"
              type="password"
              autoComplete={tab === "login" ? "current-password" : "new-password"}
              className={`${styles.input} ${error ? styles.inputError : ""}`}
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              required
              disabled={loading}
              maxLength={128}
            />
            {tab === "signup" && (
              <p className={styles.hint}>Minimum 8 characters.</p>
            )}
          </div>

          {tab === "signup" && (
            <div className={styles.fieldGroup}>
              <label className={styles.label} htmlFor="auth-confirm">
                Confirm Password
              </label>
              <input
                id="auth-confirm"
                type="password"
                autoComplete="new-password"
                className={`${styles.input} ${
                  confirm && confirm !== password ? styles.inputError : ""
                }`}
                value={confirm}
                onChange={(e) => setConfirm(e.target.value)}
                required
                disabled={loading}
                maxLength={128}
              />
            </div>
          )}

          <button
            type="submit"
            className={styles.submitBtn}
            disabled={loading || !email || !password}
          >
            {loading
              ? tab === "login"
                ? "Signing in…"
                : "Creating account…"
              : tab === "login"
              ? "Sign In"
              : "Create Account"}
          </button>
        </form>

        <p className={styles.switchText}>
          {tab === "login" ? (
            <>
              Don't have an account?{" "}
              <button
                type="button"
                className={styles.switchLink}
                onClick={() => switchTab("signup")}
              >
                Sign up
              </button>
            </>
          ) : (
            <>
              Already have an account?{" "}
              <button
                type="button"
                className={styles.switchLink}
                onClick={() => switchTab("login")}
              >
                Sign in
              </button>
            </>
          )}
        </p>
      </div>
    </div>
  );
};
