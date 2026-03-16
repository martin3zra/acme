import { BreadcrumbItem, defaultBreadcrumbs, PurchaseTransactionKind } from '@/types';

export const makeBreadcrumbs = (kind: PurchaseTransactionKind): BreadcrumbItem[] => [
  ...defaultBreadcrumbs,
  {
    title: `${kind}.title`, // e.g. "purchases.orders.title", "purchases.receipts.title"
    href: `/${kind}s`,
  },
];
