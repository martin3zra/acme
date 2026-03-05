package app

import (
	"encoding/json"
)

// GetSampleInvoiceTemplate returns a sample invoice template JSON
func GetSampleInvoiceTemplate() json.RawMessage {
	template := map[string]any{
		"page_size":   "A4",
		"page_format": "P",
		"margins": map[string]any{
			"top":    15,
			"right":  15,
			"bottom": 15,
			"left":   15,
		},
		"elements": []any{
			map[string]any{
				"type":    "text",
				"content": "{{business.name}}",
				"size":    16,
				"bold":    true,
			},
			map[string]any{
				"type":    "text",
				"content": "{{business.address}}",
				"size":    10,
			},
			map[string]any{
				"type":      "text",
				"content":   "Invoice #{{invoice.number}}",
				"size":      12,
				"bold":      true,
				"align":     "R",
			},
			map[string]any{
				"type":    "divider",
				"spacing": 5,
			},
			map[string]any{
				"type": "grid",
				"columns": []any{
					map[string]any{
						"span": 6,
						"elements": []any{
							map[string]any{
								"type":    "text",
								"content": "Bill To:",
								"bold":    true,
								"size":    10,
							},
							map[string]any{
								"type":    "text",
								"content": "{{customer.name}}",
								"size":    10,
							},
						},
					},
					map[string]any{
						"span": 6,
						"elements": []any{
							map[string]any{
								"type":    "text",
								"content": "Date: {{invoice.date}}",
								"size":    10,
								"align":   "R",
							},
							map[string]any{
								"type":    "text",
								"content": "Due: {{invoice.due_date}}",
								"size":    10,
								"align":   "R",
							},
						},
					},
				},
			},
			map[string]any{
				"type": "table",
				"headers": []any{
					map[string]any{"label": "Item", "width": 60},
					map[string]any{"label": "Qty", "width": 20},
					map[string]any{"label": "Unit Price", "width": 30},
					map[string]any{"label": "Amount", "width": 30},
				},
				"rows": []any{
					map[string]any{
						"cells": []any{
							map[string]any{"text": "{{items[0].name}}", "width": 60},
							map[string]any{"text": "{{items[0].quantity}}", "width": 20},
							map[string]any{"text": "{{currency items[0].unit_price}}", "width": 30},
							map[string]any{"text": "{{currency items[0].total}}", "width": 30},
						},
					},
				},
				"row_height": 8,
			},
			map[string]any{
				"type":    "divider",
				"spacing": 5,
			},
			map[string]any{
				"type": "grid",
				"columns": []any{
					map[string]any{
						"span": 6,
					},
					map[string]any{
						"span": 6,
						"elements": []any{
							map[string]any{
								"type":    "text",
								"content": "Subtotal: {{currency invoice.subtotal}}",
								"size":    10,
								"align":   "R",
							},
							map[string]any{
								"type":    "text",
								"content": "Tax: {{currency invoice.tax}}",
								"size":    10,
								"align":   "R",
							},
							map[string]any{
								"type":    "text",
								"content": "Total: {{currency invoice.total}}",
								"size":    12,
								"bold":    true,
								"align":   "R",
							},
						},
					},
				},
			},
		},
	}

	data, _ := json.Marshal(template)
	return data
}

// GetSampleInvoiceData returns sample data for testing invoice template
func GetSampleInvoiceData() map[string]any {
	return map[string]any{
		"business": map[string]any{
			"name":    "Acme Corporation",
			"address": "123 Main Street",
		},
		"invoice": map[string]any{
			"number":    "INV-001",
			"date":      "2026-03-05",
			"due_date":  "2026-04-05",
			"subtotal":  950.00,
			"tax":       95.00,
			"total":     1045.00,
		},
		"customer": map[string]any{
			"name": "John Doe",
		},
		"items": []any{
			map[string]any{
				"name":       "Professional Services",
				"quantity":   10,
				"unit_price": 95.00,
				"total":      950.00,
			},
		},
	}
}

// GetSampleEstimateTemplate returns a sample estimate/quote template
func GetSampleEstimateTemplate() json.RawMessage {
	template := map[string]any{
		"page_size":   "A4",
		"page_format": "P",
		"margins": map[string]any{
			"top":    15,
			"right":  15,
			"bottom": 15,
			"left":   15,
		},
		"elements": []any{
			map[string]any{
				"type":    "text",
				"content": "ESTIMATE",
				"size":    18,
				"bold":    true,
			},
			map[string]any{
				"type":    "text",
				"content": "Estimate #{{estimate.number}}",
				"size":    12,
			},
			map[string]any{
				"type": "table",
				"headers": []any{
					map[string]any{"label": "Item", "width": 70},
					map[string]any{"label": "Amount", "width": 30},
				},
				"rows": []any{
					map[string]any{
						"cells": []any{
							map[string]any{"text": "{{items[0].description}}", "width": 70},
							map[string]any{"text": "{{currency items[0].amount}}", "width": 30},
						},
					},
				},
				"row_height": 8,
			},
			map[string]any{
				"type":    "text",
				"content": "Valid until: {{estimate.valid_until}}",
				"size":    10,
			},
			map[string]any{
				"type":    "text",
				"content": "Total: {{currency estimate.total}}",
				"size":    12,
				"bold":    true,
			},
		},
	}

	data, _ := json.Marshal(template)
	return data
}

// GetSampleEstimateData returns sample data for testing estimate template
func GetSampleEstimateData() map[string]any {
	return map[string]any{
		"estimate": map[string]any{
			"number":       "EST-001",
			"valid_until":  "2026-04-05",
			"total":        2500.00,
		},
		"items": []any{
			map[string]any{
				"description": "Web Development - 50 hours @ $50/hr",
				"amount":      2500.00,
			},
		},
	}
}
