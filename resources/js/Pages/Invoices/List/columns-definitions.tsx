import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
import { TextCell } from '@/components/data-table/text-cell';
import { StatusBadge } from '@/components/status-badge';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { DiscountType, Invoice, InvoiceVerb } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (item: Invoice, action: InvoiceVerb) => void;
};

export const getColumns = ({ onDidClick }: Props): ColumnDef<Invoice>[] => {
  return [
    {
      id: 'select',
      header: ({ table }) => (
        <Checkbox
          checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && 'indeterminate')}
          onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
          aria-label="Select all"
        />
      ),
      cell: ({ row }) => <Checkbox checked={row.getIsSelected()} onCheckedChange={(value) => row.toggleSelected(!!value)} aria-label="Select row" />,
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'number',
      meta: 'ID',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="ID" alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'ncf',
      meta: 'NCF',
      enableHiding: true,
      // size: 880,
      header: (props) => {
        return <HeaderCell title="NCF" alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'customer.name',
      id: 'customer.name',
      header: ({ column }) => {
        return (
          <Button variant="ghost" onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}>
            Customer <ArrowUpDown />
          </Button>
        );
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'date',
      meta: 'Date',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Date" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'amount',
      meta: 'Amount',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Amount" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'discount',
      meta: 'Discount',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Discount" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        const discount = props.row.getValue('discount') as DiscountType;
        const suffix = discount.type === 'percentage' ? '%' : undefined;
        return <CurrencyCell columnWidth={props.column.getSize()} suffix={suffix} value={String(discount.value)} />;
      },
    },
    {
      accessorKey: 'tax',
      meta: 'Tax',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Tax" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'total',
      meta: 'Total',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Total" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'amount_due',
      meta: 'Balance',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Balance" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'status',
      size: 70,
      header: (props) => {
        return <HeaderCell title="Status" alignment="center" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <StatusBadge type="invoice" status={props.row.original.status} />;
      },
    },
    {
      accessorKey: 'paid_status',
      size: 70,
      header: (props) => {
        return <HeaderCell title="Paid Status" alignment="center" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <StatusBadge type="paid" status={props.row.original.paid_status} />;
      },
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: (props) => {
        const disabled = props.row.original.status === 'void';
        const canRecordPayment = props.row.original.paid_status === 'unpaid' || props.row.original.paid_status === 'partial';
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="size-8 p-0">
                <span className="sr-only">Open menu</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>Actions</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>View</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')} disabled={disabled}>
                Edit
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              {canRecordPayment && (
                <>
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'record-payment')}>Record payment</DropdownMenuItem>
                  <DropdownMenuSeparator />
                </>
              )}
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'void')} disabled={disabled}>
                Void
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
