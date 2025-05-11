import { BTForm, CardBrand, CardForm, CashForm, CheckForm, PaymentMethodType } from './types';

export const defaultCheckForm: CheckForm = {
  amount: 0,
  reference: '',
};

export const defaultCashForm: CashForm = {
  amount: 0,
};

export const defaultCardBrands: CardBrand[] = [
  { value: 'visa', name: 'Visa' },
  { value: 'mastercard', name: 'MasterCard' },
  { value: 'ae', name: 'American Express' },
  { value: 'unknown', name: 'Unknown' },
];

export const defaultCardForm: CardForm = {
  last4: 0,
  brand: 'unknow',
  amount: 0,
  reference: '',
};

export const defaultBTForm: BTForm = {
  amount: 0,
  reference: '',
};

export const defaultPaymentMethods: PaymentMethodType[] = [
  { value: 'cash', name: 'Cash', amount: 0, autoFocus: true },
  { value: 'ck', name: 'CK', amount: 0 },
  { value: 'card', name: 'Debit/Credit Card', amount: 0 },
  { value: 'bt', name: 'Bank Transfer', amount: 0 },
];
