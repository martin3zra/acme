import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
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
import { Customer, CustomerVerb, Replacements } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { ArrowUpDown, MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (customer: Customer, action: CustomerVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ onDidClick, t }: Props): ColumnDef<Customer>[] => {
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
      accessorKey: 'code',
      meta: t('global.code'),
      header: (props) => {
        return <HeaderCell title={t('global.code')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'name',
      meta: t('global.name'),
      header: ({ column }) => {
        return (
          <Button className="uppercase" variant="ghost" onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}>
            {t('global.name')} <ArrowUpDown />
          </Button>
        );
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'contact_name',
      meta: t('global.contact'),
      header: (props) => {
        return <HeaderCell title={t('global.contact')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'phone',
      meta: t('global.phone'),
      header: (props) => {
        return <HeaderCell title={t('global.phone')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'email',
      meta: t('global.email'),
      header: ({ column }) => {
        return (
          <Button className="uppercase" variant="ghost" onClick={() => column.toggleSorting(column.getIsSorted() === 'asc')}>
            {t('global.email')} <ArrowUpDown />
          </Button>
        );
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
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>{t('global.actions.title')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>{t('customers.viewCustomer.title')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'edit')}>{t('customers.editCustomer.title')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'trash')}>{t('customers.trashCustomer.title')}</DropdownMenuItem>
              {props.row.original.amount_due > 0 && (
                <>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'record-payment')} disabled={props.row.original.amount_due === 0}>
                    {t('global.actions.makePayment')}
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
