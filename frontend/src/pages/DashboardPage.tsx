import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Typography, Row, Col, Spin } from 'antd'
import {
  FileTextOutlined,
  FileExcelOutlined,
  FilePptOutlined,
  FolderAddOutlined,
  ArrowRightOutlined,
  FileOutlined,
  TeamOutlined,
  PlusCircleOutlined,
} from '@ant-design/icons'
import { useAuthStore } from '../store/authStore'
import { useTeamRefresh } from '../store/teamStore'
import { fileApi } from '../api/fileApi'
import { teamApi } from '../api/teamApi'
import FileIcon from '../components/common/FileIcon'
import type { FileItem } from '../types/file'
import type { Team } from '../types/team'
import { colors, spacing } from '../theme'

function formatSize(bytes: number): string {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

function timeAgo(dateStr: string): string {
  const diff = Date.now() - new Date(dateStr).getTime()
  const mins = Math.floor(diff / 60000)
  if (mins < 1) return '刚刚'
  if (mins < 60) return `${mins}分钟前`
  const hours = Math.floor(mins / 60)
  if (hours < 24) return `${hours}小时前`
  const days = Math.floor(hours / 24)
  if (days < 30) return `${days}天前`
  return new Date(dateStr).toLocaleDateString()
}

const createItems = [
  { icon: <FileTextOutlined style={{ fontSize: 32, color: colors.primary }} />, label: '新建文档', desc: '在个人空间创建' },
  { icon: <FileExcelOutlined style={{ fontSize: 32, color: '#00a870' }} />, label: '新建表格', desc: '在个人空间创建' },
  { icon: <FilePptOutlined style={{ fontSize: 32, color: '#ed7b2f' }} />, label: '新建演示', desc: '在个人空间创建' },
  { icon: <FolderAddOutlined style={{ fontSize: 32, color: '#f5a623' }} />, label: '新建文件夹', desc: '在个人空间创建' },
]

export default function DashboardPage() {
  const user = useAuthStore((s) => s.user)
  const navigate = useNavigate()
  const teamsVersion = useTeamRefresh((s) => s.version)
  const [recentFiles, setRecentFiles] = useState<FileItem[]>([])
  const [teams, setTeams] = useState<Team[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fileApi.list({ page_size: 6 })
      .then((res) => setRecentFiles(res.data.data?.items || []))
      .catch(() => {})
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    teamApi.list()
      .then((res) => setTeams(res.data.data || []))
      .catch(() => {})
  }, [teamsVersion])

  return (
    <div>
      {/* Welcome header */}
      <div style={{ marginBottom: spacing.xxl }}>
        <Typography.Title level={3} style={{ margin: 0, color: colors.textPrimary }}>
          欢迎回来，{user?.display_name}
        </Typography.Title>
        <Typography.Text style={{ color: colors.textSecondary, fontSize: 15, marginTop: 4, display: 'block' }}>
          开始你的文档协作之旅
        </Typography.Text>
      </div>

      {/* Create New */}
      <Typography.Text strong style={{ fontSize: 15, color: colors.textPrimary, marginBottom: 12, display: 'block' }}>
        在个人空间新建
      </Typography.Text>
      <Row gutter={[16, 16]} style={{ marginBottom: spacing.xxl }}>
        {createItems.map((item) => (
          <Col key={item.label} xs={24} sm={12} lg={6}>
            <div
              onClick={() => navigate('/my/files')}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 14,
                padding: '16px 20px',
                border: `1px solid ${colors.border}`,
                borderRadius: 12,
                cursor: 'pointer',
                background: colors.white,
                transition: 'all 0.2s',
              }}
              onMouseEnter={(e) => {
                e.currentTarget.style.borderColor = colors.primary
                e.currentTarget.style.boxShadow = '0 2px 12px rgba(0,82,217,0.1)'
              }}
              onMouseLeave={(e) => {
                e.currentTarget.style.borderColor = colors.border
                e.currentTarget.style.boxShadow = 'none'
              }}
            >
              <div style={{
                width: 52,
                height: 52,
                borderRadius: 12,
                background: colors.bgPage,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                flexShrink: 0,
              }}>
                {item.icon}
              </div>
              <div>
                <div style={{ fontWeight: 500, fontSize: 15, color: colors.textPrimary }}>{item.label}</div>
                <div style={{ fontSize: 12, color: colors.textTertiary, marginTop: 2 }}>{item.desc}</div>
              </div>
            </div>
          </Col>
        ))}
      </Row>

      {/* My Teams */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
        <Typography.Text strong style={{ fontSize: 15, color: colors.textPrimary }}>
          我的团队
        </Typography.Text>
        <div style={{ display: 'flex', gap: 16, alignItems: 'center' }}>
          <a
            onClick={() => navigate('/teams')}
            style={{ fontSize: 13, color: colors.primary, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}
          >
            管理团队 <ArrowRightOutlined />
          </a>
        </div>
      </div>

      {teams.length === 0 ? (
        <div style={{
          padding: 32,
          textAlign: 'center',
          border: `1px dashed ${colors.border}`,
          borderRadius: 12,
          marginBottom: spacing.xxl,
        }}>
          <TeamOutlined style={{ fontSize: 36, color: colors.textTertiary, marginBottom: 12, display: 'block' }} />
          <div style={{ marginBottom: 4, color: colors.textSecondary, fontSize: 14 }}>还没有加入任何团队</div>
          <div style={{ marginBottom: 16, color: colors.textTertiary, fontSize: 12 }}>
            创建或加入一个团队，开始协作
          </div>
          <a onClick={() => navigate('/teams')} style={{ display: 'inline-flex', alignItems: 'center', gap: 4, fontSize: 14 }}>
            <PlusCircleOutlined /> 创建或加入团队
          </a>
        </div>
      ) : (
        <Row gutter={[12, 12]} style={{ marginBottom: spacing.xxl }}>
          {teams.map((team) => (
            <Col key={team.id} xs={24} sm={12} lg={8}>
              <div
                onClick={() => navigate(`/teams/${team.id}/files`)}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  padding: '14px 18px',
                  border: `1px solid ${colors.border}`,
                  borderRadius: 10,
                  cursor: 'pointer',
                  background: colors.white,
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = colors.primary
                  e.currentTarget.style.boxShadow = '0 2px 12px rgba(0,82,217,0.1)'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = colors.border
                  e.currentTarget.style.boxShadow = 'none'
                }}
              >
                <div style={{
                  width: 44,
                  height: 44,
                  borderRadius: 10,
                  background: colors.primaryLight,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  flexShrink: 0,
                }}>
                  <TeamOutlined style={{ fontSize: 22, color: colors.primary }} />
                </div>
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{ fontWeight: 500, fontSize: 14, color: colors.textPrimary }}>
                    {team.name}
                  </div>
                  <div style={{ fontSize: 12, color: colors.textTertiary, marginTop: 2 }}>
                    {team.description || '无描述'}
                  </div>
                </div>
                <ArrowRightOutlined style={{ color: colors.textTertiary, fontSize: 14 }} />
              </div>
            </Col>
          ))}
        </Row>
      )}

      {/* Recent Files — Personal */}
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
        <Typography.Text strong style={{ fontSize: 15, color: colors.textPrimary }}>
          个人空间 — 最近文件
        </Typography.Text>
        <a
          onClick={() => navigate('/my/files')}
          style={{ fontSize: 13, color: colors.primary, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: 4 }}
        >
          查看全部 <ArrowRightOutlined />
        </a>
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 40 }}><Spin /></div>
      ) : recentFiles.length === 0 ? (
        <div style={{
          padding: 48,
          textAlign: 'center',
          border: `1px dashed ${colors.border}`,
          borderRadius: 12,
          color: colors.textSecondary,
        }}>
          <FileOutlined style={{ fontSize: 40, color: colors.textTertiary, marginBottom: 12, display: 'block' }} />
          暂无文件，点击上方按钮开始创建
        </div>
      ) : (
        <Row gutter={[12, 12]}>
          {recentFiles.map((file) => (
            <Col key={file.id} xs={24} sm={12} lg={8}>
              <div
                onDoubleClick={() => {
                  const isOffice = ['.docx', '.xlsx', '.pptx'].includes(file.file_ext)
                  if (isOffice) navigate(`/editor/${file.id}`)
                  else window.open(fileApi.downloadUrl(file.id), '_blank')
                }}
                style={{
                  display: 'flex',
                  alignItems: 'center',
                  gap: 12,
                  padding: '12px 16px',
                  border: `1px solid ${colors.border}`,
                  borderRadius: 10,
                  cursor: 'pointer',
                  background: colors.white,
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  e.currentTarget.style.borderColor = colors.primaryLight
                  e.currentTarget.style.background = '#fafbfd'
                }}
                onMouseLeave={(e) => {
                  e.currentTarget.style.borderColor = colors.border
                  e.currentTarget.style.background = colors.white
                }}
              >
                <FileIcon type={file.file_ext} size={44} />
                <div style={{ flex: 1, minWidth: 0 }}>
                  <div style={{
                    fontWeight: 500,
                    fontSize: 14,
                    color: colors.textPrimary,
                    overflow: 'hidden',
                    textOverflow: 'ellipsis',
                    whiteSpace: 'nowrap',
                  }}>
                    {file.name}
                  </div>
                  <div style={{ fontSize: 12, color: colors.textTertiary, marginTop: 2 }}>
                    {formatSize(file.file_size)} · {timeAgo(file.updated_at)}
                  </div>
                </div>
              </div>
            </Col>
          ))}
        </Row>
      )}
    </div>
  )
}
