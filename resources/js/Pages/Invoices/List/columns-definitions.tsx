import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
import { LinkCell } from '@/components/data-table/link-cell';
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
import { DiscountType, Invoice, InvoiceVerb, Replacements } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (item: Invoice, action: InvoiceVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ onDidClick, t }: Props): ColumnDef<Invoice>[] => {
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
      meta: t('global.number'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.number')} alignment="left" columnWidth={props.column.getSize()} />;
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
      meta: t('customers.customer'),
      size: 200,
      header: ({ column }) => {
        return (
          <Button className="font-semibold uppercase" variant="ghost" onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}>
            {t('global.customer')} <ArrowUpDown />
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
      accessorKey: 'date',
      meta: t('global.date'),
      size: 100,
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
      size: 100,
      header: (props) => {
        return <HeaderCell title={t('global.amount')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
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
        const discount = props.row.getValue('discount') as DiscountType;
        const suffix = discount.type === 'percentage' ? '%' : undefined;
        return <CurrencyCell columnWidth={props.column.getSize()} suffix={suffix} value={String(discount.value)} />;
      },
    },
    {
      accessorKey: 'tax',
      meta: t('global.tax'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.tax')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'total',
      meta: t('global.total'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.total')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'amount_due',
      meta: t('global.balance'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.discount')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'status',
      size: 70,
      header: (props) => {
        return <HeaderCell title={t('global.status')} alignment="center" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <StatusBadge type="invoice" status={props.row.original.status} />;
      },
    },
    {
      accessorKey: 'paid_status',
      size: 70,
      header: (props) => {
        return <HeaderCell title={t('invoices.paidStatus')} alignment="center" columnWidth={props.column.getSize()} />;
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
                <span className="sr-only">{t('global.openMenu')}</span>
                <MoreHorizontal />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>{t('global.actions.title')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>{t('global.actions.view')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')} disabled={disabled}>
                {t('global.actions.edit')}
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              {canRecordPayment && (
                <>
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'record-payment')}>
                    {t('global.actions.recordPayment')}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                </>
              )}
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'void')} disabled={disabled}>
                {t('global.actions.void')}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
