import { BreadcrumbItem, BTForm, CardBrand, CardForm, CashForm, CheckForm, PaymentMethodType, PaymentTerm } from "@/types";

export const defaultCheckForm: CheckForm = {
  amount: 0,
  reference: "",
}

export const defaultCashForm: CashForm = {
  amount: 0,
}

export const defaultCardBrands: CardBrand[] = [
  {value:"visa", name: "Visa"},
  {value:"mastercard", name: "MasterCard"},
  {value:"ae", name: "American Express"},
  {value:"unknown", name: "Unknown"},
]

export const defaultCardForm: CardForm = {
  last4: 0,
  brand: "unknow",
  amount: 0,
  reference: "",
}

export const defaultBTForm: BTForm = {
  amount: 0,
  reference: "",
}

export const defaultPaymentMethods: PaymentMethodType[] = [
  {value: "cash", name: "Cash", amount: 0, autoFocus: true},
  {value: "ck", name: "CK", amount: 0},
  {value: "card", name: "Debit/Credit Card", amount: 0},
  {value: "bt", name: "Bank Transfer", amount: 0},
]

export const paymentTerms: PaymentTerm[] = [
  {value: 1, label: "Cash"},
  {value: 7, label: "7 Days"},
  {value: 10, label: "10 Days"},
  {value: 15, label: "15 Days"},
  {value: 30, label: "30 Days"},
  {value: 60, label: "60 Days"},
  {value: 90, label: "90 Days"},
]

export const breadcrumbs: BreadcrumbItem[] = [
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