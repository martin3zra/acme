import { defaultBTForm, defaultCardForm, defaultCashForm, defaultCheckForm } from '@/constants';
import { DiscountType, PaymentForm, PaymentHeaderForm, PaymentMethodsForm } from '@/types';

export const defaultPaymentMethodsForm: PaymentMethodsForm = {
  cash: defaultCashForm,
  ck: defaultCheckForm,
  card: defaultCardForm,
  bt: defaultBTForm,
};
export const defaultDiscount: DiscountType = { value: 0, type: 'fixed' };
export const defaultHeaderForm: PaymentHeaderForm = {
  customer: undefined,
  date: undefined,
  notes: '',
  discount: 0,
};

export const defaultPaymentForm: PaymentForm = { header: defaultHeaderForm, lines: [], payment: defaultPaymentMethodsForm };
