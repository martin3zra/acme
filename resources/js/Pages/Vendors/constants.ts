import { BreadcrumbItem, defaultBreadcrumbs } from '@/types';

export const breadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  {
    title: 'vendors.title',
    href: '/vendors',
  },
];
