import { BreadcrumbItem, defaultBreadcrumbs } from '@/types';

export const breadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  {
    title: '@global.stock',
    href: '/stock-levels',
  },
];
