import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import Sparkline, { MeterBar, BatteryIcon } from '../components/Sparkline'

interface Node {
  id: string; name: string; hostname: string; os: string; arch: string
  ip: string; agent_version: string; status: string; online: boolean
  last_seen_at?: string; tags: string[]; city?: string; country?: string; is_local?: boolean
}

interface Metrics {
  cpu_percent: number; cpu_cores: number
  ram_used: number; ram_total: number
  disk_used: number; disk_total: number
  battery_pct?: number; charging?: boolean; network_type?: string
  uptime_sec: number
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

export default function Machines() {
  const [nodes, setNodes] = useState<Node[]>([])
  const [metricsMap, setMetricsMap] = useState<Map<string, Metrics>>(new Map())
  const [historyMap, setHistoryMap] = useState<Map<string, { cpu: number[]; ram: number[]; disk: number[] }>>(new Map())
  const [err, setErr] = useState('')

  useEffect(() => {
    let mounted = true
    async function load() {
      try {
        const ns = await api<Node[]>('/nodes')
        if (!mounted) return
        setNodes(ns)
        for (const n of ns) {
          if (!n.online) continue
          try {
            const r = await api<{ metrics: Metrics | null }>(`/nodes/${n.id}`)
            if (!mounted || !r.metrics) continue
            const m = r.metrics
            setMetricsMap(prev => { const next = new Map(prev); next.set(n.id, m); return next })
            setHistoryMap(prev => {
              const next = new Map(prev)
              const h = next.get(n.id) || { cpu: [], ram: [], disk: [] }
              h.cpu.push(m.cpu_percent || 0)
              h.ram.push(m.ram_total ? (m.ram_used / m.ram_total) * 100 : 0)
              h.disk.push(m.disk_total ? (m.disk_used / m.disk_total) * 100 : 0)
              if (h.cpu && h.cpu.length > 20) h.cpu.shift()
              if (h.ram && h.ram.length > 20) h.ram.shift()
              if (h.disk && h.disk.length > 20) h.disk.shift()
              next.set(n.id, h)
              return next
            })
          } catch {}
        }
      } catch (e: any) { if (mounted) setErr(e.message) }
    }
    load()
    const t = setInterval(load, 5000)
    return () => { mounted = false; clearInterval(t) }
  }, [])

  const onlineCount = (nodes || []).filter(n => n.online).length

  return (
    <>
      <h2>Machines</h2>
      <div style={{ display: 'flex', gap: '1rem', marginBottom: '1.5rem' }}>
        <div className="card" style={{ flex: 1, textAlign: 'center', marginBottom: 0 }}>
          <div className="stat-big">{onlineCount}</div>
          <div className="muted">Online</div>
        </div>
        <div className="card" style={{ flex: 1, textAlign: 'center', marginBottom: 0 }}>
          <div className="stat-big">{(nodes || []).length}</div>
          <div className="muted">Total</div>
        </div>
      </div>
      {err && <p style={{ color: 'var(--err)' }}>{err}</p>}
      <div className="grid">
        {(nodes || []).map(n => {
          const m = metricsMap.get(n.id)
          const h = historyMap.get(n.id) || { cpu: [], ram: [], disk: [] }
          const ramPct = m ? (m.ram_used / m.ram_total) * 100 : 0
          const diskPct = m ? (m.disk_used / m.disk_total) * 100 : 0
          return (
            <Link to={`/machines/${n.id}`} key={n.id} style={{ textDecoration: 'none', color: 'inherit' }}>
              <div className="card machine-card">
                <div style={{ display: 'flex', alignItems: 'center', marginBottom: '.6rem' }}>
                  <strong style={{ fontSize: '1.05rem' }}>{n.name}</strong>
                  {n.is_local && <span className="geo-badge" style={{ marginLeft: '.5rem', background: 'rgba(99,102,241,0.2)', color: 'var(--accent2)' }}>🖥️ This machine</span>}
                  <span className={`status-badge ${n.online ? 'online' : 'offline'}`} style={{ marginLeft: 'auto' }}>
                    <span className={'dot ' + (n.online ? 'online' : 'offline')} />
                    {n.online ? 'Online' : 'Offline'}
                  </span>
                </div>
                <div style={{ display: 'flex', gap: '.4rem', flexWrap: 'wrap', marginBottom: '.8rem' }}>
                  <span className="geo-badge">📍 {n.city && n.country ? `${n.city}, ${n.country}` : n.os}</span>
                  <span className="geo-badge">{n.os}/{n.arch}</span>
                  {m?.battery_pct != null && <BatteryIcon pct={m.battery_pct} charging={m.charging} />}
                </div>

                {m ? (
                  <>
                    <div className="metric-row">
                      <span className="label">CPU</span>
                      <MeterBar pct={m.cpu_percent} variant="cpu" />
                      <span className="value">{m.cpu_percent.toFixed(1)}%</span>
                    </div>
                    <div className="metric-row">
                      <span className="label">RAM</span>
                      <MeterBar pct={ramPct} variant="ram" />
                      <span className="value">{fmtBytes(m.ram_used)}</span>
                    </div>
                    <div className="metric-row">
                      <span className="label">Disk</span>
                      <MeterBar pct={diskPct} variant="disk" />
                      <span className="value">{fmtBytes(m.disk_used)}</span>
                    </div>

                    <div style={{ display: 'flex', gap: '.5rem', marginTop: '.6rem' }}>
                      <div style={{ flex: 1 }}>
                        <div className="muted" style={{ fontSize: '.65rem' }}>CPU</div>
                        <Sparkline data={h.cpu} max={100} color="#6366f1" />
                      </div>
                      <div style={{ flex: 1 }}>
                        <div className="muted" style={{ fontSize: '.65rem' }}>RAM</div>
                        <Sparkline data={h.ram} max={100} color="#c084fc" />
                      </div>
                      <div style={{ flex: 1 }}>
                        <div className="muted" style={{ fontSize: '.65rem' }}>Disk</div>
                        <Sparkline data={h.disk} max={100} color="#fbbf24" />
                      </div>
                    </div>

                    <div className="muted" style={{ marginTop: '.5rem', fontSize: '.75rem' }}>
                      {m.cpu_cores} cores · uptime {fmtUptime(m.uptime_sec)}
                      {m.network_type ? ` · ${m.network_type}` : ''}
                    </div>
                  </>
                ) : (
                  <p className="muted">Waiting for metrics…</p>
                )}
              </div>
            </Link>
          )
        })}
        {(nodes || []).length === 0 && !err && (
          <div className="card" style={{ textAlign: 'center', padding: '3rem' }}>
            <p style={{ fontSize: '2rem', marginBottom: '.5rem' }}>🖥️</p>
            <p className="muted">No machines yet. Go to <strong>Add Machine</strong> to enroll one.</p>
          </div>
        )}
      </div>
    </>
  )
}
