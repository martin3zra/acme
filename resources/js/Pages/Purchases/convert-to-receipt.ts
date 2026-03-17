import { addDays, getDaysFromTerm } from '@/lib/utils';
import type { LineForm, PurchaseForm, PurchaseSource, PurchaseWithLines } from '@/types';

export function convertToReceipt(source: PurchaseWithLines): PurchaseForm {
  const date = new Date();
  const terms = source.header.terms;

  const purchaseSource: PurchaseSource = {
    type: 'purchase_order',
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
    },
    lines: source.lines.map((line) => ({
      ...line,
      action: 'added',
    })) as unknown as LineForm[],
    kind: 'purchase_receipt',
    code: '',
    source: purchaseSource,
  };
}
