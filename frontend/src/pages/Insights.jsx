import { useState, useEffect } from 'react'
import {
  ResponsiveContainer,
  LineChart, Line,
  BarChart, Bar,
  PieChart, Pie, Cell,
  XAxis, YAxis, CartesianGrid, Tooltip, Legend,
} from 'recharts'

const COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#ec4899']
const PIE_COLORS = { completed: '#10b981', pending: '#f59e0b', paid: '#3b82f6' }

function fmt(v) {
  if (v >= 1000) return `$${(v / 1000).toFixed(1)}k`
  return `$${v}`
}

function fmtFull(v) {
  return new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 0 }).format(v)
}

function RecommendationCard({ rec }) {
  const styles = {
    warning: { border: '#fbbf24', bg: '#fffbeb', icon: '⚠️', label: 'Warning'  },
    info:    { border: '#60a5fa', bg: '#eff6ff', icon: 'ℹ️', label: 'Insight'  },
    success: { border: '#34d399', bg: '#f0fdf4', icon: '✓',  label: 'Positive' },
  }
  const s = styles[rec.type] || styles.info
  return (
    <div className="rec-card" style={{ borderLeftColor: s.border, background: s.bg }}>
      <div className="rec-header">
        <span className="rec-icon">{s.icon}</span>
        <span className="rec-type" style={{ color: s.border }}>{s.label}</span>
        <span className="rec-title">{rec.title}</span>
      </div>
      <p className="rec-desc">{rec.description}</p>
    </div>
  )
}

function CustomTooltip({ active, payload, label }) {
  if (!active || !payload?.length) return null
  return (
    <div className="chart-tooltip">
      <p className="tooltip-label">{label}</p>
      {payload.map((p, i) => (
        <p key={i} style={{ color: p.color }}>
          {p.name}: {typeof p.value === 'number' && p.value > 100 ? fmtFull(p.value) : p.value}
        </p>
      ))}
    </div>
  )
}

