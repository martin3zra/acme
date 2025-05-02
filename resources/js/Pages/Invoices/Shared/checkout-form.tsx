import { AlertDestructive } from '@/components/alert-destructive';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { cn, subtractFloats } from '@/lib/utils';
import { BTForm, CardForm, CardFormInput, CashForm, CheckForm, PaymentForm, PaymentMethod, PaymentMethodType } from '@/types';
import React from 'react';
import { defaultPaymentMethods } from '../constants';
import { BankTransferFormView } from './bank-transfer';
import { CardFormView } from './card-form';
import { CheckFormView } from './check-form';
import { InputView } from './input-view';

type errorBag = {
  [key: string]: string;
};

type CheckoutFormProps = {
  openCheckout: boolean;
  setCheckout: React.Dispatch<React.SetStateAction<boolean>>;
  paymentForm: PaymentForm;
  errors: errorBag;
  setCancelConfirmation: React.Dispatch<React.SetStateAction<boolean>>;
  totalAmount: number;
  onPlacedInvoice: (event: React.MouseEvent<HTMLButtonElement>) => void;
  processing: boolean;
  onCheckoutChange: (method: PaymentMethod, form: CashForm | CheckForm | CardForm | BTForm) => void;
  currency: (value: number | string, precision?: number, inCent?: boolean) => string;
};

type CheckoutFormState = {
  paymentMethods: PaymentMethodType[];
  activePaymentForm: PaymentMethod;
  receivedAmount: number;
  remainingBalance: number;
  cashForm: CashForm;
  ckForm: CheckForm;
  cardForm: CardForm;
  btForm: BTForm;
};

class CheckoutForm extends React.Component<CheckoutFormProps, CheckoutFormState> {
  constructor(props: CheckoutFormProps) {
    super(props);

    const paymentMethods = this.hydratePaymentMethods();
    const receivedAmount = paymentMethods.reduce((accumulator, method) => accumulator + method.amount, 0);
    const remainingBalance = props.totalAmount - receivedAmount;

    this.state = {
      activePaymentForm: 'cash',
      paymentMethods: paymentMethods,
      receivedAmount: receivedAmount,
      remainingBalance: remainingBalance,
      cashForm: props.paymentForm.cash,
      ckForm: props.paymentForm.ck,
      cardForm: props.paymentForm.card,
      btForm: props.paymentForm.bt,
    };
  }

  hydratePaymentMethods = (): PaymentMethodType[] => {
    const { paymentForm } = this.props;
    const methods = defaultPaymentMethods;

    methods.filter((p) => p.value == 'cash')[0].amount = paymentForm.cash.amount;
    methods.filter((p) => p.value == 'ck')[0].amount = paymentForm.ck.amount;
    methods.filter((p) => p.value == 'card')[0].amount = paymentForm.card.amount;
    methods.filter((p) => p.value == 'bt')[0].amount = paymentForm.bt.amount;
    return methods;
  };

  onPaymentChange = (method: PaymentMethod, value: number) => {
    this.state.paymentMethods.filter((p) => p.value == method)[0].amount = value;
  };

  handleOnChangeInputView = (method: PaymentMethod, value: number) => {
    this.setState({ activePaymentForm: method });

    if (typeof value === 'number' && method === 'cash') {
      const givenValue = isNaN(value) ? 0 : value;
      this.setState(
        (state: CheckoutFormState) => ({ cashForm: { ...state.cashForm, amount: givenValue } }),
        () => this.computeTotals('cash', this.state.cashForm),
      );
      this.onPaymentChange('cash', givenValue);
      return;
    }
  };

  handleOnChangeCheckFormView = (value: number | string) => {
    if (typeof value === 'number') {
      const givenValue = isNaN(value) ? 0 : value;
      this.setState(
        (state: CheckoutFormState) => ({ ckForm: { ...state.ckForm, amount: givenValue } }),
        () => this.computeTotals('ck', this.state.ckForm),
      );
      this.onPaymentChange('ck', givenValue);
      return;
    }

    this.setState(
      (state: CheckoutFormState) => ({ ckForm: { ...state.ckForm, reference: value } }),
      () => this.props.onCheckoutChange('ck', this.state.ckForm),
    );
  };

  handleOnChangeCardFormView = (value: number | string, key: CardFormInput) => {
    if (typeof value === 'number' && key === 'last4') {
      const givenValue = isNaN(value) ? 0 : value;
      this.setState(
        (state: CheckoutFormState) => ({ cardForm: { ...state.cardForm, last4: givenValue } }),
        () => this.props.onCheckoutChange('card', this.state.cardForm),
      );
      return;
    }
    if (key === 'amount') {
      this.setState(
        (state: CheckoutFormState) => ({ cardForm: { ...state.cardForm, [key]: Number(value) } }),
        () => this.computeTotals('card', this.state.cardForm),
      );
      this.onPaymentChange('card', Number(value));
      return;
    }

    this.setState(
      (state: CheckoutFormState) => ({ cardForm: { ...state.cardForm, [key]: value } }),
      () => this.props.onCheckoutChange('card', this.state.cardForm),
    );
  };

