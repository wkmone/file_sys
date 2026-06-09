import { useState, useEffect, useCallback, useMemo } from 'react'
import { useParams, useLocation, useNavigate } from 'react-router-dom'
import { Typography, Table, Button, message, Popconfirm, Tag, Space } from 'antd'
import { DeleteOutlined, UndoOutlined } from '@ant-design/icons'
import type { ColumnsType } from 'antd/es/table'
import apiClient from '../api/client'
import { teamApi } from '../api/teamApi'
import FileIcon from '../components/common/FileIcon'
import { colors } from '../theme'

interface TrashItem {
  id: string
  name: string
  type: 'file' | 'folder'
  deleted_at: string
  file_size?: number
  file_ext?: string
}

function formatSize(bytes: number | undefined): string {
  if (!bytes) return '-'
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / (1024 * 1024)).toFixed(1) + ' MB'
}

export default function TrashPage() {
  const { teamId } = useParams<{ teamId?: string }>()
  const location = useLocation()
  const navigate = useNavigate()

  const scope = useMemo<'personal' | 'team'>(() => {
    return location.pathname.startsWith('/teams/') ? 'team' : 'personal'
  }, [location.pathname])

  const [items, setItems] = useState<TrashItem[]>([])
  const [loading, setLoading] = useState(false)
  const [teamName, setTeamName] = useState('')

  useEffect(() => {
    if (scope === 'team' && teamId) {
      teamApi.get(teamId).then((res) => setTeamName(res.data.data?.name || ''))
        .catch(() => {})
    }
  }, [scope, teamId])

  const fetchTrash = useCallback(async () => {
    setLoading(true)
    try {
      const params: any = {}
      if (scope === 'team' && teamId) params.team_id = teamId
      const res = await apiClient.get('/trash', { params })
      const data = res.data?.data
      setItems(Array.isArray(data?.items) ? data.items : Array.isArray(data) ? data : [])
    } catch {
      message.error('获取回收站数据失败')
    } finally {
      setLoading(false)
    }
  }, [scope, teamId])

  useEffect(() => { fetchTrash() }, [fetchTrash])

  const handleRestore = async (item: TrashItem) => {
    try {
      await apiClient.post(`/trash/${item.type}/${item.id}/restore`)
      message.success(`已恢复${item.type === 'file' ? '文件' : '文件夹'}`)
      fetchTrash()
    } catch {
      message.error('恢复失败')
    }
  }

  const handlePermanentDelete = async (item: TrashItem) => {
    try {
      await apiClient.delete(`/trash/${item.type}/${item.id}`)
      message.success('已永久删除')
      fetchTrash()
    } catch {
      message.error('删除失败')
    }
  }

  const columns: ColumnsType<TrashItem> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (name: string, record) => (
        <Space>
          {record.type === 'folder'
            ? <FileIcon type="folder" size={22} />
            : <FileIcon type={record.file_ext || ''} size={22} />
          }
          <span style={{ color: colors.textPrimary }}>{name}</span>
        </Space>
      ),
    },
    {
      title: '类型',
      key: 'type',
      width: 100,
      render: (_, record) => (
        <Tag color={record.type === 'folder' ? 'orange' : 'blue'}>
          {record.type === 'folder' ? '文件夹' : record.file_ext?.replace('.', '').toUpperCase() + ' 文件' || '文件'}
        </Tag>
      ),
    },
    {
      title: '大小',
      key: 'size',
      width: 90,
      render: (_, record) => formatSize(record.file_size),
    },
    {
      title: '删除时间',
      dataIndex: 'deleted_at',
      key: 'deleted_at',
      width: 180,
      render: (val: string) => {
        const d = new Date(val)
        return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
      },
    },
    {
      title: '',
      key: 'actions',
      width: 160,
      render: (_, record) => (
        <Space>
          <Button
            type="link" size="small" icon={<UndoOutlined />}
            onClick={() => handleRestore(record)}
            style={{ color: colors.primary }}
          >
            恢复
          </Button>
          <Popconfirm
            title="永久删除后无法恢复，确定删除？"
            onConfirm={() => handlePermanentDelete(record)}
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              彻底删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ]

  const title = scope === 'team' && teamId
    ? `${teamName || '团队'} — 回收站`
    : '个人回收站'
  const subtitle = scope === 'team'
    ? '团队中已删除的文件和文件夹'
    : '你已删除的文件和文件夹'

  return (
    <div>
      <div style={{ marginBottom: 24 }}>
        <Typography.Title level={3} style={{ margin: 0, color: colors.textPrimary }}>{title}</Typography.Title>
        <Typography.Text style={{ color: colors.textSecondary, fontSize: 14 }}>{subtitle}</Typography.Text>
      </div>

      {items.length === 0 && !loading ? (
        <div style={{
          padding: 64,
          textAlign: 'center',
          border: `1px dashed ${colors.border}`,
          borderRadius: 12,
        }}>
          <DeleteOutlined style={{ fontSize: 48, color: colors.textTertiary, marginBottom: 16 }} />
          <div style={{ fontSize: 16, color: colors.textSecondary }}>回收站为空</div>
        </div>
      ) : (
        <Table
          columns={columns}
          dataSource={items.map((item) => ({ ...item, key: item.id }))}
          loading={loading}
          pagination={false}
          size="middle"
          showSorterTooltip={false}
          locale={{ emptyText: '回收站为空' }}
        />
      )}
    </div>
  )
}
