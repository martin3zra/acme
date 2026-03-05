package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/pdf"
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) listTemplatesHandler(ctx *routing.Context) {
	templates, err := s.listTemplates(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Templates/Index", map[string]any{
		"templates":    templates,
		"translations": trans("templates"),
	})
}

func (s *Server) showTemplateHandler(ctx *routing.Context) {
	templateID := ctx.Int("id")

	template, err := s.findTemplateByID(ctx.Request.Context(), templateID)
	if err != nil {
		ctx.Error(err)
		return
	}

	versions, err := s.listTemplateVersions(ctx.Request.Context(), templateID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Templates/Show", map[string]any{
		"template":     template,
		"versions":     versions,
		"translations": trans("templates"),
	})
}

func (s *Server) storeTemplateHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreTemplateForm) {
		normalizedLayoutJSON, err := NormalizeLayoutJSON(form.LayoutJSON)
		if err != nil {
			ctx.BackWithError(err)
			return
		}
		form.LayoutJSON = normalizedLayoutJSON

		service := NewTemplateService(NewTemplateRepository(s), nil)
		if err := service.ValidateTemplate(form.LayoutJSON); err != nil {
			ctx.BackWithError(err)
			return
		}

		template, err := s.storeTemplate(ctx.Request.Context(), form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.template"}))

		ctx.Redirect(fmt.Sprintf("/admin/templates/%d", template.ID))
	})
}

func (s *Server) updateTemplateHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreTemplateForm) {

		templateID := ctx.Int("id")
		normalizedLayoutJSON, err := NormalizeLayoutJSON(form.LayoutJSON)
		if err != nil {
			ctx.BackWithError(err)
			return
		}
		form.LayoutJSON = normalizedLayoutJSON

		service := NewTemplateService(NewTemplateRepository(s), nil)
		if err := service.ValidateTemplate(form.LayoutJSON); err != nil {
			ctx.BackWithError(err)
			return
		}

		_, err = s.updateTemplate(ctx.Request.Context(), templateID, form)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.template"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.template"}))

		ctx.Redirect("/admin/templates")
	})
}

func (s *Server) publishTemplateHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *PublishTemplateForm) {

		templateID := ctx.Int("id")

		version, err := s.publishTemplate(ctx.Request.Context(), templateID)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.template"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasPublished", i18n.Replacements{"subject": "@global.template"}))

		ctx.Redirect(fmt.Sprintf("/admin/templates/%d", version.TemplateID))
	})
}

