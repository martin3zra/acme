import HeadingSmall from '@/components/heading-small';
import { Badge } from '@/components/ui/badge';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { breadcrumbs } from './constants';

type InventoryMovement = {
  id: number;
  variant_id: number;
  variant_name: string;
  sku: string;
  item_name: string;
  warehouse_id: number;
  warehouse: string;
  kind: string;
  qty: number;
  unit_cost: number;
  reference_type: string;
  reference_id: number;
  created_at: string;
};

type MovementsPageProps = {
  movements: InventoryMovement[];
};

const kindLabel: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  sale: { label: 'Sale', variant: 'destructive' },
  sale_return: { label: 'Sale Return', variant: 'outline' },
  purchase_order: { label: 'Purchase Order', variant: 'secondary' },
  purchase_receipt: { label: 'Purchase Receipt', variant: 'default' },
  purchase_return: { label: 'Purchase Return', variant: 'outline' },
  vendor_bill: { label: 'Vendor Bill', variant: 'default' },
  adjustment: { label: 'Adjustment', variant: 'secondary' },
  transfer: { label: 'Transfer', variant: 'secondary' },
};

export default function Index({ auth, movements }: PageProps<MovementsPageProps>) {
  const t = useTranslation().trans;
  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="flex flex-col gap-4 p-4">
        <HeadingSmall title={t('movements.title')} description={t('movements.description')} />

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Item / Variant</TableHead>
              <TableHead>SKU</TableHead>
              <TableHead>Warehouse</TableHead>
              <TableHead>Type</TableHead>
              <TableHead className="text-right">Qty</TableHead>
              <TableHead>Reference</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {movements.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} className="text-muted-foreground py-8 text-center">
                  No inventory movements yet.
                </TableCell>
              </TableRow>
            ) : (
              movements.map((m) => {
                const info = kindLabel[m.kind] ?? { label: m.kind, variant: 'outline' as const };
                return (
                  <TableRow key={m.id}>
                    <TableCell className="text-muted-foreground text-sm">{m.created_at}</TableCell>
                    <TableCell>
                      <span className="font-medium">{m.item_name}</span>
                      {m.variant_name && m.variant_name !== m.item_name && (
                        <span className="text-muted-foreground ml-1 text-sm">· {m.variant_name}</span>
                      )}
                    </TableCell>
                    <TableCell className="text-muted-foreground">{m.sku || '—'}</TableCell>
                    <TableCell>{m.warehouse}</TableCell>
                    <TableCell>
                      <Badge variant={info.variant}>{info.label}</Badge>
                    </TableCell>
                    <TableCell className={`text-right font-mono ${m.qty < 0 ? 'text-destructive' : 'text-green-600'}`}>
                      {m.qty > 0 ? '+' : ''}
                      {m.qty.toFixed(4)}
                    </TableCell>
                    <TableCell className="text-muted-foreground text-sm capitalize">
                      {m.reference_type} #{m.reference_id}
                    </TableCell>
                  </TableRow>
                );
              })
            )}
          </TableBody>
        </Table>
      </div>
    </AppLayout>
  );
}
