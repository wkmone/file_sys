import { useState } from 'react'
import { Typography, Card, Descriptions, Button, Modal, Input, message, Avatar } from 'antd'
import { UserOutlined, LockOutlined } from '@ant-design/icons'
import { useAuthStore } from '../store/authStore'
import { authApi } from '../api/authApi'
import { colors } from '../theme'

export default function ProfilePage() {
  const user = useAuthStore((s) => s.user)
  const [passwordOpen, setPasswordOpen] = useState(false)
  const [oldPassword, setOldPassword] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [changing, setChanging] = useState(false)

  const handleChangePassword = async () => {
    if (!oldPassword || !newPassword) return
    if (newPassword.length < 8) {
      message.error('新密码至少 8 位，需包含字母和数字')
      return
    }
    setChanging(true)
    try {
      await authApi.changePassword({ old_password: oldPassword, new_password: newPassword })
      message.success('密码修改成功')
      setPasswordOpen(false)
      setOldPassword('')
      setNewPassword('')
    } catch (err: any) {
      message.error(err.response?.data?.message || '密码修改失败')
    } finally {
      setChanging(false)
    }
  }

  return (
    <div>
      <Typography.Title level={3} style={{ color: colors.textPrimary, marginBottom: 24 }}>个人设置</Typography.Title>

      {/* Profile card */}
      <Card style={{ maxWidth: 640, borderRadius: 12, border: `1px solid ${colors.border}` }}>
        <div style={{ display: 'flex', alignItems: 'center', gap: 20, marginBottom: 28 }}>
          <Avatar size={72} icon={<UserOutlined />} style={{ background: colors.primary, flexShrink: 0 }} />
          <div>
            <Typography.Title level={4} style={{ margin: 0, color: colors.textPrimary }}>
              {user?.display_name}
            </Typography.Title>
            <Typography.Text style={{ color: colors.textSecondary }}>{user?.email}</Typography.Text>
            <br />
            <Typography.Text style={{ color: colors.textTertiary, fontSize: 13 }}>
              {user?.role === 'admin' ? '管理员' : '成员'}
            </Typography.Text>
          </div>
        </div>

        <Descriptions column={1} bordered size="middle" labelStyle={{ color: colors.textSecondary, fontWeight: 500 }}>
          <Descriptions.Item label="姓名">{user?.display_name}</Descriptions.Item>
          <Descriptions.Item label="邮箱">{user?.email}</Descriptions.Item>
          <Descriptions.Item label="角色">
            {user?.role === 'admin' ? '管理员' : '成员'}
          </Descriptions.Item>
          <Descriptions.Item label="注册时间">
            {user?.created_at ? new Date(user.created_at).toLocaleDateString() : '-'}
          </Descriptions.Item>
          <Descriptions.Item label="密码">
            <Button
              type="link"
              icon={<LockOutlined />}
              onClick={() => setPasswordOpen(true)}
              style={{ padding: 0, color: colors.primary }}
            >
              修改密码
            </Button>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Change Password Modal */}
      <Modal
        title="修改密码"
        open={passwordOpen}
        onCancel={() => { setPasswordOpen(false); setOldPassword(''); setNewPassword('') }}
        onOk={handleChangePassword}
        confirmLoading={changing}
        okText="确认修改"
      >
        <div style={{ marginBottom: 4 }}>
          <Typography.Text style={{ fontSize: 13, color: colors.textSecondary }}>当前密码</Typography.Text>
        </div>
        <Input.Password
          prefix={<LockOutlined />}
          placeholder="输入当前密码"
          value={oldPassword}
          onChange={(e) => setOldPassword(e.target.value)}
          style={{ marginBottom: 16 }}
        />
        <div style={{ marginBottom: 4 }}>
          <Typography.Text style={{ fontSize: 13, color: colors.textSecondary }}>新密码</Typography.Text>
        </div>
        <Input.Password
          prefix={<LockOutlined />}
          placeholder="输入新密码（至少6位）"
          value={newPassword}
          onChange={(e) => setNewPassword(e.target.value)}
        />
      </Modal>
    </div>
  )
}
