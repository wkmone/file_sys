import { Routes, Route, Navigate } from 'react-router-dom'
import { Toaster } from 'react-hot-toast'
import LoginPage from './pages/LoginPage'
import RegisterPage from './pages/RegisterPage'
import DashboardPage from './pages/DashboardPage'
import FileBrowserPage from './pages/FileBrowserPage'
import EditorPage from './pages/EditorPage'
import TeamManagementPage from './pages/TeamManagementPage'
import TeamDetailPage from './pages/TeamDetailPage'
import TrashPage from './pages/TrashPage'
import ProfilePage from './pages/ProfilePage'
import ProtectedRoute from './components/auth/ProtectedRoute'
import AppLayout from './components/layout/AppLayout'

export default function App() {
  return (
    <>
      <Toaster position="top-center" />
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />

        <Route
          element={
            <ProtectedRoute>
              <AppLayout />
            </ProtectedRoute>
          }
        >
          {/* Home */}
          <Route path="/" element={<DashboardPage />} />

          {/* Personal space */}
          <Route path="/my/files" element={<FileBrowserPage />} />
          <Route path="/my/files/:folderId" element={<FileBrowserPage />} />
          <Route path="/my/trash" element={<TrashPage />} />

          {/* Team management */}
          <Route path="/teams" element={<TeamManagementPage />} />
          <Route path="/teams/:id" element={<TeamDetailPage />} />

          {/* Team workspace */}
          <Route path="/teams/:teamId/files" element={<FileBrowserPage />} />
          <Route path="/teams/:teamId/files/:folderId" element={<FileBrowserPage />} />
          <Route path="/teams/:teamId/trash" element={<TrashPage />} />

          {/* Profile */}
          <Route path="/profile" element={<ProfilePage />} />
        </Route>

        {/* Editor — standalone page, no layout */}
        <Route
          path="/editor/:fileId"
          element={
            <ProtectedRoute>
              <EditorPage />
            </ProtectedRoute>
          }
        />

        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </>
  )
}
