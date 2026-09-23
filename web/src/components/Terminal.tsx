import { useEffect, useRef, useState } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import type { ManagerHubWS } from '../ws'

export default function TerminalPane({ ws, nodeId, shell }: { ws: ManagerHubWS | null; nodeId: string; shell?: string }) {
  const termDivRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
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
      term.write('\r\n\x1b[32m✓ Connected — session ' + d.session_id.substring(0, 8) + '\x1b[0m\r\n')
      inputRef.current?.focus()
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
    term.write('\x1b[90m[opening session…]\x1b[0m\r\n')

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

  // Handle keyboard input via hidden input element
  function handleInput(e: React.ChangeEvent<HTMLInputElement>) {
    const data = e.target.value
    if (!data || !sessRef.current || !termRef.current) return
    // Send each character
    for (const ch of data) {
      const b64 = btoa(ch)
      ws?.send({ action: 'terminal_input', session_id: sessRef.current, data: b64 })
    }
    e.target.value = ''
  }

  // Handle special keys (Enter, Backspace, Tab, arrows, Ctrl+C...)
  function handleKeyDown(e: React.KeyboardEvent<HTMLInputElement>) {
    if (!sessRef.current) return
    let data = ''
    if (e.key === 'Enter') data = '\r'
    else if (e.key === 'Backspace') data = '\x7f'
    else if (e.key === 'Tab') data = '\t'
    else if (e.key === 'Escape') data = '\x1b'
    else if (e.key === 'ArrowUp') data = '\x1b[A'
    else if (e.key === 'ArrowDown') data = '\x1b[B'
    else if (e.key === 'ArrowRight') data = '\x1b[C'
    else if (e.key === 'ArrowLeft') data = '\x1b[D'
    else if (e.ctrlKey && e.key === 'c') data = '\x03'
    else if (e.ctrlKey && e.key === 'd') data = '\x04'
    else if (e.ctrlKey && e.key === 'z') data = '\x1a'
    else if (e.ctrlKey && e.key === 'l') data = '\x0c'
    else return // let normal characters go through handleInput

    e.preventDefault()
    const b64 = btoa(data)
    ws?.send({ action: 'terminal_input', session_id: sessRef.current, data: b64 })
  }

  return (
    <div>
      {/* Hidden input captures all keyboard input */}
      <input
        ref={inputRef}
        type="text"
        autoFocus
        style={{
          position: 'absolute',
          opacity: 0,
          width: 1,
          height: 1,
          left: -9999,
        }}
        onChange={handleInput}
        onKeyDown={handleKeyDown}
        aria-label="terminal input"
      />
      <div
        ref={termDivRef}
        onClick={() => inputRef.current?.focus()}
        style={{
          width: '100%',
          height: '500px',
          borderRadius: 8,
          overflow: 'hidden',
          background: '#0a0e17',
          cursor: 'text',
          position: 'relative',
        }}
      />
      {!ready && (
        <div style={{ marginTop: '.5rem', fontSize: '.8rem', color: 'var(--warn)' }}>
          ⏳ Waiting for connection… click on the terminal and start typing.
        </div>
      )}
    </div>
  )
}
