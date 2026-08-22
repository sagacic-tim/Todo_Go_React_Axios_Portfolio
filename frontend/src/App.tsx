// frontend/src/App.tsx
import { useEffect, useState } from "react";
import { BrowserRouter as Router, Link } from "react-router-dom";
import AppRoutes from "./routes/routes";
import { AuthProvider, useAuth } from "./contexts/AuthContext";
import { AuthModal } from "./components/AuthModal";
import styles from "./App.module.css";

// ─── Inner app (needs AuthContext already mounted) ────────────────────────────

function AppShell() {
  const { authState, logout } = useAuth();
  const [authOpen, setAuthOpen] = useState(false);

  // Open auth modal whenever a protected API call bounces with 401.
  useEffect(() => {
    const handler = () => setAuthOpen(true);
    window.addEventListener("auth:required", handler);
    return () => window.removeEventListener("auth:required", handler);
  }, []);

  const isAuthenticated = authState.status === "authenticated";
  const isLoading = authState.status === "loading";

  return (
    <>
      <nav className={styles.nav}>
        <div className={styles.navLinks}>
          <Link to="/">Home</Link>
          <Link to="/tasks">Tasks</Link>
        </div>

        <div className={styles.navTitle}>Groovey Task Manager</div>

        <div className={styles.navAuth}>
          {isLoading ? null : isAuthenticated ? (
            <>
              <span className={styles.navEmail}>
                {authState.user.email}
              </span>
              <button
                type="button"
                className={styles.navBtn}
                onClick={logout}
              >
                Sign Out
              </button>
            </>
          ) : (
            <button
              type="button"
              className={styles.navBtn}
              onClick={() => setAuthOpen(true)}
            >
              Sign In / Sign Up
            </button>
          )}
        </div>
      </nav>

      {/* Protect the main content: show a prompt when not logged in */}
      {!isLoading && !isAuthenticated ? (
        <div className={styles.guestWall}>
          <p>Sign in or create an account to manage your tasks.</p>
          <button
            type="button"
            className={styles.wallBtn}
            onClick={() => setAuthOpen(true)}
          >
            Sign In / Sign Up
          </button>
        </div>
      ) : (
        <AppRoutes />
      )}

      <AuthModal
        open={authOpen}
        onClose={() => setAuthOpen(false)}
      />
    </>
  );
}

// ─── Root ─────────────────────────────────────────────────────────────────────

const App = () => (
  <Router>
    <AuthProvider>
      <AppShell />
    </AuthProvider>
  </Router>
);

export default App;
