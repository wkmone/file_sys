import { useParams, useSearchParams } from 'react-router-dom'
import { Typography } from 'antd'
import OnlyOfficeEditor from '../components/editor/OnlyOfficeEditor'

const VALID_MODES = ['edit', 'view', 'comment', 'review', 'fillForms'] as const

export default function EditorPage() {
  const { fileId } = useParams<{ fileId: string }>()
  const [searchParams] = useSearchParams()

  if (!fileId) {
    return <Typography.Text type="danger">文件 ID 缺失</Typography.Text>
  }

  const modeParam = searchParams.get('mode')
  const mode = VALID_MODES.includes(modeParam as typeof VALID_MODES[number])
    ? (modeParam as typeof VALID_MODES[number])
    : 'edit'

  return (
    <div style={{ width: '100vw', height: '100vh' }}>
      <OnlyOfficeEditor fileId={fileId} mode={mode} />
    </div>
  )
}
