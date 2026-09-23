interface Props {
  data: number[]
  max?: number
  color?: string
  height?: number
  fill?: boolean
}

export default function Sparkline({ data, max = 100, color = '#6366f1', height = 28, fill = true }: Props) {
  if (data.length < 2) {
    return <svg width="100%" height={height} className="sparkline" />
  }
  const w = 100
  const h = height
  const pts = data.map((v, i) => {
    const x = (i / (data.length - 1)) * w
    const y = h - (Math.min(v, max) / max) * (h - 2) - 1
    return `${x},${y}`
  })
  const path = 'M' + pts.join(' L')
  const area = path + ` L${w},${h} L0,${h} Z`
  return (
    <svg width="100%" height={h} className="sparkline" viewBox={`0 0 ${w} ${h}`} preserveAspectRatio="none">
      {fill && <path d={area} fill={color} opacity={0.15} />}
      <path d={path} fill="none" stroke={color} strokeWidth={1.5} strokeLinejoin="round" />
    </svg>
  )
}

export function MeterBar({ pct, variant }: { pct: number; variant: string }) {
  return (
    <div className="meter-modern">
      <div className={`meter-${variant}`} style={{ width: `${Math.min(pct, 100)}%` }} />
    </div>
  )
}

export function BatteryIcon({ pct, charging }: { pct: number; charging?: boolean }) {
  const color = pct > 50 ? 'var(--battery)' : pct > 20 ? 'var(--warn)' : 'var(--err)'
  return (
    <span style={{ display: 'inline-flex', alignItems: 'center', gap: '.3rem' }}>
      <svg width="24" height="12" viewBox="0 0 24 12">
        <rect x="0" y="1" width="20" height="10" rx="2" fill="none" stroke={color} strokeWidth="1.5" />
        <rect x="20" y="3.5" width="3" height="5" rx="1" fill={color} />
        <rect x="2" y="3" width={`${(pct / 100) * 16}`} height="6" rx="1" fill={color} />
      </svg>
      <span style={{ fontSize: '.75rem', color }}>{pct}%{charging ? ' ⚡' : ''}</span>
    </span>
  )
}
