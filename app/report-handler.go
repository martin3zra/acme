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
		var groups []CustomerGroup
		var err error
		if form.ShowInvoices {
			groups, err = s.findCustomerWiseSalesWithInvoices(ctx.Request.Context(), form.From, form.To)
		} else {
			groups, err = s.findCustomerWiseSales(ctx.Request.Context(), form.From, form.To)
		}

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status":  "error",
				"message": err.Error(),
			})
			return
		}
		report, _ := NewSalesReportPDF(s.translator, form.From, form.To, form.ShowInvoices)
		report.Render(groups)

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
