import { defaultBTForm, defaultCardForm, defaultCashForm, defaultCheckForm } from '@/constants';
import type { BreadcrumbItem, VendorPaymentForm } from '@/types';
import { defaultBreadcrumbs } from '@/types';

export const defaultVendorPaymentForm = (): VendorPaymentForm => ({
  header: {
    vendor: undefined,
    date: new Date(),
    notes: '',
  },
  lines: [],
  payment: {
    cash: defaultCashForm,
    ck: defaultCheckForm,
    card: defaultCardForm,
    bt: defaultBTForm,
  },
  totals: {
    totalPayment: 0,
    totalRemaining: 0,
  },
});

export const createPayableBreadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  { title: 'payables.title', href: '/payables' },
  { title: 'payables.create.title', href: '' },
];

export const payablesBreadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  { title: 'payables.title', href: '/payables' },
];
