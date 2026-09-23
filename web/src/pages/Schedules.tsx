import { useEffect, useState } from 'react'
import { api } from '../api'

interface Schedule {
  id: string; name: string; cron_expr: string; enabled: boolean
  command: string; last_run_at?: string
}

export default function Schedules() {
  const [list, setList] = useState<Schedule[]>([])
  const [name, setName] = useState('')
  const [cron, setCron] = useState('0 3 * * *')
  const [command, setCommand] = useState('')
  const [err, setErr] = useState('')

  const refresh = () => api<Schedule[]>('/schedules').then(setList).catch(e => setErr(e.message))
  useEffect(() => { refresh() }, [])

  async function create(e: React.FormEvent) {
    e.preventDefault()
    setErr('')
    try {
      await api('/schedules', {
        method: 'POST',
        body: JSON.stringify({ name, cron_expr: cron, command, args: [], job_name: name }),
      })
      setName(''); setCommand('')
      refresh()
    } catch (ex: any) { setErr(ex.message) }
  }

  async function toggle(id: string, enabled: boolean) {
    await api('/schedules/' + id, { method: 'PATCH', body: JSON.stringify({ enabled: !enabled }) })
    refresh()
  }

  return (
    <>
      <h2>Schedules</h2>
      <form className="card" onSubmit={create} style={{ display: 'flex', gap: '.6rem', flexWrap: 'wrap' }}>
        <input placeholder="Name" value={name} onChange={e => setName(e.target.value)} required />
        <input placeholder="Cron (0 3 * * *)" value={cron} onChange={e => setCron(e.target.value)} required style={{ maxWidth: 140 }} />
        <input placeholder="Command" value={command} onChange={e => setCommand(e.target.value)} required />
        <button type="submit">Create</button>
      </form>
      {err && <p style={{ color: 'var(--err)' }}>{err}</p>}
      <div className="card" style={{ overflowX: 'auto' }}>
        <table>
          <thead><tr><th>Name</th><th>Cron</th><th>Command</th><th>Enabled</th><th>Last run</th><th></th></tr></thead>
          <tbody>
            {(list || []).map(s => (
              <tr key={s.id}>
                <td>{s.name}</td>
                <td><code>{s.cron_expr}</code></td>
                <td><code>{s.command}</code></td>
                <td>{s.enabled ? 'yes' : 'no'}</td>
                <td className="muted">{s.last_run_at ? new Date(s.last_run_at).toLocaleString() : '—'}</td>
                <td><button className="secondary" onClick={() => toggle(s.id, s.enabled)}>{s.enabled ? 'Disable' : 'Enable'}</button></td>
              </tr>
            ))}
          </tbody>
        </table>
        {list.length === 0 && <p className="muted">No schedules yet.</p>}
      </div>
    </>
  )
}
