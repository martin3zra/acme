import { useNumber } from '@/composables/use-number';
import { FC } from 'react';

type Props = {
  value?: string;
  columnWidth: number;
};

export const CurrencyCell: FC<Props> = ({ value, columnWidth }) => {
  const { currency } = useNumber();
  const formattedValue = value !== undefined ? currency(value) : '';
  return (
    <div className="p-1.5 text-right" title={formattedValue} style={{ width: columnWidth }}>
      {formattedValue}
    </div>
  );
};
