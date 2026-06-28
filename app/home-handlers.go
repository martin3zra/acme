package app

import (
	"strconv"
	"strings"
	"time"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/routing"
)

func (s *Server) homeHandler(ctx *routing.Context) {

	stats, err := s.findStats(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	dueInvoices, err := s.findLatestDueInvoices(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	estimates, err := s.findLatestEstimates(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	period := ctx.Query("period")
	// Default to "last12" if not provided
	if period == "" {
		period = "last12"
	}
	var chartData []*ChartData
	var totals *Totals
	if period == "last12" {
		chartData, err = s.findLastProfitOfLast12Months(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}

		totals, err = s.findTotalsProfitOfLast12Months(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}
	} else if after, ok := strings.CutPrefix(period, "year-"); ok {
		year, _ := strconv.Atoi(after)
		chartData, err = s.findLastProfitOfYear(ctx.Request.Context(), year)
		if err != nil {
			ctx.Error(err)
			return
		}
		totals, err = s.findTotalsProfitOfYear(ctx.Request.Context(), year)
		if err != nil {
			ctx.Error(err)
			return
		}
	}

	ctx.Render("Home/Index", map[string]any{
		"translations":   trans("dashboard"),
		"hasMissingData": stats.TotalCustomers+stats.TotalEstimates+stats.TotalInvoices == 0,
		"progress": map[string]any{
			"customer_created":  stats.TotalCustomers > 0,
			"products_created":  stats.TotalCustomers > 0, // grab the the total products
			"invoices_created":  stats.TotalInvoices > 0,
			"estimates_created": stats.TotalEstimates > 0,
		},
		"stats": []any{
			map[string]any{"label": "dashboard.stats.total_amount_due", "value": foundation.FormatAmount(stats.TotalDueAmount), "icon": "dollar", "bg": "bg-pink-100"},
			map[string]any{"label": "dashboard.stats.customers", "value": stats.TotalCustomers, "icon": "users", "bg": "bg-blue-100"},
			map[string]any{"label": "dashboard.stats.invoices", "value": stats.TotalInvoices, "icon": "invoice", "bg": "bg-blue-200"},
			map[string]any{"label": "dashboard.stats.estimates", "value": stats.TotalEstimates, "icon": "estimate", "bg": "bg-blue-200"},
		},
		"due_invoices": dueInvoices,
		"estimates":    estimates,
		"period":       period,
		"chart": map[string]any{
			"data":           chartData,
			"totals":         totals,
			"availableYears": availableYearsForDashboard(ctx),
		},
	})
}

func availableYearsForDashboard(ctx *routing.Context) []int {
	businessYear := CurrentCompany(ctx.Request.Context()).CreatedAt.Year()
	currentYear := time.Now().Year()
	years := make([]int, 0)
	for y := businessYear; y <= currentYear; y++ {
		years = append(years, y)
	}
	return years
}
