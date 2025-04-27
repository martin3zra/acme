import FormSection from "@/components/form-section"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { BankOperationFormProps } from "@/types"


export const CheckFormView = ({amount, reference, onChange}: BankOperationFormProps) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if(event.currentTarget.name === "ck") {
      onChange(event.currentTarget.valueAsNumber)
      return
    }

    onChange(event.currentTarget.value)
  }
  return (
    <div>
      <FormSection onSubmit={() => {}}>
        <FormSection.Title>Check payment</FormSection.Title>
        <FormSection.Description>Specify the amount of the Check and the number for future reference.</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 sm:col-span-4 space-y-2">
            <Label  htmlFor='ck' className='text-end'>Amount</Label>
            <Input
              type="number"
              min={0}
              name="ck"
              className="text-end h-12 md:text-xl"
              onChange={handleChange}
              autoFocus
              value={amount}
            />
          </div>
          <div className="col-span-6 sm:col-span-4 space-y-2">
            <Label  htmlFor='ck'>CK Number</Label>
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