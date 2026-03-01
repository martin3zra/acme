import { defaultBTForm, defaultCardForm, defaultCashForm, defaultCheckForm } from '@/constants';
import { BreadcrumbItem, defaultBreadcrumbs, DiscountType, PaymentForm, PaymentHeaderForm, PaymentMethodsForm, PaymentTotals } from '@/types';

export const defaultPaymentMethodsForm: PaymentMethodsForm = {
  cash: defaultCashForm,
  ck: defaultCheckForm,
  card: defaultCardForm,
  bt: defaultBTForm,
};
export const defaultDiscount: DiscountType = { value: 0, type: 'fixed' };
export const defaultHeaderForm: PaymentHeaderForm = {
  customer: undefined,
  date: undefined,
  notes: '',
  discount: 0,
};
export const defaultPaymentTotals: PaymentTotals = { totalPayment: 0, totalDiscount: 0, totalRemaining: 0 };

export const defaultPaymentForm: PaymentForm = {
  header: defaultHeaderForm,
  lines: [],
  payment: defaultPaymentMethodsForm,
  totals: defaultPaymentTotals,
};

export const breadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  {
    title: 'payments.title',
    href: '/payments',
  },
];

export const createPaymentBreadcrumbs: BreadcrumbItem[] = [
  ...breadcrumbs,
  {
    title: 'payments.newPayment.title',
    href: '/payments/create',
  },
];

export const editPaymentBreadcrumbs: BreadcrumbItem[] = [
  ...breadcrumbs,
  {
    title: 'payments.editPayment.title',
    href: '',
  },
];
