package check

import (
	"context"

	"github.com/jmoiron/sqlx"
)

// MissingField represents a missing field.
type MissingField struct {
	Table string

	ID     string
	Source string

	MissingField string
}

// Missing checks for missing fields in the reporting database.
func Missing(ctx context.Context, tx sqlx.QueryerContext) ([]MissingField, error) {
	var missing []MissingField

	q := `
		SELECT 'categories' as table, id, source, 'target' as missingfield FROM categories WHERE target IS NULL OR target = ''
	UNION ALL
		SELECT 'tenants' as table, id, source, 'target' as missingfield FROM tenants WHERE target IS NULL OR target = ''
	UNION ALL
		SELECT 'products' as table, id, source, 'target' as missingfield FROM products WHERE target IS NULL OR target = ''
	UNION ALL
		SELECT 'products' as table, id, source, 'amount' as missingfield FROM products WHERE amount = 0
	UNION ALL
		SELECT 'products' as table, id, source, 'unit' as missingfield FROM products WHERE unit = ''
	`

	err := sqlx.SelectContext(ctx, tx, &missing, q)
	return missing, err
}
