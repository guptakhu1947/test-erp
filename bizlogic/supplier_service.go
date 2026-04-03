package bizlogic

import "erp-app/dataaccess"

// SupplierSummary re-exports the data-access type so handlers only import bizlogic.
type SupplierSummary = dataaccess.SupplierSummary

// GetAllSuppliers returns all suppliers with their aggregated order, invoice, and spend data.
func GetAllSuppliers() ([]SupplierSummary, error) {
	return dataaccess.GetAllSupplierSummaries()
}
