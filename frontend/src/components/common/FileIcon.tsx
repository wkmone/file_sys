import { useMemo } from 'react'
import { folderColor, fileCardIconColors } from '../../theme'

interface FileIconProps {
  /** File extension including dot, e.g. ".pdf", or "folder" for folder icon */
  type: string
  size?: number
}

/** Renders a filled-block style icon: colored rounded square + white SVG shape */
export default function FileIcon({ type, size = 48 }: FileIconProps) {
  const color = useMemo(() => {
    if (type === 'folder') return folderColor
    if (type === 'logo') return '#0052D9'
    return fileCardIconColors[type] || fileCardIconColors.default
  }, [type])

  const inner = useMemo(() => {
    switch (type) {
      case 'logo':
        return <LogoShape />
      case 'folder':
        return <FolderShape />
      case '.pdf':
        return <PdfShape />
      case '.xlsx':
      case '.xls':
      case '.ods':
      case '.csv':
        return <ExcelShape />
      case '.pptx':
      case '.ppt':
      case '.odp':
        return <PptShape />
      case '.png':
      case '.jpg':
      case '.jpeg':
      case '.gif':
        return <ImageShape />
      case '.docx':
      case '.doc':
      case '.odt':
      case '.rtf':
        return <DocShape />
      case '.epub':
        return <EpubShape />
      default:
        return <GenericFileShape />
    }
  }, [type])

  return (
    <svg
      viewBox="0 0 48 48"
      width={size}
      height={size}
      style={{ display: 'block', flexShrink: 0 }}
    >
      <rect width="48" height="48" rx="10" fill={color} />
      {inner}
    </svg>
  )
}

/* ── Logo (document + folder) ────────────────── */
function LogoShape() {
  return (
    <>
      <rect x="10" y="8" width="24" height="30" rx="2" fill="white" fillOpacity={0.95} />
      <path d="M28 8l6 6H30a2 2 0 0 1-2-2V8z" fill="#b0c8e8" />
      <path d="M8 18h15l2 3h15a1.5 1.5 0 0 1 1.5 1.5v10a1.5 1.5 0 0 1-1.5 1.5H8a1.5 1.5 0 0 1-1.5-1.5V18z" fill="white" fillOpacity={0.85} />
    </>
  )
}

/* ── Folder ──────────────────────────────────── */
function FolderShape() {
  return (
    <path
      d="M10 14h11l3 3h14a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H10a2 2 0 0 1-2-2V14z"
      fill="white"
      fillOpacity={0.95}
    />
  )
}

/* ── PDF ─────────────────────────────────────── */
function PdfShape() {
  return (
    <>
      {/* document body */}
      <rect x="10" y="8" width="24" height="30" rx="2" fill="white" fillOpacity={0.95} />
      {/* corner fold */}
      <path d="M28 8l6 6H30a2 2 0 0 1-2-2V8z" fill="#e8e8e8" />
      {/* label */}
      <rect x="14" y="20" width="14" height="7" rx="1.5" fill="#e34d59" />
      <text x="21" y="25.5" textAnchor="middle" fill="white" fontSize="5" fontWeight="600" fontFamily="system-ui, sans-serif">PDF</text>
      {/* text lines */}
      <rect x="12" y="30" width="8" height="1.5" rx="0.75" fill="#ccc" />
      <rect x="12" y="33" width="6" height="1.5" rx="0.75" fill="#ccc" />
    </>
  )
}

/* ── Excel ───────────────────────────────────── */
function ExcelShape() {
  return (
    <>
      <rect x="10" y="8" width="24" height="30" rx="2" fill="white" fillOpacity={0.95} />
      <path d="M28 8l6 6H30a2 2 0 0 1-2-2V8z" fill="#e8e8e8" />
      {/* table grid */}
      <rect x="12" y="19" width="18" height="14" rx="1" fill="white" stroke="#00a870" strokeWidth={0.8} />
      <line x1="12" y1="24" x2="30" y2="24" stroke="#00a870" strokeWidth={0.5} />
      <line x1="12" y1="28" x2="30" y2="28" stroke="#00a870" strokeWidth={0.5} />
      <line x1="18" y1="19" x2="18" y2="33" stroke="#00a870" strokeWidth={0.5} />
    </>
  )
}

