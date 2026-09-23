import { useState, useRef, useEffect } from 'react'
import { api } from '../api'

export default function TerminalPane({ nodeId }: { ws?: any; nodeId: string; shell?: string }) {
  const [output, setOutput] = useState('')
  const [cmd, setCmd] = useState('')
  const [running, setRunning] = useState(false)
  const [history, setHistory] = useState<string[]>([])
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => { endRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [output])

  async function run() {
    if (!cmd || running) return
    const c = cmd
    setCmd('')
    setHistory(h => [...h, c])
    setRunning(true)
    setOutput(o => o + '\n$ ' + c + '\n')
    try {
      const job = await api<any>('/jobs', {
        method: 'POST',
        body: JSON.stringify({
          name: 'shell',
          command: '/bin/bash',
          args: ['-c', c],
          node_id: nodeId,
          timeout_sec: 30,
        }),
      })
      for (let i = 0; i < 30; i++) {
        await new Promise(r => setTimeout(r, 500))
        const j = await api<any>('/jobs/' + job.id)
        if (j.status === 'success' || j.status === 'failed' || j.status === 'cancelled' || j.status === 'timeout') {
          setOutput(o => o + (j.stdout || '') + (j.stderr || '') + '[exit ' + j.exit_code + ']\n')
          break
        }
      }
    } catch (e: any) {
      setOutput(o => o + '[error] ' + e.message + '\n')
    }
    setRunning(false)
  }

  return (
    <div style={{ display: 'flex', flexDirection: 'column', height: '550px' }}>
      <div style={{
        flex: 1,
        background: '#0a0e17',
        borderRadius: '8px 8px 0 0',
        padding: '1rem',
        overflow: 'auto',
        fontFamily: 'monospace',
        fontSize: 13,
        color: '#f1f5f9',
        whiteSpace: 'pre-wrap',
        lineHeight: 1.5,
      }}>
        {output || '\x1b[90mType a command below and press Enter…\x1b[0m'}
        <div ref={endRef} />
      </div>
      <form
        onSubmit={e => { e.preventDefault(); run() }}
        style={{
          display: 'flex', gap: '.5rem', padding: '.5rem',
          background: 'var(--panel2)', borderRadius: '0 0 8px 8px',
          border: '1px solid var(--border)', borderTop: 'none',
        }}
      >
        <span style={{ color: 'var(--accent)', alignSelf: 'center', fontFamily: 'monospace', fontWeight: 700 }}>$</span>
        <input
          type="text"
          value={cmd}
          onChange={e => setCmd(e.target.value)}
          placeholder="ls -la"
          disabled={running}
          autoFocus
          style={{ flex: 1, fontFamily: 'monospace', fontSize: 14 }}
        />
        <button type="submit" disabled={running || !cmd}>
          {running ? '...' : 'Run'}
        </button>
      </form>
    </div>
  )
}
