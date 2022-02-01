package check

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v4"
	"github.com/jmoiron/sqlx"
)

type Missing struct {
	Table string

	ID     string
	Source string

	MissingField string
}

func CheckMissing() (Missing, error) {

}

func checkString(ctx context.Context, tx *sqlx.Tx, table, column pgx.Identifier) ([]Missing, error) {
	var missing []Missing

	q := `
		SELECT 'categories' as table, id, source, 'target' as missingfield FROM categories WHERE target IS NULL OR target = ''
	UNION ALL
		SELECT 'tenants' as table, id, source, 'target' as missingfield FROM tenants WHERE target IS NULL OR target = ''
	UNION ALL
		SELECT 'products' as table, id, source, 'target' as missingfield FROM products WHERE target IS NULL OR target = ''
	UNION ALL
		SELECT 'products' as table, id, source, 'target' as missingfield FROM products WHERE target IS NULL OR target = ''

	`

	query := fmt.Sprintf("SELECT $1 FROM %s WHERE %s IS NULL OR %s == ''", table.Sanitize(), column.Sanitize())

	sqlx.SelectContext(ctx, tx, &missing, query)

	return missing, nil
}
