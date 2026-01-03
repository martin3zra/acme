import { MoneyInput } from '@/components/money-input';
import { PaymentMethod } from '@/types';
import { JSX } from 'react';

type InputViewProps = {
  value: number;
  method: PaymentMethod;
  autoFocus?: boolean;
  onChange: (method: PaymentMethod, value: number) => void;
  onFocus: (method: PaymentMethod) => void;
};

export const InputView = ({ value, method, autoFocus, onChange, onFocus }: InputViewProps): JSX.Element => {
  return (
    <div className="p-0">
      <MoneyInput
        key={method}
        className={`${method !== 'cash' && 'cursor-pointer'} border-none text-end`}
        value={value}
        autoFocus={autoFocus}
        onFocus={() => onFocus(method)}
        onChange={(c) => onChange(method, c)}
        readOnly={method !== 'cash'}
      />
    </div>
  );
};
