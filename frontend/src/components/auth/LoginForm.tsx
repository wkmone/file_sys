import { useState } from 'react'
import { useNavigate, useSearchParams, Link } from 'react-router-dom'
import { Form, Input, Button, message } from 'antd'
import { MailOutlined, LockOutlined } from '@ant-design/icons'
import { authApi } from '../../api/authApi'
import { useAuthStore } from '../../store/authStore'

export default function LoginForm() {
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const login = useAuthStore((s) => s.login)

  const onFinish = async (values: { email: string; password: string }) => {
    setLoading(true)
    try {
      const res = await authApi.login(values)
      const { access_token, user } = res.data.data
      login(access_token, user)
      message.success('登录成功')
      navigate(searchParams.get('redirect') || '/')
    } catch (err: any) {
      message.error(err.response?.data?.message || '登录失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <Form onFinish={onFinish} size="large" layout="vertical">
      <Form.Item name="email" rules={[{ required: true, type: 'email', message: '请输入有效邮箱' }]}>
        <Input prefix={<MailOutlined />} placeholder="邮箱" />
      </Form.Item>
      <Form.Item name="password" rules={[{ required: true, min: 6, message: '密码至少 6 位' }]}>
        <Input.Password prefix={<LockOutlined />} placeholder="密码" />
      </Form.Item>
      <Form.Item>
        <Button type="primary" htmlType="submit" loading={loading} block>
          登录
        </Button>
      </Form.Item>
      <div style={{ textAlign: 'center' }}>
        还没有账号？ <Link to="/register">注册</Link>
      </div>
    </Form>
  )
}
