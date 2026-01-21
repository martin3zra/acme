import { FlagSet, Receivable, ReceivableInvoiceForm } from '@/types';

export function buildReceivableState(
  receivables: Receivable[],
  invoice_uuid: string,
): {
  lines: ReceivableInvoiceForm[];
  rowSelection: FlagSet;
} {
  const lines: ReceivableInvoiceForm[] = receivables.map((receivable) => ({
    ...receivable.invoice,
    payment: invoice_uuid === receivable.invoice.uuid ? receivable.invoice.amount_due : 0,
    discount: 0,
    balance: 0,
    remaining: 0,
    original_payment: 0,
    action: 'unchanged',
  }));

  let selectedRowId = receivables.find((r) => r.invoice.uuid === invoice_uuid)?.invoice.id ?? -1;

  // When we want to record a payment from the customer list
  // and that customer only has one invoice pending and
  // not record has been selected by default we set
  // as select the existing invoice and the total
  // received would be equal to the amount due
  if (selectedRowId === -1 && receivables.length === 1) {
    selectedRowId = receivables[0].invoice.id;
    lines[0].payment = receivables[0].invoice.amount_due;
  }

  const rowSelection: FlagSet = {};
  if (selectedRowId > 0) {
    rowSelection[selectedRowId.toString()] = true;
  }

  return { lines, rowSelection };
}

export function buildRowSelection(lines: ReceivableInvoiceForm[]): FlagSet {
  const selection: FlagSet = {};
  lines.forEach((line) => {
    if (line.payment > 0) {
      selection[line.id.toString()] = true;
    }
  });
  return selection;
}

export function calculateTotals(lines: ReceivableInvoiceForm[]) {
  let totalPayment = 0;
  let totalDiscount = 0;
  let totalRemaining = 0;

  lines.forEach((line) => {
    const payment = line.payment || 0;
    const discountAmount = line.discount || 0;

    totalPayment += payment;
    totalDiscount += discountAmount;
    totalRemaining += line.remaining || 0;
  });

  return {
    totalPayment,
    totalDiscount,
    totalRemaining,
  };
}
