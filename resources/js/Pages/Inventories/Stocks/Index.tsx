import AppLayout from '@/layouts/app-layout';
import { PageProps } from '@/types';
import { breadcrumbs } from './constants';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import HeadingSmall from '@/components/heading-small';

type StockBalance = {
  variant_id: number;
  variant_name: string;
  sku: string;
  item_id: number;
  item_name: string;
  warehouse_id: number;
  warehouse: string;
  quantity: number;
  updated_at: string;
};

export default function Index({ auth, stocks }: PageProps<{ stocks: StockBalance[] }>) {
  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="flex flex-col gap-4 p-4">
        <HeadingSmall title="Stocks" description="Current inventory balances by variant and warehouse" />

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Item</TableHead>
              <TableHead>Variant</TableHead>
              <TableHead>SKU</TableHead>
              <TableHead>Warehouse</TableHead>
              <TableHead className="text-right">Qty on Hand</TableHead>
              <TableHead>Last Updated</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {stocks.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-center text-muted-foreground py-8">
                  No stock records found. Inventory balances will appear here once you receive goods or record sales.
                </TableCell>
              </TableRow>
            ) : (
              stocks.map((s) => (
                <TableRow key={`${s.variant_id}-${s.warehouse_id}`}>
                  <TableCell className="font-medium">{s.item_name}</TableCell>
                  <TableCell>{s.variant_name}</TableCell>
                  <TableCell className="text-muted-foreground">{s.sku || '—'}</TableCell>
                  <TableCell>{s.warehouse}</TableCell>
                  <TableCell className={`text-right font-mono ${s.quantity < 0 ? 'text-destructive' : ''}`}>
                    {s.quantity.toFixed(2)}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">{s.updated_at}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </AppLayout>
  );
}

