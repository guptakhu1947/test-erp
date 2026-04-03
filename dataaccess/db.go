package dataaccess

import (
	"log"

	"github.com/jmoiron/sqlx"
	_ "modernc.org/sqlite"
)

var DB *sqlx.DB

func InitDB(dataSourceName string) {
	var err error
	DB, err = sqlx.Connect("sqlite", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	migrate(DB)
	seed(DB)
	log.Println("Database initialised:", dataSourceName)
}

func migrate(db *sqlx.DB) {
	schema := `
	CREATE TABLE IF NOT EXISTS suppliers (
		id            INTEGER PRIMARY KEY AUTOINCREMENT,
		name          TEXT    NOT NULL,
		contact_email TEXT    NOT NULL,
		country       TEXT    NOT NULL
	);

	CREATE TABLE IF NOT EXISTS orders (
		id          INTEGER PRIMARY KEY AUTOINCREMENT,
		supplier_id INTEGER NOT NULL,
		order_date  TEXT    NOT NULL,
		amount      REAL    NOT NULL,
		status      TEXT    NOT NULL,
		FOREIGN KEY (supplier_id) REFERENCES suppliers(id)
	);

	CREATE TABLE IF NOT EXISTS invoices (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
		supplier_id  INTEGER NOT NULL,
		order_id     INTEGER,
		invoice_date TEXT    NOT NULL,
		amount       REAL    NOT NULL,
		status       TEXT    NOT NULL,
		FOREIGN KEY (supplier_id) REFERENCES suppliers(id),
		FOREIGN KEY (order_id)    REFERENCES orders(id)
	);
	`
	db.MustExec(schema)
}

func seed(db *sqlx.DB) {
	var count int
	db.Get(&count, "SELECT COUNT(*) FROM suppliers")
	if count > 0 {
		return
	}

	suppliers := []struct {
		name, email, country string
	}{
		{"Acme Corp", "contact@acme.com", "USA"},
		{"Global Supplies Ltd", "info@globalsupplies.com", "UK"},
		{"Tech Parts Inc", "sales@techparts.com", "Germany"},
		{"Eastern Traders", "east@traders.com", "India"},
		{"Pacific Goods Co", "orders@pacificgoods.com", "Japan"},
	}

	for _, s := range suppliers {
		db.MustExec(
			`INSERT INTO suppliers (name, contact_email, country) VALUES (?, ?, ?)`,
			s.name, s.email, s.country,
		)
	}

	// orders: supplier_id, date, amount, status
	orders := []struct {
		supplierID int
		date       string
		amount     float64
		status     string
	}{
		{1, "2024-01-05", 12500.00, "completed"},
		{1, "2024-02-10", 8300.00, "completed"},
		{1, "2024-03-15", 4200.00, "pending"},
		{2, "2024-01-20", 22000.00, "completed"},
		{2, "2024-02-28", 15600.00, "completed"},
		{3, "2024-01-08", 9400.00, "completed"},
		{3, "2024-03-01", 7800.00, "pending"},
		{4, "2024-02-14", 5100.00, "completed"},
		{4, "2024-03-22", 6700.00, "completed"},
		{4, "2024-04-01", 3200.00, "pending"},
		{5, "2024-01-30", 18900.00, "completed"},
		{5, "2024-03-10", 11200.00, "completed"},
	}

	for _, o := range orders {
		db.MustExec(
			`INSERT INTO orders (supplier_id, order_date, amount, status) VALUES (?, ?, ?, ?)`,
			o.supplierID, o.date, o.amount, o.status,
		)
	}

	// invoices: supplier_id, order_id, date, amount, status
	invoices := []struct {
		supplierID, orderID int
		date                string
		amount              float64
		status              string
	}{
		{1, 1, "2024-01-12", 12500.00, "paid"},
		{1, 2, "2024-02-18", 8300.00, "paid"},
		{1, 3, "2024-03-20", 4200.00, "pending"},
		{2, 4, "2024-01-28", 22000.00, "paid"},
		{2, 5, "2024-03-05", 15600.00, "paid"},
		{3, 6, "2024-01-15", 9400.00, "paid"},
		{3, 7, "2024-03-08", 7800.00, "pending"},
		{4, 8, "2024-02-20", 5100.00, "paid"},
		{4, 9, "2024-03-28", 6700.00, "paid"},
		{4, 10, "2024-04-05", 3200.00, "pending"},
		{5, 11, "2024-02-05", 18900.00, "paid"},
		{5, 12, "2024-03-18", 11200.00, "paid"},
	}

	for _, i := range invoices {
		db.MustExec(
			`INSERT INTO invoices (supplier_id, order_id, invoice_date, amount, status) VALUES (?, ?, ?, ?, ?)`,
			i.supplierID, i.orderID, i.date, i.amount, i.status,
		)
	}

	log.Println("Database seeded with sample supplier data")
}
