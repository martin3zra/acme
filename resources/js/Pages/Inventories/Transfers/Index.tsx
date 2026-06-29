import HeadingSmall from '@/components/heading-small';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { Link, router } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { breadcrumbs } from './constants';
import { statusVariant, TransferActions, TransferStatus } from './Shared/transfer-actions';

type Transfer = {
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

export default function Index({ auth, transfers }: PageProps<{ transfers: Transfer[] }>) {
  const t = useTranslation().trans;
  const currency = useNumber().currency;

  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="flex flex-col gap-4 p-4">
        <div className="flex items-center justify-between">
          <HeadingSmall title={t('transfers.title')} description={t('transfers.description')} />
          <Button size="sm" asChild>
            <Link href="/inventories/transfers/create">
              <Plus className="mr-1 h-4 w-4" />
              {t('transfers.newTransfer')}
            </Link>
          </Button>
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>{t('transfers.columns.date')}</TableHead>
              <TableHead>{t('transfers.columns.route')}</TableHead>
              <TableHead className="text-right">{t('transfers.footer.items')}</TableHead>
              <TableHead className="text-right">{t('transfers.footer.totalCost')}</TableHead>
              <TableHead>{t('transfers.columns.status')}</TableHead>
              <TableHead>{t('transfers.columns.requestedBy')}</TableHead>
              <TableHead className="text-right">{t('transfers.columns.actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {transfers.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-muted-foreground py-8 text-center">
                  {t('transfers.empty')}
                </TableCell>
              </TableRow>
            ) : (
              transfers.map((tr) => (
                <TableRow key={tr.id} className="cursor-pointer" onClick={() => router.visit(`/inventories/transfers/${tr.uuid}`)}>
                  <TableCell className="text-muted-foreground text-sm">{tr.created_at}</TableCell>
                  <TableCell>
                    {tr.from_warehouse} <span className="text-muted-foreground">→</span> {tr.to_warehouse}
                  </TableCell>
                  <TableCell className="text-right font-mono">{tr.line_count}</TableCell>
                  <TableCell className="text-right font-mono">{currency(tr.total_cost)}</TableCell>
                  <TableCell>
                    <Badge variant={statusVariant[tr.status]}>{t(`transfers.status.${tr.status}`)}</Badge>
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">{tr.requested_by || '—'}</TableCell>
                  <TableCell className="text-right" onClick={(e) => e.stopPropagation()}>
                    <TransferActions uuid={tr.uuid} status={tr.status} />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </AppLayout>
  );
}