/* ── PPT ─────────────────────────────────────── */
function PptShape() {
  return (
    <>
      <rect x="10" y="8" width="24" height="30" rx="2" fill="white" fillOpacity={0.95} />
      <path d="M28 8l6 6H30a2 2 0 0 1-2-2V8z" fill="#e8e8e8" />
      {/* slide area with chart */}
      <rect x="12" y="18" width="18" height="13" rx="1" fill="white" stroke="#ed7b2f" strokeWidth={0.5} />
      {/* bar chart */}
      <rect x="14" y="24" width="3" height="5" rx="0.5" fill="#ed7b2f" />
      <rect x="18" y="21" width="3" height="8" rx="0.5" fill="#ed7b2f" />
      <rect x="22" y="19" width="3" height="10" rx="0.5" fill="#ed7b2f" />
      {/* play button */}
      <circle cx="27" cy="27" r="3" fill="#ed7b2f" />
    </>
  )
}

/* ── Image ───────────────────────────────────── */
function ImageShape() {
  return (
    <>
      <rect x="10" y="8" width="24" height="30" rx="2" fill="white" fillOpacity={0.95} />
      <path d="M28 8l6 6H30a2 2 0 0 1-2-2V8z" fill="#e8e8e8" />
      {/* landscape */}
      <circle cx="26" cy="18" r="3" fill="#f0c040" />
      <path d="M12 29l5-4 4 3 3-5 6 6H12z" fill="#a0d0e0" opacity={0.7} />
      <path d="M12 28l4-4 5 4 3-3 5 3H12z" fill="#60b0d0" opacity={0.5} />
    </>
  )
}

/* ── Word / Doc ──────────────────────────────── */
function DocShape() {
  return (
    <>
      <rect x="10" y="8" width="24" height="30" rx="2" fill="white" fillOpacity={0.95} />
      <path d="M28 8l6 6H30a2 2 0 0 1-2-2V8z" fill="#e8e8e8" />
      {/* text block - heading */}
      <rect x="12" y="18" width="10" height="2" rx="1" fill="#0052D9" />
      {/* text block - body lines */}
      <rect x="12" y="22" width="18" height="1.5" rx="0.75" fill="#b0c8e8" />
      <rect x="12" y="25" width="15" height="1.5" rx="0.75" fill="#b0c8e8" />
      <rect x="12" y="28" width="12" height="1.5" rx="0.75" fill="#b0c8e8" />
      <rect x="12" y="31" width="6" height="1.5" rx="0.75" fill="#b0c8e8" />
    </>
  )
}

/* ── Epub / Book ─────────────────────────────── */
function EpubShape() {
  return (
    <>
      <rect x="10" y="8" width="24" height="30" rx="2" fill="white" fillOpacity={0.95} />
      <path d="M28 8l6 6H30a2 2 0 0 1-2-2V8z" fill="#e8e8e8" />
      {/* book spine + pages */}
      <rect x="14" y="16" width="16" height="18" rx="1" fill="white" stroke="#e34d59" strokeWidth={0.7} />
      <line x1="22" y1="16" x2="22" y2="34" stroke="#e34d59" strokeWidth={0.5} />
      {/* text lines on left page */}
      <rect x="15.5" y="18.5" width="5" height="1" rx="0.5" fill="#d0d0d0" />
      <rect x="15.5" y="20.5" width="4.5" height="1" rx="0.5" fill="#d0d0d0" />
      <rect x="15.5" y="22.5" width="5" height="1" rx="0.5" fill="#d0d0d0" />
      {/* text lines on right page */}
      <rect x="23.5" y="18.5" width="5" height="1" rx="0.5" fill="#d0d0d0" />
      <rect x="23.5" y="20.5" width="4" height="1" rx="0.5" fill="#d0d0d0" />
      <rect x="23.5" y="22.5" width="5" height="1" rx="0.5" fill="#d0d0d0" />
    </>
  )
}

/* ── Generic file ────────────────────────────── */
function GenericFileShape() {
  return (
    <>
      <rect x="10" y="8" width="24" height="30" rx="2" fill="white" fillOpacity={0.95} />
      <path d="M28 8l6 6H30a2 2 0 0 1-2-2V8z" fill="#e8e8e8" />
      {/* generic text lines */}
      <rect x="13" y="20" width="16" height="2" rx="1" fill="#888" />
      <rect x="13" y="24" width="12" height="1.5" rx="0.75" fill="#bbb" />
      <rect x="13" y="27" width="14" height="1.5" rx="0.75" fill="#bbb" />
      <rect x="13" y="30" width="8" height="1.5" rx="0.75" fill="#bbb" />
    </>
  )
}
