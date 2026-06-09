import { Typography } from 'antd'
import { FileTextOutlined } from '@ant-design/icons'
import LoginForm from '../components/auth/LoginForm'
import { colors } from '../theme'

export default function LoginPage() {
  return (
    <div style={{
      minHeight: '100vh',
      display: 'flex',
      justifyContent: 'center',
      alignItems: 'center',
      background: 'linear-gradient(135deg, #e8f0fe 0%, #d4e4fc 50%, #c5d8f8 100%)',
    }}>
      <div style={{
        width: 420,
        padding: '48px 40px',
        background: '#fff',
        borderRadius: 16,
        boxShadow: '0 8px 40px rgba(0,82,217,0.08)',
      }}>
        {/* Brand header */}
        <div style={{ textAlign: 'center', marginBottom: 36 }}>
          <div style={{
            width: 56,
            height: 56,
            borderRadius: 16,
            background: `linear-gradient(135deg, ${colors.primary}, #3370ff)`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            margin: '0 auto 16px',
            boxShadow: '0 4px 16px rgba(0,82,217,0.25)',
          }}>
            <FileTextOutlined style={{ fontSize: 28, color: '#fff' }} />
          </div>
          <Typography.Title level={3} style={{ margin: 0, color: colors.textPrimary }}>
            文件管理系统
          </Typography.Title>
          <Typography.Text style={{ color: colors.textSecondary, fontSize: 14 }}>
            团队文档协作平台
          </Typography.Text>
        </div>

        <LoginForm />
      </div>
    </div>
  )
}
