import { Column } from '@tanstack/react-table';
import { ArrowDown, ArrowUp, ArrowUpDown } from 'lucide-react';
import { ComponentProps } from 'react';

type Props<T> = {
  title: string;
  column: Column<T, unknown>;
};
export const HeaderSortCell = <T,>({ title, column, ...props }: Props<T> & ComponentProps<'div'>) => {
  const isSorted = column.getIsSorted(); // false | 'asc' | 'desc'

  return (
    <div className="hover:bg-accent flex w-fit cursor-pointer items-center gap-x-2 p-1.5 uppercase" onClick={() => column.toggleSorting()} {...props}>
      {title}
      {isSorted === 'asc' && <ArrowUp className="size-4" />}
      {isSorted === 'desc' && <ArrowDown className="size-4" />}
      {!isSorted && <ArrowUpDown className="size-4" />}
    </div>
  );
};
