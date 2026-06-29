package app

import (
	"fmt"
	"strconv"
	"testing"
)

func validVendorRecord() map[string]any {
	return map[string]any{
		"name":           "Distribuidora Andina",
		"contact_name":   "Carlos Gómez",
		"phone":          "8095550123",
		"email":          "ventas@andina.com",
		"payment_method": "CK",
		"payment_terms":  "30",
		"purchase_note":  "Entregar en almacén central",
		"lead_time_days": "7",
		"code":           "PRV-001",
		"vendor_type":    "Negocio",
	}
}

func hasWarning(warnings []ImportIssue, column string) bool {
	for _, w := range warnings {
		if w.Column == column && w.Level == IssueLevel.Warning {
			return true
		}
	}
	return false
}

func TestMapToStoreVendorForm_Valid(t *testing.T) {
	var warnings []ImportIssue
	form, code, err := mapToStoreVendorForm(1, validVendorRecord(), &warnings)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form == nil || code == nil {
		t.Fatalf("expected form and code, got form=%v code=%v", form, code)
	}
	if *code != "PRV-001" {
		t.Errorf("code: want PRV-001, got %q", *code)
	}
	if form.Name != "Distribuidora Andina" {
		t.Errorf("name: got %q", form.Name)
	}
	if form.Email != "ventas@andina.com" {
		t.Errorf("email: got %q", form.Email)
	}
	if form.Phone != "8095550123" {
		t.Errorf("phone: got %q", form.Phone)
	}
	if form.PaymentMethod != "ck" {
		t.Errorf("payment_method should be lowercased, got %q", form.PaymentMethod)
	}
	if form.PaymentTerms != "net30" {
		t.Errorf("payment_terms: want net30, got %q", form.PaymentTerms)
	}
	if form.LeadTimeDays != 7 {
		t.Errorf("lead_time_days: want 7, got %d", form.LeadTimeDays)
	}
	if form.VendorType != "business" {
		t.Errorf("vendor_type should normalize unknown label to business, got %q", form.VendorType)
	}
	if len(warnings) != 0 {
		t.Errorf("expected no warnings, got %v", warnings)
	}
}

func TestMapToStoreVendorForm_MissingNameFails(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	delete(rec, "name")

	form, _, err := mapToStoreVendorForm(2, rec, &warnings)
	if err == nil {
		t.Fatal("expected error when name is missing")
	}
	if form != nil {
		t.Errorf("expected nil form, got %v", form)
	}
}

func TestMapToStoreVendorForm_MissingCodeFailsWithWarning(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	delete(rec, "code")

	form, code, err := mapToStoreVendorForm(3, rec, &warnings)
	if err == nil {
		t.Fatal("expected error when code is missing")
	}
	if form != nil || code != nil {
		t.Errorf("expected nil form and code, got form=%v code=%v", form, code)
	}
	if !hasWarning(warnings, "code") {
		t.Errorf("expected a code warning, got %v", warnings)
	}
}

func TestMapToStoreVendorForm_InvalidEmailBlankedWithWarning(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	rec["email"] = "not-an-email"

	form, _, err := mapToStoreVendorForm(4, rec, &warnings)
	if err != nil {
		t.Fatalf("invalid email should not fail the row: %v", err)
	}
	if form.Email != "" {
		t.Errorf("invalid email should be blanked, got %q", form.Email)
	}
	if !hasWarning(warnings, "email") {
		t.Errorf("expected an email warning, got %v", warnings)
	}
}

func TestMapToStoreVendorForm_MissingEmailIsNullable(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	delete(rec, "email")

	form, _, err := mapToStoreVendorForm(5, rec, &warnings)
	if err != nil {
		t.Fatalf("missing email should not fail the row: %v", err)
	}
	if form.Email != "" {
		t.Errorf("expected empty email, got %q", form.Email)
	}
	if hasWarning(warnings, "email") {
		t.Errorf("missing (nullable) email should not warn, got %v", warnings)
	}
}

func TestMapToStoreVendorForm_MultiplePhonesKeepsFirstAndWarns(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	rec["phone"] = "8095550123 / 8295559876"

	form, _, err := mapToStoreVendorForm(6, rec, &warnings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.Phone != "8095550123" {
		t.Errorf("expected first phone, got %q", form.Phone)
	}
	if !hasWarning(warnings, "phone") {
		t.Errorf("expected a phone warning, got %v", warnings)
	}
}

func TestMapToStoreVendorForm_PaymentTermsClampedToMax(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	rec["payment_terms"] = strconv.Itoa(PaymentTermsMax + 30)

	form, _, err := mapToStoreVendorForm(7, rec, &warnings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.PaymentTerms != fmt.Sprintf("net%d", PaymentTermsMax) {
		t.Errorf("expected clamp to net%d, got %q", PaymentTermsMax, form.PaymentTerms)
	}
	if !hasWarning(warnings, "payment_terms") {
		t.Errorf("expected a payment_terms warning, got %v", warnings)
	}
}

func TestMapToStoreVendorForm_NonNumericPaymentTermsDefaultsAndWarns(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	rec["payment_terms"] = "thirty"

	form, _, err := mapToStoreVendorForm(8, rec, &warnings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.PaymentTerms != "net0" {
		t.Errorf("expected default net0, got %q", form.PaymentTerms)
	}
	if !hasWarning(warnings, "payment_terms") {
		t.Errorf("expected a payment_terms warning, got %v", warnings)
	}
}

func TestMapToStoreVendorForm_InvalidLeadTimeDefaultsAndWarns(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	rec["lead_time_days"] = "soon"

	form, _, err := mapToStoreVendorForm(9, rec, &warnings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.LeadTimeDays != 0 {
		t.Errorf("expected default 0, got %d", form.LeadTimeDays)
	}
	if !hasWarning(warnings, "lead_time_days") {
		t.Errorf("expected a lead_time_days warning, got %v", warnings)
	}
}

func TestMapToStoreVendorForm_IndividualVendorTypePreserved(t *testing.T) {
	var warnings []ImportIssue
	rec := validVendorRecord()
	rec["vendor_type"] = "Individual"

	form, _, err := mapToStoreVendorForm(10, rec, &warnings)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if form.VendorType != "individual" {
		t.Errorf("expected individual, got %q", form.VendorType)
	}
}
