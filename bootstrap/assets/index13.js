const PaidStatuses = ["paid", "unpaid", "partial", "removed", "overpaid", "pending"];
function mapPaymentLineToReceivableInvoice(paymentLine) {
  const { invoice } = paymentLine;
  return {
    id: paymentLine.id,
    uuid: invoice.uuid,
    number: invoice.number,
    ncf: invoice.ncf,
    // Placeholder since PaymentLine does not have this field
    date: new Date(invoice.date),
    due_on: new Date(invoice.due_on),
    // Placeholder, not present in PaymentLine
    total: invoice.amount,
    amount_due: invoice.amount_due,
    paid_status: invoice.paid_status,
    notes: invoice.notes,
    // Placeholder since notes don't exist in PaymentLine
    original_payment: paymentLine.payment,
    payment: paymentLine.payment,
    discount: 0,
    balance: 0,
    action: "unchanged"
    // Placeholder, as the action is not defined in PaymentLine
  };
}
const defaultBreadcrumbs = [
  {
    title: "global.navMain.dashboard",
    href: "/home"
  }
];
export {
  PaidStatuses as P,
  defaultBreadcrumbs as d,
  mapPaymentLineToReceivableInvoice as m
};
