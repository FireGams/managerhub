import { useState, useEffect } from 'react'
import { api } from '../api'
import ConfirmDialog from '../components/ConfirmDialog'

interface Token {
  id: string; label: string; expires_at: string
  used_at?: string; created_by: string; created_at: string
  token?: string
}

export default function AddMachine() {
  const [tokens, setTokens] = useState<Token[]>([])
  const [label, setLabel] = useState('')
  const [newToken, setNewToken] = useState('')
  const [newLabel, setNewLabel] = useState('')
  const [err, setErr] = useState('')
  const [copied, setCopied] = useState(false)
  const [controllerURL, setControllerURL] = useState(location.origin)

  const refresh = () => api<Token[]>('/enroll-tokens').then(setTokens).catch(() => {})
  useEffect(() => {
    refresh()
    api<{public_url: string}>('/config').then(c => {
      if (c.public_url) setControllerURL(c.public_url)
    }).catch(() => {})
  }, [])

  async function create() {
    setErr(''); setNewToken('')
    try {
      const t = await api<Token>('/enroll-tokens', { method: 'POST', body: JSON.stringify({ label }) })
      setNewToken(t.token || '')
      setNewLabel(label)
      setLabel('')
      refresh()
    } catch (e: any) { setErr(e.message) }
  }

  function copyCmd() {
    if (!newToken) return
    const cmd = `curl -fsSL https://raw.githubusercontent.com/FireGams/managerhub/refs/heads/main/scripts/install.sh?t=${Date.now()} | bash -s -- --controller ${controllerURL} --token ${newToken} --name \"${newLabel || 'my-machine'}\"`
    navigator.clipboard.writeText(cmd)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  async function deleteToken(id: string) {
    setErr('')
    try {
      await api('/enroll-tokens/' + id, { method: 'DELETE' })
      refresh()
    } catch (e: any) { setErr('Delete failed: ' + e.message) }
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
{`curl -fsSL https://raw.githubusercontent.com/FireGams/managerhub/refs/heads/main/scripts/install.sh?t=${Date.now()} | bash -s -- --controller ${controllerURL} --token ${newToken} --name \"${newLabel || 'my-machine'}\"`}
            </pre>
            <p className="muted" style={{ marginTop: '.5rem' }}>Works on Linux, macOS and Windows. Installs the agent as a system service with auto-update.</p>
            <button className="secondary" onClick={copyCmd}>{copied ? 'Copied!' : 'Copy install command'}</button>
            <p className="muted" style={{ marginTop: '.5rem' }}>Save this token — it will not be shown again. The command downloads the agent, enrolls it, and installs the service.</p>
          </div>
        )}
      </div>

      <div className="card" style={{ overflowX: 'auto' }}>
        <h3 style={{ marginBottom: '.8rem' }}>Enrollment tokens</h3>
        <table>
          <thead><tr><th>Label</th><th>Created by</th><th>Expires</th><th>Status</th><th>Actions</th></tr></thead>
          <tbody>
            {(tokens || []).map(t => (
              <tr key={t.id}>
                <td>{t.label || '—'}</td>
                <td>{t.created_by}</td>
                <td className="muted">{new Date(t.expires_at).toLocaleString()}</td>
                <td>{t.used_at ? '✓ used' : 'pending'}</td>
                <td>
                  <ConfirmDialog
                    title="Delete token"
                    message={`Delete enrollment token "${t.label || 'untitled'}"? This cannot be undone.`}
                    confirmLabel="Delete"
                    onConfirm={() => deleteToken(t.id)}
                  />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {(tokens || []).length === 0 && <p className="muted">No tokens yet.</p>}
      </div>
    </>
  )
}