  handleOnChangeBTFormView = (value: number | string) => {
    if (typeof value === 'number') {
      const givenValue = isNaN(value) ? 0 : value;
      this.setState(
        (state: CheckoutFormState) => ({ btForm: { ...state.btForm, amount: givenValue } }),
        () => this.computeTotals('bt', this.state.btForm),
      );
      this.onPaymentChange('bt', givenValue);
      return;
    }

    this.setState(
      (state: CheckoutFormState) => ({ btForm: { ...state.btForm, reference: value } }),
      () => this.props.onCheckoutChange('bt', this.state.btForm),
    );
  };

  computeTotals = (method: PaymentMethod, form: CashForm | CheckForm | CardForm | BTForm) => {
    this.setState(
      (state: CheckoutFormState) => {
        const receivedAmount = state.paymentMethods.reduce((accumulator, method) => accumulator + method.amount, 0);
        return {
          receivedAmount: receivedAmount,
          remainingBalance: subtractFloats(this.props.totalAmount, receivedAmount),
        };
      },
      () => this.props.onCheckoutChange(method, form),
    );
  };

  renderPaymentMethodForm = () => {
    const { activePaymentForm, ckForm, cardForm, btForm } = this.state;
    return {
      cash: null,
      ck: <CheckFormView {...ckForm} onChange={this.handleOnChangeCheckFormView} />,
      card: <CardFormView {...cardForm} onChange={this.handleOnChangeCardFormView} />,
      bt: <BankTransferFormView {...btForm} onChange={this.handleOnChangeBTFormView} />,
    }[activePaymentForm];
  };

  render() {
    const { receivedAmount, remainingBalance } = this.state;
    const { openCheckout, setCheckout, errors, totalAmount, onPlacedInvoice, processing, setCancelConfirmation } = this.props;
    return (
      <Sheet open={openCheckout} onOpenChange={setCheckout}>
        <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl">
          <SheetHeader>
            <SheetTitle>Checkout: {this.state.activePaymentForm}</SheetTitle>
            <SheetDescription className="text-[12px]">Checkout process</SheetDescription>
          </SheetHeader>
          <div className="grid gap-4 px-4">
            {errors.status && <AlertDestructive description={errors.status} onDestroy={() => delete errors.status} />}
            {Object.keys(errors).map((e) => (
              <AlertDestructive description={errors[e]} destroyable={false} />
            ))}
            <div className="flex w-full items-center justify-between">
              <table className="w-full table-auto">
                <thead>
                  <tr>
                    {this.state.paymentMethods.map((method) => (
                      <th
                        data-slot={`${method.value === this.state.activePaymentForm ? 'current' : 'default'}`}
                        scope="col"
                        key={method.value}
                        className={cn(
                          'w-60 border border-gray-300 px-7 text-end',
                          'data-[slot=current]:bg-primary data-[slot=current]:text-primary-foreground data-[slot=current]:border-foreground',
                        )}
                      >
                        {method.name}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    {this.state.paymentMethods.map((method) => (
                      <td key={method.value} className="border border-gray-300 px-1 text-start">
                        <InputView
                          key={method.value}
                          value={method.amount}
                          method={method.value}
                          onChange={this.handleOnChangeInputView}
                          onFocus={(pm) => this.setState({ activePaymentForm: pm })}
                        />
                      </td>
                    ))}
                  </tr>
                </tbody>
              </table>
            </div>
            <div className="pb-6">{this.renderPaymentMethodForm()}</div>
            <Separator className="" />
            <div>
              <div className="flex w-60 items-center justify-between">
                <span className="block text-2xl">To collect</span>
                <span className="block text-2xl">{this.currency(totalAmount)}</span>
              </div>
              <div className="flex w-60 items-center justify-between">
                <span className="block text-2xl">Received</span>
                <span className="block text-2xl">{this.currency(receivedAmount)}</span>
              </div>
              <div className="flex w-60 items-center justify-between">
                <span className="block text-2xl">Remaining</span>
                <span className="block text-2xl font-medium text-red-600">{this.currency(remainingBalance)}</span>
              </div>
            </div>
          </div>
          <SheetFooter>
            <div className="flex justify-end gap-x-6">
              <Button variant={'secondary'} onClick={() => setCancelConfirmation(true)}>
                Cancel
              </Button>
              <Button onClick={onPlacedInvoice} disabled={processing || remainingBalance !== 0}>
                Complete Invoice
              </Button>
            </div>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    );
  }
}

export default CheckoutForm;
