import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth, useRestoreSession } from './hooks/useAuth'
import { ProtectedRoute } from './components/ProtectedRoute'
import { LoginPage } from './pages/LoginPage'
import { ChatPage } from './pages/ChatPage'

export default function App() {
  const { auth } = useAuth()
  const { loading } = useRestoreSession()

  // Wait for the /me check to complete before rendering routes so we don't
  // flash the login page for users who already have a valid session.
  if (loading) return null

  return (
    <Routes>
      <Route path="/" element={<Navigate to={auth ? '/chat' : '/login'} replace />} />
      <Route path="/login" element={<LoginPage />} />
      <Route
        path="/chat"
        element={
          <ProtectedRoute>
            <ChatPage />
          </ProtectedRoute>
        }
      />
    </Routes>
  )
}
