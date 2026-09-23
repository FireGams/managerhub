import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../api'
import { ManagerHubWS } from '../ws'
import TerminalPane from '../components/Terminal'

const tabs = ['Overview', 'Terminal', 'Services', 'Jobs', 'Runners'] as const
type Tab = typeof tabs[number]

export default function MachineDetail() {
  const { id } = useParams<{ id: string }>()
  const [tab, setTab] = useState<Tab>('Overview')
  const [info, setInfo] = useState<any>(null)
  const wsRef = useRef<ManagerHubWS | null>(null)

  useEffect(() => {
    if (!id) return
    api('/nodes/' + id).then(setInfo).catch(() => {})
    const ws = new ManagerHubWS(localStorage.getItem('mh_token') || '')
    ws.connect()
    wsRef.current = ws
    return () => { ws.close(); wsRef.current = null }
  }, [id])

  if (!id) return <p>No node selected</p>
  const node = info?.node || {}

  return (
    <>
      <h2 style={{ display: 'flex', alignItems: 'center', gap: '.5rem' }}>
        <span className={'dot ' + (info?.online ? 'online' : 'offline')} />
        {node.name || id}
      </h2>
      <div style={{ display: 'flex', gap: '.5rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
        {tabs.map(t => (
          <button key={t} className={t === tab ? '' : 'secondary'} onClick={() => setTab(t)}>{t}</button>
        ))}
      </div>
      {tab === 'Overview' && <Overview info={info} />}
      {tab === 'Terminal' && <TerminalPane ws={wsRef.current!} nodeId={id} />}
      {tab === 'Services' && <Services ws={wsRef.current} nodeId={id} />}
      {tab === 'Jobs' && <NodeJobs ws={wsRef.current} nodeId={id} />}
      {tab === 'Runners' && <Runners ws={wsRef.current} nodeId={id} />}
    </>
  )
}

function Overview({ info }: { info: any }) {
  const n = info?.node
  const m = info?.metrics
  if (!n) return <p className="muted">Loading…</p>
  const pct = (a: number, b: number) => b > 0 ? Math.round((a / b) * 100) : 0
  return (
    <div className="grid">
      <div className="card">
        <h3>System</h3>
        <p>Hostname: {n.hostname}</p>
        <p>OS: {n.os} / {n.arch}</p>
        <p>IP: {n.ip}</p>
        <p>Agent: {n.agent_version}</p>
        <p>Status: {info.online ? 'Online' : 'Offline'}</p>
      </div>
      {m && (
        <>
          <div className="card">
            <h3>CPU</h3>
            <p>{m.cpu_percent?.toFixed(1)}% of {m.cpu_cores} cores</p>
            <div className="meter"><div style={{ width: m.cpu_percent + '%' }} /></div>
            <p className="muted">Load: {m.load1?.toFixed(2)}</p>
          </div>
          <div className="card">
            <h3>RAM</h3>
            <p>{fmtBytes(m.ram_used)} / {fmtBytes(m.ram_total)} ({pct(m.ram_used, m.ram_total)}%)</p>
            <div className="meter"><div style={{ width: pct(m.ram_used, m.ram_total) + '%' }} /></div>
          </div>
          <div className="card">
            <h3>Disk</h3>
            <p>{fmtBytes(m.disk_used)} / {fmtBytes(m.disk_total)} ({pct(m.disk_used, m.disk_total)}%)</p>
            <div className="meter"><div style={{ width: pct(m.disk_used, m.disk_total) + '%' }} /></div>
          </div>
        </>
      )}
    </div>
  )
}

function Services({ ws, nodeId }: { ws: ManagerHubWS | null; nodeId: string }) {
  const [svcs, setSvcs] = useState<any[]>([])
  useEffect(() => {
    if (!ws) return
    const off = ws.on('services_list_result', (d: any) => setSvcs(d.services || []))
    ws.send({ action: 'services_list', node_id: nodeId })
    return () => { off() }
  }, [ws, nodeId])
  return (
    <div className="card" style={{ overflowX: 'auto' }}>
      <table>
        <thead><tr><th>Name</th><th>State</th><th>Actions</th></tr></thead>
        <tbody>
          {svcs.slice(0, 50).map((s: any) => (
            <tr key={s.name}>
              <td>{s.name}</td>
              <td>{s.state}{s.sub ? ' / ' + s.sub : ''}</td>
              <td>
                {['start', 'stop', 'restart'].map(a => (
                  <button key={a} className="secondary" style={{ marginRight: '.3rem', padding: '.25rem .5rem' }}
                    onClick={() => ws?.send({ action: 'services_action', node_id: nodeId, name: a, data: s.name })}>{a}</button>
                ))}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
      {svcs.length === 0 && <p className="muted">No services found.</p>}
    </div>
  )
}

function NodeJobs({ ws, nodeId }: { ws: ManagerHubWS | null; nodeId: string }) {
  const [output, setOutput] = useState('')
  const [cmd, setCmd] = useState('')

  useEffect(() => {
    if (!ws) return
    const offs = [
      ws.on('job_output', (d: any) => setOutput(o => o + d.data)),
      ws.on('job_result', (d: any) => setOutput(o => o + `\n[exit ${d.exit_code} — ${d.status}]\n`)),
    ]
    return () => { offs.forEach(f => f()) }
  }, [ws])

  function run() {
    if (!ws || !cmd) return
    setOutput('')
    ws.send({ action: 'job_create', node_id: nodeId, name: 'manual', command: cmd, args: [] })
  }

  return (
    <>
      <div className="card" style={{ display: 'flex', gap: '.5rem', flexWrap: 'wrap' }}>
        <input placeholder="Command to run…" value={cmd} onChange={e => setCmd(e.target.value)}
          onKeyDown={e => e.key === 'Enter' && run()} />
        <button onClick={run}>Run</button>
      </div>
      <div className="card">
        <h3>Output</h3>
        <pre style={{ background: '#0a0e14', padding: '1rem', borderRadius: 6, minHeight: 120, maxHeight: 360, overflow: 'auto', fontSize: 13 }}>
          {output || 'Run a command to see output here…'}
        </pre>
      </div>
    </>
  )
}

function Runners({ ws, nodeId }: { ws: ManagerHubWS | null; nodeId: string }) {
  const [runners, setRunners] = useState<any[]>([])
  useEffect(() => {
    if (!ws) return
    const off = ws.on('runners_list_result', (d: any) => setRunners(d.runners || []))
    ws.send({ action: 'runners_list', node_id: nodeId })
    return () => { off() }
  }, [ws, nodeId])
  return (
    <div className="card">
      <table>
        <thead><tr><th>Name</th><th>Repo</th><th>Status</th><th>Service</th></tr></thead>
        <tbody>
          {runners.map((r: any) => (
            <tr key={r.name}>
              <td>{r.name}</td>
              <td>{r.repo || r.org || '—'}</td>
              <td>{r.status}</td>
              <td>{r.service || 'none'}{r.service_on ? ' (on)' : ''}</td>
            </tr>
          ))}
        </tbody>
      </table>
      {runners.length === 0 && <p className="muted">No GitHub runners detected.</p>}
    </div>
  )
}

function fmtBytes(n: number) {
  if (n > 1e12) return (n / 1e12).toFixed(1) + ' TB'
  if (n > 1e9) return (n / 1e9).toFixed(1) + ' GB'
  if (n > 1e6) return (n / 1e6).toFixed(0) + ' MB'
  return n + ' B'
}
