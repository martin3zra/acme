import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
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
import { Payment, PaymentVerb } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (item: Payment, action: PaymentVerb) => void;
};

export const getColumns = ({ onDidClick }: Props): ColumnDef<Payment>[] => {
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
      accessorKey: 'id',
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
      accessorKey: 'customer.amount_due',
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
      id: 'actions',
      enableHiding: false,
      cell: (props) => {
        const disabled = false; // props.row.original.status === 'void';
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