func (s *Server) previewTemplateHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *PreviewTemplateForm) {
		previewData := map[string]any{}
		if len(form.Data) > 0 {
			var rawPreviewData any
			if err := json.Unmarshal(form.Data, &rawPreviewData); err != nil {
				if ctx.WantsJson() {
					ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": err.Error()})
					return
				}
				ctx.BackWithError(err)
				return
			}

			switch typedPreviewData := rawPreviewData.(type) {
			case map[string]any:
				previewData = typedPreviewData
			case string:
				if strings.TrimSpace(typedPreviewData) != "" {
					if err := json.Unmarshal([]byte(typedPreviewData), &previewData); err != nil {
						if ctx.WantsJson() {
							ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": err.Error()})
							return
						}
						ctx.BackWithError(err)
						return
					}
				}
			case nil:
				previewData = map[string]any{}
			default:
				if ctx.WantsJson() {
					ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": "data must be an object or JSON object string"})
					return
				}
				ctx.BackWithError(fmt.Errorf("data must be an object or JSON object string"))
				return
			}
		}

		// Auto-populate with dummy data if preview data is empty
		if len(previewData) == 0 {
			previewData = map[string]any{
				"estimate": map[string]any{
					"number":      "EST-2024-00456",
					"date":        "March 5, 2024",
					"valid_until": "April 4, 2024",
					"project":     "Website Redesign & Development Project",
				},
				"document": map[string]any{
					"number":     "EST-2024-00456",
					"type":       "Estimate",
					"status":     "Pending",
					"issued_at":  "2024-03-05",
					"due_date":   "2024-04-04",
					"created_at": "2024-03-05T14:30:00Z",
					"page":       1,
					"pages":      1,
				},
				"company": map[string]any{
					"name":       "Acme Corporation",
					"address":    "123 Business Street",
					"city":       "Santo Domingo",
					"identifier": "123-456-789",
					"phone":      "+1 (809) 123-4567",
					"email":      "info@acmecorp.com",
					"website":    "www.acmecorp.com",
				},
				"customer": map[string]any{
					"name":    "John Smith",
					"address": "456 Client Avenue",
					"city":    "Santiago",
					"phone":   "+1 (809) 987-6543",
					"email":   "john.smith@example.com",
				},
				"items": []map[string]any{
					{
						"id":          1,
						"name":        "Professional Web Development Services",
						"description": "80 hours of full-stack development work",
						"quantity":    80,
						"unit":        "hours",
						"price":       75.00,
						"subtotal":    6000.00,
					},
					{
						"id":          2,
						"name":        "UI/UX Design & Mockups",
						"description": "Design and wireframing services",
						"quantity":    20,
						"unit":        "hours",
						"price":       85.00,
						"subtotal":    1700.00,
					},
					{
						"id":          3,
						"name":        "Project Management & Coordination",
						"description": "Project oversight and client coordination",
						"quantity":    10,
						"unit":        "hours",
						"price":       100.00,
						"subtotal":    1000.00,
					},
					{
						"id":          4,
						"name":        "Hosting & Deployment Setup",
						"description": "Server setup and deployment configuration",
						"quantity":    1,
						"unit":        "flat fee",
						"price":       500.00,
						"subtotal":    500.00,
					},
				},
				"totals": map[string]any{
					"subtotal":  9200.00,
					"tax_rate":  8.0,
					"tax_total": 736.00,
					"total":     9936.00,
					"paid":      0.00,
					"balance":   9936.00,
				},
				"notes": "This estimate is valid for 30 days from the date above. A 50% deposit is required to begin work. Remaining balance due upon project completion.",
				"terms": []string{
					"This estimate is valid for 30 days from the date above.",
					"A 50% deposit is required to begin work.",
					"Remaining balance due upon project completion.",
					"Scope changes may affect timeline and pricing.",
					"Support includes 30 days of post-launch modifications.",
				},
			}
		}

		if len(form.LayoutJSON) > 0 {
			service := NewTemplateService(NewTemplateRepository(s), nil)
			if err := service.ValidateTemplate(form.LayoutJSON); err != nil {
				if ctx.WantsJson() {
					ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": err.Error()})
					return
				}
				ctx.BackWithError(err)
				return
			}

			layout, err := UnmarshalLayoutJSON(form.LayoutJSON)
			if err != nil {
				if ctx.WantsJson() {
					ctx.JSON(http.StatusUnprocessableEntity, map[string]any{"error": err.Error()})
					return
				}
				ctx.BackWithError(err)
				return
			}

			renderer := pdf.NewRenderer()
			pdfBytes, err := renderer.Render(layout, previewData)
			if err != nil {
				if ctx.WantsJson() {
					ctx.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
					return
				}
				ctx.Error(err)
				return
			}

			ctx.Response.Header().Set("Content-Type", "application/pdf")
			ctx.Response.Header().Set("Content-Disposition", `attachment; filename="preview.pdf"`)
			ctx.Response.Write(pdfBytes)
			return
		}

		service := NewTemplateService(NewTemplateRepository(s), nil)

		pdfBytes, err := service.RenderTemplate(ctx.Request.Context(), form.TemplateID, previewData)
		if err != nil {
			if ctx.WantsJson() {
				ctx.JSON(http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			ctx.Error(err)
			return
		}

		ctx.Response.Header().Set("Content-Type", "application/pdf")
		ctx.Response.Header().Set("Content-Disposition", `attachment; filename="preview.pdf"`)
		ctx.Response.Write(pdfBytes)
	})
}

func (s *Server) deleteTemplateHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		templateID := ctx.Int("id")

		err := s.deleteTemplate(ctx.Request.Context(), templateID)
		if err != nil {
			ctx.BackWith("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.template"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.template"}))

		ctx.Redirect("/admin/templates")
	})
}

func (s *Server) duplicateTemplateHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreTemplateForm) {

		templateID := ctx.Int("id")

		template, err := s.duplicateTemplate(ctx.Request.Context(), templateID, form.Name)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasDuplicated", i18n.Replacements{"subject": "@global.template"}))

		ctx.Redirect(fmt.Sprintf("/admin/templates/%d", template.ID))
	})
}

