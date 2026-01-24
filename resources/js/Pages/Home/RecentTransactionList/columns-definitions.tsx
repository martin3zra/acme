import { CurrencyCell } from '@/components/data-table/currency-cell';
import { DateCell } from '@/components/data-table/date-cell';
import { HeaderCell } from '@/components/data-table/header-cell';
import { TextCell } from '@/components/data-table/text-cell';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { ConvertToInvoiceAction } from '@/Pages/Invoices/Shared/convert-to-invoice-action';
import { DueInvoice, Replacements, TransactionKind } from '@/types';
import { ColumnDef } from '@tanstack/react-table';
import { MoreHorizontal } from 'lucide-react';

type Props = {
  kind: TransactionKind;
  onDidClick: (item: DueInvoice, action: 'view:customer' | 'view:invoice' | 'view:estimate' | 'record:payment') => void;
  t: (key: string, replacements?: Replacements) => string;
};

export const getColumns = ({ kind, onDidClick, t }: Props): ColumnDef<DueInvoice>[] => {
  return [
    {
      accessorKey: 'due_on',
      header: (props) => {
        return (
          <HeaderCell title={t(`${kind === 'invoice' ? 'global.dueDate' : 'global.date'}`)} alignment="left" columnWidth={props.column.getSize()} />
        );
      },
      cell: (props) => {
        return <DateCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      id: 'name',
      accessorFn: (row) => row.customer.name, // prevent undefined
      header: (props) => {
        return <HeaderCell title={t('global.customer')} alignment="left" columnWidth={props.column.getSize()} />;
      },
      cell: (props) => {
        return <TextCell columnWidth={props.column.getSize()} value={props.getValue() as string} />;
      },
    },
    {
      accessorKey: 'amount_due',
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
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, kind === 'invoice' ? 'view:invoice' : 'view:estimate')}>
                {t(`dashboard.viewItem.${kind}.title`)}
              </DropdownMenuItem>
              <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'view:customer')}>
                {t('dashboard.viewItem.customer.title')}
              </DropdownMenuItem>
              {kind === 'invoice' && (
                <DropdownMenuItem onClick={() => onDidClick(props.row.original, 'record:payment')}>
                  {t('dashboard.viewItem.payment.title')}
                </DropdownMenuItem>
              )}
              {kind === 'estimate' && (
                <>
                  <DropdownMenuSeparator />
                  <ConvertToInvoiceAction id={props.row.original.uuid} renderedAs="dropdown-item" kind="estimate" />
                  <DropdownMenuItem>{t('dashboard.viewItem.estimate.markAsAccepted')}</DropdownMenuItem>
                  <DropdownMenuItem>{t('dashboard.viewItem.estimate.markAsRejected')}</DropdownMenuItem>
                </>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        );
      },
    },
  ];
};
