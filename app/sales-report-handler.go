package app

import (
	"net/http"
	"time"

	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) reportSalesHandler(ctx *routing.Context) {
	initialRange := DateRange{
		From: time.Now().AddDate(0, 0, -7).Format("2006-01-02"), // 7 days ago
		To:   time.Now().Format("2006-01-02"),                   // today
	}
	ctx.Render("Reports/Sales/Index", map[string]any{
		"translations":      trans("reports"),
		"initialRange":      initialRange,
		"dateRanges":        DateRangePresets(),
		"initialPreset":     "this_week",
		"initialReportType": "sales_by_customer",
	})
}

func (s *Server) generateSalesReportHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ReportSalesForm) {
		report, _ := NewSalesReportPDF(s.translator, form)
		groups, err := s.fetchSalesGroups(ctx.Request.Context(), form)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
			return
		}

		report.RenderReport(ctx.Request.Context(), groups, s.config.appName)
		if err := report.pdf.Output(ctx.Response); err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status":  "error",
				"message": err.Error(),
			})
		}
	})
}

func (s *Server) reportProfitLostHandler(ctx *routing.Context) {
	initialRange := DateRange{
		From: time.Now().AddDate(0, 0, -7).Format("2006-01-02"), // 7 days ago
		To:   time.Now().Format("2006-01-02"),                   // today
	}
	ctx.Render("Reports/ProfitLost/Index", map[string]any{
		"translations":  trans("reports"),
		"initialRange":  initialRange,
		"dateRanges":    DateRangePresets(),
		"initialPreset": "this_week",
	})
}

func (s *Server) reportExpensesHandler(ctx *routing.Context) {
	initialRange := DateRange{
		From: time.Now().AddDate(0, 0, -7).Format("2006-01-02"), // 7 days ago
		To:   time.Now().Format("2006-01-02"),                   // today
	}
	ctx.Render("Reports/Expenses/Index", map[string]any{
		"translations":  trans("reports"),
		"initialRange":  initialRange,
		"dateRanges":    DateRangePresets(),
		"initialPreset": "this_week",
	})
}

func (s *Server) reportTaxesHandler(ctx *routing.Context) {
	initialRange := DateRange{
		From: time.Now().AddDate(0, 0, -7).Format("2006-01-02"), // 7 days ago
		To:   time.Now().Format("2006-01-02"),                   // today
	}
	ctx.Render("Reports/Taxes/Index", map[string]any{
		"translations":  trans("reports"),
		"initialRange":  initialRange,
		"dateRanges":    DateRangePresets(),
		"initialPreset": "this_week",
	})
}

// Generic helper to convert any slice of a type that implements ReportGroup
// into a slice of ReportGroup interfaces.
func toReportGroups[T ReportGroup](groups []T) []ReportGroup {
	result := make([]ReportGroup, len(groups))
	for i, g := range groups {
		result[i] = g
	}
	return result
}
