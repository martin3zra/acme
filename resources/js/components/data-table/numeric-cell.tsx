import { FC } from 'react';

export type Props = {
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

export const NumericCell: FC<Props> = ({ value, columnWidth }) => {
  return (
    <div className="p-1.5 text-end" title={value} style={{ width: columnWidth }}>
      {value}
    </div>
  );
};
