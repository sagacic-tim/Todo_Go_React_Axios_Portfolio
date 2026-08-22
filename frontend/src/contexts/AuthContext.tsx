// src/contexts/AuthContext.tsx
import React, {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useState,
} from "react";
import { getMe, logout as apiLogout } from "../services/authService";

// ─── Types ────────────────────────────────────────────────────────────────────

export interface AuthUser {
  id: number;
  email: string;
  createdAt: string;
  updatedAt: string;
}

type AuthState =
  | { status: "loading" }
  | { status: "authenticated"; user: AuthUser }
  | { status: "unauthenticated" };

interface AuthContextValue {
  authState: AuthState;
  /** Call after a successful login/signup response to store the user. */
  setUser: (user: AuthUser) => void;
  /** Calls the logout API endpoint and clears the user. */
  logout: () => Promise<void>;
  /** True while we are waiting for the /me check on first load. */
  isLoading: boolean;
}

// ─── Context ─────────────────────────────────────────────────────────────────

const AuthContext = createContext<AuthContextValue | undefined>(undefined);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [authState, setAuthState] = useState<AuthState>({ status: "loading" });

  // Check for an existing session when the app first mounts.
  useEffect(() => {
    getMe()
      .then((user) => setAuthState({ status: "authenticated", user }))
      .catch(() => setAuthState({ status: "unauthenticated" }));
  }, []);

  const setUser = useCallback((user: AuthUser) => {
    setAuthState({ status: "authenticated", user });
  }, []);

  const logout = useCallback(async () => {
    await apiLogout().catch(() => {
      /* ignore network errors — just clear client state */
    });
    setAuthState({ status: "unauthenticated" });
  }, []);

  return (
    <AuthContext.Provider
      value={{
        authState,
        setUser,
        logout,
        isLoading: authState.status === "loading",
      }}
    >
      {children}
    </AuthContext.Provider>
  );
};

// ─── Hook ─────────────────────────────────────────────────────────────────────

export function useAuth(): AuthContextValue {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error("useAuth must be used inside <AuthProvider>");
  return ctx;
}
