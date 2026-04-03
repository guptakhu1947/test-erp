import { useState, useEffect } from 'react'

const SEV_COLOR = {
  Critical: { bg: '#fef2f2', border: '#ef4444', text: '#b91c1c', dot: '#ef4444' },
  High:     { bg: '#fffbeb', border: '#f59e0b', text: '#b45309', dot: '#f59e0b' },
  Medium:   { bg: '#eff6ff', border: '#3b82f6', text: '#1d4ed8', dot: '#3b82f6' },
  Low:      { bg: '#f0fdf4', border: '#22c55e', text: '#15803d', dot: '#22c55e' },
}

const Q_TYPES = [
  { key: 'root_cause_analysis',  label: 'Root Cause Analysis'      },
  { key: 'incident_report',      label: 'Incident Report'           },
  { key: 'supplier_performance', label: 'Supplier Performance Review'},
  { key: 'compliance_audit',     label: 'Compliance Audit'          },
]

function SeverityBadge({ sev }) {
  const c = SEV_COLOR[sev] ?? SEV_COLOR.Low
  return (
    <span style={{
      background: c.bg, color: c.text,
      border: `1px solid ${c.border}`,
      borderRadius: 999, padding: '2px 10px',
      fontSize: '.72rem', fontWeight: 700,
    }}>
      {sev}
    </span>
  )
}

function IncidentRow({ inc, selected, onClick }) {
  const c = SEV_COLOR[inc.severity] ?? SEV_COLOR.Low
  return (
    <tr
      className={`data-row${selected ? ' row-selected' : ''}`}
      onClick={onClick}
      style={{ cursor: 'pointer' }}
    >
      <td style={{ width: 32, textAlign: 'center', color: '#9ca3af', fontSize: '.8rem' }}>#{inc.id}</td>
      <td><SeverityBadge sev={inc.severity} /></td>
      <td>
        <span style={{
          background: '#f3f4f6', color: '#374151',
          borderRadius: 6, padding: '2px 8px',
          fontSize: '.72rem', fontWeight: 600,
        }}>
          {inc.root_cause}
        </span>
      </td>
      <td style={{ fontWeight: 600, color: '#111827' }}>{inc.title}</td>
      <td style={{ color: '#6b7280', fontSize: '.825rem' }}>{inc.affected_entity}</td>
    </tr>
  )
}

function DraftViewer({ draft }) {
  if (!draft) return null
  return (
    <div className="draft-viewer">
      <div className="draft-header">
        <h3 className="draft-title">{draft.title}</h3>
        <span className="draft-meta">Generated {draft.generated_at}</span>
      </div>
      {draft.sections.map((sec, si) => (
        <div key={si} className="draft-section">
          <h4 className="draft-section-title">{sec.heading}</h4>
          {sec.qa.map((qa, qi) => (
            <div key={qi} className="draft-qa">
              <p className="draft-q">{qa.question}</p>
              <p className="draft-a">{qa.answer}</p>
            </div>
          ))}
        </div>
      ))}
    </div>
  )
}

