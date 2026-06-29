import HeadingSmall from '@/components/heading-small';
import { Badge } from '@/components/ui/badge';
import { Separator } from '@/components/ui/separator';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { showBreadcrumbs } from './constants';
import { statusVariant, TransferActions, TransferStatus } from './Shared/transfer-actions';

type TransferHeader = {
  id: number;
  uuid: string;
  from_warehouse: string;
  to_warehouse: string;
  status: TransferStatus;
  notes: string;
  requested_by: string;
  line_count: number;
  total_qty: number;
  total_cost: number;
  created_at: string;
  dispatched_at: string;
  received_at: string;
};

type TransferLine = {
  id: number;
  variant_id: number;
  variant_name: string;
  sku: string;
  reference: string;
  item_name: string;
  description: string;
  unit: string;
  qty: number;
  unit_cost: number;
  line_total: number;
};

export default function Show({ auth, transfer, lines }: PageProps<{ transfer: TransferHeader; lines: TransferLine[] }>) {
  const t = useTranslation().trans;
  const currency = useNumber().currency;

  return (
    <AppLayout user={auth.user} breadcrumbs={showBreadcrumbs}>
      <AppLayout.Actions>
        <TransferActions uuid={transfer.uuid} status={transfer.status} />
      </AppLayout.Actions>

      <div className="flex flex-col gap-y-4 p-4">
        {/* Header */}
        <div className="rounded-lg border bg-white p-4 shadow-sm">
          <div className="flex items-start justify-between">
            <HeadingSmall
              title={`${transfer.from_warehouse} → ${transfer.to_warehouse}`}
              description={`${t('transfers.columns.requestedBy')}: ${transfer.requested_by || '—'}`}
            />
            <Badge variant={statusVariant[transfer.status]}>{t(`transfers.status.${transfer.status}`)}</Badge>
          </div>
          <Separator className="my-3" />
          <div className="grid grid-cols-3 gap-4 text-sm">
            <div>
              <div className="text-muted-foreground">{t('transfers.columns.date')}</div>
              <div>{transfer.created_at}</div>
            </div>
            <div>
              <div className="text-muted-foreground">{t('transfers.actions.dispatch')}</div>
              <div>{transfer.dispatched_at || '—'}</div>
            </div>
            <div>
              <div className="text-muted-foreground">{t('transfers.actions.receive')}</div>
              <div>{transfer.received_at || '—'}</div>
            </div>
          </div>
          {transfer.notes && (
            <>
              <Separator className="my-3" />
              <div className="text-sm">
                <div className="text-muted-foreground">{t('global.notes')}</div>
                <div>{transfer.notes}</div>
              </div>
            </>
          )}
        </div>

        {/* Lines */}
        <div className="rounded-lg border bg-white p-4 shadow-sm">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('transfers.line.code')}</TableHead>
                <TableHead>{t('transfers.line.item')}</TableHead>
                <TableHead>{t('transfers.line.description')}</TableHead>
                <TableHead>{t('transfers.line.unit')}</TableHead>
                <TableHead className="text-right">{t('transfers.line.qty')}</TableHead>
                <TableHead className="text-right">{t('transfers.line.cost')}</TableHead>
                <TableHead className="text-right">{t('transfers.line.total')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {lines.map((l) => (
                <TableRow key={l.id}>
                  <TableCell>{l.reference || l.sku || '—'}</TableCell>
                  <TableCell className="font-medium">
                    {l.item_name}
                    {l.variant_name && l.variant_name !== l.item_name && (
                      <span className="text-muted-foreground ml-1 text-sm">· {l.variant_name}</span>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{l.description || '—'}</TableCell>
                  <TableCell>{l.unit || '—'}</TableCell>
                  <TableCell className="text-right font-mono">{l.qty}</TableCell>
                  <TableCell className="text-right font-mono">{currency(l.unit_cost)}</TableCell>
                  <TableCell className="text-right font-mono">{currency(l.line_total)}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </div>

        {/* Footer */}
        <div className="flex justify-end">
          <div className="w-72 rounded-lg border bg-white p-4 shadow-sm">
            <div className="flex items-center justify-between">
              <span>{t('transfers.footer.items')}</span>
              <span className="font-mono">{transfer.line_count}</span>
            </div>
            <Separator className="my-2" />
            <div className="flex items-center justify-between">
              <span className="text-base">{t('transfers.footer.totalCost')}</span>
              <span className="text-base font-mono">{currency(transfer.total_cost)}</span>
            </div>
          </div>
        </div>
      </div>
    </AppLayout>
  );
}
