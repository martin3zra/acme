import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useTranslation } from '@/hooks/use-translation';
import { DueInvoice, TransactionKind } from '@/types';
import { flexRender, getCoreRowModel, getSortedRowModel, useReactTable } from '@tanstack/react-table';
import { FC } from 'react';
import { getColumns } from './columns-definitions';
type Props = {
  kind: TransactionKind;
  title: string;
  data: DueInvoice[];
  onSelectItem: (item: DueInvoice, action: 'view:customer' | 'view:invoice' | 'view:estimate' | 'record:payment') => void;
};
export const RecentTransactionList: FC<Props> = ({ kind, title, data, onSelectItem }) => {
  const t = useTranslation().trans;
  let columns = getColumns({ kind, onDidClick: onSelectItem, t });
  if (kind === 'invoice') {
    columns = columns.filter((f) => f.id !== 'status');
  }
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
    getSortedRowModel: getSortedRowModel(),
  });
  return (
    <div>
      <h1 className="py-2 text-lg">{title}</h1>
      <div className="rounded-md border">
        <Table>
          <TableHeader>
            {table.getHeaderGroups().map((headerGroup) => (
              <TableRow key={headerGroup.id}>
                {headerGroup.headers.map((header) => {
                  return (
                    <TableHead key={header.id}>
                      {header.isPlaceholder ? null : flexRender(header.column.columnDef.header, header.getContext())}
                    </TableHead>
                  );
                })}
              </TableRow>
            ))}
          </TableHeader>
          <TableBody>
            {table.getRowModel().rows?.length ? (
              table.getRowModel().rows.map((row) => (
                <TableRow key={row.id} data-state={row.getIsSelected() && 'selected'}>
                  {row.getVisibleCells().map((cell) => (
                    <TableCell key={cell.id}>{flexRender(cell.column.columnDef.cell, cell.getContext())}</TableCell>
                  ))}
                </TableRow>
              ))
            ) : (
              <TableRow>
                <TableCell colSpan={columns.length} className="h-24 text-center">
                  {t('global.noResults')}
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </div>
    </div>
  );
};