// testEstimatePDFHandler renders a professional estimate PDF preview
func (s *Server) testEstimatePDFHandler(ctx *routing.Context) {
	// Create the estimate template structure
	templateJSON := map[string]any{
		"page_size":   "A4",
		"page_format": "P",
		"margins": map[string]any{
			"top":    20,
			"right":  15,
			"bottom": 20,
			"left":   15,
		},
		"elements": []any{
			map[string]any{
				"type": "grid",
				"columns": []any{
					map[string]any{
						"span": 6,
						"elements": []any{
							map[string]any{
								"type":    "text",
								"content": "{{company.name}}",
								"size":    16,
								"bold":    true,
							},
							map[string]any{
								"type":    "text",
								"content": "{{company.address}}",
								"size":    9,
							},
							map[string]any{
								"type":    "text",
								"content": "{{company.city}}, {{company.state}} {{company.zip}}",
								"size":    9,
							},
							map[string]any{
								"type":    "text",
								"content": "Phone: {{company.phone}}",
								"size":    9,
							},
							map[string]any{
								"type":    "text",
								"content": "Email: {{company.email}}",
								"size":    9,
							},
						},
					},
					map[string]any{
						"span": 6,
						"elements": []any{
							map[string]any{
								"type":    "text",
								"content": "ESTIMATE",
								"size":    20,
								"bold":    true,
								"align":   "R",
								"color":   "0,0,100",
							},
							map[string]any{
								"type":    "text",
								"content": "",
								"size":    8,
							},
							map[string]any{
								"type":    "text",
								"content": "Est. #{{estimate.number}}",
								"size":    11,
								"bold":    true,
								"align":   "R",
							},
							map[string]any{
								"type":    "text",
								"content": "Date: {{estimate.date}}",
								"size":    10,
								"align":   "R",
							},
							map[string]any{
								"type":    "text",
								"content": "Valid Until: {{estimate.valid_until}}",
								"size":    10,
								"align":   "R",
							},
						},
					},
				},
			},
			map[string]any{
				"type":    "divider",
				"spacing": 10,
				"color":   "200,200,200",
			},
			map[string]any{
				"type": "grid",
				"columns": []any{
					map[string]any{
						"span": 6,
						"elements": []any{
							map[string]any{
								"type":    "text",
								"content": "BILL TO:",
								"size":    10,
								"bold":    true,
							},
							map[string]any{
								"type":    "text",
								"content": "{{customer.name}}",
								"size":    10,
							},
							map[string]any{
								"type":    "text",
								"content": "{{customer.email}}",
								"size":    9,
							},
							map[string]any{
								"type":    "text",
								"content": "{{customer.phone}}",
								"size":    9,
							},
						},
					},
					map[string]any{
						"span": 6,
						"elements": []any{
							map[string]any{
								"type":    "text",
								"content": "PROJECT:",
								"size":    10,
								"bold":    true,
								"align":   "R",
							},
							map[string]any{
								"type":    "text",
								"content": "{{estimate.project_name}}",
								"size":    10,
								"align":   "R",
							},
							map[string]any{
								"type":    "text",
								"content": "{{estimate.project_type}}",
								"size":    9,
								"align":   "R",
							},
						},
					},
				},
			},
			map[string]any{
				"type":    "divider",
				"spacing": 5,
			},
			map[string]any{
				"type":    "text",
				"content": "{{estimate.description}}",
				"size":    10,
			},
			map[string]any{
				"type":    "divider",
				"spacing": 8,
			},
			map[string]any{
				"type": "table",
				"headers": []any{
					map[string]any{"label": "DESCRIPTION", "width": 90},
					map[string]any{"label": "QTY", "width": 20},
					map[string]any{"label": "UNIT PRICE", "width": 35},
					map[string]any{"label": "AMOUNT", "width": 35},
				},
				"rows": []any{
					map[string]any{
						"cells": []any{
							map[string]any{"text": "Professional Web Development Services", "width": 90},
							map[string]any{"text": "80", "width": 20, "align": "C"},
							map[string]any{"text": "$75.00", "width": 35, "align": "R"},
							map[string]any{"text": "$6,000.00", "width": 35, "align": "R"},
						},
					},
					map[string]any{
						"cells": []any{
							map[string]any{"text": "UI/UX Design & Mockups", "width": 90},
							map[string]any{"text": "20", "width": 20, "align": "C"},
							map[string]any{"text": "$85.00", "width": 35, "align": "R"},
							map[string]any{"text": "$1,700.00", "width": 35, "align": "R"},
						},
					},
					map[string]any{
						"cells": []any{
							map[string]any{"text": "Project Management & Coordination", "width": 90},
							map[string]any{"text": "10", "width": 20, "align": "C"},
							map[string]any{"text": "$100.00", "width": 35, "align": "R"},
							map[string]any{"text": "$1,000.00", "width": 35, "align": "R"},
						},
					},
					map[string]any{
						"cells": []any{
							map[string]any{"text": "Hosting & Deployment Setup", "width": 90},
							map[string]any{"text": "1", "width": 20, "align": "C"},
							map[string]any{"text": "$500.00", "width": 35, "align": "R"},
							map[string]any{"text": "$500.00", "width": 35, "align": "R"},
						},
					},
				},
				"row_height":      10,
				"header_bg_color": "240,240,240",
				"border_color":    "180,180,180",
			},
			map[string]any{
				"type":    "divider",
				"spacing": 5,
			},
			map[string]any{
				"type": "grid",
				"columns": []any{
					map[string]any{
						"span": 8,
						"elements": []any{
							map[string]any{
								"type":    "text",
								"content": "NOTES:",
								"size":    10,
								"bold":    true,
							},
							map[string]any{
								"type":    "text",
								"content": "{{estimate.notes}}",
								"size":    9,
							},
						},
					},
					map[string]any{
						"span": 4,
						"elements": []any{
							map[string]any{
								"type": "grid",
								"columns": []any{
									map[string]any{
										"span": 6,
										"elements": []any{
											map[string]any{
												"type":    "text",
												"content": "Subtotal:",
												"size":    10,
												"align":   "R",
											},
											map[string]any{
												"type":    "text",
												"content": "Tax (8%):",
												"size":    10,
												"align":   "R",
											},
											map[string]any{
												"type":    "text",
												"content": "TOTAL:",
												"size":    11,
												"bold":    true,
												"align":   "R",
											},
										},
									},
									map[string]any{
										"span": 6,
										"elements": []any{
											map[string]any{
												"type":    "text",
												"content": "$9,200.00",
												"size":    10,
												"align":   "R",
											},
											map[string]any{
												"type":    "text",
												"content": "$736.00",
												"size":    10,
												"align":   "R",
											},
											map[string]any{
												"type":    "text",
												"content": "$9,936.00",
												"size":    11,
												"bold":    true,
												"align":   "R",
												"color":   "0,0,100",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			map[string]any{
				"type":    "divider",
				"spacing": 10,
			},
			map[string]any{
				"type":    "text",
				"content": "TERMS & CONDITIONS:",
				"size":    10,
				"bold":    true,
			},
			map[string]any{
				"type":    "text",
				"content": "• This estimate is valid for 30 days from the date above\n• A 50% deposit is required to begin work\n• Remaining balance due upon project completion\n• Scope changes may affect timeline and pricing\n• Support includes 30 days of post-launch modifications",
				"size":    9,
			},
		},
	}

	// Create the estimate data
	estimateData := map[string]any{
		"company": map[string]any{
			"name":    "ACME Solutions Inc.",
			"address": "456 Innovation Drive",
			"city":    "Santo Domingo",
			"state":   "DN",
			"zip":     "10101",
			"phone":   "+1-555-123-4567",
			"email":   "hello@acmesolutions.com",
		},
		"customer": map[string]any{
			"name":  "Tech Startup Co.",
			"email": "contact@techstartup.com",
			"phone": "+1-555-987-6543",
		},
		"estimate": map[string]any{
			"number":       "EST-000002",
			"date":         "2026-03-05",
			"valid_until":  "2026-04-04",
			"project_name": "E-Commerce Platform Development",
			"project_type": "Full Stack Web Development",
			"description":  "Complete design and development of a modern e-commerce platform with payment processing, inventory management, and customer analytics dashboard.",
			"notes":        "Includes 3 rounds of revisions. Additional features or significant changes may affect timeline and pricing. Timeline assumes 2 weeks for approval/feedback cycles.",
		},
	}

	// Marshal template to JSON bytes
	templateBytes, _ := json.Marshal(templateJSON)

	// Unmarshal into TemplateLayout struct
	var layout pdf.TemplateLayout
	if err := json.Unmarshal(templateBytes, &layout); err != nil {
		ctx.Error(fmt.Errorf("parse template layout: %w", err))
		return
	}

	// Create a renderer and render the PDF
	renderer := pdf.NewRenderer()
	pdfBytes, err := renderer.Render(&layout, estimateData)
	if err != nil {
		ctx.Error(fmt.Errorf("PDF rendering failed: %w", err))
		return
	}

	// Return the PDF
	ctx.Response.Header().Set("Content-Type", "application/pdf")
	ctx.Response.Header().Set("Content-Disposition", `inline; filename="estimate-EST-000002.pdf"`)
	ctx.Response.Write(pdfBytes)
}
