package app

import (
	"context"
	"time"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/playsql"
)

// vendorModel is the playsql view of the vendors table. It maps real columns
// only (the API-facing `vendor` struct carries derived/non-column fields), and
// is converted to *vendor by toVendor.
type vendorModel struct {
	ID            int        `db:"id"`
	UUID          string     `db:"uuid"`
	Code          string     `db:"code"`
	Name          string     `db:"name"`
	ContactName   string     `db:"contact_name"`
	Phone         string     `db:"phone"`
	Email         string     `db:"email"`
	Status        string     `db:"status"`
	AmountPayable float64    `db:"amount_payable"`
	PurchaseNote  string     `db:"purchase_note"`
	LeadTimeDays  int        `db:"lead_time_days"`
	VendorType    string     `db:"vendor_type"`
	PaymentMethod string     `db:"payment_method"`
	PaymentTerms  string     `db:"payment_terms"`
	CompanyID     int        `db:"company_id"`
	CreatedAt     *time.Time `db:"created_at"`
	UpdatedAt     *time.Time `db:"updated_at"`
	DeletedAt     *time.Time `db:"deleted_at"`
}

func (vendorModel) TableName() string { return "vendors" }

// listVendorsByCompany is the playsql-backed read: active vendors for a company,
// ordered by name. Kept driver-agnostic so it can run against any playsql DB.
func listVendorsByCompany(ctx context.Context, db *playsql.DB, companyID int) ([]vendorModel, error) {
	return playsql.Query[vendorModel](db).
		WhereEq("company_id", companyID).
		WhereNull("deleted_at").
		OrderBy("name", playsql.Asc).
		Get(ctx)
}

func toVendor(m vendorModel) *vendor {
	return &vendor{
		ID:            m.ID,
		UUID:          m.UUID,
		Code:          m.Code,
		Name:          m.Name,
		ContactName:   m.ContactName,
		Phone:         m.Phone,
		Email:         m.Email,
		Status:        foundation.Status(m.Status),
		AmountPayable: m.AmountPayable,
		PurchaseNote:  m.PurchaseNote,
		LeadTimeDays:  m.LeadTimeDays,
		VendorType:    m.VendorType,
		PaymentMethod: m.PaymentMethod,
		PaymentTerms:  m.PaymentTerms,
		Address:       "LOUISVILLE, Selby 3864 Johnson Street, United States of America",
		Timestamps: foundation.Timestamps{
			CreatedAt: m.CreatedAt,
			UpdatedAt: m.UpdatedAt,
			DeletedAt: m.DeletedAt,
		},
	}
}
