import { useState, useEffect } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { Layout, Menu, Button, Dropdown, Avatar, Spin } from 'antd'
import type { MenuProps } from 'antd'
import {
  DashboardOutlined,
  FolderOutlined,
  TeamOutlined,
  DeleteOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  FileTextOutlined,
  PlusCircleOutlined,
  SettingOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../../store/authStore'
import { useTeamRefresh } from '../../store/teamStore'
import { authApi } from '../../api/authApi'
import { teamApi } from '../../api/teamApi'
import FileIcon from '../common/FileIcon'
import type { Team } from '../../types/team'
import { colors, sidebar as sidebarConf } from '../../theme'

const { Header, Sider, Content } = Layout

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(false)
  const [teams, setTeams] = useState<Team[]>([])
  const [teamsLoading, setTeamsLoading] = useState(false)
  const [openKeys, setOpenKeys] = useState<string[]>([])
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()
  const teamsVersion = useTeamRefresh((s) => s.version)

  // Load user's teams for sidebar — refetches on team mutation events
  useEffect(() => {
    setTeamsLoading(true)
    teamApi.list()
      .then((res) => setTeams(res.data.data || []))
      .catch(() => {})
      .finally(() => setTeamsLoading(false))
  }, [teamsVersion])

  const getSelectedKey = () => {
    const path = location.pathname
    if (path === '/') return '/'
    if (path.startsWith('/my/files')) return '/my/files'
    if (path.startsWith('/my/trash')) return '/my/trash'
    if (path.startsWith('/teams/')) {
      // Check if it's a team workspace route
      const match = path.match(/^\/teams\/([^/]+)\/(files|trash)/)
      if (match) return `/team-${match[1]}`
      if (path === '/teams') return '/teams'
      return `/team-${path.split('/')[2]}`
    }
    if (path.startsWith('/profile')) return '/profile'
    return '/'
  }

  // Sync open keys from URL
  useEffect(() => {
    const keys: string[] = []
    const path = location.pathname
    if (path.startsWith('/my/')) keys.push('personal')
    if (path.startsWith('/teams/')) keys.push('team-space')
    setOpenKeys((prev) => {
      const merged = new Set([...prev, ...keys])
      return Array.from(merged)
    })
  }, [location.pathname])

  const handleLogout = async () => {
    try { await authApi.logout() } catch { /* ignore */ }
    logout()
    navigate('/login')
  }

  const userMenuItems = [
    { key: 'profile', icon: <UserOutlined />, label: '个人设置', onClick: () => navigate('/profile') },
    { type: 'divider' as const },
    { key: 'logout', icon: <LogoutOutlined />, label: '退出登录', onClick: handleLogout },
  ]

  // Build team submenu items
  const teamSubItems: MenuProps['items'] = [
    ...teams.map((team) => ({
      key: `/team-${team.id}`,
      icon: <TeamOutlined />,
      label: team.name,
      onClick: () => navigate(`/teams/${team.id}/files`),
    })),
    { type: 'divider' as const },
    {
      key: '/teams',
      icon: <SettingOutlined />,
      label: '管理团队',
      onClick: () => navigate('/teams'),
    },
  ]

  const menuItems: MenuProps['items'] = [
    {
      key: '/',
      icon: <DashboardOutlined />,
      label: '首页',
      onClick: () => navigate('/'),
    },
    { type: 'divider' as const },
    {
      key: 'personal',
      icon: <FolderOutlined />,
      label: '个人空间',
      children: [
        { key: '/my/files', icon: <FileTextOutlined />, label: '我的文件', onClick: () => navigate('/my/files') },
        { key: '/my/trash', icon: <DeleteOutlined />, label: '回收站', onClick: () => navigate('/my/trash') },
      ],
    },
    { type: 'divider' as const },
    {
      key: 'team-space',
      icon: <TeamOutlined />,
      label: '团队空间',
      children: teams.length === 0
        ? [{
            key: '/teams',
            icon: <PlusCircleOutlined />,
            label: teamsLoading ? '加载中...' : '创建或加入团队',
          }]
        : teamSubItems,
    },
    { type: 'divider' as const },
    {
      key: '/profile',
      icon: <UserOutlined />,
      label: '个人设置',
      onClick: () => navigate('/profile'),
    },
  ]

  return (
    <Layout style={{ minHeight: '100vh', background: colors.bgPage }}>
      <Sider
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={sidebarConf.width}
        collapsedWidth={sidebarConf.collapsedWidth}
        style={{ background: colors.white, borderRight: `1px solid ${colors.border}` }}
      >
        {/* Logo */}
        <div style={{
          height: 56,
          display: 'flex',
          alignItems: 'center',
          justifyContent: collapsed ? 'center' : 'flex-start',
          padding: collapsed ? 0 : '0 20px',
          borderBottom: `1px solid ${colors.border}`,
          cursor: 'pointer',
        }} onClick={() => navigate('/')}>
          <FileIcon type="logo" size={28} />
          {!collapsed && (
            <span style={{
              marginLeft: 10,
              fontSize: 16,
              fontWeight: 600,
              color: colors.textPrimary,
              whiteSpace: 'nowrap',
            }}>
              文件管理系统
            </span>
          )}
        </div>

        <Menu
          mode="inline"
          selectedKeys={[getSelectedKey()]}
          openKeys={openKeys}
          onOpenChange={setOpenKeys}
          items={menuItems}
          onClick={({ key }) => {
            if (key === 'personal' || key === 'team-space') return
            if (key === '/') { navigate('/'); return }
            if (key === '/my/files') navigate('/my/files')
            else if (key === '/my/trash') navigate('/my/trash')
            else if (key === '/teams') navigate('/teams')
            else if (key === '/profile') navigate('/profile')
            else if (key.startsWith('/team-')) {
              const teamId = key.replace('/team-', '')
              navigate(`/teams/${teamId}/files`)
            }
          }}
          style={{ background: 'transparent', borderRight: 'none', marginTop: 8 }}
        />
      </Sider>

      <Layout style={{ background: colors.bgPage }}>
        <Header style={{
          background: colors.white,
          padding: '0 24px',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          height: 56,
          lineHeight: '56px',
          borderBottom: `1px solid ${colors.border}`,
        }}>
          <Button
            type="text"
            icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
            onClick={() => setCollapsed(!collapsed)}
            style={{ fontSize: 16, width: 36, height: 36 }}
          />

          <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
            <div style={{
              cursor: 'pointer',
              display: 'flex',
              alignItems: 'center',
              gap: 8,
              padding: '4px 8px',
              borderRadius: 8,
            }}>
              <Avatar size={32} icon={<UserOutlined />} style={{ background: colors.primary }} />
              <span style={{ fontSize: 14, color: colors.textPrimary }}>{user?.display_name}</span>
            </div>
          </Dropdown>
        </Header>

        <Content style={{
          margin: 20,
          padding: 24,
          background: colors.white,
          borderRadius: 12,
          minHeight: 280,
          overflow: 'auto',
        }}>
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}
