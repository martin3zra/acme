import { addDays, getDaysFromTerm } from '@/lib/utils';
import type { LineForm, PurchaseForm, PurchaseSource, PurchaseWithLines } from '@/types';

export function convertToVendorBill(source: PurchaseWithLines): PurchaseForm {
  const date = new Date();
  const terms = source.header.terms;

  const purchaseSource: PurchaseSource = {
    type: 'purchase_receipt',
    id: source.header.uuid,
  };

  return {
    header: {
      vendor: source.header.vendor,
      date,
      due: terms === 'pia' ? undefined : addDays(date, getDaysFromTerm(terms)),
      terms,
      notes: source.header.notes,
      discount: source.header.discount,
      invoice_number: source.header.invoice_number ?? '',
    },
    lines: source.lines.map((line) => ({
      ...line,
      action: 'added',
    })) as unknown as LineForm[],
    kind: 'vendor_bill',
    code: '',
    source: purchaseSource,
  };
}
