package app

import (
	"log"
	"net/http"
	"time"

	"github.com/martin3zra/forge/routing"
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
		report, err := NewSalesReportPDF(s.translator, form)
		if err != nil {
			ctx.Error(err)
			return
		}
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

func (s *Server) generateProfitLostReportHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ReportForm) {
		report, err := NewProfitLostReportPDF(s.translator, form)
		if err != nil {
			ctx.Error(err)
			return
		}
		totalSales, err := s.findTotalSales(ctx.Request.Context(), form.From, form.To)
		if err != nil {
			log.Println("error generating profit and lost reports:", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
			return
		}
		expenses, err := s.findExpensesByCategories(ctx.Request.Context(),
			WithDateRange(form.From, form.To),
		)
		if err != nil {
			log.Println("error generating profit and lost reports:", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
			return
		}

		report.RenderReport(ctx.Request.Context(), totalSales, expenses, form.From, form.To, s.config.appName)
		if err := report.pdf.Output(ctx.Response); err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status":  "error",
				"message": err.Error(),
			})
		}
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

func (s *Server) generateExpensesReportHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ReportForm) {
		report, err := NewExpensesReportPDF(s.translator, form)
		if err != nil {
			ctx.Error(err)
			return
		}
		expenses, err := s.findExpensesByCategories(ctx.Request.Context(),
			WithDateRange(form.From, form.To),
		)
		if err != nil {
			log.Println("error generating expense reports:", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
			return
		}

		report.RenderReport(ctx.Request.Context(), expenses, form.From, form.To, s.config.appName)
		if err := report.pdf.Output(ctx.Response); err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status":  "error",
				"message": err.Error(),
			})
		}
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
