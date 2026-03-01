import { CellContext } from '@tanstack/react-table';
import React, { useEffect, useState } from 'react';
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

  useEffect(() => {
    setValue(getValue());
  }, [getValue]);

  const commitChange = () => {
    // table.options.meta?.updateData?.(identifier, value);
    table.options.meta?.updateData?.(row.id.toString(), column.id, value);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      commitChange();
      e.currentTarget.blur(); // optional: forces blur so UI updates consistently
    }
  };

  if (inputType === 'number') {
    return (
      <MoneyInput
        id={identifier}
        aria-label={identifier}
        value={Number(value || 0)}
        onChange={(c) => setValue(c as TValue)}
        onBlur={commitChange}
        onKeyDown={handleKeyDown}
        className="h-10 w-full border-0 px-2 py-0 text-end shadow-none"
      />
    );
  }

  return (
    <input
      id={identifier}
      aria-label={identifier}
      type={inputType}
      value={value}
      onChange={(e) => {
        if (inputType === 'number') {
          const parsed = Number(e.target.value);
          setValue((isNaN(parsed) ? 0 : parsed) as TValue);
          return;
        }
        setValue(e.target.value as TValue);
      }}
      onBlur={commitChange}
      onKeyDown={handleKeyDown}
      data-type={inputType}
      className="h-10 w-full px-2 py-0"
    />
  );
}
