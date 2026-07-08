import { AlertDestructive } from '@/components/alert-destructive';
import { DatePickerField } from '@/components/date-picker';
import InputError from '@/components/input-error';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { usePersistedState } from '@/hooks/use-persisted-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { isNotEmpty } from '@/lib/utils';
import { PageProps, Warehouse } from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import React, { useEffect, useState } from 'react';
import { createBreadcrumbs } from './constants';
import { Lines, TransferItem, TransferLine } from './Shared/lines';

type TransferFormState = {
  header: { from: string; to: string; notes: string; date: Date };
  lines: TransferLine[];
};

const emptyForm = (): TransferFormState => ({ header: { from: '', to: '', notes: '', date: new Date() }, lines: [] });

export default function Create({
  auth,
  warehouses,
  item,
  items,
}: PageProps<{
  warehouses: Warehouse[];
  item?: TransferItem;
  items?: TransferItem[];
}>) {
  const t = useTranslation().trans;
  const currency = useNumber().currency;
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;

  const [form, setForm, removeForm] = usePersistedState<TransferFormState>('transfer_create', emptyForm());
  const [currentItem, setCurrentItem] = useState<TransferItem | undefined>(undefined);
  const [amount, setAmount] = useState(0);
  const [openCancel, setOpenCancel] = useState(false);

  const referenceInputRef = React.useRef<HTMLInputElement>(null);
  const qtyInputRef = React.useRef<HTMLInputElement>(null);

  const { post, transform, processing, errors } = useForm<Record<string, any>>({
    from_warehouse_id: 0,
    to_warehouse_id: 0,
    date: new Date(),
    notes: '',
    lines: [],
  });

  // Focus the qty field after an item is picked. Deferred so the command dialog
  // finishes closing (Radix restores focus on close and would otherwise steal it).
  const focusQty = () => setTimeout(() => qtyInputRef.current?.focus(), 60);

  const applyItem = (picked: TransferItem) => {
    setCurrentItem(picked);
    if (referenceInputRef.current) referenceInputRef.current.value = picked.name;
    if (qtyInputRef.current) qtyInputRef.current.value = '1';
    setAmount(picked.cost ?? 0);
    focusQty();
  };

  useEffect(() => {
    if (item) applyItem(item);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [item]);

  // removeForm() sets the persisted state to undefined; bail out of rendering
  // until navigation completes so we never read form.header off undefined.
  // Placed after all hooks so hook order stays stable.
  if (!form) return null;

  const sameWarehouse = form.header.from !== '' && form.header.from === form.header.to;

  const searchItem = (term: string) => {
    router.reload({ only: ['item'], data: { search: term }, preserveUrl: true });
  };

  const resetLineInputs = () => {
    setCurrentItem(undefined);
    setAmount(0);
    if (referenceInputRef.current) referenceInputRef.current.value = '';
    if (qtyInputRef.current) qtyInputRef.current.value = '';
    referenceInputRef.current?.focus();
  };

  const addLine = () => {
    if (!currentItem) return;
    const qty = qtyInputRef.current?.valueAsNumber || 0;
    if (qty <= 0) return;
    setForm(() => ({ ...form, lines: [...form.lines, { ...currentItem, qty, cost: currentItem.cost ?? 0 }] }));
    resetLineInputs();
  };

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' || event.key === 'Tab') {
      if (event.currentTarget.name === 'reference' && isNotEmpty(event.currentTarget.value)) {
        event.preventDefault();
        searchItem(event.currentTarget.value);
        return;
      }
      if (event.currentTarget.name === 'qty' && currentItem) {
        event.preventDefault();
        addLine();
      }
    }
  };

  const handleRemoveLine = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    const index = Number(event.currentTarget.getAttribute('data-index'));
    setForm(() => ({ ...form, lines: form.lines.filter((_, i) => i !== index) }));
  };

  const totalCost = form.lines.reduce((acc, l) => acc + l.qty * l.cost, 0);

  const performSave = () => {
    if (sameWarehouse || form.lines.length === 0) return;
    transform(() => ({
      from_warehouse_id: Number(form.header.from),
      to_warehouse_id: Number(form.header.to),
      date: form.header.date,
      notes: form.header.notes || '',
      lines: form.lines.map((l) => ({ id: l.id, variant_id: l.variant_id, qty: l.qty, unit: l.unit?.id ?? 0, cost: l.cost, description: l.description || '' })),
    }));
    post('/inventories/transfers', {
      ...headers,
      preserveState: 'errors',
      onSuccess: () => {
        removeForm();
        router.visit('/inventories/transfers');
      },
    });
  };

  const performCancel = () => {
    router.visit('/inventories/transfers');
    setTimeout(() => removeForm(), 200);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={createBreadcrumbs}>
      <AppLayout.Actions>
        <div className="flex justify-end gap-x-6">
          <Button variant="secondary" onClick={() => setOpenCancel(true)}>
            {t('global.actions.cancel')}
          </Button>
          <Button onClick={performSave} disabled={processing || sameWarehouse || form.lines.length === 0}>
            {t('global.actions.save')}
          </Button>
        </div>
      </AppLayout.Actions>

      <div className="grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4">
        {propsErrors.status && (
          <div className="col-span-12">
            <AlertDestructive description={propsErrors.status} onDestroy={() => delete propsErrors.status} />
          </div>
        )}

        {/* Header */}
        <div className="col-span-12 grid min-h-42 grid-cols-2 gap-x-6">
          {/* Origin / destination card */}
          <div className="flex flex-col justify-center gap-y-4 rounded-lg border p-6">
            <div className="flex items-center gap-x-4">
              <Label className="w-32 whitespace-nowrap">{t('transfers.fields.fromWarehouse')}</Label>
              <Select
                value={form.header.from}
                onValueChange={(v) => setForm(() => ({ ...form, header: { ...form.header, from: v === 'none' ? '' : v } }))}
              >
                <SelectTrigger className="w-72">
                  <SelectValue placeholder={t('movements.selectWarehouse')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">{t('movements.selectWarehouse')}</SelectItem>
                  {warehouses.map((w) => (
                    <SelectItem key={w.id} value={String(w.id)} disabled={String(w.id) === form.header.to}>
                      {w.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            <div className="flex items-center gap-x-4">
              <Label className="w-32 whitespace-nowrap">{t('transfers.fields.toWarehouse')}</Label>
              <Select
                value={form.header.to}
                onValueChange={(v) => setForm(() => ({ ...form, header: { ...form.header, to: v === 'none' ? '' : v } }))}
              >
                <SelectTrigger className="w-72">
                  <SelectValue placeholder={t('movements.selectWarehouse')} />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="none">{t('movements.selectWarehouse')}</SelectItem>
                  {warehouses.map((w) => (
                    <SelectItem key={w.id} value={String(w.id)} disabled={String(w.id) === form.header.from}>
                      {w.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>
            {sameWarehouse && <p className="text-destructive text-sm">{t('movements.errors.sameWarehouse')}</p>}
            <InputError message={errors.from_warehouse_id || errors.to_warehouse_id} />
          </div>

          {/* Date */}
          <div className="grid grid-cols-12">
            <div className="col-span-6 flex flex-col gap-y-6">
              <DatePickerField
                id="date"
                label={t('global.date')}
                placeholder={t('global.datePlaceholder')}
                value={form.header.date}
                onChange={(d) => setForm(() => ({ ...form, header: { ...form.header, date: d as Date } }))}
                error={errors.date}
              />
            </div>
          </div>
        </div>

        {/* Lines */}
        <div className="col-span-12">
          <Lines
            items={items}
            lines={form.lines}
            lineError={errors.lines}
            currentItem={currentItem}
            amount={amount}
            setAmount={setAmount}
            handleRemoveLine={handleRemoveLine}
            handleKeyDown={handleKeyDown}
            handleOnSelected={applyItem}
            referenceInputRef={referenceInputRef}
            qtyInputRef={qtyInputRef}
          />
        </div>

        {/* Footer: notes + totals */}
        <div className="col-span-12 min-h-48">
          <div className="grid grid-cols-12">
            <div className="col-span-10 flex flex-col gap-y-2 py-2">
              <Label className="text-sm/6 font-medium">{t('global.notes')}</Label>
              <Textarea
                name="notes"
                rows={4}
                className="block w-1/2 resize-none rounded-lg border px-3 py-1.5 text-sm/6"
                value={form.header.notes}
                onChange={(e) => setForm(() => ({ ...form, header: { ...form.header, notes: e.currentTarget.value } }))}
              />
            </div>
            <div className="col-span-2 flex flex-col gap-y-2 rounded-lg border border-gray-300/25 bg-gray-100/10">
              <div className="grid place-content-end gap-y-4 p-2">
                <div className="flex w-60 items-center justify-between">
                  <span className="block text-base">{t('transfers.footer.items')}</span>
                  <span className="block text-base">{form.lines.length}</span>
                </div>
                <Separator />
                <div className="flex w-60 items-center justify-between">
                  <span className="block text-xl">{t('transfers.footer.totalCost')}</span>
                  <span className="block text-xl">{currency(totalCost)}</span>
                </div>
              </div>
            </div>
          </div>
        </div>

        <AlertDialog open={openCancel} onOpenChange={setOpenCancel}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{t('global.actions.cancel')}</AlertDialogTitle>
              <AlertDialogDescription>{t('global.warning.description')}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>{t('global.cancel')}</AlertDialogCancel>
              <AlertDialogAction onClick={performCancel}>{t('global.ok')}</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </AppLayout>
  );
}
