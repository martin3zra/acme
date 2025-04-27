import FormSection from "@/components/form-section"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { BankOperationFormProps } from "@/types"

type BTFormProps = BankOperationFormProps & {}

export const BankTransferFormView = ({amount, reference, onChange}: BTFormProps) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if(event.currentTarget.name === "amount") {
      onChange(event.currentTarget.valueAsNumber)
      return
    }

    onChange(event.currentTarget.value)
  }
  return (
    <div>
      <FormSection onSubmit={() => {}}>
        <FormSection.Title>Bank Transfer payment</FormSection.Title>
        <FormSection.Description>Specify the amount of the Bank Transfer and the number for future reference.</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 sm:col-span-4 space-y-2">
            <Label  htmlFor='amount' className='text-end'>Amount</Label>
            <Input
              type="number"
              min={0}
              name="amount"
              className="text-end h-12 md:text-xl"
              onChange={handleChange}
              autoFocus
              value={amount}
            />
          </div>
          <div className="col-span-6 sm:col-span-4 space-y-2">
            <Label  htmlFor='reference'>Reference</Label>
            <Input
              type="text"
              name="reference"
              className="text-start h-12 md:text-xl"
              onChange={handleChange}
              value={reference}
            />
          </div>
        </FormSection.Form>
      </FormSection>
    </div>
  )
}
