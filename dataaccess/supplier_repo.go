package dataaccess

// SupplierSummary holds aggregated supplier data returned from the DB.
type SupplierSummary struct {
	ID           int     `db:"id"            json:"id"`
	Name         string  `db:"name"          json:"name"`
	ContactEmail string  `db:"contact_email" json:"contact_email"`
	Country      string  `db:"country"       json:"country"`
	OrderCount   int     `db:"order_count"   json:"order_count"`
	InvoiceCount int     `db:"invoice_count" json:"invoice_count"`
	TotalSpend   float64 `db:"total_spend"   json:"total_spend"`
}

// GetAllSupplierSummaries returns every supplier with aggregated order, invoice, and spend data.
// Orders and invoices are pre-aggregated in subqueries to avoid fan-out multiplication when
// both tables are joined to the same supplier row.
func GetAllSupplierSummaries() ([]SupplierSummary, error) {
	query := `
		SELECT
			s.id,
			s.name,
			s.contact_email,
			s.country,
			COALESCE(o.order_count,   0) AS order_count,
			COALESCE(i.invoice_count, 0) AS invoice_count,
			COALESCE(i.total_spend,   0) AS total_spend
		FROM suppliers s
		LEFT JOIN (
			SELECT supplier_id, COUNT(*) AS order_count
			FROM orders
			GROUP BY supplier_id
		) o ON o.supplier_id = s.id
		LEFT JOIN (
			SELECT supplier_id, COUNT(*) AS invoice_count, SUM(amount) AS total_spend
			FROM invoices
			GROUP BY supplier_id
		) i ON i.supplier_id = s.id
		ORDER BY total_spend DESC
	`
	var summaries []SupplierSummary
	err := DB.Select(&summaries, query)
	return summaries, err
}
