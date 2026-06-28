package app

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
)

func mapToStoreVendorForm(rowNum int, data map[string]any, warnings *[]ImportIssue) (*StoreVendorForm, *string, error) {
	// Required fields
	name, ok := getString(data, "name")
	if !ok {
		return nil, nil, errors.New("The column `name` is missing or malformed")
	}

	code, ok := getString(data, "code")
	if !ok {
		*warnings = append(*warnings, ImportIssue{
			Row:     rowNum,
			Column:  "code",
			Level:   IssueLevel.Warning,
			Value:   fmt.Sprint(data["code"]),
			Message: "The column `code` is missing or malformed",
		})
		return nil, nil, errors.New("The column `code` is missing or malformed")
	}

	contactName, _ := getString(data, "contact_name")

	// Email is optional and nullable: a vendor row may legitimately omit it.
	// NOTE: ensure `vendors.email` is nullable at the DB level and that any
	// UNIQUE index on it ignores NULLs (partial index / NULLS NOT DISTINCT off)
	// so multiple emailless vendors can be imported without collisions.
	// Validate only when a value is present.
	email, _ := getString(data, "email")
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "email",
				Level:   IssueLevel.Warning,
				Value:   email,
				Message: "Invalid email address; imported as empty",
			})
			email = ""
		}
	}

	var phone string
	if phoneData, ok := getString(data, "phone"); ok {
		phones := parsePhone(phoneData)
		if len(phones) > 0 {
			phone = phones[0]
			if len(phones) > 1 {
				*warnings = append(*warnings, ImportIssue{
					Row:     rowNum,
					Column:  "phone",
					Level:   IssueLevel.Warning,
					Value:   phoneData,
					Message: "Multiple phone numbers detected; only the first was imported",
				})
			}
		}
	}

	purchaseNote, _ := getString(data, "purchase_note")

	leadTime := 0
	if lt, ok := getString(data, "lead_time_days"); ok && lt != "" {
		if v, err := strconv.Atoi(lt); err == nil {
			leadTime = v
		} else {
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "lead_time_days",
				Level:   IssueLevel.Warning,
				Value:   lt,
				Message: "Lead time is invalid; defaulted to 0",
			})
		}
	}

	paymentMethod, _ := getString(data, "payment_method")

	// Payment terms arrive as a number of days; normalize to the `net{n}` form.
	terms := "net0"
	if pt, ok := getString(data, "payment_terms"); ok && pt != "" {
		if val, err := strconv.Atoi(pt); err != nil {
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "payment_terms",
				Level:   IssueLevel.Warning,
				Value:   pt,
				Message: "Payment terms is invalid; defaulted to net0",
			})
		} else {
			if val > PaymentTermsMax {
				val = PaymentTermsMax
				*warnings = append(*warnings, ImportIssue{
					Row:     rowNum,
					Column:  "payment_terms",
					Level:   IssueLevel.Warning,
					Value:   pt,
					Message: "Payment terms is exceeding the max value allowed of 120 days",
				})
			}
			terms = fmt.Sprintf("net%d", val)
		}
	}

	vendorType, _ := getString(data, "vendor_type")
	vendorType = strings.ToLower(strings.TrimSpace(vendorType))
	if vendorType != "individual" && vendorType != "business" {
		vendorType = "business"
	}

	return &StoreVendorForm{
		Name:          name,
		Contact:       contactName,
		Email:         email,
		Phone:         phone,
		PurchaseNote:  purchaseNote,
		LeadTimeDays:  leadTime,
		PaymentMethod: strings.ToLower(paymentMethod),
		PaymentTerms:  terms,
		CreditLimited: false,
		CreditLimit:   0,
		VendorType:    vendorType,
		TaxReceipt:    2,
	}, &code, nil
}
