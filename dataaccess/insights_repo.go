package dataaccess

type SpendBySupplier struct {
	Name  string  `db:"name"  json:"name"`
	Spend float64 `db:"spend" json:"spend"`
}

type MonthlySpend struct {
	Month string  `db:"month" json:"month"`
	Spend float64 `db:"spend" json:"spend"`
}

type StatusCount struct {
	Status string `db:"status" json:"status"`
	Count  int    `db:"count"  json:"count"`
}

type SpendByCountry struct {
	Country string  `db:"country" json:"country"`
	Spend   float64 `db:"spend"   json:"spend"`
}

func GetSpendBySupplier() ([]SpendBySupplier, error) {
	query := `
		SELECT s.name, COALESCE(SUM(i.amount), 0) AS spend
		FROM suppliers s
		LEFT JOIN invoices i ON i.supplier_id = s.id
		GROUP BY s.id
		ORDER BY spend DESC
	`
	var result []SpendBySupplier
	err := DB.Select(&result, query)
	return result, err
}

func GetMonthlySpend() ([]MonthlySpend, error) {
	query := `
		SELECT strftime('%Y-%m', invoice_date) AS month, SUM(amount) AS spend
		FROM invoices
		GROUP BY month
		ORDER BY month
	`
	var result []MonthlySpend
	err := DB.Select(&result, query)
	return result, err
}

func GetOrderStatusCounts() ([]StatusCount, error) {
	query := `SELECT status, COUNT(*) AS count FROM orders GROUP BY status`
	var result []StatusCount
	err := DB.Select(&result, query)
	return result, err
}

func GetInvoiceStatusCounts() ([]StatusCount, error) {
	query := `SELECT status, COUNT(*) AS count FROM invoices GROUP BY status`
	var result []StatusCount
	err := DB.Select(&result, query)
	return result, err
}

func GetSpendByCountry() ([]SpendByCountry, error) {
	query := `
		SELECT s.country, COALESCE(SUM(i.amount), 0) AS spend
		FROM suppliers s
		LEFT JOIN invoices i ON i.supplier_id = s.id
		GROUP BY s.country
		ORDER BY spend DESC
	`
	var result []SpendByCountry
	err := DB.Select(&result, query)
	return result, err
}
