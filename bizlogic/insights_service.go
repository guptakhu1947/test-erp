package bizlogic

import (
	"fmt"

	"erp-app/dataaccess"
)

type Recommendation struct {
	Type        string `json:"type"`  // "warning" | "info" | "success"
	Title       string `json:"title"`
	Description string `json:"description"`
}

type InsightsDashboard struct {
	SpendBySupplier []dataaccess.SpendBySupplier `json:"spend_by_supplier"`
	MonthlySpend    []dataaccess.MonthlySpend    `json:"monthly_spend"`
	SpendByCountry  []dataaccess.SpendByCountry  `json:"spend_by_country"`
	OrderStatus     []dataaccess.StatusCount     `json:"order_status"`
	InvoiceStatus   []dataaccess.StatusCount     `json:"invoice_status"`
	Recommendations []Recommendation             `json:"recommendations"`
}

func GetInsightsDashboard() (*InsightsDashboard, error) {
	spendBySupplier, err := dataaccess.GetSpendBySupplier()
	if err != nil {
		return nil, err
	}
	monthlySpend, err := dataaccess.GetMonthlySpend()
	if err != nil {
		return nil, err
	}
	spendByCountry, err := dataaccess.GetSpendByCountry()
	if err != nil {
		return nil, err
	}
	orderStatus, err := dataaccess.GetOrderStatusCounts()
	if err != nil {
		return nil, err
	}
	invoiceStatus, err := dataaccess.GetInvoiceStatusCounts()
	if err != nil {
		return nil, err
	}

	return &InsightsDashboard{
		SpendBySupplier: spendBySupplier,
		MonthlySpend:    monthlySpend,
		SpendByCountry:  spendByCountry,
		OrderStatus:     orderStatus,
		InvoiceStatus:   invoiceStatus,
		Recommendations: generateRecommendations(spendBySupplier, orderStatus, invoiceStatus, spendByCountry),
	}, nil
}

func generateRecommendations(
	spend []dataaccess.SpendBySupplier,
	orders []dataaccess.StatusCount,
	invoices []dataaccess.StatusCount,
	countries []dataaccess.SpendByCountry,
) []Recommendation {
	var recs []Recommendation

	// Total spend
	var totalSpend float64
	for _, s := range spend {
		totalSpend += s.Spend
	}

	// Concentration risk
	if len(spend) > 0 && totalSpend > 0 {
		pct := (spend[0].Spend / totalSpend) * 100
		if pct > 35 {
			recs = append(recs, Recommendation{
				Type:  "warning",
				Title: "Supplier Concentration Risk",
				Description: fmt.Sprintf(
					"%s accounts for %.0f%% of total spend ($%.0f). Diversify to reduce supply chain vulnerability.",
					spend[0].Name, pct, spend[0].Spend,
				),
			})
		}
	}

	// Invoice backlog
	var paidInv, pendingInv int
	for _, inv := range invoices {
		switch inv.Status {
		case "paid":
			paidInv = inv.Count
		case "pending":
			pendingInv = inv.Count
		}
	}
	if total := paidInv + pendingInv; total > 0 {
		pendingPct := float64(pendingInv) / float64(total) * 100
		if pendingPct > 20 {
			recs = append(recs, Recommendation{
				Type:  "warning",
				Title: "Outstanding Invoice Backlog",
				Description: fmt.Sprintf(
					"%d of %d invoices are unpaid (%.0f%%). Review accounts payable to avoid late fees and supplier friction.",
					pendingInv, total, pendingPct,
				),
			})
		}
	}

	// Pending orders
	var completedOrd, pendingOrd int
	for _, o := range orders {
		switch o.Status {
		case "completed":
			completedOrd = o.Count
		case "pending":
			pendingOrd = o.Count
		}
	}
	if total := completedOrd + pendingOrd; total > 0 {
		pendingPct := float64(pendingOrd) / float64(total) * 100
		if pendingPct > 20 {
			recs = append(recs, Recommendation{
				Type:  "info",
				Title: "Open Orders Require Attention",
				Description: fmt.Sprintf(
					"%d orders are still open (%.0f%% of total). Follow up with suppliers to ensure timely fulfillment.",
					pendingOrd, pendingPct,
				),
			})
		}
	}

	// Top strategic supplier
	if len(spend) > 0 {
		recs = append(recs, Recommendation{
			Type:  "success",
			Title: "Top Strategic Supplier",
			Description: fmt.Sprintf(
				"%s is your highest-value supplier at $%.0f. Consider negotiating volume discounts or preferred-partner terms.",
				spend[0].Name, spend[0].Spend,
			),
		})
	}

	// Geographic risk
	if len(countries) == 1 {
		recs = append(recs, Recommendation{
			Type:  "warning",
			Title: "Single-Region Exposure",
			Description: "All suppliers are located in one country. Introduce regional alternatives to protect against geopolitical or logistics disruptions.",
		})
	} else if len(countries) >= 3 {
		recs = append(recs, Recommendation{
			Type:  "success",
			Title: "Strong Geographic Diversification",
			Description: fmt.Sprintf(
				"Suppliers span %d countries, reducing regional dependency. Maintain relationships across all regions.",
				len(countries),
			),
		})
	}

	// Invoice payment health
	if paidInv+pendingInv > 0 {
		rate := float64(paidInv) / float64(paidInv+pendingInv) * 100
		if rate >= 80 {
			recs = append(recs, Recommendation{
				Type:  "success",
				Title: "Healthy Payment Rate",
				Description: fmt.Sprintf(
					"%.0f%% of invoices are paid — strong accounts payable hygiene. Keep payment cycles consistent to maintain supplier trust.",
					rate,
				),
			})
		}
	}

	return recs
}
