package app

func mapInvoiceToStoreForm(invoice *invoice, lines []*line) *StoreInvoiceForm {
	formLines := make([]*Line, 0)
	for _, line := range lines {
		var l Line
		l.ID = int(line.ID)
		l.VariantID = int(line.ID)
		l.Unit = int(line.Unit.ID)
		l.Qty = int(line.Qty)
		l.Price = line.Price
		l.Rate = line.Tax.Rate
		l.Action = "added"
		l.tax = line.Tax.Amount
		l.amount = line.Amount
		l.discount = 0
		l.total = line.Total
		formLines = append(formLines, &l)
	}

	return &StoreInvoiceForm{
		Kind:       TransactionKinds.Template,
		termType:   TermType(invoice.Terms),
		Date:       invoice.Date,
		TaxReceipt: *invoice.TaxReceiptID,
		CustomerID: invoice.Customer.ID,
		amount:     invoice.Amount,
		amountDue:  0,
		Discount:   invoice.Discount,
		tax:        invoice.Tax,
		total:      invoice.Total,
		Notes:      invoice.Notes,
		paidStatus: PaidStatuses.UnPaid,
		Payment:    invoice.Payment,
		Source: &TransactionSource{
			ID:   invoice.UUID,
			Type: TransactionKinds.Invoice,
		},
		Lines: formLines,
	}
}
