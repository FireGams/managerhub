import { useState, useEffect } from 'react'
import { api } from '../api'

interface Token {
  id: string; label: string; expires_at: string
  used_at?: string; created_by: string; created_at: string
  token?: string
}

export default function AddMachine() {
  const [tokens, setTokens] = useState<Token[]>([])
  const [label, setLabel] = useState('')
  const [newToken, setNewToken] = useState('')
  const [err, setErr] = useState('')
  const [copied, setCopied] = useState(false)

  const refresh = () => api<Token[]>('/enroll-tokens').then(setTokens).catch(() => {})
  useEffect(() => { refresh() }, [])

  async function create() {
    setErr(''); setNewToken('')
    try {
      const t = await api<Token>('/enroll-tokens', { method: 'POST', body: JSON.stringify({ label }) })
      setNewToken(t.token || '')
      setLabel('')
      refresh()
    } catch (e: any) { setErr(e.message) }
  }

  function copyCmd() {
    if (!newToken) return
    const cmd = `export MH_AUTO_UPDATE=true\nexport MH_CONTROLLER_URL=${location.origin}\nexport MH_ENROLL_TOKEN=${newToken}\n./bin/agent`
    navigator.clipboard.writeText(cmd)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <>
      <h2>Add Machine</h2>

      <div className="card" style={{ maxWidth: 640, marginBottom: '1.5rem' }}>
        <h3 style={{ marginBottom: '.8rem' }}>Create enrollment token</h3>
        <div style={{ display: 'flex', gap: '.5rem' }}>
          <input placeholder="Label (e.g. laptop, server, VDS)" value={label} onChange={e => setLabel(e.target.value)} />
          <button onClick={create}>Generate</button>
        </div>
        {err && <p style={{ color: 'var(--err)', marginTop: '.5rem' }}>{err}</p>}
        {newToken && (
          <div style={{ marginTop: '1rem', padding: '1rem', background: 'var(--bg)', borderRadius: 8 }}>
            <p className="muted" style={{ marginBottom: '.5rem' }}>Copy this command to the target machine:</p>
            <pre style={{ fontSize: 13, whiteSpace: 'pre-wrap', wordBreak: 'break-all' }}>
{`export MH_AUTO_UPDATE=true
export MH_CONTROLLER_URL=${location.origin}
export MH_ENROLL_TOKEN=${newToken}
./bin/agent`}
            </pre>
            <button className="secondary" onClick={copyCmd}>{copied ? 'Copied!' : 'Copy command'}</button>
            <p className="muted" style={{ marginTop: '.5rem' }}>Save this token — it will not be shown again.</p>
          </div>
        )}
      </div>

      <div className="card" style={{ overflowX: 'auto' }}>
        <h3 style={{ marginBottom: '.8rem' }}>Enrollment tokens</h3>
        <table>
          <thead><tr><th>Label</th><th>Created by</th><th>Expires</th><th>Status</th></tr></thead>
          <tbody>
            {tokens.map(t => (
              <tr key={t.id}>
                <td>{t.label || '—'}</td>
                <td>{t.created_by}</td>
                <td className="muted">{new Date(t.expires_at).toLocaleString()}</td>
                <td>{t.used_at ? '✓ used' : 'pending'}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {tokens.length === 0 && <p className="muted">No tokens yet.</p>}
      </div>
    </>
  )
}
