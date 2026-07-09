package app

import (
	"context"

	"github.com/martin3zra/forge/validator"
)

// tenantExists builds an `exists` rule bound to the company making the request.
//
// The bare `exists:customers,id` rule only asks whether a row with that id exists
// anywhere in the table. Every table it is used against here is tenant-owned, so a
// request could name another company's customer, item, warehouse or tax receipt and
// pass validation. The repositories all carry `company_id` in their WHERE clauses,
// which is what actually prevents the foreign row being read or written — but that
// leaves validation asserting something weaker than it appears to.
//
// forge v0.1.4 added where clauses to `exists` (its `unique` rule always had them).
// This binds the lookup to the current company.
//
// A form validated outside a company context — account signup, company creation —
// has no company to scope to, and falls back to the unscoped rule. Those forms do
// not reference tenant-owned tables.
func tenantExists(ctx context.Context, table, column string) *validator.Exists {
	rule := validator.Rule{}.Exists(table, column)
	if company := CurrentCompany(ctx); company != nil {
		return rule.Where("company_id", company.ID)
	}
	return rule
}
