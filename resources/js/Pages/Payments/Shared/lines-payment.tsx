import { SkeletonRow } from '@/components/data-table/skeleton-row';
import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuCheckboxItem, DropdownMenuContent, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import { onValueChangeType, PaymentVerb, ReceivableInvoiceForm } from '@/types';
import {
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  RowSelectionState,
  SortingState,
  useReactTable,
  VisibilityState,
} from '@tanstack/react-table';
import { ChevronDown } from 'lucide-react';
import { FC, useState } from 'react';
import { calculateTotals } from '../build-receivables-state';
import { getColumns } from './columns-definitions';

type Props = {
  data: ReceivableInvoiceForm[];
  rowSelection: RowSelectionState;
  loading: boolean;
  setRowSelection: React.Dispatch<React.SetStateAction<RowSelectionState>>;
  onSelectPaymentLine: (receivableInvoiceForm: ReceivableInvoiceForm, action: PaymentVerb) => void;
  onValueChange?: onValueChangeType;
  onSelectionChange: (selection: RowSelectionState) => void;
};

export const List: FC<Props> = ({
  data,
  rowSelection,
  loading,
  setRowSelection,
  onSelectPaymentLine: onSelectPayment,
  onValueChange,
  onSelectionChange,
}) => {
  const currency = useNumber().currency;
  const t = useTranslation().trans;
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({});
  const totals = calculateTotals(data);
  const columns = getColumns({ onDidClick: onSelectPayment, t });

  const table = useReactTable({
    data: data,
    columns,
    getRowId: (row: ReceivableInvoiceForm) => row.id.toString(),
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: (updater) => {
      const nextSelection = typeof updater === 'function' ? updater(rowSelection) : updater;
      onSelectionChange(nextSelection);
      setRowSelection(nextSelection);
    },
    meta: {
      updateData: (inputId: string, rowIndex: number, columnId: string, value: string) => {
        onValueChange?.(inputId, value);
      },
    },
    state: {
      sorting,
      columnFilters,
      columnVisibility,
      rowSelection,
    },
  });

  return (
    <div>
      <div className="flex items-center py-4">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="outline" className="ml-auto">
              {t('global.columns')} <ChevronDown />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            {table
              .getAllColumns()
              .filter((column) => column.getCanHide())
              .map((column) => {
                return (
                  <DropdownMenuCheckboxItem
                    key={column.id}
                    className="capitalize"
                    checked={column.getIsVisible()}
                    onCheckedChange={(value) => column.toggleVisibility(!!value)}
                  >
                    {(column.columnDef.meta as string) || column.id}
                  </DropdownMenuCheckboxItem>
                );
              })}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
      <div className="rounded-md border [&_[data-type=number]]:text-right">
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
            {loading ? (
              Array.from({ length: 5 }).map((_, i) => <SkeletonRow key={i} columns={columns.length} />)
            ) : table.getRowModel().rows?.length ? (
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
          <tfoot>
            <TableRow>
              <TableCell colSpan={columns.length - 4} />
              <TableCell className="pr-4 text-end text-base font-semibold">{currency(totals.totalPayment)}</TableCell>
              <TableCell className="pr-4 text-end text-base font-semibold">{currency(totals.totalDiscount)}</TableCell>
              <TableCell className="pr-4 text-end text-base font-semibold">{currency(totals.totalRemaining)}</TableCell>
            </TableRow>
          </tfoot>
        </Table>
      </div>
      <div className="flex items-center justify-end space-x-2 py-4">
        <div className="text-muted-foreground flex-1 text-sm">
          {t('global.pagination.selectedOf', {
            rows: table.getFilteredSelectedRowModel().rows.length,
            total: table.getFilteredRowModel().rows.length,
          })}
        </div>
        <div className="space-x-2">
          <Button value="outline" size="sm" onClick={() => table.previousPage()} disabled={!table.getCanPreviousPage()}>
            {t('global.pagination.previous')}
          </Button>
          <Button value="outline" size="sm" onClick={() => table.nextPage()} disabled={!table.getCanNextPage()}>
            {t('global.pagination.next')}
          </Button>
        </div>
      </div>
    </div>
  );
};
