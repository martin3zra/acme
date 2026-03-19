import { AlertDestructive } from '@/components/alert-destructive';
import { DatePickerField } from '@/components/date-picker';
import { MoneyInput } from '@/components/money-input';
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
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { usePersistedState } from '@/hooks/use-persisted-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { BTForm, CardForm, CashForm, CheckForm, PageProps, Payable, PaymentMethod, Vendor, VendorPaymentForm } from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { RowSelectionState } from '@tanstack/table-core/build/lib/features/RowSelection';
import React, { useEffect } from 'react';
import CheckoutForm from '../Invoices/Shared/checkout-form';
import { VendorSection } from '../Purchases/Shared/vendor-section';
import { createPayableBreadcrumbs, defaultVendorPaymentForm } from './constants';
import { PayableLineRow } from './Shared/columns-definitions';
import { List } from './Shared/lines-payment';

export default function Create({ auth, vendor, vendors, payables }: PageProps<{ vendor?: Vendor; vendors: Vendor[]; payables: Payable[] }>) {
  const t = useTranslation().trans;
  const { currency } = useNumber();
  const [openCancelConfirmation, setCancelConfirmation] = React.useState(false);
  const [openCheckout, setCheckout] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [bulkPayment, setBulkPayment] = React.useState<number>(0);
  const [open, setOpen] = React.useState(false);
  const [search, setSearch] = React.useState('');
  const debouncedSearch = useDebounced(search, 500);
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({});
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;

  const [paymentForm, setPaymentForm, removePaymentForm] = usePersistedState<VendorPaymentForm>(
    'vendor_payment',
    {
      ...defaultVendorPaymentForm(),
      header: { ...defaultVendorPaymentForm().header, vendor },
    } as VendorPaymentForm,
    false,
  );

  const { post, transform, processing, errors } = useForm({
    vendor_id: 0,
    date: new Date(),
    lines: [],
  });

  const buildLines = (payables: Payable[]): PayableLineRow[] =>
    payables.map((p) => ({
      ...p,
      amount_due: p.amount_payable - p.amount_paid,
      payment: 0,
      action: 'unchanged' as const,
    }));

  useEffect(() => {
    if (!payables) return;
    setPaymentForm((prev) => ({
      ...prev,
      lines: buildLines(payables) as any,
    }));
  }, [payables]);

  useEffect(() => {
    if (!debouncedSearch) return;
    router.reload({ only: ['vendors'], data: { search: debouncedSearch }, preserveUrl: true });
  }, [debouncedSearch]);

  const totalPaid = (): number => (paymentForm.lines as any[]).reduce((acc: number, line: any) => acc + (line.payment || 0), 0);

  const handleRecordPayment = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    const payload = {
      vendor_id: paymentForm.header.vendor?.uuid,
      date: paymentForm.header.date,
      amount: totalPaid(),
      notes: paymentForm.header.notes,
      lines: (paymentForm.lines as any[])
        .filter((line: any) => line.payment > 0)
        .map((line: any) => ({
          uuid: line.invoice_uuid,
          amount_due: line.amount_due,
          payment: line.payment,
        })),
      payment: paymentForm.payment,
    };
    transform((data) => ({ ...data, ...payload }));

    post('/payables', {
      ...headers,
      preserveState: 'errors',
      onSuccess: () => {
        removePaymentForm();
        router.get('/payables');
      },
    });
  };

  const handleVendorSelection = (vendor: Vendor | undefined) => {
    setPaymentForm(() => ({ ...paymentForm, header: { ...paymentForm.header, vendor }, lines: [] }));
    setOpen(false);
    if (vendor !== undefined) {
      router.reload({
        only: ['payables'],
        data: { vendor_id: vendor.uuid },
        preserveUrl: true,
        onStart: () => setLoading(true),
        onSuccess: (page) => {
          const payables = page.props.payables as Payable[];
          setRowSelection({});
          setPaymentForm((prev) => ({ ...prev, lines: buildLines(payables) as any }));
        },
        onFinish: () => setLoading(false),
      });
    }
  };

  const handleDateChange = (date: unknown) => {
    setPaymentForm(() => ({ ...paymentForm, header: { ...paymentForm.header, date: date as Date } }));
  };

  const handleCellChange = (rowId: string, columnId: string, newValue: string | number) => {
    setPaymentForm((prev) => {
      const lines = (prev.lines as any[]).map((line: any) => {
        if (line.id.toString() === rowId) {
          const payment = columnId === 'payment' ? Number(newValue) : line.payment || 0;
          return { ...line, [columnId]: Number(newValue), remaining: (line.amount_due || 0) - payment };
        }
        return line;
      });

      const totals = lines.reduce(
        (acc: any, line: any) => {
          acc.totalPayment += line.payment || 0;
          acc.totalRemaining += (line.amount_due || 0) - (line.payment || 0);
          return acc;
        },
        { totalPayment: 0, totalRemaining: 0 },
      );

      return { ...prev, lines, totals };
    });

    setRowSelection((prev) => ({ ...prev, [rowId]: true }));
  };

  const onSelectionChange = (selection: RowSelectionState) => {
    const lines = (paymentForm.lines as any[]).map((line: any) => ({ ...line, payment: 0 }));
    Object.keys(selection).forEach((id) => {
      const index = lines.findIndex((l: any) => l.id === Number(id));
      if (index >= 0) lines[index].payment = lines[index].amount_due;
    });
    setPaymentForm((prev) => ({ ...prev, lines: [...lines] }));
  };

  const performPaymentCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    router.get('/payables');
    setTimeout(() => removePaymentForm(), 300);
  };

  const handleCheckout = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    if (totalPaid() === 0) return;
    setCheckout(true);
  };

  const handleCheckoutChange = (method: PaymentMethod, form: CashForm | CheckForm | CardForm | BTForm) => {
    setPaymentForm(() => ({ ...paymentForm, payment: { ...paymentForm.payment, [method]: form } }));
  };

  const distributePayment = (amount: number) => {
    let remaining = amount;
    setPaymentForm((prev) => {
      const updatedLines = (prev.lines as any[]).map((line: any) => {
        if (remaining <= 0) return { ...line, payment: 0 };
        if (remaining >= line.amount_due) {
          remaining -= line.amount_due;
          return { ...line, payment: line.amount_due };
        } else {
          const partial = remaining;
          remaining = 0;
          return { ...line, payment: partial };
        }
      });

      const newRowSelection: RowSelectionState = {};
      updatedLines.forEach((line: any) => {
        if (line.payment > 0) newRowSelection[line.id.toString()] = true;
      });
      setRowSelection(newRowSelection);

      return { ...prev, lines: updatedLines };
    });
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={createPayableBreadcrumbs}>
      <AppLayout.Actions>
        <div className="flex justify-end gap-x-6">
          <Button variant={'secondary'} onClick={() => setCancelConfirmation(true)}>
            {t('global.actions.cancel')}
          </Button>
          <Button onClick={handleCheckout} disabled={totalPaid() === 0 || processing}>
            {t('global.actions.checkout')}
          </Button>
        </div>
      </AppLayout.Actions>
      <div className="grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4 bg-gray-50/10">
        {!openCheckout && propsErrors.status && (
          <div className="col-span-12">
            <AlertDestructive description={propsErrors.status} onDestroy={() => delete propsErrors.status} />
          </div>
        )}
        <div className="z-50 col-span-12 grid min-h-42 grid-cols-2 gap-x-6">
          <VendorSection
            vendor={paymentForm.header.vendor}
            vendors={vendors}
            errors={errors}
            handleVendorSelection={handleVendorSelection}
            setSearch={setSearch}
            setOpen={setOpen}
            open={open}
            debouncedSearch={debouncedSearch}
          />
          <div className="grid grid-cols-12">
            <div className="col-span-6 flex flex-col gap-y-6">
              <div className="flex flex-col gap-y-2">
                <DatePickerField
                  id="date"
                  label={t('global.date')}
                  placeholder={t('global.datePlaceholder')}
                  value={paymentForm.header.date}
                  onChange={handleDateChange}
                  error={errors.date}
                />
              </div>
              <div className="flex flex-col">
                <div className="flex flex-col gap-y-2">
                  <Label className="text-sm/6 font-medium">{t('global.notes')}</Label>
                  <Textarea
                    name="notes"
                    rows={4}
                    className="focus:no-data-focus:outline-none block resize-none rounded-lg border px-3 py-1.5 text-sm/6 data-focus:outline-2 data-focus:-outline-offset-2 data-focus:outline-white/25"
                    defaultValue={paymentForm.header.notes}
                    onChange={(e) => setPaymentForm(() => ({ ...paymentForm, header: { ...paymentForm.header, notes: e.currentTarget.value } }))}
                  />
                </div>
              </div>
            </div>
            <div className="col-span-6 flex flex-col items-end space-y-4">
              <div className="flex flex-col items-end space-y-2">
                <Label htmlFor="bulkPayment">{t('global.bulkPayment')}</Label>
                <MoneyInput name="bulkPayment" value={bulkPayment} onChange={(value) => setBulkPayment(value)} className="text-end" />
              </div>
              <Button disabled={bulkPayment === 0} onClick={() => distributePayment(bulkPayment)}>
                {t('payables.applyBulkPayment')}
              </Button>
              <div className="col-span-12 grid place-items-end">
                <div className="flex flex-col gap-x-2">
                  <Label className="text-muted-foreground block text-end text-lg">{t('global.totalReceived')}</Label>
                  <Label className="block text-end text-4xl">{currency(totalPaid())}</Label>
                </div>
              </div>
            </div>
          </div>
        </div>
        <div className="col-span-12">
          <List
            loading={loading}
            data={paymentForm.lines as unknown as PayableLineRow[]}
            totals={paymentForm.totals}
            rowSelection={rowSelection}
            setRowSelection={setRowSelection}
            onSelectPayableLine={() => {}}
            onValueChange={handleCellChange}
            onSelectionChange={onSelectionChange}
          />
        </div>
        <AlertDialog open={openCancelConfirmation} onOpenChange={setCancelConfirmation}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{t('payments.confirmsCancelation.title')}</AlertDialogTitle>
              <AlertDialogDescription>{t('payments.confirmsCancelation.description')}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>{t('global.actions.cancel')}</AlertDialogCancel>
              <AlertDialogAction onClick={performPaymentCancelation}>{t('payments.confirmsCancelation.confirm')}</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
        <CheckoutForm
          action={t('global.actions.recordPayment')}
          openCheckout={openCheckout}
          setCheckout={setCheckout}
          paymentForm={paymentForm.payment}
          totalAmount={totalPaid()}
          onCompleteCheckout={handleRecordPayment}
          processing={processing}
          setCancelConfirmation={setCancelConfirmation}
          errors={propsErrors}
          onCheckoutChange={handleCheckoutChange}
          currency={currency}
          t={t}
        />
      </div>
    </AppLayout>
  );
}
