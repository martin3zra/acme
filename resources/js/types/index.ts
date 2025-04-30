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
  name: string;
  contact_name: string;
  phone: string;
  email: string;
  status: string;
  amount_due: number;
  created_at: string;
  updated_at: string;
}

export interface Item {
    id: number;
    name: string;
    price: number;
    tax: Tax;
    unit: Unit;
    description: string;
    status: string;
}

export interface Tax {
    id: number;
    name: string;
    rate: number;
}

export interface Unit {
    id: number;
    name: string;
}

export interface Unit {
    id: number;
    name: string;
}

export type DiscountType = {
    type: "fixed" | "percentage";
    value: number
  }

export interface Invoice {
    id: number;
    number: string;
    customer: Customer
    status: string
    paid_status: string
}

export interface BreadcrumbItem {
    title: string;
    href: string;
}

export type Verb = "create" | "view" | "edit" | "trash"

export interface PaymentFormType {
  amount: number;
  reference: string;
}

export type BankOperationFormProps = Partial<PaymentFormType> & {
  onChange: (value: number|string) => void;
}

export type CashForm = {
  amount: number;
}

export type CheckForm = PaymentFormType & {}

export type PaymentMethod = "cash" | "ck" | "card" | "bt";

export type CardBrand = {
  value: string;
  name: string;
}

export type CardFormInput = "last4" | "brand" | "reference" | "amount"

export type CardForm = PaymentFormType & {
  last4: number;
  brand: string;
}

export type BTForm = PaymentFormType & {}

export type PaymentMethodType = {
  value: PaymentMethod
  name: string;
  amount: number
  autoFocus?: boolean
}

export type PaymentTerm = {
  value: number
  label: string
}

export interface LineForm extends Item {
  quantity: number;
  amount: number;
}

export type PaymentForm = {
  cash: CashForm
  ck: CheckForm
  card: CardForm
  bt: BTForm
}

export interface Nameable {
  id: string | number;
  name: string
}

export type currencySignature = (value: number | string, precision?: number, inCent?: boolean) => string