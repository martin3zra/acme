import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
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
import { TextCell } from '@/components/data-table/text-cell';
import { Payable, PayableVerb, Replacements } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (item: Payable, action: PayableVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ onDidClick, t }: Props): ColumnDef<Payable>[] => {
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
      cell: ({ row }) => (
        <Checkbox checked={row.getIsSelected()} onCheckedChange={(value) => row.toggleSelected(!!value)} aria-label="Select row" />
      ),
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'invoice_number',
      meta: t('global.number'),
      header: (props) => <HeaderCell title={t('global.number')} alignment="left" columnWidth={props.column.getSize()} />,
      cell: (props) => <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      accessorKey: 'invoice_date',
      meta: t('global.date'),
      header: (props) => <HeaderCell title={t('global.date')} alignment="left" columnWidth={props.column.getSize()} />,
      cell: (props) => <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      accessorKey: 'due_date',
      meta: t('global.dueDate'),
      header: (props) => <HeaderCell title={t('global.dueDate')} alignment="left" columnWidth={props.column.getSize()} />,
      cell: (props) => {
        const dueDate = props.getValue() as string;
        const isOverdue = new Date(dueDate) < new Date() && props.row.original.paid_status !== 'paid' && props.row.original.status !== 'void';
        return (
          <span className={isOverdue ? 'text-red-600 font-medium' : ''}>
            <DateCell columnWidth={props.column.getSize()} value={dueDate} />
          </span>
        );
      },
    },
    {
      accessorKey: 'amount_payable',
      meta: t('global.amount'),
      header: (props) => <HeaderCell title={t('global.amount')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      accessorKey: 'amount_paid',
      meta: t('global.paid'),
      header: (props) => <HeaderCell title={t('global.paid')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      id: 'balance',
      meta: t('global.balance'),
      header: (props) => <HeaderCell title={t('global.balance')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => {
        const balance = props.row.original.amount_payable - props.row.original.amount_paid;
        return <CurrencyCell columnWidth={props.column.getSize()} value={balance as unknown as string} />;
      },
    },
    {
      accessorKey: 'status',
      meta: t('global.status'),
      size: 90,
      header: (props) => <HeaderCell title={t('global.status')} alignment="center" columnWidth={props.column.getSize()} />,
      cell: (props) => <StatusBadge type="payable" status={props.row.original.status} />,
    },
    {
      accessorKey: 'paid_status',
      meta: t('global.paymentStatus'),
      size: 90,
      header: (props) => <HeaderCell title={t('global.paymentStatus')} alignment="center" columnWidth={props.column.getSize()} />,
      cell: (props) => <StatusBadge type="paid" status={props.row.original.paid_status} />,
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: (props) => {
        const canVoid = props.row.original.status !== 'void' && props.row.original.paid_status !== 'paid';
        return (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="size-8 p-0">
                <span className="sr-only">{t('global.openMenu')}</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="**:data-[slot=dropdown-menu-item]:cursor-pointer">
              <DropdownMenuLabel>{t('global.actions.title')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>
                {t('payables.viewPayable.title')}
              </DropdownMenuItem>
              {canVoid && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'void')}>
                    {t('payables.voidPayable.title')}
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
