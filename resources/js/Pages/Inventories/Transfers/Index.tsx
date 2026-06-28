import AppLayout from '@/layouts/app-layout';
import { PageProps, Warehouse } from '@/types';
import { breadcrumbs } from './constants';
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table';
import { Badge } from '@/components/ui/badge';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '@/components/ui/dialog';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import { useForm } from '@inertiajs/react';
import { useState } from 'react';
import { Plus } from 'lucide-react';
import { useTranslation } from '@/hooks/use-translation';

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

type ItemVariant = {
  id: number;
  name: string;
  item_name: string;
  sku: string;
};

type TransferPageProps = {
  movements: InventoryMovement[];
  variants: ItemVariant[];
  warehouses: Warehouse[];
};

type TransferForm = {
  variant_id: string;
  from_warehouse_id: string;
  to_warehouse_id: string;
  qty: string;
  notes: string;
};

const kindLabel: Record<string, { label: string; variant: 'default' | 'secondary' | 'destructive' | 'outline' }> = {
  sale:             { label: 'Sale',             variant: 'destructive' },
  sale_return:      { label: 'Sale Return',      variant: 'outline' },
  purchase_order:   { label: 'Purchase Order',   variant: 'secondary' },
  purchase_receipt: { label: 'Purchase Receipt', variant: 'default' },
  purchase_return:  { label: 'Purchase Return',  variant: 'outline' },
  vendor_bill:      { label: 'Vendor Bill',      variant: 'default' },
  adjustment:       { label: 'Adjustment',       variant: 'secondary' },
  transfer:         { label: 'Transfer',         variant: 'secondary' },
};

function NewTransferDialog({ variants, warehouses }: { variants: ItemVariant[]; warehouses: Warehouse[] }) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(false);
  const { data, setData, post, reset, errors, processing } = useForm<TransferForm>({
    variant_id: '',
    from_warehouse_id: '',
    to_warehouse_id: '',
    qty: '',
    notes: '',
  });

  const sameWarehouse =
    data.from_warehouse_id !== '' && data.from_warehouse_id === data.to_warehouse_id;

  const submit = (e: React.FormEvent) => {
    e.preventDefault();
    if (sameWarehouse) return;
    post('/inventories/transfers', {
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
          <Plus className="h-4 w-4 mr-1" />
          {t('movements.newTransfer')}
        </Button>
      </DialogTrigger>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>{t('movements.dialogTitle')}</DialogTitle>
        </DialogHeader>
        <form onSubmit={submit} className="flex flex-col gap-4 mt-2">
          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">{t('movements.variant')}</label>
            <Select value={data.variant_id} onValueChange={(v) => setData('variant_id', v)}>
              <SelectTrigger>
                <SelectValue placeholder={t('movements.selectVariant')} />
              </SelectTrigger>
              <SelectContent>
                {variants.map((v) => (
                  <SelectItem key={v.id} value={String(v.id)}>
                    {v.item_name} — {v.name} {v.sku ? `(${v.sku})` : ''}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.variant_id && <p className="text-sm text-destructive">{errors.variant_id}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">{t('movements.fromWarehouse')}</label>
            <Select value={data.from_warehouse_id} onValueChange={(v) => setData('from_warehouse_id', v)}>
              <SelectTrigger>
                <SelectValue placeholder={t('movements.selectWarehouse')} />
              </SelectTrigger>
              <SelectContent>
                {warehouses.map((w) => (
                  <SelectItem key={w.id} value={String(w.id)}>
                    {w.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {errors.from_warehouse_id && <p className="text-sm text-destructive">{errors.from_warehouse_id}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">{t('movements.toWarehouse')}</label>
            <Select value={data.to_warehouse_id} onValueChange={(v) => setData('to_warehouse_id', v)}>
              <SelectTrigger>
                <SelectValue placeholder={t('movements.selectWarehouse')} />
              </SelectTrigger>
              <SelectContent>
                {warehouses.map((w) => (
                  <SelectItem key={w.id} value={String(w.id)}>
                    {w.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {sameWarehouse && <p className="text-sm text-destructive">{t('movements.errors.sameWarehouse')}</p>}
            {errors.to_warehouse_id && <p className="text-sm text-destructive">{errors.to_warehouse_id}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">{t('movements.qty')}</label>
            <Input
              type="number"
              step="any"
              min="0"
              value={data.qty}
              onChange={(e) => setData('qty', e.target.value)}
              placeholder={t('movements.qtyPlaceholder')}
            />
            {errors.qty && <p className="text-sm text-destructive">{errors.qty}</p>}
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-sm font-medium">{t('movements.notes')}</label>
            <Input
              value={data.notes}
              onChange={(e) => setData('notes', e.target.value)}
              placeholder={t('movements.notesPlaceholder')}
            />
          </div>

          <Button type="submit" disabled={processing || sameWarehouse}>
            {processing ? t('global.saving') : t('movements.save')}
          </Button>
        </form>
      </DialogContent>
    </Dialog>
  );
}

export default function Index({ auth, movements, variants, warehouses }: PageProps<TransferPageProps>) {
  const t = useTranslation().trans;
  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="flex flex-col gap-4 p-4">
        <div className="flex items-center justify-between">
          <HeadingSmall title={t('movements.title')} description={t('movements.description')} />
          <NewTransferDialog variants={variants ?? []} warehouses={warehouses ?? []} />
        </div>

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
                <TableCell colSpan={7} className="text-center text-muted-foreground py-8">
                  No inventory movements yet.
                </TableCell>
              </TableRow>
            ) : (
              movements.map((m) => {
                const info = kindLabel[m.kind] ?? { label: m.kind, variant: 'outline' as const };
                return (
                  <TableRow key={m.id}>
                    <TableCell className="text-sm text-muted-foreground">{m.created_at}</TableCell>
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
                      {m.qty > 0 ? '+' : ''}{m.qty.toFixed(4)}
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
