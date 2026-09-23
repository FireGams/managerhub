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
      convertEol: true,
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.open(ref.current)
    fit.fit()

    // Focus on click
    ref.current.addEventListener('click', () => term.focus())
    term.focus()

    const offs: (() => void)[] = []

    offs.push(ws.on('terminal_ready', (d: any) => {
      sessRef.current = d.session_id
      term.write('\r\n\x1b[32m✓ Connected — session ' + d.session_id.substring(0, 8) + '\x1b[0m\r\n')
    }))

    offs.push(ws.on('terminal_output', (d: any) => {
      if (d.session_id === sessRef.current) {
        try {
          // decode base64
          const bytes = Uint8Array.from(atob(d.data), c => c.charCodeAt(0))
          term.write(bytes)
        } catch {
          term.write(d.data)
        }
      }
    }))

    offs.push(ws.on('terminal_closed', (d: any) => {
      if (d.session_id === sessRef.current) {
        term.write('\r\n\x1b[33m[session closed: ' + (d.reason || 'unknown') + ']\x1b[0m\r\n')
        sessRef.current = ''
      }
    }))

    // Open session
    ws.send({ action: 'terminal_open', node_id: nodeId, shell: shell || '', cols: term.cols, rows: term.rows })
    term.write('\x1b[90m[opening session…]\x1b[0m\r\n')

    // Input: send base64-encoded bytes
    term.onData((data: string) => {
      if (!sessRef.current) {
        term.write('\x1b[31m[no session — wait for connected]\x1b[0m')
        return
      }
      // Local echo for debugging (remove once confirmed working)
      term.write(data)
      // Send as base64
      const b64 = btoa(data)
      ws.send({ action: 'terminal_input', session_id: sessRef.current, data: b64 })
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
      <div ref={ref} style={{ width: '100%', height: '500px', borderRadius: 8, overflow: 'hidden', background: '#0a0e17', cursor: 'text' }} />
    </div>
  )
}
