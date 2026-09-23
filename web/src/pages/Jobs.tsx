import { useEffect, useState } from 'react'
import { api } from '../api'

interface Job {
  id: string; name: string; type: string; node_id?: string
  command: string; status: string; exit_code?: number
  created_at: string; started_at?: string; finished_at?: string
  stdout?: string; stderr?: string; error?: string
}

export default function Jobs() {
  const [jobs, setJobs] = useState<Job[]>([])
  const [name, setName] = useState('')
  const [command, setCommand] = useState('')
  const [err, setErr] = useState('')
  const [broadcastResult, setBroadcastResult] = useState('')

  const refresh = () => api<Job[]>('/jobs').then(setJobs).catch(e => setErr(e.message))
  useEffect(() => { refresh(); const t = setInterval(refresh, 4000); return () => clearInterval(t) }, [])

  async function create(e: React.FormEvent) {
    e.preventDefault()
    setErr('')
    try {
      await api('/jobs', { method: 'POST', body: JSON.stringify({ name, command, args: [] }) })
      setName(''); setCommand('')
      refresh()
    } catch (ex: any) { setErr(ex.message) }
  }

  async function broadcastToAll() {
    if (!command) return
    setErr(''); setBroadcastResult('')
    try {
      const r = await api<any>('/jobs/multi', {
        method: 'POST',
        body: JSON.stringify({ name: name || 'broadcast', command, all_nodes: true, timeout_sec: 300 }),
      })
      setBroadcastResult(`Sent to ${r.count} machine(s)`)
      refresh()
    } catch (e: any) { setErr(e.message) }
  }

  return (
    <>
      <h2>Jobs</h2>
      <form className="card" onSubmit={create} style={{ display: 'flex', gap: '.6rem', flexWrap: 'wrap' }}>
        <input placeholder="Job name" value={name} onChange={e => setName(e.target.value)} required />
        <input placeholder="Command (e.g. /bin/echo hello)" value={command} onChange={e => setCommand(e.target.value)} required />
        <button type="submit">Run on 1</button>
        <button type="button" className="secondary" onClick={broadcastToAll}>🚀 Broadcast to ALL</button>
      </form>
      {broadcastResult && <p style={{ color: 'var(--ok)', margin: '.5rem 0' }}>{broadcastResult}</p>}
      {err && <p style={{ color: 'var(--err)' }}>{err}</p>}
      <div className="card" style={{ overflowX: 'auto' }}>
        <table>
          <thead><tr><th>Name</th><th>Command</th><th>Status</th><th>Exit</th><th>Created</th></tr></thead>
          <tbody>
            {jobs.map(j => (
              <tr key={j.id}>
                <td>{j.name}</td>
                <td><code>{j.command}</code></td>
                <td>{j.status}</td>
                <td>{j.exit_code ?? '—'}</td>
                <td className="muted">{new Date(j.created_at).toLocaleString()}</td>
              </tr>
            ))}
          </tbody>
        </table>
        {jobs.length === 0 && <p className="muted">No jobs yet.</p>}
      </div>
    </>
  )
}
