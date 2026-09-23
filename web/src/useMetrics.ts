import { useState, useEffect, useRef } from 'react'
import { api } from './api'

export interface MetricsSnapshot {
  cpu_percent: number
  cpu_cores: number
  ram_used: number
  ram_total: number
  disk_used: number
  disk_total: number
  uptime_sec: number
  battery_pct?: number
  charging?: boolean
  network_type?: string
  city?: string
  country?: string
}

interface MetricsHistory {
  cpu: number[]
  ram: number[]
  disk: number[]
  latest: MetricsSnapshot | null
}

const MAX_POINTS = 30

export function useMetrics(nodeId: string | null, intervalMs = 5000): MetricsHistory {
  const [latest, setLatest] = useState<MetricsSnapshot | null>(null)
  const historyRef = useRef<MetricsHistory>({ cpu: [], ram: [], disk: [], latest: null })

  useEffect(() => {
    if (!nodeId) return
    let mounted = true

    async function fetchMetrics() {
      if (!nodeId) return
      try {
        const r = await api<{ metrics: MetricsSnapshot | null }>(`/nodes/${nodeId}`)
        if (!mounted || !r.metrics) return
        const m = r.metrics
        setLatest(m)
        const h = historyRef.current
        h.latest = m
        if (!h.cpu) h.cpu = []
        if (!h.ram) h.ram = []
        if (!h.disk) h.disk = []
        h.cpu.push(m.cpu_percent || 0)
        h.ram.push(m.ram_total ? (m.ram_used / m.ram_total) * 100 : 0)
        h.disk.push(m.disk_total ? (m.disk_used / m.disk_total) * 100 : 0)
        if (h.cpu.length > MAX_POINTS) h.cpu.shift()
        if (h.ram.length > MAX_POINTS) h.ram.shift()
        if (h.disk.length > MAX_POINTS) h.disk.shift()
      } catch {}
    }

    fetchMetrics()
    const t = setInterval(fetchMetrics, intervalMs)
    return () => { mounted = false; clearInterval(t) }
  }, [nodeId, intervalMs])

  return { ...historyRef.current, latest }
}
