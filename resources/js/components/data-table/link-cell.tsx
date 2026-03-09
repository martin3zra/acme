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
    <div className="w-full min-w-0 p-1.5" title={value} style={{ width: columnWidth }}>
      <Link className="text-primary block truncate font-medium hover:underline hover:underline-offset-4" href={href}>
        {value}
      </Link>
    </div>
  );
};