export default function HelpDesk() {
  const [incidents, setIncidents]   = useState([])
  const [loading, setLoading]       = useState(true)
  const [error, setError]           = useState(null)
  const [selected, setSelected]     = useState(null)   // incident object
  const [qType, setQType]           = useState('root_cause_analysis')
  const [draft, setDraft]           = useState(null)
  const [draftLoading, setDL]       = useState(false)

  useEffect(() => {
    fetch('/api/helpdesk/incidents')
      .then(r => { if (!r.ok) throw new Error('Failed to load incidents'); return r.json() })
      .then(data => { setIncidents(data); })
      .catch(e => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  function fetchDraft(type, incidentID) {
    setDL(true)
    setDraft(null)
    const id = incidentID ?? 0
    fetch(`/api/helpdesk/questionnaire?type=${type}&incident_id=${id}`)
      .then(r => r.json())
      .then(setDraft)
      .catch(e => setError(e.message))
      .finally(() => setDL(false))
  }

  const counts = {
    Critical: incidents.filter(i => i.severity === 'Critical').length,
    High:     incidents.filter(i => i.severity === 'High').length,
    Medium:   incidents.filter(i => i.severity === 'Medium').length,
    Low:      incidents.filter(i => i.severity === 'Low').length,
  }

  if (loading) return (
    <div className="page-center">
      <div className="spinner" />
      <p className="status-text">Running incident scan…</p>
    </div>
  )
  if (error) return (
    <div className="page-center"><p className="error-text">Error: {error}</p></div>
  )

  return (
    <div className="content">
      <div className="page-header">
        <div>
          <h1 className="page-title">Help Desk — Data Incident Analyser</h1>
          <p className="page-sub">Automated root-cause detection · questionnaire drafter · strategic recommendations</p>
        </div>
      </div>

      {/* Severity KPIs */}
      <div className="kpi-row">
        {Object.entries(counts).map(([sev, n]) => {
          const c = SEV_COLOR[sev]
          return (
            <div key={sev} className="kpi-card" style={{ borderColor: c.border }}>
              <span className="kpi-label" style={{ color: c.text }}>{sev}</span>
              <span className="kpi-value" style={{ color: c.text }}>{n}</span>
            </div>
          )
        })}
        <div className="kpi-card accent">
          <span className="kpi-label">Total Incidents</span>
          <span className="kpi-value">{incidents.length}</span>
        </div>
      </div>

      <div className="hd-layout">

        {/* Left: incident list */}
        <div className="hd-left">
          <div className="panel-header">
            <span className="panel-title">Prioritised Incident Register</span>
            <span className="panel-sub">Click a row to scope the questionnaire</span>
          </div>
          <div className="table-wrapper" style={{ borderRadius: '0 0 10px 10px' }}>
            <table className="supplier-table">
              <thead>
                <tr>
                  <th>#</th>
                  <th>Severity</th>
                  <th>Root Cause</th>
                  <th>Title</th>
                  <th>Entity</th>
                </tr>
              </thead>
              <tbody>
                {incidents.length === 0 && (
                  <tr><td colSpan={5} className="empty-row">No incidents detected.</td></tr>
                )}
                {incidents.map(inc => (
                  <IncidentRow
                    key={inc.id}
                    inc={inc}
                    selected={selected?.id === inc.id}
                    onClick={() => setSelected(prev => prev?.id === inc.id ? null : inc)}
                  />
                ))}
              </tbody>
            </table>
          </div>

          {/* Incident detail panel */}
          {selected && (
            <div className="incident-detail" style={{
              borderLeftColor: SEV_COLOR[selected.severity]?.border,
              background: SEV_COLOR[selected.severity]?.bg,
            }}>
              <div className="detail-row">
                <SeverityBadge sev={selected.severity} />
                <span className="detail-rc">{selected.root_cause}</span>
              </div>
              <p className="detail-title">{selected.title}</p>
              <p className="detail-desc">{selected.description}</p>
              <div className="detail-evidence">
                <span className="evidence-label">Evidence:</span> {selected.evidence}
              </div>
              <div className="detail-action">
                <span className="action-label">Recommended Action:</span> {selected.recommended_action}
              </div>
            </div>
          )}
        </div>

        {/* Right: questionnaire drafter */}
        <div className="hd-right">
          <div className="panel-header">
            <span className="panel-title">Questionnaire Drafter</span>
            <span className="panel-sub">{selected ? `Scoped to: ${selected.title}` : 'Global — all incidents'}</span>
          </div>
          <div className="drafter-body">
            <div className="drafter-controls">
              <div className="drafter-type-grid">
                {Q_TYPES.map(qt => (
                  <button
                    key={qt.key}
                    className={`type-btn${qType === qt.key ? ' active' : ''}`}
                    onClick={() => setQType(qt.key)}
                  >
                    {qt.label}
                  </button>
                ))}
              </div>
              <button
                className="generate-btn"
                onClick={() => fetchDraft(qType, selected?.id)}
                disabled={draftLoading}
              >
                {draftLoading ? 'Generating…' : 'Generate Draft'}
              </button>
            </div>

            {!draft && !draftLoading && (
              <div className="draft-empty">
                <p>Select a questionnaire type above and click <strong>Generate Draft</strong>.</p>
                <p style={{ marginTop: '.5rem', fontSize: '.8rem', color: '#9ca3af' }}>
                  Optionally click an incident row on the left to scope the response to that specific incident.
                </p>
              </div>
            )}
            {draftLoading && (
              <div className="page-center" style={{ height: 200 }}>
                <div className="spinner" />
              </div>
            )}
            {draft && <DraftViewer draft={draft} />}
          </div>
        </div>

      </div>
    </div>
  )
}
