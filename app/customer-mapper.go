package app

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

func mapToStoreCustomerForm(rowNum int, data map[string]any, warnings *[]ImportIssue) (*StoreCustomerForm, *string, error) {
	// Required fields
	name, ok := getString(data, "name")
	if !ok {
		return nil, nil, errors.New("The column `name` is missing or malformed")
	}
	contactName, _ := getString(data, "contact_name")
	code, ok := getString(data, "code")
	if !ok {
		*warnings = append(*warnings, ImportIssue{
			Row:     rowNum,
			Column:  "code",
			Level:   IssueLevel.Warning,
			Value:   data["code"].(string),
			Message: "The column `code` is missing or malformed",
		})
		return nil, nil, errors.New("The column `code` is missing or malformed")
	}

	email, _ := getString(data, "email")
	if !ok {
		*warnings = append(*warnings, ImportIssue{
			Row:     rowNum,
			Column:  "email",
			Level:   IssueLevel.Warning,
			Value:   data["email"].(string),
			Message: "Invalid or empty Email address",
		})
	}
	var phone string
	phoneData, ok := getString(data, "phone")
	if ok {
		phones := parsePhone(phoneData)
		if len(phones) > 0 {
			phone = phones[0]
			if len(phones) > 1 {
				*warnings = append(*warnings, ImportIssue{
					Row:     rowNum,
					Column:  "phone",
					Level:   IssueLevel.Warning,
					Value:   data["phone"].(string),
					Message: "Multiple phone numbers detected; only the first was imported",
				})
			}
		}
	}

	paymentMethod, ok := getString(data, "payment_method")
	if !ok {
		return nil, nil, errors.New("The column `payment_method` is missing or malformed")
	}
	paymentTerms, ok := getString(data, "payment_terms")
	if !ok {
		return nil, nil, errors.New("The column `payment_terms` is missing or malformed")
	} else {
		if val, err := strconv.Atoi(paymentTerms); err != nil {
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "payment_terms",
				Level:   IssueLevel.Warning,
				Value:   data["payment_terms"].(string),
				Message: "Payment terms is invalid",
			})
		} else if val > PaymentTermsMax {
			paymentTerms = strconv.Itoa(PaymentTermsMax)
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "payment_terms",
				Level:   IssueLevel.Warning,
				Value:   data["payment_terms"].(string),
				Message: "Payment terms is exceeding the max value allowed of 120 days",
			})
		}
	}
	creditLimited, ok := getBoolean(data, "credit_limited")
	if !ok {
		return nil, nil, errors.New("The column `credit_limited` is missing or malformed")
	}

	creditLimit, ok, err := getFloat64(data, "credit_limit")
	if !ok && err == nil {
		return nil, nil, errors.New("The column `credit_limit` is missing or malformed")
	} else if err != nil {
		switch err {
		case ErrMaskedValue:
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "credit_limit",
				Level:   IssueLevel.Warning,
				Value:   data["credit_limit"].(string),
				Message: "Masked value ignored",
			})
		case ErrInvalidNum:
			*warnings = append(*warnings, ImportIssue{
				Row:     rowNum,
				Column:  "credit_limit",
				Level:   IssueLevel.Warning,
				Value:   data["credit_limit"].(string),
				Message: "Invalid number",
			})
		}
	}

	return &StoreCustomerForm{
		Name:          name,
		Contact:       contactName,
		Email:         email,
		Phone:         phone,
		PaymentMethod: strings.ToLower(paymentMethod),
		PaymentTerms:  fmt.Sprintf("net%s", paymentTerms),
		CreditLimited: creditLimited,
		CreditLimit:   creditLimit,
		TaxReceipt:    2,
		CustomerType:  "business",
	}, &code, nil
}

func parsePhone(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	// Common separators
	separators := []string{"/", ",", ";", "|"}
	for _, sep := range separators {
		value = strings.ReplaceAll(value, sep, " ")
	}

	parts := strings.Fields(value)

	var phones []string
	for _, p := range parts {
		clean := normalizePhone(p)
		if clean != "" {
			phones = append(phones, clean)
		}
	}

	return phones
}

func normalizePhone(p string) string {
	var b strings.Builder
	for _, r := range p {
		if unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}

	phone := b.String()

	// basic sanity check
	if len(phone) < 7 {
		return ""
	}

	// enforce DB limit
	if len(phone) > 20 {
		return phone[:20]
	}

	return phone
}
