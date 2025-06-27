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
import { PaymentVerb, ReceivableInvoiceForm, Replacements } from '@/types';
import { Link } from '@inertiajs/react';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (item: ReceivableInvoiceForm, action: PaymentVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ onDidClick, t }: Props): ColumnDef<ReceivableInvoiceForm, string | number>[] => {
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
      meta: t('payments.single.invoice'),
      header: ({ column }) => {
        return (
          <Button className="uppercase" variant="ghost" onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}>
            {t('payments.single.invoice')} <ArrowUpDown />
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
      accessorKey: 'due_on',
      meta: t('global.dueDate'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.dueDate')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'total',
      meta: t('global.amount'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.amount')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'balance',
      meta: t('global.balance'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.balance')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={props.row.original.amount_due as unknown as string} />,
    },
    {
      accessorKey: 'payment',
      meta: t('global.payment'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.payment')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <EditableCell {...props} identifier={props.row.original.uuid} inputType="number" />;
      },
    },
    {
      accessorKey: 'discount',
      meta: t('global.discount'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.discount')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'remaining',
      meta: t('global.balance'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.balance')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        const remaining = props.row.original.amount_due - props.row.original.payment;
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
                <span className="sr-only">{t('global.openMenu')}</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="[&_[data-slot=dropdown-menu-item]]:cursor-pointer">
              <DropdownMenuLabel>{t('global.actions.title')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>
                <Link href={`/invoices?id=${props.row.original.uuid}`}>{t('payments.single.viewInvoice')}</Link>
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'trash')}>{t('global.actions.delete')}</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
