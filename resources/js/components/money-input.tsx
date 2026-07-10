import { formatMoney, parseMoney } from '@/lib/money';
import React from 'react';
import { Input } from './ui/input';

type Props = Omit<React.ComponentPropsWithoutRef<'input'>, 'value' | 'onChange'> & {
  value: number;
  onChange: (cents: number) => void;
};

export const MoneyInput = React.forwardRef<HTMLInputElement, Props>(function ({ value, onChange, ...inputProps }, ref) {
  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    onChange(parseMoney(e.target.value) / 100);
  };
  return <Input {...inputProps} ref={ref} min={0} inputMode="numeric" value={formatMoney(value * 100)} onChange={handleChange} />;
});
