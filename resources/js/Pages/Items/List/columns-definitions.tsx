import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
import { HeaderSortCell } from '@/components/data-table/header-sort-cell';
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
import { Item, Replacements, Verb } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MoreHorizontal, Package, Wrench } from 'lucide-react';

type Props = {
  onDidClick: (item: Item, action: Verb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ onDidClick, t }: Props): ColumnDef<Item>[] => {
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
      cell: (props) => {
        return (
          <div className="flex items-center space-x-2">
            <Checkbox checked={props.row.getIsSelected()} onCheckedChange={(value) => props.row.toggleSelected(!!value)} aria-label="Select row" />
            {props.row.original.item_type === 'service' ? <Wrench className="size-4" /> : <Package className="size-4" />}
          </div>
        );
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      id: 'identifiers.reference',
      accessorFn: (row) => row.identifiers?.reference ?? '', // prevent undefined
      meta: t('global.reference'),
      size: 100,
      header: (props) => {
        return <HeaderCell title={t('global.reference')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'name',
      meta: t('global.name'),
      size: 320,
      header: ({ column }) => {
        return <HeaderSortCell<Item> title={t('global.name')} column={column} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      id: 'identifiers.code',
      accessorFn: (row) => row.identifiers?.code ?? '', // prevent undefined
      meta: t('global.code'),
      header: (props) => {
        return <HeaderCell title={t('global.code')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      id: 'identifiers.sku',
      accessorFn: (row) => row.identifiers?.sku ?? '', // prevent undefined
      meta: t('global.sku'),
      header: (props) => {
        return <HeaderCell title={t('global.sku')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      id: 'identifiers.barcode',
      accessorFn: (row) => row.identifiers?.barcode ?? '', // prevent undefined
      meta: t('global.barcode'),
      header: (props) => {
        return <HeaderCell title={t('global.barcode')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      id: 'identifiers.vendor_reference',
      accessorFn: (row) => row.identifiers?.vendor_reference ?? '', // prevent undefined
      meta: t('items.single.vendor_reference'),
      header: (props) => {
        return <HeaderCell title={t('items.single.vendor_reference')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'description',
      meta: t('global.description'),
      header: (props) => {
        return <HeaderCell title={t('global.description')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'unit.name',
      meta: t('global.unit'),
      size: 84,
      header: (props) => {
        return <HeaderCell title={t('global.unit')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'tax.name',
      meta: t('global.taxRate'),
      size: 105,
      header: (props) => {
        return <HeaderCell title={t('global.taxRate')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'status',
      meta: t('global.status'),
      size: 70,
      header: (props) => {
        return <HeaderCell title={t('global.status')} alignment="center" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <StatusBadge type="status" status={props.row.original.status} />;
      },
    },
    {
      accessorKey: 'price',
      meta: t('global.price'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.price')} alignment="right" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <CurrencyCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'created_at',
      meta: t('global.addedAt'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.addedAt')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
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
            <DropdownMenuContent align="end" className="**:data-[slot=dropdown-menu-item]:cursor-pointer">
              <DropdownMenuLabel>{t('global.actions.title')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>{t('items.viewItem.title')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')}>{t('items.editItem.title')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'trash')}>{t('items.trashItem.title')}</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
