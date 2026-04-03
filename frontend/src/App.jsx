import { useState } from 'react'
import Suppliers from './pages/Suppliers'
import Insights  from './pages/Insights'
import HelpDesk  from './pages/HelpDesk'
import Chat      from './pages/Chat'
import './App.css'

const PAGES = [
  { id: 'suppliers', label: 'Suppliers' },
  { id: 'insights',  label: 'Insights'  },
  { id: 'helpdesk',  label: 'Help Desk' },
  { id: 'glow',      label: '✦ Glow'    },
]

export default function App() {
  const [page, setPage] = useState('suppliers')

  return (
    <div className="app">
      <header className="top-bar">
        <div className="logo">ERP System</div>
        <nav className="nav">
          {PAGES.map(p => (
            <button
              key={p.id}
              className={`nav-link ${page === p.id ? 'active' : ''}${p.id === 'glow' ? ' nav-glow' : ''}`}
              onClick={() => setPage(p.id)}
            >
              {p.label}
            </button>
          ))}
        </nav>
      </header>

      <main className={page === 'glow' ? 'main-chat' : ''}>
        {page === 'suppliers' && <Suppliers />}
        {page === 'insights'  && <Insights  />}
        {page === 'helpdesk'  && <HelpDesk  />}
        {page === 'glow'      && <Chat      />}
      </main>
    </div>
  )
}
