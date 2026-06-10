import { useEffect, useRef, useState } from 'react'
import { Spin } from 'antd'
import { onlyofficeApi } from '../../api/onlyofficeApi'

interface OnlyOfficeEditorProps {
  fileId: string
  mode: 'edit' | 'view' | 'comment' | 'review' | 'fillForms'
  customization?: Partial<{
    autosave: boolean
    chat: boolean
    comments: boolean
    compactHeader: boolean
    compactToolbar: boolean
    forcesave: boolean
    help: boolean
    hideRightMenu: boolean
    hideRulers: boolean
    spellcheck: boolean
    uiTheme: string
    toolbarHideFileName: boolean
    zoom: number
    macros: boolean
    plugins: boolean
  }>
}

declare global {
  interface Window {
    DocsAPI: any
  }
}

// 全局缓存 - 存储编辑器配置和脚本加载状态
const configCache = new Map<string, { config: any; token: string; timestamp: number }>()
const CACHE_DURATION = 30 * 60 * 1000 // 30 分钟缓存

let docsAPIPromise: Promise<void> | null = null
let docsAPILoaded = false

// 在应用启动时预加载 DocsAPI（延迟执行以避免阻塞首屏）
function preloadDocsAPI(dsUrl: string) {
  if (docsAPILoaded || docsAPIPromise) return
  setTimeout(() => {
    loadDocsAPI(dsUrl).catch(err => console.warn('Preload DocsAPI warning:', err))
  }, 1000)
}

function loadDocsAPI(dsUrl: string): Promise<void> {
  if (window.DocsAPI) {
    docsAPILoaded = true
    return Promise.resolve()
  }
  if (docsAPIPromise) return docsAPIPromise

  docsAPIPromise = new Promise<void>((resolve, reject) => {
    const script = document.createElement('script')
    script.src = `${dsUrl}/web-apps/apps/api/documents/api.js`
    script.async = true
    script.defer = true

    const timeout = setTimeout(() => {
      docsAPIPromise = null
      reject(new Error('连接 OnlyOffice 服务超时，请检查网络或服务状态'))
    }, 30000) // 增加到 30 秒

    script.onload = () => {
      clearTimeout(timeout)
      if (window.DocsAPI) { 
        docsAPILoaded = true
        resolve()
        return 
      }
      let n = 0
      const poll = setInterval(() => {
        if (window.DocsAPI) { 
          clearInterval(poll)
          docsAPILoaded = true
          resolve() 
        }
        else if (++n > 60) { // 增加到 60 次轮询（6秒）
          clearInterval(poll)
          docsAPIPromise = null
          reject(new Error('编辑器初始化超时'))
        }
      }, 100)
    }

    script.onerror = () => {
      clearTimeout(timeout)
      docsAPIPromise = null
      reject(new Error('无法连接 OnlyOffice 服务，请稍后重试'))
    }

    document.head.appendChild(script)
  })

  return docsAPIPromise
}

function getCachedConfig(fileId: string, mode: string) {
  const key = `${fileId}-${mode}`
  const cached = configCache.get(key)
  if (cached && (Date.now() - cached.timestamp) < CACHE_DURATION) {
    return cached
  }
  if (cached) {
    configCache.delete(key)
  }
  return null
}

function setCachedConfig(fileId: string, mode: string, config: any, token: string) {
  const key = `${fileId}-${mode}`
  configCache.set(key, { config, token, timestamp: Date.now() })
}

// 在组件导入时触发预加载
const defaultDsUrl = import.meta.env.VITE_ONLYOFFICE_DS_URL || 'http://localhost:9980'
preloadDocsAPI(defaultDsUrl)

export default function OnlyOfficeEditor({ fileId, mode, customization }: OnlyOfficeEditorProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const instanceRef = useRef<any>(null)
  const mountedRef = useRef(true)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [loadProgress, setLoadProgress] = useState(0)

  const editorId = `oo-editor-${fileId}`

  useEffect(() => {
    mountedRef.current = true
    let cancelled = false

    async function init() {
      setLoading(true)
      setError(null)
      setLoadProgress(10)

      if (instanceRef.current?.destroyEditor) {
        try { instanceRef.current.destroyEditor() } catch { /* ignore */ }
        instanceRef.current = null
      }

      if (containerRef.current) {
        containerRef.current.innerHTML = ''
      }

      try {
        const dsUrl = onlyofficeApi.getDocServerUrl()
        setLoadProgress(20)

        // 并行加载：尝试从缓存获取配置 + 加载脚本
        const cachedConfig = getCachedConfig(fileId, mode)
        setLoadProgress(30)

        let config: any
        let token: string

        if (cachedConfig) {
          // 使用缓存，直接加载
          config = cachedConfig.config
          token = cachedConfig.token
          setLoadProgress(60)
        } else {
          // 无缓存，同时请求配置和加载脚本
          const [configRes] = await Promise.all([
            onlyofficeApi.getEditorConfig(fileId, mode),
            loadDocsAPI(dsUrl),
          ])
          config = configRes.data.data.config
          token = configRes.data.data.token
          setCachedConfig(fileId, mode, config, token)
          setLoadProgress(60)
        }

        // 确保 DocsAPI 已加载
        if (!window.DocsAPI) {
          await loadDocsAPI(dsUrl)
        }
        setLoadProgress(80)

        if (cancelled || !mountedRef.current || !containerRef.current) return

        const editorConfig: Record<string, any> = { 
          type: 'desktop', 
          width: '100%', 
          height: '100%', 
          token,
        }
        if (config) {
          Object.assign(editorConfig, config)
        }
        if (customization && editorConfig.editorConfig?.customization) {
          Object.assign(editorConfig.editorConfig.customization, customization)
        }

        setLoadProgress(90)
        const editor = new window.DocsAPI.DocEditor(editorId, editorConfig)
        instanceRef.current = editor
        setLoadProgress(100)
        setLoading(false)
      } catch (err: any) {
        if (cancelled || !mountedRef.current) return
        setError(err.response?.data?.message || err.message || '编辑器加载失败，请刷新页面重试')
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
  }, [fileId, mode, customization])

  return (
    <div style={{ position: 'relative', width: '100%', height: '100%' }}>
      {loading && (
        <div style={{
          position: 'absolute', inset: 0,
          display: 'flex', flexDirection: 'column',
          justifyContent: 'center', alignItems: 'center',
          background: '#fff', zIndex: 1, gap: '16px',
        }}>
          <Spin size="large" tip={loadProgress < 60 ? '加载编辑器组件...' : '初始化文档...'} />
          <div style={{ 
            width: '200px', 
            height: '4px', 
            background: '#f0f0f0', 
            borderRadius: '2px',
            overflow: 'hidden',
          }}>
            <div style={{
              width: `${loadProgress}%`,
              height: '100%',
              background: '#1890ff',
              transition: 'width 0.3s ease',
            }} />
          </div>
        </div>
      )}
      {error && (
        <div style={{
          position: 'absolute', inset: 0,
          display: 'flex', flexDirection: 'column',
          justifyContent: 'center', alignItems: 'center',
          background: '#fff', zIndex: 1, gap: '12px',
        }}>
          <p style={{ color: '#e34d59', fontSize: 16, margin: 0 }}>{error}</p>
          <button
            onClick={() => window.location.reload()}
            style={{
              padding: '8px 16px',
              background: '#1890ff',
              color: '#fff',
              border: 'none',
              borderRadius: '4px',
              cursor: 'pointer',
              fontSize: 14,
            }}
          >
            刷新重试
          </button>
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
