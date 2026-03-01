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
import { Expense, Replacements, Verb } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (item: Expense, action: Verb) => void;
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
      accessorKey: 'category.name',
      meta: t('global.category'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.category')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return (
          <div className="[&_[data-slot=has-notes]]:-px-6 relative flex">
            <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />
          </div>
        );
      },
    },
    {
      accessorKey: 'notes',
      meta: t('global.notes'),
      // size: 880,
      header: (props) => {
        return <HeaderCell title={t('global.notes')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return (
          <div className="[&_[data-slot=has-notes]]:-px-6 relative flex">
            <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />
          </div>
        );
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
    {
      id: 'actions',
      enableHiding: false,
      cell: (props) => {
        const isDeleted = Boolean(props.row.original.deleted_at);
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
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>{t('expenses.viewExpense.title')}</DropdownMenuItem>
              {!isDeleted && (
                <>
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')} disabled={isDeleted}>
                    {t('expenses.editExpense.title')}
                  </DropdownMenuItem>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'trash')} disabled={isDeleted}>
                    {t('expenses.trashExpense.title')}
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
