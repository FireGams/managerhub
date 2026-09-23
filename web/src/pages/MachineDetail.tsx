import { useEffect, useRef, useState } from 'react'
import { useParams } from 'react-router-dom'
import { api } from '../api'
import { ManagerHubWS } from '../ws'
import TerminalPane from '../components/Terminal'
import Sparkline, { MeterBar, BatteryIcon } from '../components/Sparkline'
import ConfirmDialog from '../components/ConfirmDialog'
import { useNavigate } from 'react-router-dom'
import { useMetrics } from '../useMetrics'

const tabs = ['Overview', 'Terminal', 'Services', 'Docker', 'Jobs', 'Runners'] as const
type Tab = typeof tabs[number]

export default function MachineDetail() {
  const { id } = useParams<{ id: string }>()
  const nav = useNavigate()
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
        <span className={`status-badge ${info?.online ? 'online' : 'offline'}`} style={{ marginLeft: '.5rem' }}>
          {info?.online ? 'Online' : 'Offline'}
        </span>
        {node.city && node.country && (
          <span className="geo-badge" style={{ marginLeft: '.5rem' }}>📍 {node.city}, {node.country}</span>
        )}
        <span style={{ marginLeft: 'auto' }}>
          <ConfirmDialog
            title="Delete machine"
            message={`Remove "${node.name}" from ManagerHub? The agent will need to re-enroll.`}
            confirmLabel="Delete machine"
            onConfirm={async () => {
              try {
                await api('/nodes/' + id, { method: 'DELETE' })
                nav('/machines')
              } catch (e: any) {
                alert('Delete failed: ' + e.message)
              }
            }}
          />
        </span>
      </h2>
      <div style={{ display: 'flex', gap: '.5rem', marginBottom: '1rem', flexWrap: 'wrap' }}>
        {tabs.map(t => (
          <button key={t} className={t === tab ? '' : 'secondary'} onClick={() => setTab(t)}>{t}</button>
        ))}
      </div>
      {tab === 'Overview' && <Overview nodeId={id} info={info} />}
      {tab === 'Terminal' && <TerminalPane ws={wsRef.current!} nodeId={id} />}
      {tab === 'Services' && <Services ws={wsRef.current} nodeId={id} />}
      {tab === 'Docker' && <DockerTab ws={wsRef.current} nodeId={id} />}
      {tab === 'Jobs' && <NodeJobs ws={wsRef.current} nodeId={id} />}
      {tab === 'Runners' && <Runners ws={wsRef.current} nodeId={id} />}
    </>
  )
}

