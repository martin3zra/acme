import { InvoiceForm, InvoiceLine, InvoiceWithLines, PaymentTermValue, TransactionKind } from '@/types';

export function convertToInvoice(
  kind: TransactionKind,
  source: InvoiceWithLines,
  extra: { terms: PaymentTermValue; taxReceipt: number; ncf: string },
): InvoiceForm {
  return {
    kind: 'invoice',
    header: {
      customer: source.header.customer,
      date: new Date(source.header.date),
      due: source.header.due_on ? new Date(source.header.due_on) : undefined,
      terms: extra.terms,
      taxReceipt: extra.taxReceipt,
      notes: source.header.notes,
      discount: source.header.discount,
    },
    lines: source.lines.map((line: InvoiceLine) => {
      return { ...line, total: 0, tax: { ...line.tax, amount: 0 } };
    }),
    payment: source.header.payment,
    source: { type: kind, id: source.header.id },
  };
}
