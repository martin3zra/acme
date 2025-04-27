import { AlertDestructive } from "@/components/alert-destructive"
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from "@/components/ui/sheet"
import { InputView } from "./input-view"
import { CardForm, CardFormInput, PaymentFormType, PaymentMethod, PaymentMethodType } from "@/types"
import { useNumber } from "@/composables/use-number"
import { CheckFormView } from "./check-form"
import { CardFormView } from "./card-form"
import { BankTransferFormView } from "./bank-transfer"
import { Button } from "@/components/ui/button"
import { Separator } from "@/components/ui/separator"

type errorBag = {
  [key:string]: string
}

type CheckoutFormProps = {
  openCheckout: boolean,
  setCheckout: React.Dispatch<React.SetStateAction<boolean>>
  paymentMethods: PaymentMethodType[]
  errors: errorBag
  handleOnChangeInputView: (method: PaymentMethod, value: number) => void
  activePaymentForm: PaymentMethod
  setActivePaymentForm: React.Dispatch<React.SetStateAction<PaymentMethod>>
  setCancelConfirmation: React.Dispatch<React.SetStateAction<boolean>>
  ckForm: PaymentFormType
  cardForm: CardForm
  btForm: PaymentFormType
  handleOnChangeCheckFormView: (value: number | string) => void
  handleOnChangeCardFormView: (value: number | string, key: CardFormInput) => void
  handleOnChangeBTFormView: (value: number | string) => void
  remainingBalance: number
  totalAmount: number
  receivedAmount: number
  onPlacedInvoice: (event: React.MouseEvent<HTMLButtonElement>) => void
  processing: boolean
}

export const CheckoutForm = ({
  openCheckout,
  setCheckout,
  paymentMethods,
  errors,
  handleOnChangeInputView,
  activePaymentForm,
  setActivePaymentForm,
  ckForm,
  cardForm,
  btForm,
  handleOnChangeCheckFormView,
  handleOnChangeCardFormView,
  handleOnChangeBTFormView,
  remainingBalance,
  totalAmount,
  receivedAmount,
  setCancelConfirmation,
  onPlacedInvoice,
  processing
}: CheckoutFormProps) => {
  const currency = useNumber().currency;

  const CashFormView = () => <></>

  const renderPaymentMethodForm = () => {
    if (activePaymentForm === "ck") return <CheckFormView {...ckForm} onChange={handleOnChangeCheckFormView}/>
    if (activePaymentForm === "card") return <CardFormView {...cardForm} onChange={handleOnChangeCardFormView} />
    if (activePaymentForm === "bt") return <BankTransferFormView {...btForm} onChange={handleOnChangeBTFormView}  />
    return <CashFormView />
  }
  return (
    <Sheet open={openCheckout} onOpenChange={setCheckout}>
      <SheetContent side='right' className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl">
        <SheetHeader>
          <SheetTitle>Checkout</SheetTitle>
          <SheetDescription className="text-[12px]">Checkout process</SheetDescription>
        </SheetHeader>
        <div className="grid gap-4 px-4">
          {errors.status && <AlertDestructive description={errors.status} onDestroy={() => delete errors.status }/>}
          <h4>Payment detail</h4>
          <div className='flex justify-between items-center w-full'>
              <table className="w-full table-auto">
              <thead>
                <tr>
                {paymentMethods.map((method) =>
                  <th scope="col" key={method.value} className="w-60 border border-gray-300">{method.name}</th>
                )}
                </tr>
              </thead>
              <tbody>
                <tr>
                {paymentMethods.map((method) =>
                  <td key={method.value} className="border px-1 border-gray-300 text-start">
                    <InputView
                      key={method.value}
                      value={method.amount}
                      method={method.value}
                      onChange={handleOnChangeInputView}
                      onFocus={(methodType) => setActivePaymentForm(methodType)}
                      />
                  </td>
                )}
                </tr>
              </tbody>
            </table>

          </div>
          <div className='pb-6'>
            {renderPaymentMethodForm()}
          </div>
          <Separator className='' />
          <div>
            <div className='flex justify-between items-center w-60'>
              <span className="block text-2xl">To collect</span>
              <span className="block text-2xl">{currency(totalAmount)}</span>
            </div>
            <div className='flex justify-between items-center w-60'>
              <span className="block text-2xl">Received</span>
              <span className="block text-2xl">{currency(receivedAmount)}</span>
            </div>
            <div className='flex justify-between items-center w-60'>
              <span className="block text-2xl">Remaining</span>
              <span className="block text-2xl text-red-600 font-medium">{currency(remainingBalance)}</span>
            </div>
          </div>
        </div>
        <SheetFooter>
          {remainingBalance !== 0 && <AlertDestructive description="The amount collected must be equals to the Invoice total amount." destroyable={false} />}
          <div className='flex justify-end gap-x-6'>
            <Button variant={"secondary"} onClick={() => setCancelConfirmation(true)}>Cancel</Button>
            <Button onClick={onPlacedInvoice} disabled={processing || remainingBalance !== 0}>Complete Invoice</Button>
          </div>
        </SheetFooter>
      </SheetContent>
    </Sheet>
  )
}