import { Navigate, Route, Routes, NavLink, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import Machines from './pages/Machines'
import Jobs from './pages/Jobs'
import Schedules from './pages/Schedules'
import MachineDetail from './pages/MachineDetail'
import AddMachine from './pages/AddMachine'
import AIAssistant from './pages/AIAssistant'
import Settings from './pages/Settings'
import RunnersPage from './pages/Runners'
import Login from './pages/Login'

const tokenKey = 'mh_token'

export function getToken() { return localStorage.getItem(tokenKey) || '' }
export function setToken(t: string) {
  if (t) localStorage.setItem(tokenKey, t)
  else localStorage.removeItem(tokenKey)
}

function Shell({ children, onLogout }: { children: React.ReactNode; onLogout: () => void }) {
  const nav = useNavigate()
  const links = [
    ['/', 'Dashboard'],
    ['/machines', 'Machines'],
    ['/add-machine', 'Add Machine'],
    ['/jobs', 'Jobs'],
    ['/runners', 'GitHub Runners'],
    ['/schedules', 'Schedules'],
    ['/ai', 'AI Assistant'],
    ['/settings', 'Settings'],
  ] as const
  return (
    <div className="layout">
      <aside className="sidebar">
        <h1>ManagerHub</h1>
        {links.map(([to, label]) => (
          <NavLink key={to} to={to} end={to === '/'}>{label}</NavLink>
        ))}
        <button className="secondary" style={{ marginTop: 'auto' }}
          onClick={() => { setToken(''); onLogout(); nav('/login') }}>
          Logout
        </button>
      </aside>
      <main className="main">{children}</main>
    </div>
  )
}

function Dashboard() {
  return (
    <>
      <h2>Dashboard</h2>
      <div className="card">
        <p className="muted">Fleet overview — go to Machines for details.</p>
      </div>
    </>
  )
}

export default function App() {
  const [authed, setAuthed] = useState(!!getToken())
  if (!authed) {
    return (
      <Routes>
        <Route path="*" element={<Login onLogin={() => setAuthed(true)} />} />
      </Routes>
    )
  }
  return (
    <Shell onLogout={() => setAuthed(false)}>
      <Routes>
        <Route path="/" element={<Dashboard />} />
        <Route path="/machines" element={<Machines />} />
        <Route path="/machines/:id" element={<MachineDetail />} />
        <Route path="/add-machine" element={<AddMachine />} />
        <Route path="/ai" element={<AIAssistant />} />
        <Route path="/settings" element={<Settings />} />
        <Route path="/runners" element={<RunnersPage />} />
        <Route path="/jobs" element={<Jobs />} />
        <Route path="/schedules" element={<Schedules />} />
        <Route path="/login" element={<Navigate to="/" />} />
        <Route path="*" element={<Navigate to="/" />} />
      </Routes>
    </Shell>
  )
}
