import { createContext, useState, useCallback, type ReactNode } from 'react'
import { setToken } from '../api/client'

interface AuthState {
  token: string
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
    setToken(state.token)
    setAuth(state)
  }, [])

  const signOut = useCallback(() => {
    setToken(null)
    setAuth(null)
  }, [])

  return (
    <AuthContext.Provider value={{ auth, signIn, signOut }}>
      {children}
    </AuthContext.Provider>
  )
}
