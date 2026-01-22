import { LucideIcon } from 'lucide-react';
import { IconName } from './icons';

export interface LinkedCompany {
  uuid: string;
  role: string;
}
export interface User {
  id: number;
  uuid: string;
  name: string;
  email: string;
  pending_email: string;
  avatar?: string;
  email_verified_at: string | null;
  status: string;
  linked: number;
  linkedCompanies: LinkedCompany[];
  created_at: string;
  updated_at: string;
  [key: string]: unknown; // This allows for additional properties
}

export interface Auth {
  user: User;
  company: Company;
  account: AuthAccount;
}

export interface AuthAccount {
  uuid: string;
  owner: boolean;
}

export type SequenceField = string | number;

export type ModuleType = 'invoices' | 'customers' | 'payments';

export type SequenceTypeKey = 'prefix' | 'suffix' | 'next';

export type SequenceConfig = {
  prefix: string;
  next: number;
  padding: number;
  [key: string]: SequenceField;
};

export type Sequences = {
  [module: string]:
    | {
        [type: string]: SequenceConfig;
      }
    | SequenceConfig;
};

export interface Company {
  id: number;
  uuid: string;
  name: string;
  address: string;
  identifier: string;
  city: string;
  sequences: Sequences;
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

export const CustomerTypes = ['individual', 'business'] as const;

export type CustomerType = (typeof CustomerTypes)[number];

export type CustomerTypeFilter = CustomerType | 'all';

export type InvoiceTypeFilter = 'all' | 'cash' | 'credit';

export type TransactionKind = 'invoice' | 'estimate' | 'order';

export interface OpenBalance {
  invoice_id: number;
  date: Date;
  amount: number;
}
export interface Customer {
  id: number;
  uuid: string;
  code: string;
  name: string;
  contact_name: string;
  phone: string;
  email: string;
  address: string;
  status: string;
  payment_method: string;
  payment_terms: string;
  amount_due: number;
  credit_limited: boolean;
  credit_limit: number;
  customer_type: 'individual' | 'business';
  tax_receipt: number;
  open_balance: OpenBalance;
  open_balance_as_of: Date;
  created_at: string;
  updated_at: string;
}

export const ItemTypes = ['product', 'service'] as const;

export type ItemType = (typeof ItemTypes)[number];

export type ItemTypeFilter = ItemType | 'all';

export interface ItemIdentifiers {
  reference: string;
  code: string;
  sku: string;
  barcode: string;
  vendor_reference: string;
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
  item_type?: ItemType; // This can be 'product' or 'service'
  identifiers?: ItemIdentifiers; // Optional identifiers for the item
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

export const InvoiceStatuses = ['open', 'draft', 'sent', 'viewed', 'overdue', 'completed', 'void', 'partial'] as const;

export type InvoiceStatus = (typeof InvoiceStatuses)[number];

export const PaidStatuses = ['paid', 'unpaid', 'partial', 'removed', 'overpaid', 'pending'] as const;

export type PaidStatus = (typeof PaidStatuses)[number];

export const Statuses = ['enabled', 'disabled'] as const;
export type Status = (typeof Statuses)[number];

export type StatusType = 'paid' | 'invoice' | 'status' | 'payment';

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
  pdfURL: string;
}

export interface BreadcrumbItem {
  title: string;
  href: string;
}

export type Verb = 'create' | 'view' | 'edit' | 'trash';

export type InvoiceVerb = Exclude<Verb, 'trash'> | 'void' | 'record-payment';

export type PaymentVerb = Verb | 'void';

export type CustomerVerb = Verb | 'record-payment' | 'issue-invoice';

export type UserVerb = Verb | 'permission';

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

export const PaymentMethods = ['cash', 'ck', 'card', 'bt'] as const;
export type PaymentMethod = (typeof PaymentMethods)[number];

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
  value: string;
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
  terms: string;
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
  code: string;
  date: string;
  amount: number;
  invoices: number;
  status: string;
  created_at: string;
  updated_at: string;
  notes: string;
  customer: Customer;
  payment: PaymentMethodsForm;
};

export type PaymentHeaderForm = {
  customer: Customer | undefined;
  date: Date | undefined;
  notes: string;
  discount: number;
};

export type FlagSet = Record<string, boolean>;

export type ReceivableInvoiceForm = ReceivableInvoice & {
  original_payment: number;
  payment: number;
  discount: number;
  balance: number;
  remaining: number;
  action: LineAction;
};

export type PaymentForm = {
  header: PaymentHeaderForm;
  lines: ReceivableInvoiceForm[];
  payment: PaymentMethodsForm;
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
  date: Date;
  due_on: Date;
  total: number;
  amount_due: number;
  paid_status: string;
  notes: string;
}

export interface PaymentLine {
  id: number;
  created_at: string;
  payment: number;
  invoice: {
    uuid: string;
    code: string;
    amount: number;
    amount_due: number;
    date: string;
    due_on: string;
    paid_status: PaidStatus;
    ncf: string;
    notes: string;
  };
}

export interface PaymentWithLines {
  header: Payment;
  lines: PaymentLine[];
  pdfURL: string;
}

export type onValueChangeType = (inputId: string, newValue: string | number) => void;

export function mapPaymentLineToReceivableInvoice(paymentLine: PaymentLine): ReceivableInvoiceForm {
  const { invoice } = paymentLine;

  return {
    id: paymentLine.id,
    uuid: invoice.uuid,
    number: invoice.code,
    ncf: invoice.ncf, // Placeholder since PaymentLine does not have this field
    date: new Date(invoice.date),
    due_on: new Date(invoice.due_on), // Placeholder, not present in PaymentLine
    total: invoice.amount,
    amount_due: invoice.amount_due,
    paid_status: invoice.paid_status,
    notes: invoice.notes, // Placeholder since notes don't exist in PaymentLine
    original_payment: paymentLine.payment,
    payment: paymentLine.payment,
    discount: 0,
    balance: 0,
    action: 'unchanged', // Placeholder, as the action is not defined in PaymentLine
  };
}

export type Replacements = { [key: string]: string | number };

export const defaultBreadcrumbs: BreadcrumbItem[] = [
  {
    title: 'global.navMain.dashboard',
    href: '/home',
  },
];

export interface NavItem {
  title: string;
  url: string;
  icon?: LucideIcon | null;
  isActive?: boolean;
  requiredAbility?: string;
  match?: string[];
  // components: string[];
}

export type Role = {
  id: string;
  label: string;
  description: string;
};

export type RoleType = 'developer' | 'owner' | 'admin' | 'supervisor' | 'standard';

export interface SlotProps {
  children: React.ReactNode;
}

export interface StatItem {
  label: string;
  value: string;
  icon: IconName | string;
  bg: string;
}

export interface DueInvoice {
  uuid: string;
  due_on: string;
  customer: {
    uuid: string;
    name: string;
  };
  amount: number;
}

export interface ChartPoint {
  month: string;
  sales: number;
  expenses: number;
}

export interface Totals {
  totalSales: number;
  totalReceipts: number;
  totalExpenses: number;
  netIncome: number;
}
