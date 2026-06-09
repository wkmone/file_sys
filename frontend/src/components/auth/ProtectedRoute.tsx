import { useEffect, useState } from 'react'
import { Navigate } from 'react-router-dom'
import { useAuthStore } from '../../store/authStore'
import { authApi } from '../../api/authApi'
import { Spin } from 'antd'

export default function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const [loading, setLoading] = useState(!isAuthenticated)

  useEffect(() => {
    // Already authenticated, skip
    if (isAuthenticated) {
      setLoading(false)
      return
    }

    // Try to restore session via stored token
    if (useAuthStore.getState().accessToken) {
      authApi.me()
        .then((res) => {
          useAuthStore.getState().setUser(res.data.data)
          useAuthStore.setState({ isAuthenticated: true })
        })
        .catch(() => {
          useAuthStore.getState().logout()
        })
        .finally(() => setLoading(false))
    } else {
      // No token at all, just stop loading
      setLoading(false)
    }
  }, [isAuthenticated])

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!useAuthStore.getState().isAuthenticated) {
    const redirect = encodeURIComponent(window.location.pathname + window.location.search)
    return <Navigate to={`/login?redirect=${redirect}`} replace />
  }

  return <>{children}</>
}