export default function Insights() {
  const [data, setData]     = useState(null)
  const [loading, setLoading] = useState(true)
  const [error, setError]   = useState(null)

  useEffect(() => {
    fetch('/api/insights')
      .then(r => { if (!r.ok) throw new Error('Failed to load insights'); return r.json() })
      .then(setData)
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  if (loading) return (
    <div className="page-center">
      <div className="spinner" />
      <p className="status-text">Analysing supplier data…</p>
    </div>
  )
  if (error) return (
    <div className="page-center"><p className="error-text">Error: {error}</p></div>
  )

  const totalSpend = data.spend_by_supplier.reduce((s, r) => s + r.spend, 0)

  return (
    <div className="content">
      <div className="page-header">
        <div>
          <h1 className="page-title">Insights Dashboard</h1>
          <p className="page-sub">Automated analysis · trend visualizations · strategic recommendations</p>
        </div>
      </div>

      {/* ── KPIs ── */}
      <div className="kpi-row">
        <div className="kpi-card accent">
          <span className="kpi-label">Total Spend</span>
          <span className="kpi-value">{fmtFull(totalSpend)}</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Top Supplier</span>
          <span className="kpi-value" style={{ fontSize: '1.1rem' }}>
            {data.spend_by_supplier[0]?.name ?? '—'}
          </span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Countries Covered</span>
          <span className="kpi-value">{data.spend_by_country.length}</span>
        </div>
        <div className="kpi-card">
          <span className="kpi-label">Recommendations</span>
          <span className="kpi-value">{data.recommendations.length}</span>
        </div>
      </div>

      {/* ── Charts grid ── */}
      <div className="charts-grid">

        {/* Monthly spend trend */}
        <div className="chart-card wide">
          <h2 className="chart-title">Monthly Spend Trend</h2>
          <p className="chart-sub">Invoice amounts over time</p>
          <ResponsiveContainer width="100%" height={240}>
            <LineChart data={data.monthly_spend} margin={{ top: 10, right: 20, left: 10, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="month" tick={{ fontSize: 12 }} />
              <YAxis tickFormatter={fmt} tick={{ fontSize: 12 }} width={55} />
              <Tooltip content={<CustomTooltip />} />
              <Line
                type="monotone"
                dataKey="spend"
                name="Spend"
                stroke="#3b82f6"
                strokeWidth={2.5}
                dot={{ r: 4, fill: '#3b82f6' }}
                activeDot={{ r: 6 }}
              />
            </LineChart>
          </ResponsiveContainer>
        </div>

        {/* Spend by supplier */}
        <div className="chart-card wide">
          <h2 className="chart-title">Spend by Supplier</h2>
          <p className="chart-sub">Total invoiced amount per supplier</p>
          <ResponsiveContainer width="100%" height={240}>
            <BarChart data={data.spend_by_supplier} margin={{ top: 10, right: 20, left: 10, bottom: 0 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" />
              <XAxis dataKey="name" tick={{ fontSize: 11 }} />
              <YAxis tickFormatter={fmt} tick={{ fontSize: 12 }} width={55} />
              <Tooltip content={<CustomTooltip />} />
              <Bar dataKey="spend" name="Spend" radius={[4, 4, 0, 0]}>
                {data.spend_by_supplier.map((_, i) => (
                  <Cell key={i} fill={COLORS[i % COLORS.length]} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Spend by country */}
        <div className="chart-card">
          <h2 className="chart-title">Spend by Country</h2>
          <p className="chart-sub">Geographic distribution</p>
          <ResponsiveContainer width="100%" height={220}>
            <BarChart
              layout="vertical"
              data={data.spend_by_country}
              margin={{ top: 5, right: 20, left: 10, bottom: 0 }}
            >
              <CartesianGrid strokeDasharray="3 3" stroke="#f0f0f0" horizontal={false} />
              <XAxis type="number" tickFormatter={fmt} tick={{ fontSize: 12 }} />
              <YAxis type="category" dataKey="country" tick={{ fontSize: 12 }} width={70} />
              <Tooltip content={<CustomTooltip />} />
              <Bar dataKey="spend" name="Spend" radius={[0, 4, 4, 0]}>
                {data.spend_by_country.map((_, i) => (
                  <Cell key={i} fill={COLORS[i % COLORS.length]} />
                ))}
              </Bar>
            </BarChart>
          </ResponsiveContainer>
        </div>

        {/* Status donut charts */}
        <div className="chart-card donut-pair">
          <div className="donut-item">
            <h2 className="chart-title">Order Status</h2>
            <ResponsiveContainer width="100%" height={180}>
              <PieChart>
                <Pie
                  data={data.order_status}
                  dataKey="count"
                  nameKey="status"
                  cx="50%"
                  cy="50%"
                  innerRadius={45}
                  outerRadius={72}
                  paddingAngle={3}
                >
                  {data.order_status.map((entry, i) => (
                    <Cell key={i} fill={PIE_COLORS[entry.status] ?? COLORS[i]} />
                  ))}
                </Pie>
                <Tooltip formatter={(v, name) => [v, name]} />
                <Legend iconType="circle" iconSize={9} wrapperStyle={{ fontSize: 12 }} />
              </PieChart>
            </ResponsiveContainer>
          </div>
          <div className="donut-item">
            <h2 className="chart-title">Invoice Status</h2>
            <ResponsiveContainer width="100%" height={180}>
              <PieChart>
                <Pie
                  data={data.invoice_status}
                  dataKey="count"
                  nameKey="status"
                  cx="50%"
                  cy="50%"
                  innerRadius={45}
                  outerRadius={72}
                  paddingAngle={3}
                >
                  {data.invoice_status.map((entry, i) => (
                    <Cell key={i} fill={PIE_COLORS[entry.status] ?? COLORS[i]} />
                  ))}
                </Pie>
                <Tooltip formatter={(v, name) => [v, name]} />
                <Legend iconType="circle" iconSize={9} wrapperStyle={{ fontSize: 12 }} />
              </PieChart>
            </ResponsiveContainer>
          </div>
        </div>

      </div>

      {/* ── Recommendations ── */}
      <div className="section-header">
        <h2 className="section-title">Strategic Recommendations</h2>
        <p className="section-sub">Generated from current supplier data patterns</p>
      </div>
      <div className="recs-grid">
        {data.recommendations.map((rec, i) => (
          <RecommendationCard key={i} rec={rec} />
        ))}
      </div>
    </div>
  )
}
