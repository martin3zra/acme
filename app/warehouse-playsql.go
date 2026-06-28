package app

import (
	"context"
	"time"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

// warehouseModel is the playsql view of the warehouses table.
type warehouseModel struct {
	ID        int        `db:"id"`
	UUID      string     `db:"uuid"`
	Name      string     `db:"name"`
	Location  *string    `db:"location"`
	Status    string     `db:"status"`
	CompanyID int        `db:"company_id"`
	CreatedAt *time.Time `db:"created_at"`
	UpdatedAt *time.Time `db:"updated_at"`
	DeletedAt *time.Time `db:"deleted_at"`
}

func (warehouseModel) TableName() string { return "warehouses" }

// listWarehousesByCompany returns active warehouses for a company, ordered by
// name. No enum predicate here, so the query runs entirely through playsql.
func listWarehousesByCompany(ctx context.Context, db *playsql.DB, companyID int) ([]warehouseModel, error) {
	return playsql.Query[warehouseModel](db).
		WhereEq("company_id", companyID).
		WhereNull("deleted_at").
		OrderBy("name", playsql.Asc).
		Get(ctx)
}

func toWarehouse(m warehouseModel) *warehouse {
	location := ""
	if m.Location != nil {
		location = *m.Location
	}
	return &warehouse{
		ID:       m.ID,
		UUID:     m.UUID,
		Name:     m.Name,
		Location: location,
		Status:   foundation.Status(m.Status),
		Timestamps: foundation.Timestamps{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
	}
}
