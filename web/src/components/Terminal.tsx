import { useEffect, useRef } from 'react'
import { Terminal as XTerm } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import type { ManagerHubWS } from '../ws'

export default function TerminalPane({ ws, nodeId, shell }: { ws: ManagerHubWS | null; nodeId: string; shell?: string }) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (!ref.current || !ws) return
    const term = new XTerm({
      fontSize: 13,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: { background: '#0f1419', foreground: '#e7ecf3', cursor: '#3b82f6' },
      cursorBlink: true,
    })
    const fit = new FitAddon()
    term.loadAddon(fit)
    term.open(ref.current)
    fit.fit()

    let sessionId = ''
    const offs: (() => void)[] = []

    offs.push(ws.on('terminal_ready', (d: any) => { sessionId = d.session_id }))
    offs.push(ws.on('terminal_output', (d: any) => {
      if (d.session_id === sessionId) term.write(atob(d.data))
    }))
    offs.push(ws.on('terminal_closed', (d: any) => {
      if (d.session_id === sessionId) term.write('\r\n\x1b[33m[session closed: ' + (d.reason || 'unknown') + ']\x1b[0m\r\n')
    }))

    ws.send({ action: 'terminal_open', node_id: nodeId, shell: shell || '', cols: term.cols, rows: term.rows })

    term.onData((data: string) => {
      if (sessionId) ws.send({ action: 'terminal_input', session_id: sessionId, data: btoa(data) })
    })

    const resizeObs = new ResizeObserver(() => {
      fit.fit()
      if (sessionId) ws.send({ action: 'terminal_resize', session_id: sessionId, cols: term.cols, rows: term.rows })
    })
    resizeObs.observe(ref.current)

    return () => {
      offs.forEach(f => f())
      resizeObs.disconnect()
      if (sessionId) ws.send({ action: 'terminal_close', session_id: sessionId })
      term.dispose()
    }
  }, [ws, nodeId, shell])

  return <div ref={ref} style={{ width: '100%', height: '500px', borderRadius: 8, overflow: 'hidden' }} />
}
