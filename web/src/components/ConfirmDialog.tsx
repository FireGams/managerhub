import { useState } from 'react'

interface Props {
  title: string
  message: string
  confirmLabel?: string
  danger?: boolean
  onConfirm: () => void
}

export default function ConfirmDialog({ title, message, confirmLabel = 'Delete', danger = true, onConfirm }: Props) {
  const [open, setOpen] = useState(false)

  if (open) {
    return (
      <span style={{
        position: 'fixed' as const, inset: 0, zIndex: 1000,
        background: 'rgba(0,0,0,0.6)', backdropFilter: 'blur(4px)',
        display: 'flex', alignItems: 'center', justifyContent: 'center',
      }} onClick={() => setOpen(false)}>
        <div className="card" style={{ maxWidth: 400, width: '90%', margin: 0 }}
          onClick={e => e.stopPropagation()}>
          <h3 style={{ marginBottom: '.8rem' }}>{title}</h3>
          <p className="muted" style={{ marginBottom: '1.2rem' }}>{message}</p>
          <div style={{ display: 'flex', gap: '.5rem', justifyContent: 'flex-end' }}>
            <button className="secondary" onClick={() => setOpen(false)}>Cancel</button>
            <button
              style={{ background: danger ? 'var(--err)' : 'var(--accent)' }}
              onClick={() => { onConfirm(); setOpen(false) }}
            >
              {confirmLabel}
            </button>
          </div>
        </div>
      </span>
    )
  }

  return (
    <button
      className="secondary"
      style={danger ? { background: 'rgba(248,113,113,0.15)', color: 'var(--err)', padding: '.25rem .6rem', fontSize: '.8rem' } : { padding: '.25rem .6rem', fontSize: '.8rem' }}
      onClick={() => setOpen(true)}
    >
      🗑 {confirmLabel}
    </button>
  )
}
