import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useHeader } from '@/composables/use-headers';
import AppLayout from '@/layouts/app-layout';
import { PageProps, Warehouse } from '@/types';
import { useForm } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { useState } from 'react';
import { breadcrumbs } from './constants';

type AdjustmentRow = {
  id: number;
  variant_id: number;
  variant_name: string;
  sku: string;
  item_name: string;
  warehouse_id: number;
  warehouse: string;
  qty: number;
  reason: string;
  notes: string;
  created_at: string;
};

type ItemVariant = {
  id: number;
  name: string;
  item_name: string;
  sku: string;
};

type AdjustmentPageProps = {
  adjustments: AdjustmentRow[];
  variants: ItemVariant[];
  warehouses: Warehouse[];
  defaultWarehouseId?: number | null;
};

type AdjustmentForm = {
  variant_id: string;
  warehouse_id: string;
  qty: string;
  reason: string;
  notes: string;
};

function NewAdjustmentDialog({
  variants,
  warehouses,
  defaultWarehouseId,
}: {
  variants: ItemVariant[];
  warehouses: Warehouse[];
  defaultWarehouseId?: number | null;
}) {
  const { headers } = useHeader();
  const [open, setOpen] = useState(false);
  const { data, setData, post, transform, reset, errors, processing } = useForm<AdjustmentForm>({
    variant_id: '',
    warehouse_id: defaultWarehouseId ? String(defaultWarehouseId) : '',
    qty: '',
    reason: '',
    notes: '',
  });

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    // Backend expects numeric ids/qty; the form holds them as strings (Select/
    // Input values), so coerce before the JSON body is sent.
    transform((d) => ({
      variant_id: Number(d.variant_id),
      warehouse_id: Number(d.warehouse_id),
      qty: Number(d.qty),
      reason: d.reason,
      notes: d.notes,
    }));
    post('/inventories/adjustments', {
      ...headers,
      onSuccess: () => {
        reset();
        setOpen(false);
      },
    });
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button size="sm">
          <Plus className="mr-1 h-4 w-4" />
          New Adjustment
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Record Stock Adjustment</DialogTitle>
        </DialogHeader>
        <form onSubmit={submit} className="mt-2 flex flex-col gap-4">
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">Item Variant</label>
            <Select value={data.variant_id} onValueChange={(v) => setData('variant_id', v)}>
              <SelectTrigger>
                <SelectValue placeholder="Select variant…" />
              </SelectTrigger>
              <SelectContent>
                {variants.map((v) => (
                  <SelectItem key={v.id} value={String(v.id)}>
                    {v.item_name} — {v.name} {v.sku ? `(${v.sku})` : ''}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.variant_id && <p className="text-destructive text-sm">{errors.variant_id}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">Warehouse</label>
            <Select value={data.warehouse_id} onValueChange={(v) => setData('warehouse_id', v)}>
              <SelectTrigger>
                <SelectValue placeholder="Select warehouse…" />
              </SelectTrigger>
              <SelectContent>
                {warehouses.map((w) => (
                  <SelectItem key={w.id} value={String(w.id)}>
                    {w.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.warehouse_id && <p className="text-destructive text-sm">{errors.warehouse_id}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">Quantity (use negative to remove stock)</label>
            <Input type="number" step="any" value={data.qty} onChange={(e) => setData('qty', e.target.value)} placeholder="e.g. 10 or -5" />
            {errors.qty && <p className="text-destructive text-sm">{errors.qty}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">Reason</label>
            <Input value={data.reason} onChange={(e) => setData('reason', e.target.value)} placeholder="e.g. Physical count correction" />
            {errors.reason && <p className="text-destructive text-sm">{errors.reason}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">Notes (optional)</label>
            <Input value={data.notes} onChange={(e) => setData('notes', e.target.value)} placeholder="Additional notes…" />
          </div>

          <Button type="submit" disabled={processing}>
            {processing ? 'Saving…' : 'Save Adjustment'}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function Index({ auth, adjustments, variants, warehouses, defaultWarehouseId }: PageProps<AdjustmentPageProps>) {
  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="flex flex-col gap-4 p-4">
        <div className="flex items-center justify-between">
          <HeadingSmall title="Stock Adjustments" description="Manual inventory adjustments" />
          <NewAdjustmentDialog variants={variants ?? []} warehouses={warehouses ?? []} defaultWarehouseId={defaultWarehouseId} />
        </div>

        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Date</TableHead>
              <TableHead>Item / Variant</TableHead>
              <TableHead>SKU</TableHead>
              <TableHead>Warehouse</TableHead>
              <TableHead className="text-right">Qty</TableHead>
              <TableHead>Reason</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {adjustments.length === 0 ? (
              <TableRow>
                <TableCell colSpan={6} className="text-muted-foreground py-8 text-center">
                  No adjustments recorded yet.
                </TableCell>
              </TableRow>
            ) : (
              adjustments.map((a) => (
                <TableRow key={a.id}>
                  <TableCell className="text-muted-foreground text-sm">{a.created_at}</TableCell>
                  <TableCell>
                    <span className="font-medium">{a.item_name}</span>
                    {a.variant_name && a.variant_name !== a.item_name && (
                      <span className="text-muted-foreground ml-1 text-sm">· {a.variant_name}</span>
                    )}
                  </TableCell>
                  <TableCell className="text-muted-foreground">{a.sku || '—'}</TableCell>
                  <TableCell>{a.warehouse}</TableCell>
                  <TableCell className={`text-right font-mono ${a.qty < 0 ? 'text-destructive' : 'text-green-600'}`}>
                    {a.qty > 0 ? '+' : ''}
                    {a.qty.toFixed(4)}
                  </TableCell>
                  <TableCell className="text-muted-foreground text-sm">{a.reason}</TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>
    </AppLayout>
  );
}
