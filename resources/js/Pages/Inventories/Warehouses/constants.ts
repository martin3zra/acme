import { BreadcrumbItem, defaultBreadcrumbs, TransactionKind } from '@/types';

export const makeBreadcrumbs = (kind: TransactionKind): BreadcrumbItem[] => [
  ...defaultBreadcrumbs,
  {
    title: `${kind}s.title`, // e.g. "invoices.title", "estimates.title"
    href: `/${kind}s`,
  },
];
