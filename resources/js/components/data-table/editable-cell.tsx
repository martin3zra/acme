import { CellContext } from '@tanstack/react-table';
import React, { useState } from 'react';
import { MoneyInput } from '../money-input';

type EditableCellProps<TData, TValue> = CellContext<TData, TValue> & {
  identifier: string;
  inputType?: React.HTMLInputTypeAttribute; // 'text' | 'number' | 'email' | etc.
};

export function EditableCell<TData, TValue extends string | number>({
  getValue,
  row,
  column,
  table,
  identifier,
  inputType = 'text',
}: EditableCellProps<TData, TValue>) {
  const initialValue = getValue();
  const [value, setValue] = useState(initialValue);

  const onBlur = () => {
    table.options.meta?.updateData(identifier, row.index, column.id, value);
  };

  if (inputType === 'number') {
    return (
      <MoneyInput
        id={identifier}
        value={Number(value || 0)}
        onChange={(c) => setValue(c as TValue)}
        onBlur={onBlur}
        className="h-10 w-full px-2 py-0 text-end border-0 shadow-none"
      />
    )
  }

  return (
    <input
      id={identifier}
      type={inputType}
      value={value}
      onChange={(e) => {
        const val = inputType === 'number' ? Number(e.target.value) : e.target.value;
        setValue(val as TValue);
      }}
      onBlur={onBlur}
      data-type={inputType}
      className="h-10 w-full px-2 py-0"
    />
  );
}
