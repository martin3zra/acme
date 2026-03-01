import { cn } from '@/lib/utils';
import { ComponentProps, FC } from 'react';

type Props = {
  title: string;
  columnWidth: number;
  alignment: 'left' | 'center' | 'right';
};

export const HeaderCell: FC<Props & ComponentProps<'div'>> = ({ title, columnWidth, alignment, className = '', ...props }) => {
  return (
    <div
      className={cn(
        'p-1.5 font-semibold uppercase',
        alignment === 'left' && 'text-start',
        alignment === 'center' && 'text-center',
        alignment === 'right' && 'text-end',
        className,
      )}
      style={{ width: columnWidth }}
      {...props}
    >
      {title}
    </div>
  );
};
