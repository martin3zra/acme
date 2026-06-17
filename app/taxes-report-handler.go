package app

import (
	"net/http"
	"time"

	"github.com/martin3zra/forge/routing"
)

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

func (s *Server) generateTaxesReportHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ReportForm) {
		report, err := NewTaxesReportPDF(s.translator, form)
		if err != nil {
			ctx.Error(err)
			return
		}
		groups, err := s.findTaxesWiseSales(ctx.Request.Context(), form.From, form.To)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{"status": "error", "message": err.Error()})
			return
		}

		report.RenderReport(ctx.Request.Context(), groups, form.From, form.To, s.config.appName)
		if err := report.pdf.Output(ctx.Response); err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status":  "error",
				"message": err.Error(),
			})
		}
	})
}
