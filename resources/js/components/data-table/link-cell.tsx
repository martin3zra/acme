import { Link } from '@inertiajs/react';
import { FC } from 'react';

export type Props = {
  href: string;
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
};

export const LinkCell: FC<Props> = ({ href, value, columnWidth }) => {
  return (
    <div className="text-primary p-1.5 font-medium hover:underline hover:underline-offset-4" title={value} style={{ width: columnWidth }}>
      <Link href={href}>{value}</Link>
    </div>
  );
};
