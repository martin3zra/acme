import { d as defaultBTForm, a as defaultCardForm, b as defaultCheckForm, c as defaultCashForm } from "./constants2.js";
import { d as defaultBreadcrumbs } from "./index13.js";
const paymentTerms = [
  { value: 1, label: "Cash" },
  { value: 7, label: "7 Days" },
  { value: 10, label: "10 Days" },
  { value: 15, label: "15 Days" },
  { value: 30, label: "30 Days" },
  { value: 60, label: "60 Days" },
  { value: 90, label: "90 Days" }
];
const breadcrumbs = [
  ...defaultBreadcrumbs,
  {
    title: "invoices.title",
    href: "/invoices"
  }
];
const createBreadcrumbs = [
  ...breadcrumbs,
  {
    title: "invoices.newInvoice.title",
    href: "/invoices/create"
  }
];
const editBreadcrumbs = [
  ...breadcrumbs,
  {
    title: "invoices.editInvoice.title",
    href: ""
  }
];
const defaultPaymentMethodsForm = {
  cash: defaultCashForm,
  ck: defaultCheckForm,
  card: defaultCardForm,
  bt: defaultBTForm
};
const defaultDiscount = { value: 0, type: "fixed" };
const defaultHeaderForm = {
  customer: void 0,
  date: void 0,
  due: void 0,
  terms: 0,
  taxReceipt: 0,
  notes: void 0,
  discount: defaultDiscount
};
const defaultInvoiceForm = { header: defaultHeaderForm, lines: [], payment: defaultPaymentMethodsForm };
export {
  defaultInvoiceForm as a,
  breadcrumbs as b,
  createBreadcrumbs as c,
  defaultDiscount as d,
  editBreadcrumbs as e,
  paymentTerms as p
};
