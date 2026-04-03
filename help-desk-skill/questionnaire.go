package helpdesk

import (
	"fmt"
	"strings"
	"time"
)

// QuestionnaireType identifies which questionnaire template to use.
type QuestionnaireType string

const (
	QTypeIncidentReport       QuestionnaireType = "incident_report"
	QTypeSupplierPerformance  QuestionnaireType = "supplier_performance"
	QTypeComplianceAudit      QuestionnaireType = "compliance_audit"
	QTypeRootCauseAnalysis    QuestionnaireType = "root_cause_analysis"
)

// DraftRequest is the input to the questionnaire drafter.
type DraftRequest struct {
	Type     QuestionnaireType `json:"type"`
	Incident *Incident         `json:"incident,omitempty"` // nil → use global summary
}

// DraftResponse is the generated questionnaire response.
type DraftResponse struct {
	Title     string            `json:"title"`
	Type      QuestionnaireType `json:"type"`
	GeneratedAt string          `json:"generated_at"`
	Sections  []Section         `json:"sections"`
}

// Section is one block (heading + Q&A pairs) inside a draft.
type Section struct {
	Heading string `json:"heading"`
	QA      []QA   `json:"qa"`
}

type QA struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// DraftQuestionnaire generates a structured questionnaire response for the given request.
func DraftQuestionnaire(req DraftRequest, allIncidents []Incident) DraftResponse {
	switch req.Type {
	case QTypeIncidentReport:
		return draftIncidentReport(req.Incident)
	case QTypeSupplierPerformance:
		return draftSupplierPerformance(req.Incident, allIncidents)
	case QTypeComplianceAudit:
		return draftComplianceAudit(allIncidents)
	case QTypeRootCauseAnalysis:
		return draftRootCauseAnalysis(req.Incident, allIncidents)
	default:
		return draftRootCauseAnalysis(req.Incident, allIncidents)
	}
}

// ── Templates ────────────────────────────────────────────────────────────────

func draftIncidentReport(inc *Incident) DraftResponse {
	now := time.Now().Format("2006-01-02")

	var title, entity, desc, evidence, action string
	var sev Severity
	var rc RootCauseCategory

	if inc != nil {
		title    = inc.Title
		entity   = inc.AffectedEntity
		desc     = inc.Description
		evidence = inc.Evidence
		action   = inc.RecommendedAction
		sev      = inc.Severity
		rc       = inc.RootCause
	} else {
		title    = "General Data Incident Report"
		entity   = "Multiple Suppliers"
		desc     = "Multiple data quality and financial incidents detected across the supplier base."
		evidence = "See incident list for full details."
		action   = "Review all open incidents and assign owners."
		sev      = SeverityHigh
		rc       = RootCauseProcess
	}

	return DraftResponse{
		Title:       fmt.Sprintf("Incident Report — %s", title),
		Type:        QTypeIncidentReport,
		GeneratedAt: now,
		Sections: []Section{
			{
				Heading: "1. Incident Overview",
				QA: []QA{
					{Question: "What is the incident title?", Answer: title},
					{Question: "Which entity is affected?", Answer: entity},
					{Question: "What is the incident severity?", Answer: string(sev)},
					{Question: "When was this report generated?", Answer: now},
				},
			},
			{
				Heading: "2. Incident Description",
				QA: []QA{
					{Question: "Describe the incident.", Answer: desc},
					{Question: "What data evidence supports this finding?", Answer: evidence},
					{Question: "What is the identified root cause category?", Answer: string(rc)},
				},
			},
			{
				Heading: "3. Impact Assessment",
				QA: []QA{
					{Question: "What is the business impact?", Answer: impactStatement(sev, rc)},
					{Question: "Are there any regulatory or compliance implications?", Answer: complianceImplication(rc)},
					{Question: "What is the financial exposure estimate?", Answer: "Pending detailed review by Finance team."},
				},
			},
			{
				Heading: "4. Resolution & Actions",
				QA: []QA{
					{Question: "What immediate action is recommended?", Answer: action},
					{Question: "Who is the responsible owner?", Answer: "Assign to: Procurement Lead / AP Manager (update as appropriate)"},
					{Question: "What is the target resolution date?", Answer: fmt.Sprintf("Within 5 business days of %s", now)},
					{Question: "How will resolution be verified?", Answer: "Re-run the automated incident scan. Incident should no longer appear once data has been corrected."},
				},
			},
		},
	}
}

