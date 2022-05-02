// Package invoice allows generating invoices from a filled report database.
package invoice

import (
	"context"
	"fmt"
	"time"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/jmoiron/sqlx"
)

// Invoice represents an invoice for a tenant.
type Invoice struct {
	Tenant Tenant

	PeriodStart time.Time
	PeriodEnd   time.Time

	Categories []Category
	// Total represents the total accumulated cost of the invoice.
	Total float64
}

// Category represents a category of the invoice i.e. a namespace.
type Category struct {
	ID     string
	Source string
	Target string
	Items  []Item
	// Total represents the total accumulated cost per category.
	Total float64
}

// Item represents a line in the invoice.
type Item struct {
	// Description describes the line item.
	Description string
	// Product describes the product this item is based on.
	ProductRef
	// Quantity represents the amount of the resource used.
	Quantity float64
	// QuantityMin represents the minimum amount of the resource used.
	QuantityMin float64
	// QuantityAvg represents the average amount of the resource used.
	QuantityAvg float64
	// QuantityMax represents the maximum amount of the resource used.
	QuantityMax float64

	// Unit represents the unit of the item. e.g. MiB
	Unit string
	// PricePerUnit represents the price per unit in Rappen
	PricePerUnit float64
	// Discount represents a discount in percent. 0.3 discount equals price per unit * 0.7
	Discount float64
	// Total represents the total accumulated cost.
	// (hour1 * quantity * price per unit * discount) + (hour2 * quantity * price
	// per unit * discount)
	Total float64
}

// Tenant represents a tenant in the invoice.
type Tenant struct {
	ID     string
	Source string
	Target string
}

// ProductRef represents a product reference in the invoice.
type ProductRef struct {
	ID     string `db:"product_ref_id"`
	Source string `db:"product_ref_source"`
	Target string `db:"product_ref_target"`
}

// Generate generates invoices for the given month.
// No data is written to the database. The transaction can be read-only.
func Generate(ctx context.Context, tx *sqlx.Tx, year int, month time.Month) ([]Invoice, error) {
	tenants, err := tenantsForPeriod(ctx, tx, year, month)
	if err != nil {
		return nil, err
	}

	invoices := make([]Invoice, 0, len(tenants))
	for _, tenant := range tenants {
		invoice, err := invoiceForTenant(ctx, tx, tenant, year, month)
		if err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

func invoiceForTenant(ctx context.Context, tx *sqlx.Tx, tenant db.Tenant, year int, month time.Month) (Invoice, error) {
	var categories []db.Category
	err := sqlx.SelectContext(ctx, tx, &categories,
		`SELECT DISTINCT categories.*
			FROM categories
			INNER JOIN facts ON (facts.category_id = categories.id)
			INNER JOIN date_times ON (facts.date_time_id = date_times.id)
			WHERE date_times.year = $1 AND date_times.month = $2
			AND facts.tenant_id = $3
			ORDER BY categories.source
			`,
		year, int(month), tenant.Id)

	if err != nil {
		return Invoice{}, fmt.Errorf("failed to load categories for %q at %d %s: %w", tenant.Source, year, month.String(), err)
	}

	invCategories := make([]Category, 0, len(categories))
	for _, category := range categories {
		items, err := itemsForCategory(ctx, tx, tenant, category, year, month)
		if err != nil {
			return Invoice{}, err
		}
		invCategories = append(invCategories, Category{
			ID:     category.Id,
			Source: category.Source,
			Target: category.Target.String,
			Items:  items,
			Total:  sumCategoryTotal(items),
		})
	}

	return Invoice{
		Tenant:      Tenant{ID: tenant.Id, Source: tenant.Source, Target: tenant.Target.String},
		PeriodStart: time.Date(year, month, 1, 0, 0, 0, 0, time.UTC),
		PeriodEnd:   time.Date(year, month, 1, 0, 0, 0, 0, time.UTC).AddDate(0, 1, -1),
		Categories:  invCategories,
		Total:       sumInvoiceTotal(invCategories),
	}, nil
}

func itemsForCategory(ctx context.Context, tx *sqlx.Tx, tenant db.Tenant, category db.Category, year int, month time.Month) ([]Item, error) {
	var items []Item
	err := sqlx.SelectContext(ctx, tx, &items,
		`SELECT queries.description,
				SUM(facts.quantity) as quantity, MIN(facts.quantity) as quantitymin, AVG(facts.quantity) as quantityavg, MAX(facts.quantity) as quantitymax,
				products.unit, products.amount AS pricePerUnit, discounts.discount,
				products.id as product_ref_id, products.source as product_ref_source, COALESCE(products.target,''::text) as product_ref_target,
				SUM( facts.quantity * products.amount * ( 1::double precision - discounts.discount ) ) AS total
			FROM facts
				INNER JOIN tenants    ON (facts.tenant_id = tenants.id)
				INNER JOIN queries    ON (facts.query_id = queries.id)
				LEFT  JOIN subqueries ON (subqueries.query_id = queries.id)
				INNER JOIN discounts  ON (facts.discount_id = discounts.id)
				INNER JOIN products   ON (facts.product_id = products.id)
				INNER JOIN date_times ON (facts.date_time_id = date_times.id)
			WHERE date_times.year = $1 AND date_times.month = $2
				AND facts.tenant_id = $3
				AND facts.category_id = $4
        AND subqueries.parent_id is NULL
			GROUP BY queries.description, products.amount, products.unit, products.id, products.source, products.target, discounts.discount
		`,
		year, int(month), tenant.Id, category.Id)

	if err != nil {
		return nil, fmt.Errorf("failed to load item for %q/%q at %d %s: %w", tenant.Source, category.Source, year, month.String(), err)
	}

	return items, nil
}

func tenantsForPeriod(ctx context.Context, tx *sqlx.Tx, year int, month time.Month) ([]db.Tenant, error) {
	var tenants []db.Tenant

	err := sqlx.SelectContext(ctx, tx, &tenants,
		`SELECT DISTINCT tenants.*
			FROM tenants
				INNER JOIN facts ON (facts.tenant_id = tenants.id)
				INNER JOIN date_times ON (facts.date_time_id = date_times.id)
			WHERE date_times.year = $1 AND date_times.month = $2
			ORDER BY tenants.source
		`,
		year, int(month))

	if err != nil {
		return nil, fmt.Errorf("failed to load tenants for %d %s: %w", year, month.String(), err)
	}
	return tenants, nil
}

func sumCategoryTotal(itms []Item) (sum float64) {
	for _, itm := range itms {
		sum += itm.Total
	}
	return
}

func sumInvoiceTotal(cat []Category) (sum float64) {
	for _, itm := range cat {
		sum += itm.Total
	}
	return
}
