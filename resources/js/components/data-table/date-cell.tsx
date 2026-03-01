import { clampISODate } from '@/lib/utils';
import { FC } from 'react';

export type Props = {
  /**
   * Date as an ISO string
   * @example
   * new Date().toISOString()
   */
  value?: string;
  /**
   * Set the width of a column in pixels
   * @example
   * { header: props => <Cell columnWidth={props.column.getSize()} /> }
   */
  columnWidth: number;
};

export const DateCell: FC<Props> = ({ value, columnWidth }) => {
  /**
   * Intl.DateTimeFormat is a standard browser built-in object
   * @see https://developer.mozilla.org/en-US/docs/Web/JavaScript/Reference/Global_Objects/Intl/DateTimeFormat
   */
  const clamped = clampISODate(value);

  const formattedValue = clamped
    ? clamped.split('-').reverse().join('-') // 27-02-2026
    : '';

  return (
    <div className="truncate p-1.5 text-left tabular-nums" title={formattedValue} style={{ width: columnWidth }}>
      {formattedValue}
    </div>
  );
};
