import { d as defaultBTForm, a as defaultCardForm, b as defaultCheckForm, c as defaultCashForm } from "./constants2.js";
import { d as defaultBreadcrumbs } from "./index13.js";
const defaultPaymentMethodsForm = {
  cash: defaultCashForm,
  ck: defaultCheckForm,
  card: defaultCardForm,
  bt: defaultBTForm
};
const defaultHeaderForm = {
  customer: void 0,
  date: void 0,
  notes: "",
  discount: 0
};
const defaultPaymentForm = { header: defaultHeaderForm, lines: [], payment: defaultPaymentMethodsForm };
const breadcrumbs = [
  ...defaultBreadcrumbs,
  {
    title: "payments.title",
    href: "/payments"
  }
];
const createPaymentBreadcrumbs = [
  ...breadcrumbs,
  {
    title: "payments.newPayment.title",
    href: "/payments/create"
  }
];
const editPaymentBreadcrumbs = [
  ...breadcrumbs,
  {
    title: "payments.editPayment.title",
    href: ""
  }
];
export {
  breadcrumbs as b,
  createPaymentBreadcrumbs as c,
  defaultPaymentForm as d,
  editPaymentBreadcrumbs as e
};
