import { cn } from '@/lib/utils';
import { ComponentProps, FC } from 'react';

export interface Props extends ComponentProps<'div'> {
  /**
   * Provide the value to render in the cell
   */
  value?: string;
  /**
   * Set the width of a column in pixels
   * @example
   * { header: props => <Cell columnWidth={props.column.getSize()} /> }
   */
  columnWidth: number;
}

export const TextCell: FC<Props> = ({ value, className, columnWidth }) => {
  return (
    <div className={cn(className, 'p-1.5')} title={value} style={{ width: columnWidth }}>
      {value}
    </div>
  );
};
