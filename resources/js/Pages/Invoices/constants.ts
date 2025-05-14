import { defaultBTForm, defaultCardForm, defaultCashForm, defaultCheckForm } from '@/constants';
import { BreadcrumbItem, DiscountType, HeaderForm, InvoiceForm, PaymentMethodsForm, PaymentTerm } from '@/types';

export const paymentTerms: PaymentTerm[] = [
  { value: 1, label: 'Cash' },
  { value: 7, label: '7 Days' },
  { value: 10, label: '10 Days' },
  { value: 15, label: '15 Days' },
  { value: 30, label: '30 Days' },
  { value: 60, label: '60 Days' },
  { value: 90, label: '90 Days' },
];

export const createBreadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Home',
    href: '/home',
  },
  {
    title: 'Invoices',
    href: '/invoices',
  },
  {
    title: 'New Invoice',
    href: '/invoices/create',
  },
];

export const editBreadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Home',
    href: '/home',
  },
  {
    title: 'Invoices',
    href: '/invoices',
  },
  {
    title: 'Edit Invoice',
    href: '',
  },
];

export const defaultPaymentMethodsForm: PaymentMethodsForm = {
  cash: defaultCashForm,
  ck: defaultCheckForm,
  card: defaultCardForm,
  bt: defaultBTForm,
};
export const defaultDiscount: DiscountType = { value: 0, type: 'fixed' };
export const defaultHeaderForm: HeaderForm = {
  customer: undefined,
  date: undefined,
  due: undefined,
  terms: 0,
  taxReceipt: 0,
  notes: undefined,
  discount: defaultDiscount,
};

export const defaultInvoiceForm: InvoiceForm = { header: defaultHeaderForm, lines: [], payment: defaultPaymentMethodsForm };
