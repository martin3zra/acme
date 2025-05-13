import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { EditableCell } from '@/components/data-table/editable-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
import { TextCell } from '@/components/data-table/text-cell';
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
import { PaymentVerb, ReceivableInvoiceForm } from '@/types';
import { Link } from '@inertiajs/react';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (item: ReceivableInvoiceForm, action: PaymentVerb) => void;
};

export const getColumns = ({ onDidClick }: Props): ColumnDef<ReceivableInvoiceForm, string | number>[] => {
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
      id: 'number',
      header: ({ column }) => {
        return (
          <Button variant="ghost" onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}>
            Invoice <ArrowUpDown />
          </Button>
        );
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'ncf',
      meta: 'NCF',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="NCF" alignment="left" columnWidth={props.column.getSize()} />;
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
        return <HeaderCell title="Date" alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'due_on',
      meta: 'Due On',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Due On" alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'total',
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
      accessorKey: 'balance',
      meta: 'Balance',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Balance" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.row.original.amount_due as unknown as string} />;
      },
    },
    {
      accessorKey: 'payment',
      meta: 'Payment',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Payment" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <EditableCell {...props} identifier={props.row.original.uuid} inputType="number" />;
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
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'remaining',
      meta: 'Remaining',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Remaining" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        let remaining = props.row.original.amount_due - props.row.original.payment;
        // if (props.row.original.action !== undefined) {
        //   remaining = props.row.original.amount_due;
        // }
        // if (props.row.original.action === 'updated') {
        //   remaining = props.row.original.amount_due - props.row.original.payment;
        // }
        return <CurrencyCell columnWidth={props.column.getSize()} value={remaining as unknown as string} />;
      },
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: (props) => {
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
              <DropdownMenuItem>
                <Link href={`/invoices?id=${props.row.original.uuid}`}>View</Link>
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')}>Edit</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
