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
import { Replacements, Verb, Warehouse } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MoreHorizontal, Warehouse as WarehouseIcon } from 'lucide-react';

type Props = {
  onDidClick: (warehouse: Warehouse, action: Verb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ onDidClick, t }: Props): ColumnDef<Warehouse>[] => {
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
            <WarehouseIcon className="size-4" />
          </div>
        );
      },
      enableSorting: false,
      enableHiding: false,
    },
    {
      accessorKey: 'name',
      meta: t('global.name'),
      size: 320,
      header: ({ column }) => {
        return <HeaderSortCell<Warehouse> title={t('global.name')} column={column} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'location',
      meta: t('global.location'),
      header: (props) => {
        return <HeaderCell title={t('global.location')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={(props.getValue() as string) || ''} />;
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
      accessorKey: 'created_at',
      meta: t('global.addedAt'),
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
            <DropdownMenuContent align="end" className="[&_[data-slot=dropdown-menu-item]]:cursor-pointer">
              <DropdownMenuLabel>{t('global.actions.title')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>{t('warehouses.viewWarehouse.title')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')}>{t('warehouses.editWarehouse.title')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'trash')}>{t('warehouses.trashWarehouse.title')}</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
