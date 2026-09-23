import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'

interface Node {
  id: string; name: string; hostname: string; os: string; arch: string
  ip: string; agent_version: string; status: string; online: boolean
  last_seen_at?: string; tags: string[]
}

function fmt(n: number) {
  if (n > 1e12) return (n / 1e12).toFixed(1) + ' TB'
  if (n > 1e9) return (n / 1e9).toFixed(0) + ' GB'
  if (n > 1e6) return (n / 1e6).toFixed(0) + ' MB'
  return n + ' B'
}

export default function Machines() {
  const [nodes, setNodes] = useState<Node[]>([])
  const [err, setErr] = useState('')

  useEffect(() => {
    api<Node[]>('/nodes').then(setNodes).catch(e => setErr(e.message))
    const t = setInterval(() => {
      api<Node[]>('/nodes').then(setNodes).catch(() => {})
    }, 5000)
    return () => clearInterval(t)
  }, [])

  return (
    <>
      <h2>Machines</h2>
      {err && <p style={{ color: 'var(--err)' }}>{err}</p>}
      <div className="grid">
        {nodes.map(n => (
          <Link to={`/machines/${n.id}`} key={n.id} style={{ textDecoration: 'none', color: 'inherit' }}>
          <div className="card">
            <div style={{ display: 'flex', alignItems: 'center', marginBottom: '.5rem' }}>
              <span className={'dot ' + (n.online ? 'online' : 'offline')} />
              <strong>{n.name}</strong>
              <span className="muted" style={{ marginLeft: 'auto' }}>{n.online ? 'Online' : 'Offline'}</span>
            </div>
            <p className="muted">{n.hostname} · {n.os}/{n.arch}</p>
            <p className="muted">{n.ip} · agent {n.agent_version}</p>
            {n.last_seen_at && <p className="muted">last seen {new Date(n.last_seen_at).toLocaleString()}</p>}
            {n.tags?.length > 0 && (
              <p className="muted">tags: {n.tags.join(', ')}</p>
            )}
          </div>
          </Link>
        ))}
        {nodes.length === 0 && !err && <p className="muted">No machines registered yet.</p>}
      </div>
    </>
  )
}

export { fmt }
