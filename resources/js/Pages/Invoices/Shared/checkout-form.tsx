import { AlertDestructive } from "@/components/alert-destructive"
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { InputView } from "./input-view"
import { BTForm, CardForm, CardFormInput, CashForm, CheckForm, PaymentForm, PaymentMethod, PaymentMethodType } from "@/types"
import { useNumber } from "@/composables/use-number"
import { CheckFormView } from "./check-form"
import { CardFormView } from "./card-form"
import { BankTransferFormView } from "./bank-transfer"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"
import React from "react"
import { defaultBTForm, defaultCardForm, defaultCashForm, defaultCheckForm } from "../constants"

type errorBag = {
  [key:string]: string
}

type CheckoutFormProps = {
  openCheckout: boolean,
  setCheckout: React.Dispatch<React.SetStateAction<boolean>>
  paymentMethods: PaymentMethodType[]
  paymentForm: PaymentForm
  errors: errorBag
  setCancelConfirmation: React.Dispatch<React.SetStateAction<boolean>>
  totalAmount: number
  onPlacedInvoice: (event: React.MouseEvent<HTMLButtonElement>) => void
  processing: boolean
  onPaymentChange: (method: PaymentMethod, value: number) => void
  onCheckoutChange: (method: PaymentMethod, form: CashForm | CheckForm | CardForm | BTForm) => void
}

type CheckoutFormState = {
  activePaymentForm: PaymentMethod
  cashForm: CashForm
  ckForm: CheckForm
  cardForm: CardForm
  btForm: BTForm
}

const CashFormView = () => <></>

class CheckoutForm extends React.Component<CheckoutFormProps, CheckoutFormState> {
  currency = useNumber().currency;
  constructor(props: CheckoutFormProps){
    super(props)
    // get initial value from
    console.log(props.paymentForm)
    this.state = {
      activePaymentForm: "cash",
      cashForm: props.paymentForm.cash,
      ckForm: props.paymentForm.ck,
      cardForm: props.paymentForm.card,
      btForm: props.paymentForm.bt,
    }
  }

   handleOnChangeInputView = (method: PaymentMethod, value: number) => {
    this.setState({activePaymentForm: method})

    if (typeof value  === "number" && method === "cash") {
      const givenValue =  isNaN(value) ? 0 : value
      this.setState((state: CheckoutFormState) => ({ cashForm: {...state.cashForm, amount: givenValue}}), () => {this.props.onCheckoutChange("cash",this.state.cashForm)})
      this.props.onPaymentChange("cash", givenValue)
      return
    }
  }

  handleOnChangeCheckFormView = (value: number|string) => {
    if (typeof value  === "number") {
      const givenValue =  isNaN(value) ? 0 : value
      this.setState((state: CheckoutFormState) => ({ ckForm: {...state.ckForm, amount: givenValue}}), () => {this.props.onCheckoutChange("ck", this.state.ckForm)})
      this.props.onPaymentChange("ck", givenValue)
      return
    }

    this.setState((state: CheckoutFormState) => ({ ckForm: {...state.ckForm, reference: value}}), () => {this.props.onCheckoutChange("ck", this.state.ckForm)})
  }

  handleOnChangeCardFormView = (value: number | string, key: CardFormInput) => {
    if (typeof value  === "number" && key === "last4") {
      const givenValue =  isNaN(value) ? 0 : value
      this.setState((state: CheckoutFormState) => ({ cardForm: {...state.cardForm, last4: givenValue}}), () => {this.props.onCheckoutChange("card", this.state.cardForm)})
      return
    }
    if (key === "amount"){
      this.setState((state: CheckoutFormState) => ({ cardForm: {...state.cardForm, [key]: Number(value)}}), () => {this.props.onCheckoutChange("card", this.state.cardForm)})
      this.props.onPaymentChange("card", Number(value))
      return
    }

    this.setState((state: CheckoutFormState) => ({ cardForm: {...state.cardForm, [key]: value}}), () => {this.props.onCheckoutChange("card", this.state.cardForm)})
  }

