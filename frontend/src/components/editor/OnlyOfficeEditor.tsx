import { useEffect, useRef, useState } from 'react'
import { Spin } from 'antd'
import { onlyofficeApi } from '../../api/onlyofficeApi'

interface OnlyOfficeEditorProps {
  fileId: string
  mode: 'edit' | 'view'
}

declare global {
  interface Window {
    DocsAPI: any
  }
}

let docsAPIPromise: Promise<void> | null = null

function loadDocsAPI(dsUrl: string): Promise<void> {
  if (window.DocsAPI) return Promise.resolve()
  if (docsAPIPromise) return docsAPIPromise

  docsAPIPromise = new Promise<void>((resolve, reject) => {
    const script = document.createElement('script')
    script.src = `${dsUrl}/web-apps/apps/api/documents/api.js`
    script.async = true

    const timeout = setTimeout(() => {
      docsAPIPromise = null
      reject(new Error('连接 OnlyOffice 服务超时'))
    }, 15000)

    script.onload = () => {
      clearTimeout(timeout)
      if (window.DocsAPI) { resolve(); return }
      let n = 0
      const poll = setInterval(() => {
        if (window.DocsAPI) { clearInterval(poll); resolve() }
        else if (++n > 30) { clearInterval(poll); docsAPIPromise = null; reject(new Error('编辑器初始化超时')) }
      }, 100)
    }

    script.onerror = () => {
      clearTimeout(timeout)
      docsAPIPromise = null
      reject(new Error('无法连接 OnlyOffice 服务'))
    }

    document.head.appendChild(script)
  })

  return docsAPIPromise
}

export default function OnlyOfficeEditor({ fileId, mode }: OnlyOfficeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const instanceRef = useRef<any>(null)
  const mountedRef = useRef(true)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const editorId = `oo-editor-${fileId}`

  useEffect(() => {
    mountedRef.current = true
    let cancelled = false

    async function init() {
      setLoading(true)
      setError(null)

      if (instanceRef.current?.destroyEditor) {
        try { instanceRef.current.destroyEditor() } catch { /* ignore */ }
        instanceRef.current = null
      }

      if (containerRef.current) {
        containerRef.current.innerHTML = ''
      }

      try {
        const dsUrl = onlyofficeApi.getDocServerUrl()

        const [configRes] = await Promise.all([
          onlyofficeApi.getEditorConfig(fileId, mode),
          loadDocsAPI(dsUrl),
        ])

        if (cancelled || !mountedRef.current || !containerRef.current) return

        const { config, token } = configRes.data.data

        const editorConfig: Record<string, any> = { type: 'desktop', width: '100%', height: '100%', token }
        if (config) {
          Object.assign(editorConfig, config)
        }

        const editor = new window.DocsAPI.DocEditor(editorId, editorConfig)
        instanceRef.current = editor
        setLoading(false)
      } catch (err: any) {
        if (cancelled || !mountedRef.current) return
        setError(err.response?.data?.message || err.message || '编辑器加载失败')
        setLoading(false)
      }
    }

    init()

    return () => {
      cancelled = true
      mountedRef.current = false
      if (instanceRef.current?.destroyEditor) {
        try { instanceRef.current.destroyEditor() } catch { /* ignore */ }
        instanceRef.current = null
      }
    }
  }, [fileId, mode])

  return (
    <div style={{ position: 'relative', width: '100%', height: '100%' }}>
      {loading && (
        <div style={{
          position: 'absolute', inset: 0,
          display: 'flex', justifyContent: 'center', alignItems: 'center',
          background: '#fff', zIndex: 1,
        }}>
          <Spin size="large" />
        </div>
      )}
      {error && (
        <div style={{
          position: 'absolute', inset: 0,
          display: 'flex', justifyContent: 'center', alignItems: 'center',
          background: '#fff', zIndex: 1,
        }}>
          <p style={{ color: '#e34d59', fontSize: 16 }}>{error}</p>
        </div>
      )}
      <div
        ref={containerRef}
        id={editorId}
        style={{ width: '100%', height: '100%' }}
      />
    </div>
  )
}
