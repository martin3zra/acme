import { useNumber } from '@/composables/use-number';
import { FC } from 'react';

type Props = {
  value?: string;
  columnWidth: number;
  suffix?: string;
};

export const CurrencyCell: FC<Props> = ({ value, columnWidth, suffix }) => {
  const { currency } = useNumber();
  let formattedValue = value !== undefined ? currency(value) : '$0.00';
  formattedValue = suffix !== undefined ? formattedValue.substring(1) : formattedValue;
  return (
    <div className="p-1.5 text-right" title={formattedValue} style={{ width: columnWidth }}>
      {formattedValue}
      {suffix && <span>{suffix}</span>}
    </div>
  );
};
