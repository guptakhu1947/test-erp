import { useState, useEffect } from 'react'

function fmt(amount) {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0,
  }).format(amount)
}

export default function Suppliers() {
  const [suppliers, setSuppliers] = useState([])
  const [loading, setLoading]     = useState(true)
  const [error, setError]         = useState(null)
  const [sort, setSort]           = useState({ key: 'total_spend', dir: 'desc' })
  const [search, setSearch]       = useState('')

  useEffect(() => {
    fetch('/api/suppliers')
      .then(r => { if (!r.ok) throw new Error('Failed to load suppliers'); return r.json() })
      .then(setSuppliers)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  function handleSort(key) {
    setSort(prev => ({ key, dir: prev.key === key && prev.dir === 'asc' ? 'desc' : 'asc' }))
  }

  const filtered = suppliers.filter(s =>
    s.name.toLowerCase().includes(search.toLowerCase()) ||
    s.country.toLowerCase().includes(search.toLowerCase())
  )

  const sorted = [...filtered].sort((a, b) => {
    const av = a[sort.key], bv = b[sort.key]
    const cmp = typeof av === 'string' ? av.localeCompare(bv) : av - bv
    return sort.dir === 'asc' ? cmp : -cmp
  })

  const totalSpend    = suppliers.reduce((s, r) => s + r.total_spend,    0)
  const totalOrders   = suppliers.reduce((s, r) => s + r.order_count,    0)
  const totalInvoices = suppliers.reduce((s, r) => s + r.invoice_count,  0)

  function SortIcon({ col }) {
    if (sort.key !== col) return <span className="sort-icon muted">↕</span>
    return <span className="sort-icon active">{sort.dir === 'asc' ? '↑' : '↓'}</span>
  }

  if (loading) return (
    <div className="page-center">
      <div className="spinner" />
      <p className="status-text">Loading suppliers…</p>
    </div>
  )

  if (error) return (
    <div className="page-center"><p className="error-text">Error: {error}</p></div>
  )

  return (
    <div className="content">
      <div className="page-header">
        <div>
          <h1 className="page-title">Suppliers</h1>
          <p className="page-sub">{suppliers.length} suppliers · fiscal year overview</p>
        </div>
      </div>

      <div className="kpi-row">
        <div className="kpi-card">
          <span className="kpi-label">Total Suppliers</span>
          <span className="kpi-value">{suppliers.length}</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Total Orders</span>
          <span className="kpi-value">{totalOrders}</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Total Invoices</span>
          <span className="kpi-value">{totalInvoices}</span>
        </div>
        <div className="kpi-card accent">
          <span className="kpi-label">Total Spend</span>
          <span className="kpi-value">{fmt(totalSpend)}</span>
        </div>
      </div>

      <div className="toolbar">
        <input
          className="search-input"
          placeholder="Search by name or country…"
          value={search}
          onChange={e => setSearch(e.target.value)}
        />
      </div>

      <div className="table-wrapper">
        <table className="supplier-table">
          <thead>
            <tr>
              {[
                { key: 'name',          label: 'Supplier'    },
                { key: 'contact_email', label: 'Email'       },
                { key: 'country',       label: 'Country'     },
                { key: 'order_count',   label: 'Orders'      },
                { key: 'invoice_count', label: 'Invoices'    },
                { key: 'total_spend',   label: 'Total Spend' },
              ].map(col => (
                <th key={col.key} onClick={() => handleSort(col.key)} className="sortable-th">
                  {col.label} <SortIcon col={col.key} />
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {sorted.length === 0 && (
              <tr><td colSpan={6} className="empty-row">No suppliers found.</td></tr>
            )}
            {sorted.map(s => (
              <tr key={s.id} className="data-row">
                <td className="td-name">{s.name}</td>
                <td className="td-email">{s.contact_email}</td>
                <td><span className="badge badge-country">{s.country}</span></td>
                <td className="td-num">{s.order_count}</td>
                <td className="td-num">{s.invoice_count}</td>
                <td className="td-spend">{fmt(s.total_spend)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
