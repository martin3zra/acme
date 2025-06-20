import { defaultBTForm, defaultCardForm, defaultCashForm, defaultCheckForm } from '@/constants';
import { BreadcrumbItem, defaultBreadcrumbs, DiscountType, HeaderForm, InvoiceForm, PaymentMethodsForm, PaymentTerm } from '@/types';

export const paymentTerms: PaymentTerm[] = [
  // { value: 1, label: ':cash' },
  // { value: 7, label: '7 :days' },
  // { value: 10, label: '10 :days' },
  // { value: 15, label: '15 :days' },
  // { value: 30, label: '30 :days' },
  // { value: 60, label: '60 :days' },
  // { value: 90, label: '90 :days' },
  { value: 'pia', label: 'Payment In Advance' },
  { value: 'net7', label: 'Net 7' },
  { value: 'net10', label: 'Net 10' },
  { value: 'net15', label: 'Net 15' },
  { value: 'net30', label: 'Net 30' },
  { value: 'net60', label: 'Net 60' },
  { value: 'net90', label: 'Net 90' },
  { value: 'net120', label: 'Net 120' },
];

export const breadcrumbs: BreadcrumbItem[] = [
  ...defaultBreadcrumbs,
  {
    title: 'invoices.title',
    href: '/invoices',
  },
];

export const createBreadcrumbs: BreadcrumbItem[] = [
  ...breadcrumbs,
  {
    title: 'invoices.newInvoice.title',
    href: '/invoices/create',
  },
];

export const editBreadcrumbs: BreadcrumbItem[] = [
  ...breadcrumbs,
  {
    title: 'invoices.editInvoice.title',
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