func draftSupplierPerformance(inc *Incident, all []Incident) DraftResponse {
	now := time.Now().Format("2006-01-02")
	entity := "All Suppliers"
	if inc != nil {
		entity = inc.AffectedEntity
	}

	// Count incidents for this entity
	var entityIncidents []Incident
	for _, i := range all {
		if inc == nil || i.AffectedEntity == entity {
			entityIncidents = append(entityIncidents, i)
		}
	}

	return DraftResponse{
		Title:       fmt.Sprintf("Supplier Performance Review — %s", entity),
		Type:        QTypeSupplierPerformance,
		GeneratedAt: now,
		Sections: []Section{
			{
				Heading: "1. Supplier Profile",
				QA: []QA{
					{Question: "Supplier name", Answer: entity},
					{Question: "Review period", Answer: "Current fiscal year to date"},
					{Question: "Number of open incidents", Answer: fmt.Sprintf("%d", len(entityIncidents))},
				},
			},
			{
				Heading: "2. Order Fulfilment Performance",
				QA: []QA{
					{Question: "Are orders being fulfilled on time?", Answer: orderFulfilmentAnswer(entityIncidents)},
					{Question: "What percentage of orders are currently pending?", Answer: "Refer to Supplier Summary dashboard for real-time figures."},
					{Question: "Have there been any fulfilment incidents in this period?", Answer: incidentSummaryByRoot(entityIncidents, RootCauseOperational)},
				},
			},
			{
				Heading: "3. Financial & Invoice Compliance",
				QA: []QA{
					{Question: "Is the supplier submitting invoices on time?", Answer: invoiceComplianceAnswer(entityIncidents)},
					{Question: "Are there any unpaid invoices beyond payment terms?", Answer: "Review accounts payable ageing report for this supplier."},
					{Question: "Have there been any financial incidents?", Answer: incidentSummaryByRoot(entityIncidents, RootCauseFinancial)},
				},
			},
			{
				Heading: "4. Strategic Assessment",
				QA: []QA{
					{Question: "Does this supplier represent a concentration risk?", Answer: incidentSummaryByRoot(entityIncidents, RootCauseStrategic)},
					{Question: "Is a backup supplier in place?", Answer: "To be confirmed by Procurement team."},
					{Question: "Overall performance rating (1–5)", Answer: performanceRating(entityIncidents)},
				},
			},
			{
				Heading: "5. Action Plan",
				QA: []QA{
					{Question: "What corrective actions are required?", Answer: buildActionPlan(entityIncidents)},
					{Question: "Next review date", Answer: "Quarterly — schedule for 90 days from " + now},
				},
			},
		},
	}
}

func draftComplianceAudit(all []Incident) DraftResponse {
	now := time.Now().Format("2006-01-02")

	critical := filterBySeverity(all, SeverityCritical)
	high      := filterBySeverity(all, SeverityHigh)

	return DraftResponse{
		Title:       "Procurement Compliance Audit Response",
		Type:        QTypeComplianceAudit,
		GeneratedAt: now,
		Sections: []Section{
			{
				Heading: "1. Audit Scope",
				QA: []QA{
					{Question: "Audit period covered", Answer: "Current fiscal year to date"},
					{Question: "Number of suppliers in scope", Answer: "All active suppliers in ERP system"},
					{Question: "Audit date", Answer: now},
				},
			},
			{
				Heading: "2. Purchase-to-Pay (P2P) Controls",
				QA: []QA{
					{Question: "Are all invoices matched to a purchase order?", Answer: dataQualityAnswer(all)},
					{Question: "Are payment terms being adhered to?", Answer: financialComplianceAnswer(all)},
					{Question: "Is the three-way match process (PO/GR/Invoice) enforced?", Answer: "Partial — see data-integrity incidents for gaps requiring remediation."},
				},
			},
			{
				Heading: "3. Supplier Risk Management",
				QA: []QA{
					{Question: "Have concentration risks been identified?", Answer: strategicRiskAnswer(all)},
					{Question: "Is supplier spend monitored regularly?", Answer: "Yes — automated insight scans generate real-time spend concentration alerts."},
					{Question: "Number of Critical incidents open", Answer: fmt.Sprintf("%d", len(critical))},
					{Question: "Number of High incidents open", Answer: fmt.Sprintf("%d", len(high))},
				},
			},
			{
				Heading: "4. Remediation Status",
				QA: []QA{
					{Question: "List all open incidents requiring remediation", Answer: buildFullIncidentList(all)},
					{Question: "What is the overall remediation timeline?", Answer: "All Critical items: 48 hours. High: 5 business days. Medium/Low: next sprint cycle."},
					{Question: "Who is accountable for sign-off?", Answer: "Head of Procurement and CFO for Critical/High; Procurement Lead for Medium/Low."},
				},
			},
		},
	}
}

