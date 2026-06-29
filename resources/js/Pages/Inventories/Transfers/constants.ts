import { BreadcrumbItem, defaultBreadcrumbs } from '@/types';

export const breadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  {
    title: 'transfers.title',
    href: '/inventories/transfers',
  },
];

export const createBreadcrumbs: BreadcrumbItem[] = [
  ...breadcrumbs,
  {
    title: 'transfers.create.title',
    href: '/inventories/transfers/create',
  },
];

export const showBreadcrumbs: BreadcrumbItem[] = [
  ...breadcrumbs,
  {
    title: 'transfers.show.title',
    href: '#',
  },
];
