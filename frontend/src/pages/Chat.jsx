import { useState, useRef, useEffect } from 'react'

const TOOL_LABELS = {
  get_supplier_overview:   { icon: '🏢', label: 'Supplier overview' },
  get_incidents:           { icon: '🔍', label: 'Incident database'  },
  get_questionnaire_draft: { icon: '📝', label: 'Drafting response'  },
  get_insights:            { icon: '📊', label: 'Spend insights'     },
}

const SUGGESTIONS = [
  'What are our highest-risk incidents right now?',
  'Show me a root cause analysis for all open incidents',
  'Which supplier accounts for the most spend?',
  'Draft a compliance audit response',
  'Give me a performance review for Global Supplies Ltd',
  'What trends do you see in our monthly spend?',
]

function ToolPill({ tool }) {
  const meta = TOOL_LABELS[tool.name] ?? { icon: '⚙️', label: tool.display }
  return (
    <span className="tool-pill">
      <span>{meta.icon}</span>
      <span>{tool.display || meta.label}</span>
    </span>
  )
}

function GlowAvatar() {
  return (
    <div className="glow-avatar">
      <span className="glow-avatar-inner">G</span>
    </div>
  )
}

// Render markdown-lite: bold, bullet lists, numbered lists, inline code
function GlowText({ text }) {
  const lines = text.split('\n')
  const elements = []
  let i = 0

  while (i < lines.length) {
    const line = lines[i]

    if (!line.trim()) {
      elements.push(<div key={i} className="msg-spacer" />)
      i++
      continue
    }

    // Heading (## or ###)
    if (/^#{2,3}\s/.test(line)) {
      const content = line.replace(/^#{2,3}\s/, '')
      elements.push(<p key={i} className="msg-heading">{inlineFormat(content)}</p>)
      i++
      continue
    }

    // Bullet list block
    if (/^[-*]\s/.test(line)) {
      const items = []
      while (i < lines.length && /^[-*]\s/.test(lines[i])) {
        items.push(<li key={i}>{inlineFormat(lines[i].replace(/^[-*]\s/, ''))}</li>)
        i++
      }
      elements.push(<ul key={`ul-${i}`} className="msg-list">{items}</ul>)
      continue
    }

    // Numbered list block
    if (/^\d+\.\s/.test(line)) {
      const items = []
      while (i < lines.length && /^\d+\.\s/.test(lines[i])) {
        items.push(<li key={i}>{inlineFormat(lines[i].replace(/^\d+\.\s/, ''))}</li>)
        i++
      }
      elements.push(<ol key={`ol-${i}`} className="msg-list msg-list-ol">{items}</ol>)
      continue
    }

    elements.push(<p key={i} className="msg-line">{inlineFormat(line)}</p>)
    i++
  }

  return <div className="glow-text">{elements}</div>
}

function inlineFormat(text) {
  // Split on **bold**, *italic*, `code`
  const parts = text.split(/(\*\*[^*]+\*\*|\*[^*]+\*|`[^`]+`)/g)
  return parts.map((part, i) => {
    if (part.startsWith('**') && part.endsWith('**'))
      return <strong key={i}>{part.slice(2, -2)}</strong>
    if (part.startsWith('*') && part.endsWith('*'))
      return <em key={i}>{part.slice(1, -1)}</em>
    if (part.startsWith('`') && part.endsWith('`'))
      return <code key={i} className="inline-code">{part.slice(1, -1)}</code>
    return part
  })
}

function Message({ msg }) {
  const isUser = msg.role === 'user'

  if (isUser) {
    return (
      <div className="msg-row msg-row-user">
        <div className="bubble bubble-user">{msg.content}</div>
      </div>
    )
  }

  if (msg.role === 'thinking') {
    return (
      <div className="msg-row msg-row-glow">
        <GlowAvatar />
        <div className="bubble bubble-thinking">
          <span className="thinking-dots"><span /><span /><span /></span>
          {msg.tools?.map((t, i) => <ToolPill key={i} tool={t} />)}
        </div>
      </div>
    )
  }

  return (
    <div className="msg-row msg-row-glow">
      <GlowAvatar />
      <div className="bubble bubble-glow">
        {msg.tools?.length > 0 && (
          <div className="tool-pills-row">
            {msg.tools.map((t, i) => <ToolPill key={i} tool={t} />)}
          </div>
        )}
        <GlowText text={msg.content} />
      </div>
    </div>
  )
}

export default function Chat() {
  const [messages, setMessages] = useState([
    {
      role: 'glow',
      content: "Hi, I'm **Glow** — your ERP procurement analyst. I have live access to your supplier data, incident database, and insights.\n\nAsk me anything about your suppliers, open incidents, spend trends, or I can draft compliance and audit responses for you.",
      tools: [],
    }
  ])
  const [input, setInput]       = useState('')
  const [loading, setLoading]   = useState(false)
  const bottomRef               = useRef(null)
  const inputRef                = useRef(null)

  // API conversation history (only user + assistant turns)
  const historyRef = useRef([])

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  async function send(text) {
    const userText = (text ?? input).trim()
    if (!userText || loading) return
    setInput('')

    // Add user bubble
    setMessages(prev => [...prev, { role: 'user', content: userText }])

    // Add to API history
    historyRef.current = [...historyRef.current, { role: 'user', content: userText }]

    // Show thinking indicator
    setMessages(prev => [...prev, { role: 'thinking', tools: [] }])
    setLoading(true)

    try {
      const res = await fetch('/api/glow/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ messages: historyRef.current }),
      })

      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Unknown error' }))
        throw new Error(err.error ?? 'Request failed')
      }

      const data = await res.json()

      // Update thinking → real response
      setMessages(prev => [
        ...prev.filter(m => m.role !== 'thinking'),
        { role: 'glow', content: data.content, tools: data.tool_uses ?? [] },
      ])

      // Add assistant turn to history
      historyRef.current = [...historyRef.current, { role: 'assistant', content: data.content }]

    } catch (err) {
      setMessages(prev => [
        ...prev.filter(m => m.role !== 'thinking'),
        { role: 'glow', content: `Sorry, I ran into an error: ${err.message}`, tools: [] },
      ])
    } finally {
      setLoading(false)
      inputRef.current?.focus()
    }
  }

  function handleKey(e) {
    if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); send() }
  }

  return (
    <div className="chat-page">
      {/* Header */}
      <div className="chat-header">
        <GlowAvatar />
        <div>
          <p className="chat-header-name">Glow</p>
          <p className="chat-header-sub">ERP Procurement Analyst · Live data access</p>
        </div>
        <button className="chat-clear" onClick={() => {
          historyRef.current = []
          setMessages([{
            role: 'glow',
            content: "Conversation cleared. How can I help you?",
            tools: [],
          }])
        }}>Clear</button>
      </div>

      {/* Messages */}
      <div className="chat-messages">
        {messages.map((msg, i) => <Message key={i} msg={msg} />)}
        <div ref={bottomRef} />
      </div>

      {/* Suggestions (only when just the greeting is shown) */}
      {messages.length === 1 && (
        <div className="suggestions">
          {SUGGESTIONS.map((s, i) => (
            <button key={i} className="suggestion-chip" onClick={() => send(s)}>
              {s}
            </button>
          ))}
        </div>
      )}

      {/* Input */}
      <div className="chat-input-row">
        <textarea
          ref={inputRef}
          className="chat-input"
          placeholder="Ask Glow about suppliers, incidents, spend trends, or request a questionnaire draft…"
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKey}
          rows={1}
          disabled={loading}
        />
        <button
          className="chat-send"
          onClick={() => send()}
          disabled={loading || !input.trim()}
        >
          {loading ? '…' : '↑'}
        </button>
      </div>
    </div>
  )
}
