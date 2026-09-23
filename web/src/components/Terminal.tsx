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
      term.write('\r\n\x1b[32m[OK] Connected — you can type commands\x1b[0m\r\n')
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

    // Open session
    ws.send({ action: 'terminal_open', node_id: nodeId, shell: shell || '', cols: term.cols, rows: term.rows })
    term.write('\x1b[90m[connecting…]\x1b[0m\r\n')

    // Capture ALL keyboard events on document when terminal is active
    function onKeyDown(e: KeyboardEvent) {
      if (!sessRef.current) return
      // Don't capture if user is typing in an input/textarea elsewhere
      const tag = (e.target as HTMLElement)?.tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT') return

      let data = ''
      if (e.key === 'Enter') data = '\r'
      else if (e.key === 'Backspace') data = '\x7f'
      else if (e.key === 'Tab') { data = '\t'; e.preventDefault() }
      else if (e.key === 'Escape') data = '\x1b'
      else if (e.key === 'ArrowUp') { data = '\x1b[A'; e.preventDefault() }
      else if (e.key === 'ArrowDown') { data = '\x1b[B'; e.preventDefault() }
      else if (e.key === 'ArrowRight') { data = '\x1b[C'; e.preventDefault() }
      else if (e.key === 'ArrowLeft') { data = '\x1b[D'; e.preventDefault() }
      else if (e.ctrlKey && e.key === 'c') { data = '\x03'; e.preventDefault() }
      else if (e.ctrlKey && e.key === 'd') { data = '\x04'; e.preventDefault() }
      else if (e.ctrlKey && e.key === 'z') { data = '\x1a'; e.preventDefault() }
      else if (e.ctrlKey && e.key === 'l') { data = '\x0c'; e.preventDefault() }
      else if (e.ctrlKey && e.key === 'u') { data = '\x15'; e.preventDefault() }
      else if (e.ctrlKey && e.key === 'a') { data = '\x01'; e.preventDefault() }
      else if (e.ctrlKey && e.key === 'e') { data = '\x05'; e.preventDefault() }
      else if (e.key.length === 1 && !e.ctrlKey && !e.metaKey) {
        data = e.key
      } else return

      if (data && sessRef.current) {
        const b64 = btoa(data)
        ws?.send({ action: 'terminal_input', session_id: sessRef.current, data: b64 })
      }
    }

    document.addEventListener('keydown', onKeyDown)

    const resizeObs = new ResizeObserver(() => {
      fit.fit()
      if (sessRef.current) {
        ws.send({ action: 'terminal_resize', session_id: sessRef.current, cols: term.cols, rows: term.rows })
      }
    })
    resizeObs.observe(termDivRef.current)

    return () => {
      document.removeEventListener('keydown', onKeyDown)
      offs.forEach(f => f())
      resizeObs.disconnect()
      if (sessRef.current) ws.send({ action: 'terminal_close', session_id: sessRef.current })
      term.dispose()
      termRef.current = null
    }
  }, [ws, nodeId, shell])

  return (
    <div>
      <div
        ref={termDivRef}
        style={{
          width: '100%',
          height: '500px',
          borderRadius: 8,
          overflow: 'hidden',
          background: '#0a0e17',
          cursor: 'text',
        }}
      />
      <div style={{ marginTop: '.5rem', fontSize: '.8rem', color: ready ? 'var(--ok)' : 'var(--warn)' }}>
        {ready ? '[OK] Session active — just type commands' : '[...] Connecting...'}
      </div>
    </div>
  )
}
