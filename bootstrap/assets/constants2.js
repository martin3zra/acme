const defaultCheckForm = {
  amount: 0,
  reference: ""
};
const defaultCashForm = {
  amount: 0
};
const defaultCardBrands = [
  { value: "visa", name: "Visa" },
  { value: "mastercard", name: "MasterCard" },
  { value: "ae", name: "American Express" },
  { value: "unknown", name: "Unknown" }
];
const defaultCardForm = {
  last4: 0,
  brand: "unknow",
  amount: 0,
  reference: ""
};
const defaultBTForm = {
  amount: 0,
  reference: ""
};
const defaultPaymentMethods = [
  { value: "cash", name: "Cash", amount: 0, autoFocus: true },
  { value: "ck", name: "CK", amount: 0 },
  { value: "card", name: "Debit/Credit Card", amount: 0 },
  { value: "bt", name: "Bank Transfer", amount: 0 }
];
export {
  defaultCardForm as a,
  defaultCheckForm as b,
  defaultCashForm as c,
  defaultBTForm as d,
  defaultPaymentMethods as e,
  defaultCardBrands as f
};
