import { BreadcrumbItem, defaultBreadcrumbs, PaymentTerm } from '@/types';

export const paymentTerms: PaymentTerm[] = [
  { value: 'pia', label: 'Payment In Advance' },
  { value: 'net0', label: 'Due on Receipt' },
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
    title: 'vendors.title',
    href: '/vendors',
  },
];
