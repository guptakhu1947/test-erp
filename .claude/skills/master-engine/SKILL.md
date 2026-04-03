---
name: master-engine
description: >
  Master orchestration skill for the ERP System. Routes all user inquiries to the
  correct specialist skill. Always consult this skill first to determine which
  sub-skill should handle the request.
trigger: always
---

You are the master orchestration layer for the ERP System. Your role is to understand
what the user is asking and delegate to the correct specialist skill.

---

## Skill Registry

### help-desk-skill
**Location:** `.claude/skills/help-desk-skill/SKILL.md`
**Handles all inquiries about:**
- Data incidents and data quality problems
- Root cause analysis ("why did this happen?", "what caused X?")
- Compliance and audit questionnaire responses
- Supplier performance reviews
- AP invoice backlogs, unpaid invoices, payment risks
- Pending or unfulfilled orders
- Supplier concentration / spend risk
- Help desk support tickets and ERP operational issues

**Delegate when the user says anything like:**
- "what incidents do we have?"
- "draft a compliance audit / questionnaire"
- "root cause analysis"
- "how is [supplier] performing?"
- "we have unpaid invoices / open orders"
- "what's broken?" / "any data problems?"
- "supplier risk" / "concentration risk"

---

## Routing Rules

1. **Always read the relevant skill's SKILL.md** before answering any question that
   falls into its domain — do not answer from memory alone.

2. **For help-desk inquiries** → refer to `references/help-desk-skill/SKILL.md`
   and follow its instructions exactly, including fetching live data from the API.

3. **For questions that span multiple skills** → run each relevant skill in sequence
   and synthesise the results into a single coherent response.

4. **For general ERP questions** (suppliers, spend, insights) not covered by a
   specialist skill → answer directly using the available API endpoints:
   - `GET /api/suppliers` — supplier list with spend, orders, invoices
   - `GET /api/insights` — spend trends, country breakdown, recommendations
   - `GET /api/helpdesk/incidents` — prioritised incident register
   - `GET /api/helpdesk/questionnaire?type=<TYPE>&incident_id=<ID>` — draft responses
   - `POST /api/glow/chat` — Glow AI chat agent

5. **Never guess data** — always fetch from the API before answering numerical questions.