func draftRootCauseAnalysis(inc *Incident, all []Incident) DraftResponse {
	now := time.Now().Format("2006-01-02")

	targets := all
	title   := "Root Cause Analysis — All Open Incidents"
	if inc != nil {
		targets = []Incident{*inc}
		title   = fmt.Sprintf("Root Cause Analysis — %s", inc.Title)
	}

	// Group by root cause
	groups := map[RootCauseCategory][]Incident{}
	for _, i := range targets {
		groups[i.RootCause] = append(groups[i.RootCause], i)
	}

	var sections []Section
	sections = append(sections, Section{
		Heading: "1. Executive Summary",
		QA: []QA{
			{Question: "Total incidents analysed", Answer: fmt.Sprintf("%d", len(targets))},
			{Question: "Distinct root cause categories identified", Answer: fmt.Sprintf("%d", len(groups))},
			{Question: "Highest severity present", Answer: string(highestSeverity(targets))},
			{Question: "Analysis date", Answer: now},
		},
	})

	order := []RootCauseCategory{
		RootCauseFinancial, RootCauseStrategic, RootCauseProcess,
		RootCauseOperational, RootCauseDataQuality,
	}
	secNum := 2
	for _, rc := range order {
		items, ok := groups[rc]
		if !ok {
			continue
		}
		var affected []string
		for _, i := range items {
			affected = append(affected, i.AffectedEntity)
		}
		sections = append(sections, Section{
			Heading: fmt.Sprintf("%d. Root Cause: %s", secNum, rc),
			QA: []QA{
				{Question: "Number of incidents in this category", Answer: fmt.Sprintf("%d", len(items))},
				{Question: "Affected entities", Answer: strings.Join(unique(affected), ", ")},
				{Question: "What systemic condition allows this issue to recur?", Answer: systemicCause(rc)},
				{Question: "What is the recommended permanent fix?", Answer: permanentFix(rc)},
				{Question: "What is the preventative control going forward?", Answer: preventativeControl(rc)},
			},
		})
		secNum++
	}

	sections = append(sections, Section{
		Heading: fmt.Sprintf("%d. Prioritised Action Register", secNum),
		QA:      buildPrioritisedActions(targets),
	})

	return DraftResponse{
		Title:       title,
		Type:        QTypeRootCauseAnalysis,
		GeneratedAt: now,
		Sections:    sections,
	}
}

// ── helper text generators ────────────────────────────────────────────────────

func impactStatement(sev Severity, rc RootCauseCategory) string {
	switch sev {
	case SeverityCritical:
		return fmt.Sprintf("Critical business impact. The %s issue requires immediate executive attention and may affect payment terms, supplier relationships, or regulatory standing.", rc)
	case SeverityHigh:
		return fmt.Sprintf("High business impact. The %s issue should be resolved within the current week to prevent escalation.", rc)
	case SeverityMedium:
		return fmt.Sprintf("Moderate business impact. The %s issue should be addressed in the current sprint to prevent downstream effects.", rc)
	default:
		return fmt.Sprintf("Low business impact. The %s issue should be logged and resolved in the next maintenance window.", rc)
	}
}

func complianceImplication(rc RootCauseCategory) string {
	switch rc {
	case RootCauseFinancial:
		return "Possible breach of payment terms. Review supplier contract SLAs and statutory payment obligations."
	case RootCauseDataQuality:
		return "Invoices without matching POs may breach internal controls policy and external audit requirements."
	case RootCauseStrategic:
		return "Concentration risk may need to be disclosed in financial risk registers or board reports."
	default:
		return "No immediate regulatory implication identified. Monitor for escalation."
	}
}

func orderFulfilmentAnswer(incidents []Incident) string {
	for _, i := range incidents {
		if i.RootCause == RootCauseOperational {
			return fmt.Sprintf("Issue detected: %s. %s", i.Title, i.RecommendedAction)
		}
	}
	return "No open fulfilment incidents detected for this period."
}

func invoiceComplianceAnswer(incidents []Incident) string {
	for _, i := range incidents {
		if i.RootCause == RootCauseFinancial {
			return fmt.Sprintf("Issue detected: %s. %s", i.Title, i.RecommendedAction)
		}
	}
	return "Invoice submission appears within normal parameters. Continue monitoring."
}

func performanceRating(incidents []Incident) string {
	critical := len(filterBySeverity(incidents, SeverityCritical))
	high      := len(filterBySeverity(incidents, SeverityHigh))
	switch {
	case critical > 0:
		return "1 / 5 — Requires immediate improvement plan"
	case high > 1:
		return "2 / 5 — Below expectations; corrective action required"
	case high == 1:
		return "3 / 5 — Meets minimum requirements; improvement needed"
	default:
		return "4 / 5 — Performing well; minor optimisations possible"
	}
}

func buildActionPlan(incidents []Incident) string {
	if len(incidents) == 0 {
		return "No corrective actions required at this time."
	}
	var lines []string
	for _, i := range incidents {
		lines = append(lines, fmt.Sprintf("[%s] %s → %s", i.Severity, i.Title, i.RecommendedAction))
	}
	return strings.Join(lines, "\n")
}

