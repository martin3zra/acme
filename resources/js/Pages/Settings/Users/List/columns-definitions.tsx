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
import { Replacements, User, UserVerb } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { BadgeCheck, Link, MoreHorizontal } from 'lucide-react';

type Props = {
  onDidClick: (user: User, action: UserVerb) => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ onDidClick, t }: Props): ColumnDef<User>[] => {
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
      accessorKey: 'name',
      meta: t('global.name'),
      header: ({ column }) => {
        return <HeaderSortCell<User> title={t('global.name')} column={column} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'email',
      meta: t('global.email'),
      // size: 70,
      header: (props) => {
        return <HeaderCell title={t('global.email')} alignment="center" columnWidth={props.column.getSize()} />;
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
      accessorKey: 'email_verified_at',
      meta: t('global.verified'),
      size: 70,
      header: (props) => {
        return <HeaderCell title={t('global.verified')} alignment="center" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        if (props.row.original.email_verified_at !== null) {
          return (
            <div className="flex justify-center">
              <BadgeCheck size={22} />
            </div>
          );
        }

        return null;
      },
    },
    {
      accessorKey: 'linked',
      meta: t('global.link'),
      size: 70,
      header: (props) => {
        return <HeaderCell title={t('global.link')} alignment="center" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        if (props.row.original.linked === 0) {
          return null;
        }

        return (
          <div className="flex justify-center">
            <Link size={22} />
          </div>
        );
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
            <DropdownMenuContent align="end" className="[&_[data-slot=dropdown-menu-item]]:cursor-pointer">
              <DropdownMenuLabel>{t('global.actions.title')}</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view')}>{t('users.viewUser.title')}</DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'trash')}>{t('users.trashUser.title')}</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'permission')}>{t('users.permissionUser.title')}</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
