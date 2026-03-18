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
import type { Purchase, PurchaseTransactionKind, Replacements } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MoreHorizontal } from 'lucide-react';
import { ConvertToReceiptAction } from '../Shared/convert-to-receipt-action';
import { PurchaseSourceIcon } from '../Shared/purchase-source-icon';

export type PurchaseVerb = 'view' | 'edit' | 'delete';

type Props = {
  kind: PurchaseTransactionKind;
  onDidClick: (item: Purchase, action: PurchaseVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ kind, onDidClick, t }: Props): ColumnDef<Purchase>[] => {
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
            {source && <PurchaseSourceIcon source={source} kind={kind} />}
          </div>
        );
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'number',
      meta: t('global.number'),
      header: (props) => <HeaderCell title={t('global.number')} alignment="left" columnWidth={props.column.getSize()} />,
      cell: (props) => <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      accessorKey: 'vendor.name',
      id: 'vendor.name',
      meta: t('global.vendor'),
      size: 200,
      header: ({ column }) => <HeaderSortCell<Purchase> title={t('global.vendor')} column={column} />,
      cell: (props) => (
        <LinkCell href={`/vendors?id=${props.row.original.vendor.uuid}`} columnWidth={props.column.getSize()} value={props.getValue() as string} />
      ),
    },
    {
      accessorKey: 'date',
      meta: t('global.date'),
      size: 110,
      header: (props) => <HeaderCell title={t('global.date')} alignment="left" columnWidth={props.column.getSize()} />,
      cell: (props) => <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      accessorKey: 'amount',
      meta: t('global.amount'),
      size: 110,
      header: (props) => <HeaderCell title={t('global.amount')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      accessorKey: 'total',
      meta: t('global.total'),
      size: 110,
      header: (props) => <HeaderCell title={t('global.total')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      accessorKey: 'amount_due',
      meta: t('global.balance'),
      size: 110,
      header: (props) => <HeaderCell title={t('global.balance')} alignment="right" columnWidth={props.column.getSize()} />,
      cell: (props) => <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />,
    },
    {
      accessorKey: 'payment_status',
      meta: t('global.payment'),
      size: 100,
      header: (props) => <HeaderCell title={t('global.payment')} alignment="center" columnWidth={props.column.getSize()} />,
      cell: (props) => {
        const status = props.getValue() as string;
        return <StatusBadge type="paid" status={status} />;
      },
    },
    {
      accessorKey: 'status',
      size: 70,
      header: (props) => <HeaderCell title={t('global.status')} alignment="center" columnWidth={props.column.getSize()} />,
      cell: (props) => <StatusBadge kind={kind} type="purchase" status={props.row.original.status} />,
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
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>{t('global.actions.view')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')}>{t('global.actions.edit')}</DropdownMenuItem>

              {props.row.original.transaction_kind === 'purchase_order' && (
                <>
                  <DropdownMenuSeparator />
                  <ConvertToReceiptAction id={props.row.original.uuid} title={t('global.convertToReceipt')} />
                </>
              )}

              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'delete')}>{t('global.actions.delete')}</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