func dataQualityAnswer(all []Incident) string {
	for _, i := range all {
		if i.RootCause == RootCauseDataQuality {
			return fmt.Sprintf("Gap identified: %s — %s", i.Title, i.RecommendedAction)
		}
	}
	return "No orphaned invoices detected. Three-way match appears to be functioning."
}

func financialComplianceAnswer(all []Incident) string {
	count := 0
	for _, i := range all {
		if i.RootCause == RootCauseFinancial {
			count++
		}
	}
	if count > 0 {
		return fmt.Sprintf("%d financial incident(s) detected. Payment term compliance should be reviewed.", count)
	}
	return "Payment terms appear to be adhered to within current data set."
}

func strategicRiskAnswer(all []Incident) string {
	for _, i := range all {
		if i.RootCause == RootCauseStrategic {
			return fmt.Sprintf("Yes — %s: %s", i.Title, i.Description)
		}
	}
	return "No strategic concentration risk detected above threshold."
}

func buildFullIncidentList(all []Incident) string {
	if len(all) == 0 {
		return "No open incidents."
	}
	var lines []string
	for _, i := range all {
		lines = append(lines, fmt.Sprintf("#%d [%s] %s (%s)", i.ID, i.Severity, i.Title, i.AffectedEntity))
	}
	return strings.Join(lines, "\n")
}

func systemicCause(rc RootCauseCategory) string {
	switch rc {
	case RootCauseFinancial:
		return "Lack of automated payment-run scheduling and invoice-aging alerts in the ERP system."
	case RootCauseStrategic:
		return "Procurement strategy has not enforced dual-sourcing policies for high-spend categories."
	case RootCauseProcess:
		return "The AP process lacks automated approval routing and exception handling for unmatched invoices."
	case RootCauseOperational:
		return "Supplier SLA monitoring is manual; no automated alerts exist for overdue orders."
	case RootCauseDataQuality:
		return "The system allows invoice creation without mandatory PO reference, bypassing three-way match."
	default:
		return "Root cause under investigation."
	}
}

func permanentFix(rc RootCauseCategory) string {
	switch rc {
	case RootCauseFinancial:
		return "Implement automated AP ageing reports with escalation workflows. Set payment-term SLA alerts."
	case RootCauseStrategic:
		return "Enforce dual-sourcing policy for any supplier exceeding 30% of category spend. Update procurement policy."
	case RootCauseProcess:
		return "Configure mandatory PO-backed invoice workflow in ERP. Add three-way match validation at invoice entry."
	case RootCauseOperational:
		return "Implement order tracking with automated supplier chasing at T+lead-time. Integrate delivery confirmation."
	case RootCauseDataQuality:
		return "Add database constraints to require order_id on all invoices. Run a back-fill migration for historical data."
	default:
		return "Define corrective action with the process owner."
	}
}

func preventativeControl(rc RootCauseCategory) string {
	switch rc {
	case RootCauseFinancial:
		return "Monthly AP ageing review; automated alert when any invoice exceeds 30 days unpaid."
	case RootCauseStrategic:
		return "Quarterly spend-concentration dashboard review by Procurement Director."
	case RootCauseProcess:
		return "Weekly AP exception report reviewed by Finance Manager."
	case RootCauseOperational:
		return "Automated order-status dashboard reviewed by Procurement team daily."
	case RootCauseDataQuality:
		return "Daily automated data-quality scan with incident creation for any orphaned records."
	default:
		return "Periodic manual review by process owner."
	}
}

func buildPrioritisedActions(incidents []Incident) []QA {
	if len(incidents) == 0 {
		return []QA{{Question: "Actions required", Answer: "None — no incidents detected."}}
	}
	var qa []QA
	for _, i := range incidents {
		qa = append(qa, QA{
			Question: fmt.Sprintf("[%s] #%d — %s", i.Severity, i.ID, i.Title),
			Answer:   fmt.Sprintf("Entity: %s | Action: %s", i.AffectedEntity, i.RecommendedAction),
		})
	}
	return qa
}

func incidentSummaryByRoot(incidents []Incident, rc RootCauseCategory) string {
	var found []Incident
	for _, i := range incidents {
		if i.RootCause == rc {
			found = append(found, i)
		}
	}
	if len(found) == 0 {
		return "None detected."
	}
	var titles []string
	for _, i := range found {
		titles = append(titles, i.Title)
	}
	return strings.Join(titles, "; ")
}

func filterBySeverity(incidents []Incident, sev Severity) []Incident {
	var out []Incident
	for _, i := range incidents {
		if i.Severity == sev {
			out = append(out, i)
		}
	}
	return out
}

func highestSeverity(incidents []Incident) Severity {
	for _, sev := range []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow} {
		for _, i := range incidents {
			if i.Severity == sev {
				return sev
			}
		}
	}
	return SeverityLow
}

func unique(s []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range s {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}
