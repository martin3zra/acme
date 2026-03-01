import { BreadcrumbItem, defaultBreadcrumbs } from '@/types';

export const breadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  {
    title: 'expenses.title',
    href: '/expenses',
  },
];

export const createExpenseBreadcrumbs: BreadcrumbItem[] = [
  ...breadcrumbs,
  {
    title: 'expenses.newExpense.title',
    href: '/expenses/create',
  },
];

export const editExpenseBreadcrumbs: BreadcrumbItem[] = [
  ...breadcrumbs,
  {
    title: 'expenses.editExpense.title',
    href: '',
  },
];
