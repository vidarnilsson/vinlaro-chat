import { Routes, Route, Navigate } from 'react-router-dom'
import { useAuth } from './hooks/useAuth'
import { ProtectedRoute } from './components/ProtectedRoute'
import { LoginPage } from './pages/LoginPage'
import { ChatPage } from './pages/ChatPage'

export default function App() {
  const { auth } = useAuth()

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