  handleOnChangeBTFormView = (value: number|string) => {
    if (typeof value  === "number") {
      const givenValue =  isNaN(value) ? 0 : value
      this.setState((state: CheckoutFormState) => ({ btForm: {...state.btForm, amount: givenValue}}), () => {this.props.onCheckoutChange("bt", this.state.btForm)})
      this.props.onPaymentChange("bt", givenValue)
      return
    }

    this.setState((state: CheckoutFormState) => ({ btForm: {...state.btForm, reference: value}}), () => {this.props.onCheckoutChange("bt", this.state.btForm)})
  }

  receivedAmount = this.props.paymentMethods.reduce((accumulator, method) => accumulator + method.amount, 0)

  computeRemainingBalance = (): number => {
    return this.props.totalAmount - this.receivedAmount
  }


  renderPaymentMethodForm = () => {
    if (this.state.activePaymentForm === "ck") return <CheckFormView {...this.state.ckForm} onChange={this.handleOnChangeCheckFormView}/>
    if (this.state.activePaymentForm === "card") return <CardFormView {...this.state.cardForm} onChange={this.handleOnChangeCardFormView} />
    if (this.state.activePaymentForm === "bt") return <BankTransferFormView {...this.state.btForm} onChange={this.handleOnChangeBTFormView}  />
    return <CashFormView />
  }
  render() {
    const { openCheckout, setCheckout, errors, totalAmount, onPlacedInvoice, processing, setCancelConfirmation } = this.props
    return (
      <Sheet open={openCheckout} onOpenChange={setCheckout}>
        <SheetContent side='right' className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl">
          <SheetHeader>
            <SheetTitle>Checkout</SheetTitle>
            <SheetDescription className="text-[12px]">Checkout process</SheetDescription>
          </SheetHeader>
          <div className="grid gap-4 px-4">
            {JSON.stringify(this.props.paymentMethods)}
            {this.props.errors.status && <AlertDestructive description={this.props.errors.status} onDestroy={() => delete errors.status }/>}
            <h4>Payment detail</h4>
            <div className='flex justify-between items-center w-full'>
                <table className="w-full table-auto">
                <thead>
                  <tr>
                  {this.props.paymentMethods.map((method) =>
                    <th scope="col" key={method.value} className="w-60 border border-gray-300">{method.name}</th>
                  )}
                  </tr>
                </thead>
                <tbody>
                  <tr>
                  {this.props.paymentMethods.map((method) =>
                    <td key={method.value} className="border px-1 border-gray-300 text-start">
                      <InputView
                        key={method.value}
                        value={method.amount}
                        method={method.value}
                        onChange={this.handleOnChangeInputView}
                        onFocus={(methodType) => this.setState({activePaymentForm: methodType})}
                        />
                    </td>
                  )}
                  </tr>
                </tbody>
              </table>

            </div>
            <div className='pb-6'>
              {this.renderPaymentMethodForm()}
            </div>
            <Separator className='' />
            <div>
              <div className='flex justify-between items-center w-60'>
                <span className="block text-2xl">To collect</span>
                <span className="block text-2xl">{this.currency(totalAmount)}</span>
              </div>
              <div className='flex justify-between items-center w-60'>
                <span className="block text-2xl">Received</span>
                <span className="block text-2xl">{this.currency(this.receivedAmount)}</span>
              </div>
              <div className='flex justify-between items-center w-60'>
                <span className="block text-2xl">Remaining</span>
                <span className="block text-2xl text-red-600 font-medium">{this.currency(this.computeRemainingBalance())}</span>
              </div>
            </div>
          </div>
          <SheetFooter>
            {this.computeRemainingBalance() !== 0 && <AlertDestructive description="The amount collected must be equals to the Invoice total amount." destroyable={false} />}
            <div className='flex justify-end gap-x-6'>
              <Button variant={"secondary"} onClick={() => setCancelConfirmation(true)}>Cancel</Button>
              <Button onClick={onPlacedInvoice} disabled={processing || this.computeRemainingBalance() !== 0}>Complete Invoice</Button>
            </div>
          </SheetFooter>
        </SheetContent>
      </Sheet>
    )
  }
}

export default CheckoutForm