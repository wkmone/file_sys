import { useEffect, useState, useCallback } from 'react'
import { Modal, Select, Segmented, Button, List, Tag, message, Spin } from 'antd'
import { DeleteOutlined, UserOutlined, TeamOutlined } from '@ant-design/icons'
import { fileApi, filePermissionApi } from '../../api/fileApi'
import { folderApi, folderPermissionApi } from '../../api/folderApi'
import { userApi } from '../../api/userApi'
import { permissionApi } from '../../api/permissionApi'
import type { FilePermission } from '../../types/file'
import { colors } from '../../theme'

interface ShareDialogProps {
  open: boolean
  resourceType: 'file' | 'folder'
  resourceId: string
  resourceName: string
  scope: 'personal' | 'team'
  onClose: () => void
}

export default function ShareDialog({ open, resourceType, resourceId, resourceName, scope, onClose }: ShareDialogProps) {
  const [permissions, setPermissions] = useState<FilePermission[]>([])
  const [loading, setLoading] = useState(false)
  const [searching, setSearching] = useState(false)
  const [searchResults, setSearchResults] = useState<{ id: string; label: string }[]>([])
  const [selectedUser, setSelectedUser] = useState<string | undefined>()
  const [permLevel, setPermLevel] = useState<string>('read')
  const [saving, setSaving] = useState(false)

  const fetchList = resourceType === 'file' ? filePermissionApi.list : folderPermissionApi.list
  const shareApi = resourceType === 'file' ? fileApi.share : folderApi.share

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const res = await fetchList(resourceId)
      setPermissions(res.data.data || [])
    } catch { /* ignore */ }
    finally { setLoading(false) }
  }, [resourceId, fetchList])

  useEffect(() => {
    if (open) load()
  }, [open, load])

  const handleSearch = async (query: string) => {
    if (!query || query.length < 1) { setSearchResults([]); return }
    setSearching(true)
    try {
      const res = await userApi.search(query)
      const users = (res.data.data || []).map((u) => ({ id: u.id, label: `${u.display_name} (${u.email})` }))
      setSearchResults(users)
    } catch { setSearchResults([]) }
    finally { setSearching(false) }
  }

  const handleAdd = async () => {
    if (!selectedUser || !permLevel) return
    setSaving(true)
    try {
      await shareApi(resourceId, { user_id: selectedUser, permission: permLevel })
      message.success('共享成功')
      setSelectedUser(undefined)
      load()
    } catch { message.error('共享失败') }
    finally { setSaving(false) }
  }

  const handleRemove = async (permId: string) => {
    try {
      await permissionApi.delete(permId)
      message.success('已移除权限')
      load()
    } catch { message.error('移除失败') }
  }

  return (
    <Modal
      title={`共享: ${resourceName}`}
      open={open}
      onCancel={onClose}
      footer={null}
      width={480}
    >
      <div style={{ marginBottom: 20, padding: 16, background: colors.bgPage, borderRadius: 8 }}>
        <div style={{ fontWeight: 500, marginBottom: 10, color: colors.textPrimary }}>添加共享</div>
        <Select
          showSearch
          placeholder="搜索用户..."
          style={{ width: '100%', marginBottom: 10 }}
          filterOption={false}
          onSearch={handleSearch}
          onChange={(val) => setSelectedUser(val)}
          value={selectedUser}
          options={searchResults.map((u) => ({ value: u.id, label: u.label }))}
          notFoundContent={searching ? <Spin size="small" /> : '无匹配用户'}
        />
        <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
          <Segmented
            value={permLevel}
            onChange={(val) => setPermLevel(val as string)}
            options={[
              { label: '可读', value: 'read' },
              { label: '可写', value: 'write' },
              { label: '管理', value: 'admin' },
            ]}
          />
          <Button type="primary" onClick={handleAdd} loading={saving} disabled={!selectedUser}>
            添加
          </Button>
        </div>
      </div>

      <div style={{ fontWeight: 500, marginBottom: 8, color: colors.textPrimary }}>当前共享</div>
      {loading ? (
        <div style={{ textAlign: 'center', padding: 20 }}><Spin /></div>
      ) : permissions.length === 0 ? (
        <div style={{ color: colors.textTertiary, padding: 16, textAlign: 'center', border: `1px dashed ${colors.border}`, borderRadius: 8 }}>
          暂无共享用户
        </div>
      ) : (
        <List
          dataSource={permissions}
          renderItem={(p) => (
            <List.Item
              actions={[
                <Button
                  key="remove"
                  type="text"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={() => handleRemove(p.id)}
                />,
              ]}
            >
              <List.Item.Meta
                avatar={p.user_id ? <UserOutlined style={{ fontSize: 18, color: colors.textSecondary }} /> : <TeamOutlined style={{ fontSize: 18, color: colors.textSecondary }} />}
                title={p.user_name || p.team_name}
                description={<Tag>{p.permission === 'admin' ? '管理' : p.permission === 'write' ? '可写' : '可读'}</Tag>}
              />
            </List.Item>
          )}
        />
      )}
    </Modal>
  )
}
