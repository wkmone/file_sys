// 腾讯文档风格主题配置
// 使用 antd token + 内联 style 常量统一管理

// 主色 token (传递给 ConfigProvider)
export const themeToken = {
  colorPrimary: '#0052D9',
  borderRadius: 8,
  fontFamily: "-apple-system, BlinkMacSystemFont, 'PingFang SC', 'Microsoft YaHei', 'Helvetica Neue', sans-serif",
  colorBgContainer: '#ffffff',
  colorBorderSecondary: '#e5e6eb',
}

// 品牌色
export const colors = {
  primary: '#0052D9',
  primaryLight: '#e8f0fe',
  primaryHover: '#0044b8',
  textPrimary: '#1d1d1f',
  textSecondary: '#88888a',
  textTertiary: '#b0b0b2',
  border: '#e5e6eb',
  bgPage: '#f5f6f7',
  bgHover: '#f0f2f5',
  bgSidebar: '#ffffff',
  white: '#ffffff',
  danger: '#e34d59',
  success: '#00a870',
  warning: '#ed7b2f',
}

// 文件类型大图标颜色
export const fileCardIconColors: Record<string, string> = {
  '.pdf': '#e34d59',
  '.xlsx': '#00a870',
  '.xls': '#00a870',
  '.pptx': '#ed7b2f',
  '.ppt': '#ed7b2f',
  '.png': '#0052D9',
  '.jpg': '#0052D9',
  '.jpeg': '#0052D9',
  '.gif': '#0052D9',
  '.docx': '#0052D9',
  '.doc': '#0052D9',
  default: colors.textSecondary,
}

// 文件夹图标颜色
export const folderColor = '#f5a623'

// 间距体系
export const spacing = {
  xs: 4,
  sm: 8,
  md: 12,
  lg: 16,
  xl: 24,
  xxl: 32,
}

// 侧边栏
export const sidebar = {
  width: 220,
  collapsedWidth: 64,
}
