import { useState, useRef, useEffect } from 'react'

interface Props {
  title: string
  message: string
  confirmLabel?: string
  danger?: boolean
  onConfirm: () => void
}

export default function ConfirmDialog({ title, message, confirmLabel = 'Delete', danger = true, onConfirm }: Props) {
  const [open, setOpen] = useState(false)
  const ref = useRef<HTMLDialogElement>(null)

  useEffect(() => {
    if (open) ref.current?.showModal()
    else ref.current?.close()
  }, [open])

  return (
    <>
      <button
        className={danger ? 'secondary' : ''}
        style={danger ? { background: 'rgba(248,113,113,0.15)', color: 'var(--err)', padding: '.25rem .6rem', fontSize: '.8rem' } : { padding: '.25rem .6rem', fontSize: '.8rem' }}
        onClick={() => setOpen(true)}
      >
        🗑 {confirmLabel}
      </button>
      <dialog ref={ref} style={{
        background: 'var(--panel)',
        border: '1px solid var(--border)',
        borderRadius: '16px',
        padding: '1.5rem',
        color: 'var(--text)',
        maxWidth: 400,
        width: '90%',
        backdropFilter: 'blur(12px)',
      }}>
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
      </dialog>
    </>
  )
}
