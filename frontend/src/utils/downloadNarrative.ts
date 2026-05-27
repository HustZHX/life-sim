export type NarrativeExportFormat = 'txt' | 'md' | 'pdf'

function escapeHtml(text: string): string {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
}

/** 用于本地文件名的安全片段 */
export function safeExportFilename(name: string): string {
  const trimmed = name.trim().replace(/[<>:"/\\|?*\x00-\x1f]/g, '_')
  return trimmed.slice(0, 80) || '轻小说'
}

function triggerBlobDownload(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}

function formatTxt(title: string, content: string): string {
  const header = title.trim() ? `${title.trim()}\n${'='.repeat(Math.min(title.length, 40))}\n\n` : ''
  return `${header}${content.trim()}\n`
}

function formatMd(title: string, content: string): string {
  const heading = title.trim() ? `# ${title.trim()}\n\n` : ''
  return `${heading}${content.trim()}\n`
}

export function downloadNarrativeAsTxt(title: string, content: string, baseName: string) {
  const text = formatTxt(title, content)
  const blob = new Blob([text], { type: 'text/plain;charset=utf-8' })
  triggerBlobDownload(blob, `${safeExportFilename(baseName)}.txt`)
}

export function downloadNarrativeAsMd(title: string, content: string, baseName: string) {
  const text = formatMd(title, content)
  const blob = new Blob([text], { type: 'text/markdown;charset=utf-8' })
  triggerBlobDownload(blob, `${safeExportFilename(baseName)}.md`)
}

export async function downloadNarrativeAsPdf(
  title: string,
  content: string,
  baseName: string
): Promise<void> {
  const { jsPDF } = await import('jspdf')
  const doc = new jsPDF({ unit: 'mm', format: 'a4', orientation: 'portrait' })
  const wrapper = document.createElement('div')
  wrapper.style.width = '800px'
  wrapper.style.padding = '24px'
  wrapper.style.fontFamily = '"PingFang SC", "Microsoft YaHei", "Noto Sans SC", sans-serif'
  wrapper.style.fontSize = '14px'
  wrapper.style.lineHeight = '1.75'
  wrapper.style.color = '#303133'
  wrapper.innerHTML = `
    <h1 style="font-size:20px;margin:0 0 16px;font-weight:600">${escapeHtml(title || '轻小说')}</h1>
    <div style="white-space:pre-wrap;word-break:break-word">${escapeHtml(content.trim())}</div>
    <p style="margin-top:24px;font-size:12px;color:#909399">由 Life-Sim 导出 · AI 生成内容仅供参考</p>
  `
  document.body.appendChild(wrapper)

  try {
    await new Promise<void>((resolve) => {
      let settled = false
      const done = () => {
        if (settled) return
        settled = true
        resolve()
      }
      doc.html(wrapper, {
        callback: (pdf) => {
          pdf.save(`${safeExportFilename(baseName)}.pdf`)
          done()
        },
        margin: [12, 12, 12, 12],
        autoPaging: 'text',
        width: 186,
        windowWidth: 800,
        x: 12,
        y: 12,
      })
      setTimeout(done, 15000)
    })
  } finally {
    document.body.removeChild(wrapper)
  }
}

export async function downloadNarrative(
  format: NarrativeExportFormat,
  title: string,
  content: string,
  baseName: string
): Promise<void> {
  if (!content.trim()) {
    throw new Error('暂无内容可导出')
  }
  switch (format) {
    case 'txt':
      downloadNarrativeAsTxt(title, content, baseName)
      break
    case 'md':
      downloadNarrativeAsMd(title, content, baseName)
      break
    case 'pdf':
      await downloadNarrativeAsPdf(title, content, baseName)
      break
    default:
      throw new Error('不支持的导出格式')
  }
}
