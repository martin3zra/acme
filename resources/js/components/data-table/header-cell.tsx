import { FC } from 'react';

type Props = {
  title: string;
  columnWidth: number;
  alignment: 'left' | 'center' | 'right';
};

export const HeaderCell: FC<Props> = ({ title, columnWidth, alignment }) => {
  return (
    <div className={`p-1.5 font-semibold text-${alignment} uppercase`} style={{ width: columnWidth }}>
      {title}
    </div>
  );
};
