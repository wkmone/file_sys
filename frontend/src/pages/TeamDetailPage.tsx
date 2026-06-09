import { useState, useEffect, useCallback } from 'react'
import { useParams, useNavigate } from 'react-router-dom'
import { Typography, Button, List, Modal, Select, message, Popconfirm, Tag, Space, Spin, Input } from 'antd'
import { ArrowLeftOutlined, UserAddOutlined, DeleteOutlined, CrownOutlined, TeamOutlined, FolderOutlined } from '@ant-design/icons'
import { teamApi } from '../api/teamApi'
import { authApi } from '../api/authApi'
import { useTeamRefresh } from '../store/teamStore'
import type { Team, TeamMember } from '../types/team'
import type { User } from '../types/user'
import { colors } from '../theme'

const roleColors: Record<string, string> = { owner: 'gold', admin: 'blue', member: 'green' }
const roleLabels: Record<string, string> = { owner: '拥有者', admin: '管理员', member: '成员' }

export default function TeamDetailPage() {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const bumpTeams = useTeamRefresh((s) => s.bump)

  const [team, setTeam] = useState<Team | null>(null)
  const [members, setMembers] = useState<TeamMember[]>([])
  const [loading, setLoading] = useState(false)
  const [addOpen, setAddOpen] = useState(false)
  const [users, setUsers] = useState<User[]>([])
  const [searching, setSearching] = useState(false)
  const [selectedUserId, setSelectedUserId] = useState<string>()
  const [selectedRole, setSelectedRole] = useState('member')
  const [adding, setAdding] = useState(false)
  const [editOpen, setEditOpen] = useState(false)
  const [editName, setEditName] = useState('')
  const [editDesc, setEditDesc] = useState('')
  const [saving, setSaving] = useState(false)

  const fetchTeam = useCallback(async () => {
    if (!id) return
    setLoading(true)
    try {
      const [teamRes, membersRes] = await Promise.all([
        teamApi.get(id),
        teamApi.members(id),
      ])
      setTeam(teamRes.data.data)
      setMembers(membersRes.data.data || [])
    } catch {
      message.error('获取团队信息失败')
      navigate('/teams')
    } finally {
      setLoading(false)
    }
  }, [id, navigate])

  useEffect(() => { fetchTeam() }, [fetchTeam])

  const handleSearchUsers = async (q: string) => {
    if (!q) { setUsers([]); return }
    setSearching(true)
    try {
      const res = await authApi.listUsers({ page_size: 50 })
      const allUsers = res.data.data?.items || []
      setUsers(allUsers.filter((u) =>
        u.email.includes(q) || u.display_name.includes(q)
      ))
    } catch {
      message.error('搜索用户失败')
    } finally {
      setSearching(false)
    }
  }

  const handleAddMember = async () => {
    if (!selectedUserId || !id) return
    setAdding(true)
    try {
      await teamApi.addMember(id, { user_id: selectedUserId, role: selectedRole })
      message.success('成员已添加')
      bumpTeams()
      setAddOpen(false)
      setSelectedUserId(undefined)
      setSelectedRole('member')
      fetchTeam()
    } catch {
      message.error('添加失败')
    } finally {
      setAdding(false)
    }
  }

  const handleRemoveMember = async (userId: string) => {
    if (!id) return
    try {
      await teamApi.removeMember(id, userId)
      bumpTeams()
      setMembers((prev) => prev.filter((m) => m.user_id !== userId))
      message.success('成员已移除')
    } catch {
      message.error('移除失败')
    }
  }

  const handleSaveEdit = async () => {
    if (!id || !editName.trim()) return
    setSaving(true)
    try {
      await teamApi.update(id, { name: editName, description: editDesc })
      message.success('团队信息已更新')
      bumpTeams()
      setEditOpen(false)
      fetchTeam()
    } catch {
      message.error('更新失败')
    } finally {
      setSaving(false)
    }
  }

  const openEditModal = () => {
    if (!team) return
    setEditName(team.name)
    setEditDesc(team.description || '')
    setEditOpen(true)
  }

  if (loading) {
    return (
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: 300 }}>
        <Spin size="large" />
      </div>
    )
  }

  if (!team) return null

  return (
    <div>
      <Button
        type="text"
        icon={<ArrowLeftOutlined />}
        onClick={() => navigate('/teams')}
        style={{ marginBottom: 20, padding: 0, color: colors.textSecondary, fontSize: 14 }}
      >
        返回团队列表
      </Button>

      {/* Team header */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start', marginBottom: 28 }}>
        <div style={{ display: 'flex', gap: 16, alignItems: 'flex-start' }}>
          <div style={{
            width: 56,
            height: 56,
            borderRadius: 14,
            background: colors.primaryLight,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            flexShrink: 0,
          }}>
            <TeamOutlined style={{ fontSize: 28, color: colors.primary }} />
          </div>
          <div>
            <Typography.Title level={3} style={{ margin: 0, color: colors.textPrimary }}>{team.name}</Typography.Title>
            <Typography.Text style={{ color: colors.textSecondary, fontSize: 14 }}>
              {team.description || '无描述'}
            </Typography.Text>
            <br />
            <Typography.Text style={{ color: colors.textTertiary, fontSize: 13 }}>
              创建于 {new Date(team.created_at).toLocaleDateString()}
            </Typography.Text>
          </div>
        </div>
        <Space>
          <Button type="primary" icon={<FolderOutlined />} onClick={() => navigate(`/teams/${team.id}/files`)}>
            团队文件
          </Button>
          <Button icon={<UserAddOutlined />} onClick={() => setAddOpen(true)}>
            添加成员
          </Button>
          <Button onClick={openEditModal}>编辑团队</Button>
        </Space>
      </div>

      {/* Members section */}
      <div style={{ marginBottom: 12 }}>
        <Typography.Text strong style={{ fontSize: 15, color: colors.textPrimary }}>
          团队成员 ({members.length})
        </Typography.Text>
      </div>

      <List
        dataSource={members}
        locale={{ emptyText: '暂无成员' }}
        renderItem={(m) => (
          <List.Item
            style={{ padding: '12px 0' }}
            actions={[
              <Tag key="role" color={roleColors[m.role]}>
                {m.role === 'owner' && <CrownOutlined style={{ marginRight: 4 }} />}
                {roleLabels[m.role]}
              </Tag>,
              m.role !== 'owner' && (
                <Popconfirm key="remove" title="确定移除该成员？" onConfirm={() => handleRemoveMember(m.user_id)}>
                  <Button size="small" danger icon={<DeleteOutlined />}>移除</Button>
                </Popconfirm>
              ),
            ]}
          >
            <List.Item.Meta
              avatar={
                <div style={{
                  width: 40,
                  height: 40,
                  borderRadius: 10,
                  background: colors.primaryLight,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: colors.primary,
                  fontWeight: 600,
                  fontSize: 16,
                }}>
                  {m.display_name.charAt(0)}
                </div>
              }
              title={<span style={{ color: colors.textPrimary }}>{m.display_name}</span>}
              description={`${m.email} · 加入: ${new Date(m.joined_at).toLocaleDateString()}`}
            />
          </List.Item>
        )}
      />

      {/* Add Member Modal */}
      <Modal
        title="添加成员"
        open={addOpen}
        onCancel={() => { setAddOpen(false); setUsers([]); setSelectedUserId(undefined) }}
        onOk={handleAddMember}
        confirmLoading={adding}
        okText="添加"
        okButtonProps={{ disabled: !selectedUserId }}
      >
        <div style={{ marginBottom: 4 }}>
          <Typography.Text style={{ fontSize: 13, color: colors.textSecondary }}>搜索用户</Typography.Text>
        </div>
        <Select
          showSearch
          placeholder="输入邮箱或昵称搜索..."
          filterOption={false}
          onSearch={handleSearchUsers}
          onChange={(val) => setSelectedUserId(val)}
          value={selectedUserId}
          style={{ width: '100%', marginBottom: 12 }}
          notFoundContent={searching ? <Spin size="small" /> : '无匹配用户'}
          options={users.map((u) => ({
            label: `${u.display_name} (${u.email})`,
            value: u.id,
          }))}
        />
        <div style={{ marginBottom: 4 }}>
          <Typography.Text style={{ fontSize: 13, color: colors.textSecondary }}>角色</Typography.Text>
        </div>
        <Select
          value={selectedRole}
          onChange={(val) => setSelectedRole(val)}
          style={{ width: '100%' }}
          options={[
            { label: '管理员', value: 'admin' },
            { label: '成员', value: 'member' },
          ]}
        />
      </Modal>

      {/* Edit Team Modal */}
      <Modal
        title="编辑团队"
        open={editOpen}
        onCancel={() => setEditOpen(false)}
        onOk={handleSaveEdit}
        confirmLoading={saving}
        okText="保存"
      >
        <div style={{ marginBottom: 4 }}>
          <Typography.Text style={{ fontSize: 13, color: colors.textSecondary }}>团队名称</Typography.Text>
        </div>
        <Input
          value={editName}
          onChange={(e) => setEditName(e.target.value)}
          style={{ marginBottom: 16 }}
        />
        <div style={{ marginBottom: 4 }}>
          <Typography.Text style={{ fontSize: 13, color: colors.textSecondary }}>团队描述</Typography.Text>
        </div>
        <Input.TextArea
          value={editDesc}
          onChange={(e) => setEditDesc(e.target.value)}
          rows={3}
        />
      </Modal>
    </div>
  )
}
