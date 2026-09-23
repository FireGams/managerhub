import { useEffect, useState } from 'react'
import { api } from '../api'

export default function Settings() {
  const [apiKey, setApiKey] = useState('')
  const [model, setModel] = useState('openai/gpt-4o-mini')
  const [maxTokens, setMaxTokens] = useState(2048)
  const [temperature, setTemperature] = useState(0.7)
  const [models, setModels] = useState<string[]>([])
  const [msg, setMsg] = useState('')
  const [err, setErr] = useState('')

  useEffect(() => {
    api<any>('/ai/config').then(c => {
      setApiKey(c.api_key || '')
      setModel(c.model || 'openai/gpt-4o-mini')
      setMaxTokens(c.max_tokens || 2048)
      setTemperature(c.temperature ?? 0.7)
    }).catch(() => {})
    api<string[]>('/ai/models').then(setModels).catch(() => {})
  }, [])

  async function save() {
    setMsg(''); setErr('')
    try {
      await api('/ai/config', {
        method: 'POST',
        body: JSON.stringify({ api_key: apiKey, model, max_tokens: maxTokens, temperature }),
      })
      setMsg('Configuration saved.')
    } catch (e: any) { setErr(e.message) }
  }

  return (
    <>
      <h2>Settings</h2>
      <div className="card" style={{ maxWidth: 560 }}>
        <h3 style={{ marginBottom: '1rem' }}>OpenRouter AI</h3>
        <div style={{ marginBottom: '.8rem' }}>
          <label className="muted">API Key</label>
          <input type="password" value={apiKey} onChange={e => setApiKey(e.target.value)}
            placeholder="sk-or-v1-..." />
        </div>
        <div style={{ marginBottom: '.8rem' }}>
          <label className="muted">Model</label>
          <select value={model} onChange={e => setModel(e.target.value)}>
            {models.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
        </div>
        <div style={{ display: 'flex', gap: '.8rem', marginBottom: '.8rem' }}>
          <div style={{ flex: 1 }}>
            <label className="muted">Max tokens</label>
            <input type="number" value={maxTokens} onChange={e => setMaxTokens(+e.target.value)} />
          </div>
          <div style={{ flex: 1 }}>
            <label className="muted">Temperature</label>
            <input type="number" step={0.1} min={0} max={2} value={temperature}
              onChange={e => setTemperature(+e.target.value)} />
          </div>
        </div>
        {msg && <p style={{ color: 'var(--ok)', marginBottom: '.5rem' }}>{msg}</p>}
        {err && <p style={{ color: 'var(--err)', marginBottom: '.5rem' }}>{err}</p>}
        <button onClick={save}>Save</button>
      </div>
    </>
  )
}
