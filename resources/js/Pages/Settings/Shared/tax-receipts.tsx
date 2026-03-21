'use client';

import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog';
import { Input } from '@/components/ui/input';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { PageProps, TaxReceiptForSetup } from '@/types';
import { router, usePage } from '@inertiajs/react';
import * as React from 'react';

interface Props {
  uuid: string;
  taxReceipts: TaxReceiptForSetup[];
}

export function TaxReceipts({ uuid, taxReceipts }: Props) {
  const { auth } = usePage<PageProps>().props;
  const { headers } = useHeader();
  const t = useTranslation().trans;
  const [active, setActive] = React.useState<Record<number | string, boolean>>({});
  const [ranges, setRanges] = React.useState<Record<number | string, { start: number; end: number }>>({});
  const [open, setOpen] = React.useState(false);

  React.useEffect(() => {
    const initialActive: Record<number | string, boolean> = {};
    const initialRanges: Record<number | string, { start: number; end: number }> = {};
    taxReceipts.forEach((r) => {
      initialActive[r.id] = (r.sequence_start ?? 0) > 0 && (r.sequence_end ?? 0) > 0;
      initialRanges[r.id] = { start: r.sequence_start ?? 0, end: r.sequence_end ?? 0 };
    });
    setActive(initialActive);
    setRanges(initialRanges);
  }, [taxReceipts]);

  const handleToggle = (id: number | string) => {
    setActive((prev) => ({ ...prev, [id]: !prev[id] }));
  };

  const handleChange = (id: number | string, field: 'start' | 'end', value: number) => {
    setRanges((prev) => ({
      ...prev,
      [id]: { ...prev[id], [field]: Number.isNaN(value) ? 0 : value },
    }));
  };

  const validateRange = (start: number, end: number) => {
    if (!start || !end) return false;
    return !isNaN(start) && !isNaN(end) && start <= end;
  };

  const selectedReceipts = taxReceipts.filter((r) => active[r.id]);
  const hasInvalidSelectedRanges = selectedReceipts.some((r) => {
    const range = ranges[r.id] ?? { start: 0, end: 0 };
    return !validateRange(range.start, range.end);
  });

  const handleSave = () => {
    // Send data to server via Inertia
    router.put(
      `/settings/${auth.account.uuid}/companies/${uuid}/tax-receipts`,
      {
        receipts: selectedReceipts.map((r) => ({
          id: r.id,
          start: ranges[r.id]?.start,
          end: ranges[r.id]?.end,
        })),
      },
      { ...headers, preserveState: 'errors' },
    );
    setOpen(false);
  };

  return (
    <>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Activar</TableHead>
            <TableHead>Nombre</TableHead>
            <TableHead>Serie</TableHead>
            <TableHead>Tipo</TableHead>
            <TableHead>Secuencia Inicio</TableHead>
            <TableHead>Secuencia Fin</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {taxReceipts.map((receipt) => {
            const isActive = active[receipt.id] ?? false;
            const range = ranges[receipt.id] ?? { start: '', end: '' };
            const isValid = validateRange(range.start, range.end);

            return (
              <TableRow key={receipt.id}>
                <TableCell>
                  <Checkbox checked={isActive} onCheckedChange={() => handleToggle(receipt.id)} />
                </TableCell>
                <TableCell>{receipt.name}</TableCell>
                <TableCell>{receipt.serie}</TableCell>
                <TableCell className="capitalize">{receipt.type}</TableCell>
                <TableCell>
                  <Input
                    type="number"
                    value={range.start}
                    onChange={(e) => handleChange(receipt.id, 'start', e.target.valueAsNumber)}
                    disabled={!isActive}
                    className={cn(isActive && !isValid && 'border-red-500')}
                  />
                </TableCell>
                <TableCell>
                  <Input
                    type="number"
                    value={range.end}
                    onChange={(e) => handleChange(receipt.id, 'end', e.target.valueAsNumber)}
                    disabled={!isActive}
                    className={cn(isActive && !isValid && 'border-red-500')}
                  />
                </TableCell>
              </TableRow>
            );
          })}
        </TableBody>
      </Table>

      <div className="mt-4">
        <Button onClick={() => setOpen(true)} disabled={selectedReceipts.length === 0 || hasInvalidSelectedRanges}>
          {t('global.save')}
        </Button>
      </div>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Confirmar Comprobantes</DialogTitle>
          </DialogHeader>
          <div className="space-y-2">
            {selectedReceipts.map((r) => (
              <div key={r.id} className="flex justify-between">
                <span>
                  {r.name} ({r.serie})
                </span>
                <span>
                  {ranges[r.id]?.start} → {ranges[r.id]?.end}
                </span>
              </div>
            ))}
          </div>
          <DialogFooter>
            <Button variant="secondary" onClick={() => setOpen(false)}>
              {t('global.cancel')}
            </Button>
            <Button onClick={handleSave}>{t('global.confirmAndSave')}</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  );
}
