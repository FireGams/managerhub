import { useRef, useState, useEffect } from 'react'
import { api } from '../api'

interface Msg { role: 'user' | 'assistant'; content: string }

export default function AIAssistant() {
  const [msgs, setMsgs] = useState<Msg[]>([])
  const [input, setInput] = useState('')
  const [busy, setBusy] = useState(false)
  const [model, setModel] = useState('')
  const [models, setModels] = useState<string[]>([])
  const endRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    api<string[]>('/ai/models').then(m => { setModels(m); setModel(m[0]) }).catch(() => {})
  }, [])
  useEffect(() => { endRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [msgs])

  async function send() {
    const text = input.trim()
    if (!text || busy) return
    setInput('')
    setMsgs(m => [...m, { role: 'user', content: text }])
    setBusy(true)
    try {
      const r = await api<{ content: string; model: string }>('/ai/chat', {
        method: 'POST',
        body: JSON.stringify({ message: text, model }),
      })
      setMsgs(m => [...m, { role: 'assistant', content: r.content }])
    } catch (e: any) {
      setMsgs(m => [...m, { role: 'assistant', content: '⚠ ' + e.message }])
    } finally {
      setBusy(false)
    }
  }

  return (
    <>
      <h2>AI Assistant</h2>
      <div style={{ display: 'flex', gap: '.5rem', marginBottom: '.8rem', alignItems: 'center' }}>
        <span className="muted">Model:</span>
        <select value={model} onChange={e => setModel(e.target.value)} style={{ maxWidth: 300 }}>
          {(models || []).map(m => <option key={m} value={m}>{m}</option>)}
        </select>
      </div>
      <div className="card" style={{ height: 'calc(100vh - 260px)', display: 'flex', flexDirection: 'column' }}>
        <div style={{ flex: 1, overflowY: 'auto', padding: '.5rem' }}>
          {msgs.length === 0 && (
            <div className="muted" style={{ textAlign: 'center', marginTop: '3rem' }}>
              <p style={{ fontSize: '2rem', marginBottom: '.5rem' }}>🤖</p>
              <p>Ask anything — scripts, logs analysis, architecture, troubleshooting.</p>
            </div>
          )}
          {(msgs || []).map((m, i) => (
            <div key={i} style={{
              marginBottom: '1rem', padding: '.75rem 1rem', borderRadius: 8,
              background: m.role === 'user' ? 'var(--border)' : 'var(--bg)',
              borderLeft: m.role === 'user' ? '3px solid var(--accent)' : '3px solid var(--ok)',
              whiteSpace: 'pre-wrap', fontSize: 14, lineHeight: 1.5,
            }}>
              <strong style={{ fontSize: 11, color: 'var(--muted)' }}>{m.role === 'user' ? 'You' : 'AI'}</strong>
              <div style={{ marginTop: '.3rem' }}>{m.content}</div>
            </div>
          ))}
          {busy && <p className="muted">Thinking…</p>}
          <div ref={endRef} />
        </div>
        <div style={{ display: 'flex', gap: '.5rem', paddingTop: '.5rem', borderTop: '1px solid var(--border)' }}>
          <input
            placeholder="Describe your problem or ask for a script…"
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && !e.shiftKey && send()}
            disabled={busy}
          />
          <button onClick={send} disabled={busy}>Send</button>
        </div>
      </div>
    </>
  )
}
