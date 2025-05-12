import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
import { LinkCell } from '@/components/data-table/link-cell';
import { NumericCell } from '@/components/data-table/numeric-cell';
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
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Payment, PaymentVerb } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, MessageCircleMore, MoreHorizontal } from 'lucide-react';

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
      accessorKey: 'number',
      meta: 'Number',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Number" alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        const hasNotes = !!props.row.original.notes;
        return (
          <div className="[&_[data-slot=has-notes]]:-px-6 relative flex [&_[data-slot=has-notes]]:block [&_[data-slot=has-notes]]:text-red-500">
            <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />
            <TooltipProvider>
              <Tooltip>
                <TooltipTrigger>
                  <MessageCircleMore
                    className="absolute inset-0 -top-0 left-[56%] hidden size-5 -translate-x-1/2 -translate-y-1/2 transform cursor-pointer"
                    data-slot={hasNotes ? 'has-notes' : 'default'}
                  />
                </TooltipTrigger>
                <TooltipContent>{props.row.original.notes}</TooltipContent>
              </Tooltip>
            </TooltipProvider>
          </div>
        );
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
        return (
          <LinkCell
            href={`/customers?id=${props.row.original.customer.uuid}`}
            columnWidth={props.column.getSize()}
            value={props.getValue() as string}
          />
        );
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
        return <HeaderCell title="Date" alignment="left" columnWidth={props.column.getSize()} />;
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
      accessorKey: 'invoices',
      meta: 'Invoices',
      // size: 880,
      header: (props) => {
        return <HeaderCell title="Invoices" alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <NumericCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'status',
      size: 70,
      header: (props) => {
        return <HeaderCell title="Status" alignment="center" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <StatusBadge type="payment" status={props.row.original.status} />;
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
              {props.row.original.status !== 'void' && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'void')} disabled={disabled}>
                    Void
                  </DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
