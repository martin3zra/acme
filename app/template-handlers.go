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
