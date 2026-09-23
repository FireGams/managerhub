import { useEffect, useRef, useState } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import type { ManagerHubWS } from '../ws'

export default function TerminalPane({ ws, nodeId, shell }: { ws: ManagerHubWS | null; nodeId: string; shell?: string }) {
  const termDivRef = useRef<HTMLDivElement>(null)
  const sessRef = useRef('')
  const termRef = useRef<XTerm | null>(null)
  const [ready, setReady] = useState(false)
  const [cmd, setCmd] = useState('')

  useEffect(() => {
    if (!termDivRef.current || !ws) return
    const term = new XTerm({
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: { background: '#0a0e17', foreground: '#f1f5f9', cursor: '#6366f1' },
      cursorBlink: true,
      convertEol: true,
    })
    termRef.current = term
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.open(termDivRef.current)
    fit.fit()

    const offs: (() => void)[] = []

    offs.push(ws.on('terminal_ready', (d: any) => {
      sessRef.current = d.session_id
      setReady(true)
      term.write('\x1b[32m[OK] Connected\x1b[0m\r\n')
    }))

    offs.push(ws.on('terminal_output', (d: any) => {
      if (d.session_id === sessRef.current) {
        try {
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
        setReady(false)
      }
    }))

    ws.send({ action: 'terminal_open', node_id: nodeId, shell: shell || '', cols: term.cols, rows: term.rows })
    term.write('\x1b[90m[connecting…]\x1b[0m\r\n')

    const resizeObs = new ResizeObserver(() => {
      fit.fit()
      if (sessRef.current) {
        ws.send({ action: 'terminal_resize', session_id: sessRef.current, cols: term.cols, rows: term.rows })
      }
    })
    resizeObs.observe(termDivRef.current)

    return () => {
      offs.forEach(f => f())
      resizeObs.disconnect()
      if (sessRef.current) ws.send({ action: 'terminal_close', session_id: sessRef.current })
      term.dispose()
      termRef.current = null
    }
  }, [ws, nodeId, shell])

  function sendCommand(e: React.FormEvent) {
    e.preventDefault()
    if (!cmd || !sessRef.current || !ws) return
    // Send each character + Enter
    for (const ch of cmd) {
      ws.send({ action: 'terminal_input', session_id: sessRef.current, data: btoa(ch) })
    }
    ws.send({ action: 'terminal_input', session_id: sessRef.current, data: btoa('\r') })
    setCmd('')
  }

  return (
    <div>
      <div
        ref={termDivRef}
        style={{
          width: '100%',
          height: '450px',
          borderRadius: '8px 8px 0 0',
          overflow: 'hidden',
          background: '#0a0e17',
        }}
      />
      <form
        onSubmit={sendCommand}
        style={{
          display: 'flex',
          gap: '.5rem',
          padding: '.5rem',
          background: 'var(--panel2)',
          borderRadius: '0 0 8px 8px',
          border: '1px solid var(--border)',
          borderTop: 'none',
        }}
      >
        <span style={{ color: 'var(--accent)', alignSelf: 'center', fontFamily: 'monospace' }}>$</span>
        <input
          type="text"
          value={cmd}
          onChange={e => setCmd(e.target.value)}
          placeholder={ready ? 'Type a command and press Enter…' : 'Waiting for connection…'}
          disabled={!ready}
          autoFocus
          style={{ flex: 1, fontFamily: 'monospace', fontSize: 14 }}
        />
        <button type="submit" disabled={!ready || !cmd}>Run</button>
      </form>
      <div style={{ marginTop: '.3rem', fontSize: '.75rem', color: ready ? 'var(--ok)' : 'var(--warn)' }}>
        {ready ? 'Session active — type commands below' : 'Connecting…'}
      </div>
    </div>
  )
}
