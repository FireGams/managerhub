import { useEffect, useRef } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import type { ManagerHubWS } from '../ws'

export default function TerminalPane({ ws, nodeId, shell }: { ws: ManagerHubWS | null; nodeId: string; shell?: string }) {
  const ref = useRef<HTMLDivElement>(null)
  const sessRef = useRef('')

  useEffect(() => {
    if (!ref.current || !ws) return
    const term = new XTerm({
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: { background: '#0a0e17', foreground: '#f1f5f9', cursor: '#6366f1' },
      cursorBlink: true,
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.open(ref.current)
    fit.fit()
    term.focus()

    const offs: (() => void)[] = []

    offs.push(ws.on('terminal_ready', (d: any) => {
      sessRef.current = d.session_id
      term.write('\x1b[32m[connected — session ' + d.session_id.substring(0, 8) + ']\x1b[0m\r\n')
    }))

    offs.push(ws.on('terminal_output', (d: any) => {
      if (d.session_id === sessRef.current) {
        try { term.write(atob(d.data)) } catch { term.write(d.data) }
      }
    }))

    offs.push(ws.on('terminal_closed', (d: any) => {
      if (d.session_id === sessRef.current) {
        term.write('\r\n\x1b[33m[session closed: ' + (d.reason || 'unknown') + ']\x1b[0m\r\n')
        sessRef.current = ''
      }
    }))

    // Open terminal session
    ws.send({ action: 'terminal_open', node_id: nodeId, shell: shell || '', cols: term.cols, rows: term.rows })

    // Send raw text input (no base64 — simpler and more reliable)
    term.onData((data: string) => {
      if (sessRef.current) {
        ws.send({ action: 'terminal_input', session_id: sessRef.current, data: data })
      } else {
        term.write('\x1b[31m[not connected yet]\x1b[0m')
      }
    })

    const resizeObs = new ResizeObserver(() => {
      fit.fit()
      if (sessRef.current) {
        ws.send({ action: 'terminal_resize', session_id: sessRef.current, cols: term.cols, rows: term.rows })
      }
    })
    resizeObs.observe(ref.current)

    return () => {
      offs.forEach(f => f())
      resizeObs.disconnect()
      if (sessRef.current) ws.send({ action: 'terminal_close', session_id: sessRef.current })
      term.dispose()
    }
  }, [ws, nodeId, shell])

  return (
    <div>
      <div ref={ref} style={{ width: '100%', height: '500px', borderRadius: 8, overflow: 'hidden', background: '#0a0e17' }} />
    </div>
  )
}
