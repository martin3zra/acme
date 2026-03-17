import { capitalize } from '@/lib/utils';
import type { BreadcrumbItem, DiscountType, PaymentTerm, PurchaseForm, PurchaseTransactionKind } from '@/types';
import { defaultBreadcrumbs } from '@/types';

export const paymentTerms: PaymentTerm[] = [
  { value: 'pia', label: 'Payment In Advance' },
  { value: 'net0', label: 'Due on Receipt' },
  { value: 'net7', label: 'Net 7' },
  { value: 'net10', label: 'Net 10' },
  { value: 'net15', label: 'Net 15' },
  { value: 'net30', label: 'Net 30' },
  { value: 'net60', label: 'Net 60' },
  { value: 'net90', label: 'Net 90' },
  { value: 'net120', label: 'Net 120' },
];

export const defaultDiscount: DiscountType = { value: 0, type: 'fixed' };

export const purchaseKindMeta = (kind: PurchaseTransactionKind) => {
  switch (kind) {
    case 'purchase_order':
      return {
        key: 'purchases.orders',
        listUrl: '/purchases/orders',
        createUrl: '/purchases/orders/create',
      };
    case 'purchase_receipt':
      return {
        key: 'purchases.receipts',
        listUrl: '/purchases/receipts',
        createUrl: '/purchases/receipts/create',
      };
    case 'vendor_bill':
    default:
      return {
        key: 'purchases.vendor-bills',
        listUrl: '/purchases/vendor-bills',
        createUrl: '/purchases/vendor-bills/create',
      };
  }
};

export const makeBreadcrumbs = (kind: PurchaseTransactionKind): BreadcrumbItem[] => {
  const meta = purchaseKindMeta(kind);
  return [
    ...defaultBreadcrumbs,
    {
      title: `${meta.key}.title`,
      href: meta.listUrl,
    },
  ];
};

export const makeCreateBreadcrumbs = (kind: PurchaseTransactionKind): BreadcrumbItem[] => {
  const meta = purchaseKindMeta(kind);
  return [
    ...makeBreadcrumbs(kind),
    {
      title: `${meta.key}.new${capitalize(meta.key.split('.').slice(-1)[0] || 'Purchase')}.title`,
      href: meta.createUrl,
    },
  ];
};

export const makeEditBreadcrumbs = (kind: PurchaseTransactionKind): BreadcrumbItem[] => {
  const meta = purchaseKindMeta(kind);
  return [
    ...makeBreadcrumbs(kind),
    {
      title: `${meta.key}.edit${capitalize(meta.key.split('.').slice(-1)[0] || 'Purchase')}.title`,
      href: '',
    },
  ];
};

export const defaultPurchaseForm = (kind: PurchaseTransactionKind): PurchaseForm => ({
  header: {
    vendor: undefined,
    date: new Date(),
    due: undefined,
    terms: 'pia',
    notes: undefined,
    discount: defaultDiscount,
  },
  lines: [],
  kind,
  source: { type: kind, id: '', code: '' },
});
