export interface User {
  id: number;
  first_name: string;
  last_name: string;
  email: string;
  avatar?: string;
  email_verified_at: string | null;
  created_at: string;
  updated_at: string;
  [key: string]: unknown; // This allows for additional properties
}

export interface Auth {
  user: User;
  company: Company;
}

export interface Company {
  id: number;
  name: string;
  address: string;
  identifier: string;
  city: string;
  created_at: string;
  updated_at: string;
}

export interface Flash {
  [key: string]: unknown; // This allows for additional properties
}

export type PageProps<T extends Record<string, unknown> = Record<string, unknown>> = T & {
  auth: Auth;
  flash: Flash;
  csrf_token: string;
};

export interface SharedData {
  auth: Auth;
  flash: Flash;
  csrf_token: string;
}

export interface Customer {
  id: number;
  uuid: string;
  name: string;
  contact_name: string;
  phone: string;
  email: string;
  address: string;
  status: string;
  amount_due: number;
  created_at: string;
  updated_at: string;
}

export interface Item {
  id: number;
  uuid: string;
  name: string;
  price: number;
  tax: Tax;
  unit: Unit;
  description: string;
  status: string;
}

export type LineAction = 'added' | 'updated' | 'deleted' | 'unchanged';

export interface InvoiceLine extends Item {
  qty: number;
  amount: number;
  total: number;
  tax: TaxWithAmount;
  action: LineAction;
}

export interface Tax {
  id: number;
  name: string;
  rate: number;
}

export interface TaxWithAmount extends Tax {
  amount: number;
}

export interface Unit {
  id: number;
  name: string;
}

export type DiscountType = {
  type: 'fixed' | 'percentage';
  value: number;
};

export const InvoiceStatuses = ['draft', 'sent', 'viewed', 'overdue', 'completed', 'void'] as const;

export type InvoiceStatus = (typeof InvoiceStatuses)[number];

export const PaidStatuses = ['paid', 'unpaid', 'partial', 'removed', 'overpaid', 'pending'] as const;

export type PaidStatus = (typeof PaidStatuses)[number];

export const Statuses = ['enabled', 'disabled'] as const;
export type Status = (typeof Statuses)[number];

export type StatusType = 'paid' | 'invoice' | 'status';

export interface Invoice {
  id: number;
  uuid: string;
  number: string;
  ncf: string;
  customer: Customer;
  date: string;
  due_on?: string;
  terms: number;
  tax_receipt_id: number;
  amount: number;
  discount: DiscountType;
  tax: number;
  total: number;
  amount_due: number;
  payment: PaymentMethodsForm;
  status: string;
  paid_status: PaidStatus;
  notes: string;
}

export interface InvoiceWithLines {
  header: Invoice;
  lines: InvoiceLine[];
}

export interface BreadcrumbItem {
  title: string;
  href: string;
}

export type Verb = 'create' | 'view' | 'edit' | 'trash';

export type InvoiceVerb = Exclude<Verb, 'trash'> | 'void' | 'record-payment';

export type PaymentVerb = Exclude<Verb, 'trash'> | 'void';

export interface PaymentFormType {
  amount: number;
  reference: string;
}

export type BankOperationFormProps = Partial<PaymentFormType> & {
  onChange: (value: number | string) => void;
};

export type CashForm = {
  amount: number;
};

export type CheckForm = PaymentFormType & {};

export type PaymentMethod = 'cash' | 'ck' | 'card' | 'bt';

export type CardBrand = {
  value: string;
  name: string;
};

export type CardFormInput = 'last4' | 'brand' | 'reference' | 'amount';

export type CardForm = PaymentFormType & {
  last4: number;
  brand: string;
};

export type BTForm = PaymentFormType & {};

export type PaymentMethodType = {
  value: PaymentMethod;
  name: string;
  amount: number;
  autoFocus?: boolean;
};

export type PaymentTerm = {
  value: number;
  label: string;
};

export interface LineForm extends Item {
  qty: number;
  amount: number;
  action: LineAction;
}

export type PaymentMethodsForm = {
  cash: CashForm;
  ck: CheckForm;
  card: CardForm;
  bt: BTForm;
};

export interface Nameable {
  id: string | number;
  name: string;
}

export interface TaxReceipt extends Nameable {
  available: boolean;
}

export type currencySignature = (value: number | string, precision?: number, inCent?: boolean) => string;

export type HeaderForm = {
  customer: Customer | undefined;
  date: Date | undefined;
  due: Date | undefined;
  terms: number;
  taxReceipt: number;
  notes: string | undefined;
  discount: DiscountType;
};

export type InvoiceForm = {
  header: HeaderForm;
  lines: LineForm[];
  payment: PaymentMethodsForm;
};

export type Payment = {
  id: number;
  uuid: string;
  number: string;
  date: Date | undefined;
  amount: number;
  created_at: string;
  updated_at: string;
  customer: {
    uuid: string;
    name: string;
    amount_due: string;
  };
};

export type PaymentHeaderForm = {
  customer: Customer | undefined;
  date: Date | undefined;
};

export type ReceivableInvoiceForm = ReceivableInvoice & {
  payment: number;
  discount: number;
  balance: number;
};

export type PaymentForm = {
  header: PaymentHeaderForm;
  lines: ReceivableInvoiceForm[];
};

export interface Receivable {
  id: number;
  uuid: string;
  invoice: ReceivableInvoice;
}

export interface ReceivableInvoice {
  id: number;
  uuid: string;
  number: string;
  ncf: string;
  date: string;
  due_on: string;
  total: number;
  amount_due: number;
  paid_status: string;
  notes: string;
}

export interface ReceivableTransfomer extends Receivable {
  transform(): ReceivableInvoiceForm;
}
