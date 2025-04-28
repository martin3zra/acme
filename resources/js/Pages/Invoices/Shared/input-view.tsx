import { Input } from "@/components/ui/input";
import { PaymentMethod } from "@/types";
import { JSX } from "react";

type InputViewProps = {
  value: number;
  method: PaymentMethod;
  autoFocus?: boolean;
  onChange: (method: PaymentMethod, value: number) => void;
  onFocus: (method: PaymentMethod) => void;
}

export const InputView = ({ value, method, autoFocus, onChange, onFocus }: InputViewProps): JSX.Element => {

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange(method, event.currentTarget.valueAsNumber)
  }

  const handleOnFocus = (event: React.FocusEvent<HTMLInputElement>) => {
    onFocus(method)
  }

  return (
    <div className='p-0'>
      <Input
        key={method}
        type="number"
        min={0}
        className={`${method !== "cash" && 'cursor-pointer'} text-end border-none`}
        value={value}
        autoFocus={autoFocus}
        onFocus={handleOnFocus}
        onChange={handleChange}
        readOnly={method !== "cash"}
      />
    </div>
  )
}