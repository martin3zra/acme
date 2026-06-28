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
import { Payable, PayableForm, PayableVerb, Replacements, VendorPaymentTotals } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MoreHorizontal } from 'lucide-react';

type Props = {
  totals: VendorPaymentTotals;
  onDidClick: (item: PayableForm, action: PayableVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

// We join Payable + PayableForm for the row data
export type PayableLineRow = Payable & PayableForm;

export const getColumns = ({ totals, onDidClick, t }: Props): ColumnDef<PayableLineRow, string | number>[] => {
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
      footer: () => null,
    },
    {
      accessorKey: 'invoice_number',
      meta: t('payables.single.bill'),
      size: 140,
      header: ({ column }) => <HeaderCell title={t('payables.single.bill')} alignment="left" columnWidth={column.getSize()} />,
      cell: (props) => <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
      footer: () => null,
    },
    {
      accessorKey: 'invoice_date',
      meta: t('global.date'),
      size: 100,
      header: (props) => <HeaderCell title={t('global.date')} alignment="left" columnWidth={props.column.getSize()} />,
      cell: (props) => <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
      footer: () => null,
    },
    {
      accessorKey: 'due_date',
      meta: t('global.dueDate'),
      header: (props) => <HeaderCell title={t('global.dueDate')} alignment="left" columnWidth={props.column.getSize()} />,
      cell: (props) => {
        const dueDate = props.getValue() as string;
        const isOverdue = new Date(dueDate) < new Date();
        return (
          <span className={isOverdue ? 'text-red-600 font-medium' : ''}>
            <DateCell columnWidth={props.column.getSize()} value={dueDate} />
          </span>
        );
      },
      footer: () => null,
    },
    {
      accessorKey: 'amount_payable',
      meta: t('global.amount'),
      header: (props) => <HeaderCell title={t('global.amount')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
      footer: () => null,
    },
    {
      accessorKey: 'amount_due',
      meta: t('global.amount_due'),
      header: (props) => <HeaderCell title={t('global.amount_due')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={props.row.original.amount_due as unknown as string} />,
      footer: () => null,
    },
    {
      accessorKey: 'payment',
      meta: t('global.payment'),
      header: (props) => <HeaderCell title={t('global.payment')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => <EditableCell {...props} identifier={props.row.original.invoice_uuid} inputType="number" />,
      footer: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={totals.totalPayment as unknown as string} />,
    },
    {
      id: 'balance_after',
      meta: t('global.remainingBalance'),
      header: (props) => <HeaderCell title={t('global.remainingBalance')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => {
        const remaining = props.row.original.amount_due - (props.row.original.payment || 0);
        return <CurrencyCell columnWidth={props.column.getSize()} value={remaining as unknown as string} />;
      },
      footer: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={totals.totalRemaining as unknown as string} />,
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: (props) => (
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
            <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'void')}>
              {t('global.actions.delete')}
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      ),
      footer: () => null,
    },
  ];
};
