import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { Typography, Button, List, Modal, Input, message, Popconfirm, Tag, Tabs, Spin, Badge } from 'antd'
import {
  PlusOutlined, TeamOutlined, DeleteOutlined, EyeOutlined,
  CrownOutlined, LoginOutlined, FolderOutlined, SearchOutlined,
  CheckOutlined, CloseOutlined, ClockCircleOutlined,
} from '@ant-design/icons'
import { teamApi } from '../api/teamApi'
import { useAuthStore } from '../store/authStore'
import { useTeamRefresh } from '../store/teamStore'
import type { Team, TeamMember, JoinRequest } from '../types/team'
import { colors } from '../theme'

const roleColors: Record<string, string> = { owner: 'gold', admin: 'blue', member: 'green' }
const roleLabels: Record<string, string> = { owner: '拥有者', admin: '管理员', member: '成员' }

export default function TeamManagementPage() {
  const navigate = useNavigate()
  const user = useAuthStore((s) => s.user)
  const bumpTeams = useTeamRefresh((s) => s.bump)
  const [teams, setTeams] = useState<Team[]>([])
  const [loading, setLoading] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [newTeamName, setNewTeamName] = useState('')
  const [newTeamDesc, setNewTeamDesc] = useState('')
  const [creating, setCreating] = useState(false)

  // Discover teams
  const [discoverOpen, setDiscoverOpen] = useState(false)
  const [allTeams, setAllTeams] = useState<Team[]>([])
  const [discovering, setDiscovering] = useState(false)
  const [pendingMap, setPendingMap] = useState<Record<string, boolean>>({})
  const [joiningMap, setJoiningMap] = useState<Record<string, boolean>>({})

  // Members
  const [memberModalOpen, setMemberModalOpen] = useState(false)
  const [selectedTeam, setSelectedTeam] = useState<Team | null>(null)
  const [members, setMembers] = useState<TeamMember[]>([])

  // Join requests (admin)
  const [requestsModalOpen, setRequestsModalOpen] = useState(false)
  const [joinRequests, setJoinRequests] = useState<JoinRequest[]>([])
  const [requestsLoading, setRequestsLoading] = useState(false)

  const fetchTeams = useCallback(async () => {
    setLoading(true)
    try {
      const res = await teamApi.list()
      setTeams(res.data.data || [])
    } catch {
      message.error('获取团队列表失败')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => { fetchTeams() }, [fetchTeams])

  const myTeamIds = new Set(teams.map((t) => t.id))

  const handleCreate = async () => {
    if (!newTeamName.trim()) return
    setCreating(true)
    try {
      await teamApi.create({ name: newTeamName, description: newTeamDesc })
      message.success('团队创建成功')
      bumpTeams()
      setCreateOpen(false)
      setNewTeamName('')
      setNewTeamDesc('')
      fetchTeams()
    } catch {
      message.error('创建失败')
    } finally {
      setCreating(false)
    }
  }

  const handleDelete = async (id: string) => {
    try {
      await teamApi.delete(id)
      message.success('团队已删除')
      bumpTeams()
      fetchTeams()
    } catch {
      message.error('删除失败')
    }
  }

  const openDiscover = async () => {
    setDiscoverOpen(true)
    setDiscovering(true)
    try {
      const res = await teamApi.discover()
      const all: Team[] = res.data.data || []
      setAllTeams(all)
      // Check pending requests for non-member teams
      const pm: Record<string, boolean> = {}
      await Promise.all(
        all.filter((t) => !myTeamIds.has(t.id)).map(async (t) => {
          try {
            const r = await teamApi.getPendingRequest(t.id)
            if (r.data.data) pm[t.id] = true
          } catch { /* ignore */ }
        })
      )
      setPendingMap(pm)
    } catch {
      message.error('获取团队列表失败')
    } finally {
      setDiscovering(false)
    }
  }

  const handleRequestJoin = async (teamId: string) => {
    setJoiningMap((prev) => ({ ...prev, [teamId]: true }))
    try {
      await teamApi.requestJoin(teamId)
      message.success('已发送加入申请，等待管理员审批')
      setPendingMap((prev) => ({ ...prev, [teamId]: true }))
    } catch (err: any) {
      message.error(err.response?.data?.message || '申请失败')
    } finally {
      setJoiningMap((prev) => ({ ...prev, [teamId]: false }))
    }
  }

  const openMembers = async (team: Team) => {
    setSelectedTeam(team)
    setMemberModalOpen(true)
    try {
      const res = await teamApi.members(team.id)
      setMembers(res.data.data || [])
    } catch {
      message.error('获取成员列表失败')
    }
  }

  const handleRemoveMember = async (userId: string) => {
    if (!selectedTeam) return
    try {
      await teamApi.removeMember(selectedTeam.id, userId)
      setMembers((prev) => prev.filter((m) => m.user_id !== userId))
      message.success('成员已移除')
    } catch {
      message.error('移除失败')
    }
  }

  const openJoinRequests = async (team: Team) => {
    setSelectedTeam(team)
    setRequestsModalOpen(true)
    setRequestsLoading(true)
    try {
      const res = await teamApi.listJoinRequests(team.id)
      setJoinRequests(res.data.data || [])
    } catch {
      message.error('获取申请列表失败')
    } finally {
      setRequestsLoading(false)
    }
  }

  const handleApprove = async (requestId: string) => {
    if (!selectedTeam) return
    try {
      await teamApi.handleJoinRequest(selectedTeam.id, requestId, 'approved')
      message.success('已批准加入')
      setJoinRequests((prev) => prev.filter((r) => r.id !== requestId))
      bumpTeams()
      fetchTeams()
    } catch {
      message.error('操作失败')
    }
  }

  const handleReject = async (requestId: string) => {
    if (!selectedTeam) return
    try {
      await teamApi.handleJoinRequest(selectedTeam.id, requestId, 'rejected')
      message.success('已拒绝')
      setJoinRequests((prev) => prev.filter((r) => r.id !== requestId))
    } catch {
      message.error('操作失败')
    }
  }

  const ownedTeams = teams.filter((t) => t.owner_id === user?.id)
  const joinedTeams = teams.filter((t) => t.owner_id !== user?.id)
  const adminTeams = teams.filter((t) => {
    // Teams where I'm owner or admin — used for join requests tab
    return t.owner_id === user?.id
  })

  const renderTeamCard = (team: Team, showRequestBadge = false) => (
    <div
      style={{
        border: `1px solid ${colors.border}`,
        borderRadius: 12,
        padding: 20,
        background: colors.white,
        transition: 'all 0.2s',
        position: 'relative',
      }}
      onMouseEnter={(e) => {
        e.currentTarget.style.borderColor = colors.primary
        e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,82,217,0.1)'
      }}
      onMouseLeave={(e) => {
        e.currentTarget.style.borderColor = colors.border
        e.currentTarget.style.boxShadow = 'none'
      }}
    >
      {showRequestBadge && pendingMap[team.id] && (
        <Badge
          count={<ClockCircleOutlined style={{ color: '#faad14', fontSize: 14 }} />}
          offset={[-6, 6]}
          style={{ zIndex: 1 }}
        >
          <span />
        </Badge>
      )}
      <div style={{
        width: 48, height: 48, borderRadius: 12,
        background: colors.primaryLight,
        display: 'flex', alignItems: 'center', justifyContent: 'center',
        marginBottom: 14,
      }}>
        <TeamOutlined style={{ fontSize: 24, color: colors.primary }} />
      </div>
      <div style={{ marginBottom: 12 }}>
        <div style={{ fontWeight: 600, fontSize: 15, color: colors.textPrimary, marginBottom: 4 }}>
          {team.name}
          {team.owner_id === user?.id && <CrownOutlined style={{ color: '#faad14', marginLeft: 6, fontSize: 13 }} />}
        </div>
        <div style={{ fontSize: 13, color: colors.textTertiary, minHeight: 18 }}>
          {team.description || '无描述'}
        </div>
      </div>
      <div style={{ display: 'flex', gap: 8, borderTop: `1px solid ${colors.border}`, paddingTop: 12 }}>
        <Button type="text" size="small" icon={<EyeOutlined />}
          onClick={() => navigate(`/teams/${team.id}`)}
          style={{ color: colors.primary }}>
          详情
        </Button>
        <Button type="text" size="small" icon={<FolderOutlined />}
          onClick={() => navigate(`/teams/${team.id}/files`)}>
          文件
        </Button>
        <Button type="text" size="small" icon={<TeamOutlined />}
          onClick={(e) => { e.stopPropagation(); openMembers(team) }}>
          成员
        </Button>
        {team.owner_id === user?.id && (
          <Popconfirm title="确定删除此团队？" onConfirm={(e) => { e?.stopPropagation(); handleDelete(team.id) }}>
            <Button type="text" size="small" danger icon={<DeleteOutlined />} onClick={(e) => e.stopPropagation()}>
              删除
            </Button>
          </Popconfirm>
        )}
      </div>
    </div>
  )

  return (
    <div>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
        <div>
          <Typography.Title level={3} style={{ margin: 0, color: colors.textPrimary }}>团队管理</Typography.Title>
          <Typography.Text style={{ color: colors.textSecondary, fontSize: 14 }}>
            管理你的团队或发现并加入新的团队
          </Typography.Text>
        </div>
        <div style={{ display: 'flex', gap: 12 }}>
          <Button icon={<SearchOutlined />} onClick={openDiscover}>发现团队</Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>创建团队</Button>
        </div>
      </div>

      <Tabs
        items={[
          {
            key: 'all',
            label: `全部团队 (${teams.length})`,
            children: teams.length === 0 ? (
              <div style={{ padding: 64, textAlign: 'center', border: `1px dashed ${colors.border}`, borderRadius: 12 }}>
                <TeamOutlined style={{ fontSize: 48, color: colors.textTertiary, marginBottom: 16 }} />
                <div style={{ fontSize: 16, color: colors.textSecondary, marginBottom: 8 }}>还没有加入任何团队</div>
                <div style={{ fontSize: 13, color: colors.textTertiary, marginBottom: 16 }}>创建一个新团队或点击"发现团队"加入已有团队</div>
                <div style={{ display: 'flex', gap: 12, justifyContent: 'center' }}>
                  <Button icon={<SearchOutlined />} onClick={openDiscover}>发现团队</Button>
                  <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>创建团队</Button>
                </div>
              </div>
            ) : (
              <List
                loading={loading}
                grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 3, xl: 4 }}
                dataSource={teams}
                renderItem={(team) => <List.Item>{renderTeamCard(team)}</List.Item>}
              />
            ),
          },
          {
            key: 'owned',
            label: `我创建的 (${ownedTeams.length})`,
            children: ownedTeams.length === 0 ? (
              <div style={{ padding: 48, textAlign: 'center', color: colors.textSecondary }}>
                你还没有创建任何团队
              </div>
            ) : (
              <List
                loading={loading}
                grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 3, xl: 4 }}
                dataSource={ownedTeams}
                renderItem={(team) => <List.Item>{renderTeamCard(team)}</List.Item>}
              />
            ),
          },
          {
            key: 'joined',
            label: `我加入的 (${joinedTeams.length})`,
            children: joinedTeams.length === 0 ? (
              <div style={{ padding: 48, textAlign: 'center', color: colors.textSecondary }}>
                你还没有加入其他团队，点击上方"发现团队"按钮浏览可加入的团队
              </div>
            ) : (
              <List
                loading={loading}
                grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 3, xl: 4 }}
                dataSource={joinedTeams}
                renderItem={(team) => <List.Item>{renderTeamCard(team)}</List.Item>}
              />
            ),
          },
          ...(adminTeams.length > 0 ? [{
            key: 'requests',
            label: `入团审批`,
            children: (
              <List
                loading={loading}
                grid={{ gutter: 16, xs: 1, sm: 2, md: 3, lg: 3, xl: 4 }}
                dataSource={adminTeams}
                renderItem={(team) => (
                  <List.Item>
                    <div
                      style={{
                        border: `1px solid ${colors.border}`,
                        borderRadius: 12,
                        padding: 20,
                        background: colors.white,
                        cursor: 'pointer',
                        transition: 'all 0.2s',
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.borderColor = colors.primary
                        e.currentTarget.style.boxShadow = '0 4px 16px rgba(0,82,217,0.1)'
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.borderColor = colors.border
                        e.currentTarget.style.boxShadow = 'none'
                      }}
                      onClick={() => openJoinRequests(team)}
                    >
                      <div style={{
                        width: 48, height: 48, borderRadius: 12,
                        background: colors.primaryLight,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        marginBottom: 14,
                      }}>
                        <TeamOutlined style={{ fontSize: 24, color: colors.primary }} />
                      </div>
                      <div style={{ fontWeight: 600, fontSize: 15, color: colors.textPrimary }}>
                        {team.name}
                      </div>
                      <div style={{ fontSize: 13, color: colors.textSecondary, marginTop: 4 }}>
                        点击查看入团申请
                      </div>
                    </div>
                  </List.Item>
                )}
              />
            ),
          }] : []),
        ]}
      />

      {/* Create Team Modal */}
      <Modal
        title="创建团队"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={handleCreate}
        confirmLoading={creating}
        okText="创建"
      >
        <div style={{ marginBottom: 4 }}>
          <Typography.Text style={{ fontSize: 13, color: colors.textSecondary }}>团队名称</Typography.Text>
        </div>
        <Input
          placeholder="输入团队名称"
          value={newTeamName}
          onChange={(e) => setNewTeamName(e.target.value)}
          style={{ marginBottom: 16 }}
        />
        <div style={{ marginBottom: 4 }}>
          <Typography.Text style={{ fontSize: 13, color: colors.textSecondary }}>团队描述（可选）</Typography.Text>
        </div>
        <Input.TextArea
          placeholder="输入团队描述"
          value={newTeamDesc}
          onChange={(e) => setNewTeamDesc(e.target.value)}
          rows={3}
        />
      </Modal>

      {/* Discover Teams Modal */}
      <Modal
        title="发现团队"
        open={discoverOpen}
        onCancel={() => setDiscoverOpen(false)}
        footer={null}
        width={720}
      >
        <Spin spinning={discovering}>
          {allTeams.length === 0 && !discovering ? (
            <div style={{ padding: 48, textAlign: 'center', color: colors.textSecondary }}>
              暂无可加入的团队
            </div>
          ) : (
            <List
              dataSource={allTeams}
              renderItem={(team) => {
                const isMember = myTeamIds.has(team.id)
                const isPending = pendingMap[team.id]
                const isJoining = joiningMap[team.id]
                return (
                  <List.Item
                    actions={[
                      isMember ? (
                        <Tag color="green">已加入</Tag>
                      ) : isPending ? (
                        <Tag icon={<ClockCircleOutlined />} color="orange">等待审批</Tag>
                      ) : (
                        <Button
                          type="primary"
                          size="small"
                          icon={<LoginOutlined />}
                          loading={isJoining}
                          onClick={() => handleRequestJoin(team.id)}
                        >
                          申请加入
                        </Button>
                      ),
                    ]}
                  >
                    <List.Item.Meta
                      avatar={
                        <div style={{
                          width: 40, height: 40, borderRadius: 10,
                          background: colors.primaryLight,
                          display: 'flex', alignItems: 'center', justifyContent: 'center',
                        }}>
                          <TeamOutlined style={{ fontSize: 20, color: colors.primary }} />
                        </div>
                      }
                      title={
                        <span style={{ color: colors.textPrimary }}>
                          {team.name}
                          {team.owner_id === user?.id && <CrownOutlined style={{ color: '#faad14', marginLeft: 6 }} />}
                        </span>
                      }
                      description={
                        <span style={{ color: colors.textTertiary }}>
                          {team.description || '无描述'}
                        </span>
                      }
                    />
                  </List.Item>
                )
              }}
            />
          )}
        </Spin>
      </Modal>

      {/* Members Modal */}
      <Modal
        title={`${selectedTeam?.name} - 成员管理`}
        open={memberModalOpen}
        onCancel={() => setMemberModalOpen(false)}
        footer={null}
        width={560}
      >
        <List
          dataSource={members}
          locale={{ emptyText: '暂无成员' }}
          renderItem={(m) => (
            <List.Item
              actions={[
                <Tag key="role" color={roleColors[m.role]}>
                  {m.role === 'owner' && <CrownOutlined style={{ marginRight: 4 }} />}
                  {roleLabels[m.role]}
                </Tag>,
                m.role !== 'owner' && selectedTeam?.owner_id === user?.id && (
                  <Popconfirm key="remove" title="移除该成员？" onConfirm={() => handleRemoveMember(m.user_id)}>
                    <Button size="small" danger>移除</Button>
                  </Popconfirm>
                ),
              ]}
            >
              <List.Item.Meta
                title={<span style={{ color: colors.textPrimary }}>{m.display_name}</span>}
                description={`${m.email} · 加入: ${new Date(m.joined_at).toLocaleDateString()}`}
              />
            </List.Item>
          )}
        />
      </Modal>

      {/* Join Requests Modal */}
      <Modal
        title={`${selectedTeam?.name} - 入团申请`}
        open={requestsModalOpen}
        onCancel={() => setRequestsModalOpen(false)}
        footer={null}
        width={560}
      >
        <Spin spinning={requestsLoading}>
          {joinRequests.length === 0 && !requestsLoading ? (
            <div style={{ padding: 32, textAlign: 'center', color: colors.textSecondary }}>
              暂无待审批的申请
            </div>
          ) : (
            <List
              dataSource={joinRequests}
              renderItem={(r) => (
                <List.Item
                  actions={[
                    r.status === 'pending' ? (
                      <div style={{ display: 'flex', gap: 8 }}>
                        <Button
                          size="small"
                          type="primary"
                          icon={<CheckOutlined />}
                          onClick={() => handleApprove(r.id)}
                        >
                          批准
                        </Button>
                        <Button
                          size="small"
                          danger
                          icon={<CloseOutlined />}
                          onClick={() => handleReject(r.id)}
                        >
                          拒绝
                        </Button>
                      </div>
                    ) : (
                      <Tag color={r.status === 'approved' ? 'green' : 'red'}>
                        {r.status === 'approved' ? '已批准' : '已拒绝'}
                      </Tag>
                    ),
                  ]}
                >
                  <List.Item.Meta
                    avatar={
                      <div style={{
                        width: 36, height: 36, borderRadius: '50%',
                        background: colors.primaryLight,
                        display: 'flex', alignItems: 'center', justifyContent: 'center',
                        color: colors.primary, fontWeight: 600, fontSize: 14,
                      }}>
                        {(r.display_name || '?')[0]}
                      </div>
                    }
                    title={<span style={{ color: colors.textPrimary }}>{r.display_name}</span>}
                    description={`${r.email} · ${new Date(r.created_at).toLocaleString()}`}
                  />
                </List.Item>
              )}
            />
          )}
        </Spin>
      </Modal>
    </div>
  )
}
