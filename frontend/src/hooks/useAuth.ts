import { useContext, useEffect, useState } from 'react'
import { AuthContext } from '../context/AuthContext'
import { getMe } from '../api/auth'

export function useAuth() {
  return useContext(AuthContext)
}

// useRestoreSession checks for an existing session cookie on app mount.
// Call once at the top level (main.tsx or App.tsx) before rendering routes.
export function useRestoreSession() {
  const { signIn } = useContext(AuthContext)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    getMe()
      .then(signIn)
      .catch(() => {}) // 401 means no session — stay logged out
      .finally(() => setLoading(false))
  }, []) // eslint-disable-line react-hooks/exhaustive-deps

  return { loading }
}
