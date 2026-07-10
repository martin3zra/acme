import { defaultBTForm, defaultCardForm, defaultCashForm, defaultCheckForm } from '@/constants';
import {
  BreadcrumbItem,
  defaultBreadcrumbs,
  DiscountType,
  PaymentForm,
  PaymentHeaderForm,
  PaymentMethodsForm,
  PaymentTotals,
  ReceivableInvoiceForm,
} from '@/types';

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

// Totals are derived from the lines, never stored independently, so Create and
// Edit cannot drift apart on what a payment adds up to.
export const computeTotals = (lines: ReceivableInvoiceForm[]): PaymentTotals =>
  lines.reduce(
    (acc, line) => {
      acc.totalPayment += line.payment || 0;
      acc.totalDiscount += line.discount || 0;
      acc.totalRemaining += line.remaining || 0;
      return acc;
    },
    { totalPayment: 0, totalDiscount: 0, totalRemaining: 0 },
  );

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
