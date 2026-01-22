import { Button } from '@/components/ui/button';
import { DropdownMenu, DropdownMenuCheckboxItem, DropdownMenuContent, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useTranslation } from '@/hooks/use-translation';
import { Invoice, InvoiceTypeFilter, InvoiceVerb, TransactionKind } from '@/types';
import {
  ColumnFiltersState,
  flexRender,
  getCoreRowModel,
  getFilteredRowModel,
  getPaginationRowModel,
  getSortedRowModel,
  SortingState,
  useReactTable,
  VisibilityState,
} from '@tanstack/react-table';
import { ChevronDown } from 'lucide-react';
import { FC, useState } from 'react';
import { getColumns } from './columns-definitions';

type Props = {
  kind: TransactionKind;
  data: Invoice[];
  onSelectInvoice: (invoice: Invoice, action: InvoiceVerb) => void;
  currentInvoiceTypeFilter: InvoiceTypeFilter;
  onInvoiceTypeFilterChanges: (value: InvoiceTypeFilter) => void;
};

const hiddenColumns: string[] = ['paid_status', 'amount_due', 'ncf'];
export const List: FC<Props> = ({ kind, data, currentInvoiceTypeFilter, onSelectInvoice, onInvoiceTypeFilterChanges }) => {
  const isInvoice = kind === 'invoice';
  const t = useTranslation().trans;
  const [sorting, setSorting] = useState<SortingState>([]);
  const [columnFilters, setColumnFilters] = useState<ColumnFiltersState>([]);
  const [columnVisibility, setColumnVisibility] = useState<VisibilityState>({
    ncf: false,
    amount_due: isInvoice,
    paid_status: isInvoice,
  });
  const [rowSelection, setRowSelection] = useState({});

  const columns = getColumns({ kind, onDidClick: onSelectInvoice, t });
  const table = useReactTable({
    data,
    columns,
    onSortingChange: setSorting,
    onColumnFiltersChange: setColumnFilters,
    getCoreRowModel: getCoreRowModel(),
    getPaginationRowModel: getPaginationRowModel(),
    getSortedRowModel: getSortedRowModel(),
    getFilteredRowModel: getFilteredRowModel(),
    onColumnVisibilityChange: setColumnVisibility,
    onRowSelectionChange: setRowSelection,
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
        {/* <Input
          placeholder="Filter names..."
          value={(table.getColumn('customer.name')?.getFilterValue() as string) ?? ''}
          onChange={(event) => table.getColumn('customer.name')?.setFilterValue(event.target.value)}
          className="max-w-sm"
        /> */}
        {kind === 'invoice' && (
          <Tabs
            value={currentInvoiceTypeFilter}
            onValueChange={(value: string) => onInvoiceTypeFilterChanges(value as InvoiceTypeFilter)}
            className="[&_[data-slot=tabs-trigger]:not([data-state=active])]:cursor-pointer"
          >
            <TabsList>
              <TabsTrigger value="all">{t('global.all')}</TabsTrigger>
              <TabsTrigger value="credit">{t('global.credit')}</TabsTrigger>
              <TabsTrigger value="cash">{t('global.cash')}</TabsTrigger>
            </TabsList>
          </Tabs>
        )}
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
              .filter((column) => {
                if (kind === 'invoice') {
                  return true;
                }
                return !hiddenColumns.includes(column.id);
              })
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
