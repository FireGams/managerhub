import { useState } from 'react'
import { api } from '../api'
import { setToken } from '../App'

export default function Login({ onLogin }: { onLogin: () => void }) {
  const [username, setUsername] = useState('admin')
  const [password, setPassword] = useState('')
  const [err, setErr] = useState('')

  async function submit(e: React.FormEvent) {
    e.preventDefault()
    setErr('')
    try {
      const r = await api<{ token: string }>('/auth/login', {
        method: 'POST',
        body: JSON.stringify({ username, password }),
      })
      setToken(r.token)
      onLogin()
    } catch (ex: any) {
      setErr(ex.message)
    }
  }

  return (
    <div className="login-wrap">
      <form className="card login-card" onSubmit={submit}>
        <h2>ManagerHub</h2>
        <div className="field">
          <label>Username</label>
          <input value={username} onChange={e => setUsername(e.target.value)} autoFocus />
        </div>
        <div className="field">
          <label>Password</label>
          <input type="password" value={password} onChange={e => setPassword(e.target.value)} />
        </div>
        {err && <p style={{ color: 'var(--err)', marginBottom: '.8rem' }}>{err}</p>}
        <button type="submit">Sign in</button>
      </form>
    </div>
  )
}