function Overview({ nodeId, info }: { nodeId: string; info: any }) {
  const n = info?.node
  const { latest: m, cpu, ram, disk } = useMetrics(nodeId, 5000)
  if (!n) return <p className="muted">Loading…</p>

  const ramPct = m ? (m.ram_used / m.ram_total) * 100 : 0
  const diskPct = m ? (m.disk_used / m.disk_total) * 100 : 0

  return (
    <>
      <div className="chart-grid" style={{ marginBottom: '1rem' }}>
        <div className="card" style={{ marginBottom: 0 }}>
          <h3 style={{ fontSize: '.8rem', color: 'var(--muted)', marginBottom: '.5rem' }}>CPU</h3>
          <div className="stat-big">{m?.cpu_percent?.toFixed(1) || '—'}%</div>
          <MeterBar pct={m?.cpu_percent || 0} variant="cpu" />
          <div className="muted" style={{ marginTop: '.3rem' }}>{m?.cpu_cores || n.cpu_cores} cores</div>
          <div style={{ marginTop: '.5rem' }}><Sparkline data={cpu} max={100} color="#6366f1" height={40} /></div>
        </div>
        <div className="card" style={{ marginBottom: 0 }}>
          <h3 style={{ fontSize: '.8rem', color: 'var(--muted)', marginBottom: '.5rem' }}>RAM</h3>
          <div className="stat-big">{ramPct.toFixed(0)}%</div>
          <MeterBar pct={ramPct} variant="ram" />
          <div className="muted" style={{ marginTop: '.3rem' }}>{fmtBytes(m?.ram_used || 0)} / {fmtBytes(m?.ram_total || 0)}</div>
          <div style={{ marginTop: '.5rem' }}><Sparkline data={ram} max={100} color="#c084fc" height={40} /></div>
        </div>
        <div className="card" style={{ marginBottom: 0 }}>
          <h3 style={{ fontSize: '.8rem', color: 'var(--muted)', marginBottom: '.5rem' }}>Disk</h3>
          <div className="stat-big">{diskPct.toFixed(0)}%</div>
          <MeterBar pct={diskPct} variant="disk" />
          <div className="muted" style={{ marginTop: '.3rem' }}>{fmtBytes(m?.disk_used || 0)} / {fmtBytes(m?.disk_total || 0)}</div>
          <div style={{ marginTop: '.5rem' }}><Sparkline data={disk} max={100} color="#fbbf24" height={40} /></div>
        </div>
        {m?.battery_pct != null && (
          <div className="card" style={{ marginBottom: 0 }}>
            <h3 style={{ fontSize: '.8rem', color: 'var(--muted)', marginBottom: '.5rem' }}>Battery</h3>
            <div className="stat-big">{m.battery_pct}%</div>
            <MeterBar pct={m.battery_pct} variant="battery" />
            <div style={{ marginTop: '.3rem' }}><BatteryIcon pct={m.battery_pct} charging={m.charging} /></div>
            {m.network_type && <div className="muted" style={{ marginTop: '.3rem' }}>📶 {m.network_type}</div>}
          </div>
        )}
      </div>

      <div className="card">
        <h3 style={{ marginBottom: '.8rem' }}>System Info</h3>
        <div className="chart-grid">
          <div>
            <p className="muted">Hostname</p>
            <p>{n.hostname}</p>
          </div>
          <div>
            <p className="muted">OS / Arch</p>
            <p>{n.os} / {n.arch}</p>
          </div>
          <div>
            <p className="muted">IP</p>
            <p>{n.ip || '—'}</p>
          </div>
          <div>
            <p className="muted">Agent version</p>
            <p>{n.agent_version}</p>
          </div>
          <div>
            <p className="muted">Uptime</p>
            <p>{fmtUptime(m?.uptime_sec || 0)}</p>
          </div>
          <div>
            <p className="muted">Location</p>
            <p>{n.city && n.country ? `${n.city}, ${n.country}` : 'Unknown'}</p>
          </div>
          {n.tags?.length > 0 && (
            <div>
              <p className="muted">Tags</p>
              <p>{n.tags.join(', ')}</p>
            </div>
          )}
        </div>
      </div>
    </>
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

function DockerTab({ ws, nodeId }: { ws: ManagerHubWS | null; nodeId: string }) {
  const [containers, setContainers] = useState<any[]>([])
  const [available, setAvailable] = useState(false)
  const [logs, setLogs] = useState('')
  useEffect(() => {
    if (!ws) return
    const offs = [
      ws.on('docker_list_result', (d: any) => {
        setAvailable(d.available)
        setContainers(d.containers || [])
      }),
      ws.on('docker_action_result', (d: any) => {
        if (d.action === 'logs') setLogs(d.logs || d.error || '')
        else ws?.send({ action: 'docker_list', node_id: nodeId })
      }),
    ]
    ws.send({ action: 'docker_list', node_id: nodeId })
    return () => { offs.forEach(f => f()) }
  }, [ws, nodeId])

  if (!available) return <div className="card"><p className="muted">Docker not available on this machine.</p></div>
  return (
    <>
      <div className="card" style={{ overflowX: 'auto' }}>
        <table>
          <thead><tr><th>Name</th><th>Image</th><th>State</th><th>Actions</th></tr></thead>
          <tbody>
            {containers.map((c: any) => (
              <tr key={c.id}>
                <td>{c.name}</td>
                <td><code>{c.image}</code></td>
                <td>{c.state}</td>
                <td>
                  {['start', 'stop', 'restart', 'logs'].map(a => (
                    <button key={a} className="secondary" style={{ marginRight: '.3rem', padding: '.25rem .5rem' }}
                      onClick={() => ws?.send({ action: 'docker_action', node_id: nodeId, name: a, data: c.name })}>{a}</button>
                  ))}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
        {containers.length === 0 && <p className="muted">No containers.</p>}
      </div>
      {logs && (
        <div className="card">
          <h3>Logs</h3>
          <pre style={{ background: '#0a0e14', padding: '1rem', borderRadius: 6, maxHeight: 300, overflow: 'auto', fontSize: 12 }}>{logs}</pre>
          <button className="secondary" onClick={() => setLogs('')}>Clear</button>
        </div>
      )}
    </>
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

function fmtUptime(sec: number) {
  const d = Math.floor(sec / 86400)
  const h = Math.floor((sec % 86400) / 3600)
  const m = Math.floor((sec % 3600) / 60)
  if (d > 0) return `${d}d ${h}h`
  if (h > 0) return `${h}h ${m}m`
  return `${m}m`
}
