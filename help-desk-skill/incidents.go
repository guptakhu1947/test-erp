package helpdesk

import (
	"fmt"

	"erp-app/dataaccess"
)

type Severity string

const (
	SeverityCritical Severity = "Critical"
	SeverityHigh     Severity = "High"
	SeverityMedium   Severity = "Medium"
	SeverityLow      Severity = "Low"
)

type RootCauseCategory string

const (
	RootCauseProcess     RootCauseCategory = "Process Gap"
	RootCauseFinancial   RootCauseCategory = "Financial Risk"
	RootCauseOperational RootCauseCategory = "Operational Failure"
	RootCauseStrategic   RootCauseCategory = "Strategic Risk"
	RootCauseDataQuality RootCauseCategory = "Data Integrity"
)

// Incident represents a detected data issue with a prioritised root cause.
type Incident struct {
	ID              int               `json:"id"`
	Severity        Severity          `json:"severity"`
	RootCause       RootCauseCategory `json:"root_cause"`
	AffectedEntity  string            `json:"affected_entity"`
	Title           string            `json:"title"`
	Description     string            `json:"description"`
	Evidence        string            `json:"evidence"`
	RecommendedAction string          `json:"recommended_action"`
	SeverityScore   int               `json:"-"` // used for sorting only
}

// DetectAndPrioritise scans all supplier data and returns incidents ranked by severity.
func DetectAndPrioritise() ([]Incident, error) {
	summaries, err := dataaccess.GetAllSupplierSummaries()
	if err != nil {
		return nil, err
	}
	orderCounts, err := dataaccess.GetOrderStatusCounts()
	if err != nil {
		return nil, err
	}
	invoiceCounts, err := dataaccess.GetInvoiceStatusCounts()
	if err != nil {
		return nil, err
	}
	spendBySupplier, err := dataaccess.GetSpendBySupplier()
	if err != nil {
		return nil, err
	}

	var incidents []Incident
	idSeq := 1

	// ── Aggregate totals ──────────────────────────────────────────────────────
	var totalSpend float64
	for _, s := range spendBySupplier {
		totalSpend += s.Spend
	}

	var totalOrders, pendingOrders, totalInvoices, pendingInvoices int
	for _, o := range orderCounts {
		totalOrders += o.Count
		if o.Status == "pending" {
			pendingOrders = o.Count
		}
	}
	for _, i := range invoiceCounts {
		totalInvoices += i.Count
		if i.Status == "pending" {
			pendingInvoices = i.Count
		}
	}

	// ── Per-supplier incidents ────────────────────────────────────────────────
	for _, s := range summaries {
		// Supplier with zero orders
		if s.OrderCount == 0 {
			incidents = append(incidents, Incident{
				ID:             idSeq,
				Severity:       SeverityHigh,
				SeverityScore:  3,
				RootCause:      RootCauseOperational,
				AffectedEntity: s.Name,
				Title:          "Supplier Has No Orders",
				Description:    fmt.Sprintf("%s has no recorded orders. This may indicate a dormant relationship or a data entry gap.", s.Name),
				Evidence:       "order_count = 0",
				RecommendedAction: "Verify if the supplier relationship is still active. If active, audit the orders table for missing records.",
			})
			idSeq++
		}

		// Pending invoice ratio > 25 %
		if s.InvoiceCount > 0 {
			pendingInvForSupplier, err := pendingInvoiceCountForSupplier(s.ID)
			if err == nil && pendingInvForSupplier > 0 {
				ratio := float64(pendingInvForSupplier) / float64(s.InvoiceCount) * 100
				if ratio >= 50 {
					incidents = append(incidents, Incident{
						ID:             idSeq,
						Severity:       SeverityCritical,
						SeverityScore:  4,
						RootCause:      RootCauseFinancial,
						AffectedEntity: s.Name,
						Title:          "Critical Invoice Backlog",
						Description:    fmt.Sprintf("%s has %.0f%% unpaid invoices (%d of %d). This poses an immediate cash-flow and supplier-trust risk.", s.Name, ratio, pendingInvForSupplier, s.InvoiceCount),
						Evidence:       fmt.Sprintf("pending_invoices=%d, total_invoices=%d", pendingInvForSupplier, s.InvoiceCount),
						RecommendedAction: "Escalate to accounts payable immediately. Engage the supplier to agree a payment schedule.",
					})
					idSeq++
				} else if ratio >= 25 {
					incidents = append(incidents, Incident{
						ID:             idSeq,
						Severity:       SeverityHigh,
						SeverityScore:  3,
						RootCause:      RootCauseFinancial,
						AffectedEntity: s.Name,
						Title:          "Elevated Pending Invoices",
						Description:    fmt.Sprintf("%s has %.0f%% of invoices unpaid. Review the accounts-payable cycle for this supplier.", s.Name, ratio),
						Evidence:       fmt.Sprintf("pending_invoices=%d, total_invoices=%d", pendingInvForSupplier, s.InvoiceCount),
						RecommendedAction: "Review payment terms and schedule outstanding invoices for the next payment run.",
					})
					idSeq++
				}
			}
		}

		// Pending orders ratio > 30 %
		if s.OrderCount > 0 {
			pendingOrdForSupplier, err := pendingOrderCountForSupplier(s.ID)
			if err == nil && pendingOrdForSupplier > 0 {
				ratio := float64(pendingOrdForSupplier) / float64(s.OrderCount) * 100
				if ratio >= 30 {
					incidents = append(incidents, Incident{
						ID:             idSeq,
						Severity:       SeverityMedium,
						SeverityScore:  2,
						RootCause:      RootCauseOperational,
						AffectedEntity: s.Name,
						Title:          "High Open Order Rate",
						Description:    fmt.Sprintf("%s has %.0f%% of orders still open (%d of %d). Delivery timelines may be at risk.", s.Name, ratio, pendingOrdForSupplier, s.OrderCount),
						Evidence:       fmt.Sprintf("pending_orders=%d, total_orders=%d", pendingOrdForSupplier, s.OrderCount),
						RecommendedAction: "Follow up with supplier on expected delivery dates. Flag for procurement review if overdue by >7 days.",
					})
					idSeq++
				}
			}
		}
	}

	// ── Global incidents ──────────────────────────────────────────────────────

	// Spend concentration risk
	if len(spendBySupplier) > 0 && totalSpend > 0 {
		pct := spendBySupplier[0].Spend / totalSpend * 100
		sev, score := SeverityMedium, 2
		if pct > 50 {
			sev, score = SeverityCritical, 4
		} else if pct > 35 {
			sev, score = SeverityHigh, 3
		}
		if pct > 30 {
			incidents = append(incidents, Incident{
				ID:             idSeq,
				Severity:       sev,
				SeverityScore:  score,
				RootCause:      RootCauseStrategic,
				AffectedEntity: spendBySupplier[0].Name,
				Title:          "Supplier Spend Concentration",
				Description:    fmt.Sprintf("%s accounts for %.0f%% of total spend ($%.0f of $%.0f). Single-source dependency creates supply-chain fragility.", spendBySupplier[0].Name, pct, spendBySupplier[0].Spend, totalSpend),
				Evidence:       fmt.Sprintf("supplier_spend=%.0f, total_spend=%.0f, concentration=%.0f%%", spendBySupplier[0].Spend, totalSpend, pct),
				RecommendedAction: "Identify at least one alternative supplier for the same category. Negotiate a dual-sourcing agreement.",
			})
			idSeq++
		}
	}

	// Global invoice backlog
	if totalInvoices > 0 {
		pendingPct := float64(pendingInvoices) / float64(totalInvoices) * 100
		if pendingPct >= 25 {
			incidents = append(incidents, Incident{
				ID:             idSeq,
				Severity:       SeverityHigh,
				SeverityScore:  3,
				RootCause:      RootCauseProcess,
				AffectedEntity: "All Suppliers",
				Title:          "Organisation-Wide Invoice Processing Backlog",
				Description:    fmt.Sprintf("%.0f%% of all invoices (%d of %d) are unpaid. This points to a systemic accounts-payable process gap.", pendingPct, pendingInvoices, totalInvoices),
				Evidence:       fmt.Sprintf("pending=%d, total=%d", pendingInvoices, totalInvoices),
				RecommendedAction: "Audit the AP process: check approval bottlenecks, missing PO matches, and ERP workflow configuration.",
			})
			idSeq++
		}
	}

	// Global order fulfilment
	if totalOrders > 0 {
		pendingPct := float64(pendingOrders) / float64(totalOrders) * 100
		if pendingPct >= 25 {
			incidents = append(incidents, Incident{
				ID:             idSeq,
				Severity:       SeverityMedium,
				SeverityScore:  2,
				RootCause:      RootCauseOperational,
				AffectedEntity: "All Suppliers",
				Title:          "Elevated Unfulfilled Order Rate",
				Description:    fmt.Sprintf("%.0f%% of all orders (%d of %d) remain open. Review fulfilment SLAs across the supplier base.", pendingPct, pendingOrders, totalOrders),
				Evidence:       fmt.Sprintf("pending=%d, total=%d", pendingOrders, totalOrders),
				RecommendedAction: "Run an aged-orders report. Escalate any order older than the contracted lead time to the relevant supplier manager.",
			})
			idSeq++
		}
	}

	// Data-quality check: invoices with no matching order
	orphaned, err := countOrphanedInvoices()
	if err == nil && orphaned > 0 {
		incidents = append(incidents, Incident{
			ID:             idSeq,
			Severity:       SeverityHigh,
			SeverityScore:  3,
			RootCause:      RootCauseDataQuality,
			AffectedEntity: "Invoice Records",
			Title:          "Invoices Without Matching Orders",
			Description:    fmt.Sprintf("%d invoice(s) have no corresponding purchase order. This may indicate unauthorised spend or missing order records.", orphaned),
			Evidence:       fmt.Sprintf("orphaned_invoices=%d", orphaned),
			RecommendedAction: "Audit invoices where order_id IS NULL or references a non-existent order. Require PO-backed invoices going forward.",
		})
		idSeq++
	}

	sortBySeverity(incidents)
	return incidents, nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func pendingInvoiceCountForSupplier(supplierID int) (int, error) {
	var count int
	err := dataaccess.DB.Get(&count,
		`SELECT COUNT(*) FROM invoices WHERE supplier_id = ? AND status = 'pending'`, supplierID)
	return count, err
}

func pendingOrderCountForSupplier(supplierID int) (int, error) {
	var count int
	err := dataaccess.DB.Get(&count,
		`SELECT COUNT(*) FROM orders WHERE supplier_id = ? AND status = 'pending'`, supplierID)
	return count, err
}

func countOrphanedInvoices() (int, error) {
	var count int
	err := dataaccess.DB.Get(&count,
		`SELECT COUNT(*) FROM invoices WHERE order_id IS NULL OR order_id NOT IN (SELECT id FROM orders)`)
	return count, err
}

func sortBySeverity(incidents []Incident) {
	// Simple insertion sort — list is always short
	for i := 1; i < len(incidents); i++ {
		for j := i; j > 0 && incidents[j].SeverityScore > incidents[j-1].SeverityScore; j-- {
			incidents[j], incidents[j-1] = incidents[j-1], incidents[j]
		}
	}
}
