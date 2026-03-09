import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
import { HeaderSortCell } from '@/components/data-table/header-sort-cell';
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
import { capitalize } from '@/lib/utils';
import { DiscountType, Invoice, InvoiceVerb, Replacements, TransactionKind } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MoreHorizontal } from 'lucide-react';
import { ConvertToInvoiceAction } from '../Shared/convert-to-invoice-action';
import { SourceIcon } from '../Shared/source-icon';

type Props = {
  kind: TransactionKind;
  onDidClick: (item: Invoice, action: InvoiceVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ kind, onDidClick, t }: Props): ColumnDef<Invoice>[] => {
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
      cell: ({ row }) => {
        const source = row.original.source;
        return (
          <div className="flex space-x-3">
            <Checkbox checked={row.getIsSelected()} onCheckedChange={(value) => row.toggleSelected(!!value)} aria-label="Select row" />
            {source && <SourceIcon source={source} kind={kind} />}
          </div>
        );
      },
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
      meta: t('global.customer'),
      minSize: 200,
      maxSize: 260,
      header: ({ column }) => {
        return <HeaderSortCell<Invoice> title={t('global.customer')} column={column} />;
      },
      cell: (props) => {
        const customerName = String(props.getValue() ?? '');

        return (
          <div title={customerName} className="min-w-0">
            <LinkCell href={`/customers?id=${props.row.original.customer.uuid}`} columnWidth={props.column.getSize()} value={customerName} />
          </div>
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
        return <HeaderCell title={t('global.balance')} alignment="right" columnWidth={props.column.getSize()} />;
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
        return <StatusBadge kind={kind} type="invoice" status={props.row.original.status} />;
      },
    },
    {
      accessorKey: 'paid_status',
      size: 70,
      meta: t(`${kind}s.paidStatus`),
      header: (props) => {
        return <HeaderCell title={t(`${kind}s.paidStatus`)} alignment="center" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <StatusBadge type="paid" status={props.row.original.paid_status} />;
      },
    },
    {
      id: 'actions',
      enableHiding: false,
      cell: (props) => {
        const disabled = props.row.original.status === 'void' || props.row.original.status === 'closed';
        const canRecordPayment = (kind === 'invoice' && props.row.original.paid_status === 'unpaid') || props.row.original.paid_status === 'partial';
        const canBeVoid = kind === 'invoice';
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
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>
                {t(`${kind}s.view${capitalize(kind)}.title`)}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')} disabled={disabled}>
                {t(`${kind}s.edit${capitalize(kind)}.title`)}
              </DropdownMenuItem>
              {kind === 'invoice' && props.row.original.status !== 'void' && (
                <>
                  <ConvertToInvoiceAction mode="duplicate" id={props.row.original.uuid} title={t('global.duplicateInvoice')} kind={kind} />
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'mark-as-recurrent')}>
                    {t('global.actions.markAsRecurrent')}
                  </DropdownMenuItem>
                </>
              )}
              {canRecordPayment && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'record-payment')}>
                    {t('global.actions.recordPayment')}
                  </DropdownMenuItem>
                </>
              )}
              {canBeVoid && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'void')} disabled={disabled}>
                    {t(`${kind}s.void${capitalize(kind)}.title`)}
                  </DropdownMenuItem>
                </>
              )}
              {(kind === 'estimate' || kind === 'order') && props.row.original.status !== 'closed' && props.row.original.status !== 'void' && (
                <>
                  <DropdownMenuSeparator />
                  <ConvertToInvoiceAction id={props.row.original.uuid} title={t('global.convertToInvoice')} kind={kind} />
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
