import { calculateDueDate } from '@/lib/utils';
import { InvoiceForm, InvoiceLine, InvoiceWithLines, PaymentTermValue, TransactionKind } from '@/types';

export function convertToInvoice(
  kind: TransactionKind,
  source: InvoiceWithLines,
  extra: { terms: PaymentTermValue; taxReceipt: number; ncf: string },
  clonedFrom?: number,
): InvoiceForm {
  return {
    kind: 'invoice',
    header: {
      customer: source.header.customer,
      date: new Date(),
      // date: new Date(source.header.date),
      due: extra.terms === 'pia' ? undefined : calculateDueDate(extra.terms),
      // due: source.header.due_on ? new Date(source.header.due_on) : undefined,
      terms: extra.terms,
      taxReceipt: extra.taxReceipt,
      notes: source.header.notes,
      discount: source.header.discount,
    },
    lines: source.lines.map((line: InvoiceLine) => {
      return { ...line, total: 0, tax: { ...line.tax, amount: 0 } };
    }),
    payment: source.header.payment,
    source: { type: kind, id: source.header.uuid },
    clonedFrom,
  };
}
