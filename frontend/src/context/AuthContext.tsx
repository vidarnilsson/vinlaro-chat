import { createContext, useState, useCallback, type ReactNode } from 'react'

interface AuthState {
  user_id: string
  username: string
}

interface AuthContextValue {
  auth: AuthState | null
  signIn: (state: AuthState) => void
  signOut: () => void
}

export const AuthContext = createContext<AuthContextValue>({
  auth: null,
  signIn: () => {},
  signOut: () => {},
})

export function AuthProvider({ children }: { children: ReactNode }) {
  const [auth, setAuth] = useState<AuthState | null>(null)

  const signIn = useCallback((state: AuthState) => {
    setAuth(state)
  }, [])

  const signOut = useCallback(() => {
    setAuth(null)
  }, [])

  return (
    <AuthContext.Provider value={{ auth, signIn, signOut }}>
      {children}
    </AuthContext.Provider>
  )
}
