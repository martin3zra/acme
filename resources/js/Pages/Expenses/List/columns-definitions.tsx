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
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from '@/components/ui/tooltip';
import { Expense, PaymentVerb, Replacements } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MessageCircleMore, MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (item: Expense, action: PaymentVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ onDidClick, t }: Props): ColumnDef<Expense>[] => {
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
      meta: t('global.number'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.number')} alignment="left" columnWidth={props.column.getSize()} />;
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
                    className="absolute inset-0 -top-0 left-[90%] hidden size-5 -translate-x-1/2 -translate-y-1/2 transform cursor-pointer"
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
    // {
    //   accessorKey: 'customer.name',
    //   id: 'customer.name',
    //   meta: t('global.customer'),
    //   header: ({ column }) => {
    //     return <HeaderSortCell<Payment> title={t('global.customer')} column={column} />;
    //   },
    //   cell: (props) => {
    //     return (
    //       <LinkCell
    //         href={`/customers?id=${props.row.original.customer.uuid}`}
    //         columnWidth={props.column.getSize()}
    //         value={props.getValue() as string}
    //       />
    //     );
    //   },
    // },
    // {
    //   accessorKey: 'customer.amount_due',
    //   meta: t('global.balance'),
    //   // size: 880,
    //   header: (props) => {
    //     return <HeaderCell title={t('global.balance')} alignment="right" columnWidth={props.column.getSize()} />;
    //   },
    //   cell: (props) => {
    //     return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
    //   },
    // },
    {
      accessorKey: 'date',
      meta: t('global.date'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.date')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'amount',
      meta: t('global.amount'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.amount')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    // {
    //   accessorKey: 'invoices',
    //   meta: t('global.navMain.invoices'),
    //   // size: 880,
    //   header: (props) => {
    //     return <HeaderCell title={t('global.navMain.invoices')} alignment="right" columnWidth={props.column.getSize()} />;
    //   },
    //   cell: (props) => {
    //     return <NumericCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
    //   },
    // },
    // {
    //   accessorKey: 'status',
    //   meta: t('global.status'),
    //   size: 70,
    //   header: (props) => {
    //     return <HeaderCell title={t('global.status')} alignment="center" columnWidth={props.column.getSize()} />;
    //   },
    //   cell: (props) => {
    //     return <StatusBadge type="payment" status={props.row.original.status} />;
    //   },
    // },
    {
      id: 'actions',
      enableHiding: false,
      cell: (props) => {
        const disabled = false; // props.row.original.status === 'void';
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="size-8 p-0">
                <span className="sr-only">{t('global.openMenu')}</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="[&_[data-slot=dropdown-menu-item]]:cursor-pointer">
              <DropdownMenuLabel>{t('global.actions.title')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>{t('payments.viewPayment.title')}</DropdownMenuItem>
              {/* {props.row.original.status !== 'void' && (
                <>
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')} disabled={disabled}>
                    {t('payments.editPayment.title')}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'void')} disabled={disabled}>
                    {t('payments.voidPayment.title')}
                  </DropdownMenuItem>
                </>
              )} */}
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
