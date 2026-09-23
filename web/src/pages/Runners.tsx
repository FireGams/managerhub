import { useEffect, useState } from 'react'
import { api } from '../api'

interface Node { id: string; name: string; online: boolean }

export default function RunnersPage() {
  const [nodes, setNodes] = useState<Node[]>([])
  const [selected, setSelected] = useState<Set<string>>(new Set())
  const [allNodes, setAllNodes] = useState(false)
  const [repoURL, setRepoURL] = useState('https://github.com/FireGams/managerhub')
  const [token, setToken] = useState('')
  const [name, setName] = useState('')
  const [labels, setLabels] = useState('self-hosted,linux')
  const [result, setResult] = useState('')
  const [err, setErr] = useState('')

  useEffect(() => { api<Node[]>('/nodes').then(setNodes).catch(() => {}) }, [])

  function toggle(id: string) {
    setSelected(s => { const n = new Set(s); n.has(id) ? n.delete(id) : n.add(id); return n })
  }

  async function install() {
    setErr(''); setResult('')
    try {
      const body = {
        node_ids: allNodes ? [] : [...selected],
        all_nodes: allNodes,
        repo_url: repoURL,
        token,
        name,
        labels,
      }
      const r = await api<any>('/runners/install', { method: 'POST', body: JSON.stringify(body) })
      setResult(`Installation command sent to ${r.count} machine(s)`)
    } catch (e: any) { setErr(e.message) }
  }

  return (
    <>
      <h2>GitHub Runners — Install</h2>
      <div className="card" style={{ maxWidth: 640 }}>
        <p className="muted" style={{ marginBottom: '1rem' }}>
          Install GitHub Actions runners on selected machines.
          Use <strong>org URL</strong> (<code>https://github.com/ORG</code>) to make runners available to all repos.
        </p>

        <div style={{ marginBottom: '.8rem' }}>
          <label className="muted">Repository or Org URL</label>
          <input value={repoURL} onChange={e => setRepoURL(e.target.value)}
            placeholder="https://github.com/owner/repo or https://github.com/org" />
        </div>

        <div style={{ marginBottom: '.8rem' }}>
          <label className="muted">Registration Token</label>
          <input value={token} onChange={e => setToken(e.target.value)}
            placeholder="Get from Settings → Actions → Runners → New runner" />
        </div>

        <div style={{ display: 'flex', gap: '.8rem', marginBottom: '.8rem' }}>
          <div style={{ flex: 1 }}>
            <label className="muted">Runner name (optional)</label>
            <input value={name} onChange={e => setName(e.target.value)} placeholder="auto-generated" />
          </div>
          <div style={{ flex: 1 }}>
            <label className="muted">Labels</label>
            <input value={labels} onChange={e => setLabels(e.target.value)} />
          </div>
        </div>

        <div style={{ marginBottom: '.8rem' }}>
          <label className="muted">Target machines</label>
          <div style={{ marginTop: '.3rem' }}>
            <label style={{ display: 'flex', alignItems: 'center', gap: '.4rem', marginBottom: '.3rem' }}>
              <input type="checkbox" checked={allNodes} onChange={e => setAllNodes(e.target.checked)} style={{ width: 'auto' }} />
              All online machines
            </label>
            {!allNodes && (
              <div style={{ maxHeight: 150, overflow: 'auto', border: '1px solid var(--border)', borderRadius: 6, padding: '.5rem' }}>
                {nodes.map(n => (
                  <label key={n.id} style={{ display: 'flex', alignItems: 'center', gap: '.4rem', marginBottom: '.2rem' }}>
                    <input type="checkbox" checked={selected.has(n.id)} onChange={() => toggle(n.id)} style={{ width: 'auto' }} />
                    <span className={'dot ' + (n.online ? 'online' : 'offline')} />
                    {n.name}
                  </label>
                ))}
              </div>
            )}
          </div>
        </div>

        {err && <p style={{ color: 'var(--err)', marginBottom: '.5rem' }}>{err}</p>}
        {result && <p style={{ color: 'var(--ok)', marginBottom: '.5rem' }}>{result}</p>}
        <button onClick={install} disabled={!token || (!allNodes && selected.size === 0)}>
          Install Runner{allNodes ? ' on ALL machines' : ` on ${selected.size} machine(s)`}
        </button>
      </div>
    </>
  )
}
