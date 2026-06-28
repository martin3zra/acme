import { BreadcrumbItem, defaultBreadcrumbs } from '@/types';

export const breadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  {
    title: 'warehouses.title',
    href: '/inventories/warehouses',
  },
];
