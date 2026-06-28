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
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { usePersistedState } from '@/hooks/use-persisted-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { BTForm, CardForm, CashForm, CheckForm, Customer, PageProps, PaymentForm, PaymentMethod, Receivable, ReceivableInvoiceForm } from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { RowSelectionState } from '@tanstack/table-core/build/lib/features/RowSelection';
import React, { useEffect } from 'react';
import CheckoutForm from '../Invoices/Shared/checkout-form';
import { CustomerSection } from '../Invoices/Shared/customer-section';
import { createPaymentBreadcrumbs } from '../Payments/constants';
import { buildReceivableState, buildRowSelection } from './build-receivables-state';
import { defaultPaymentForm } from './constants';
import { List } from './Shared/lines-payment';

export default function Create({
  auth,
  customer,
  customers,
  receivables,
  invoice_uuid,
  forceInitial,
}: PageProps<{ customer: Customer; customers: Customer[]; receivables: Receivable[]; invoice_uuid: string; forceInitial: boolean }>) {
  const t = useTranslation().trans;
  const { currency } = useNumber();
  const [openCancelConfirmation, setCancelConfirmation] = React.useState(false);
  const [openCheckout, setCheckout] = React.useState(false);
  const [loading, setLoading] = React.useState(false);
  const [bulkPayment, setBulkPayment] = React.useState<number>(0);
  const [bulkDiscount, setBulkDiscount] = React.useState<number>(0);
  const [open, setOpen] = React.useState(false);
  const [search, setSearch] = React.useState('');
  const dedbouncedSearch = useDebounced(search, 500);
  const [paymentForm, setPaymentForm, removePaymentForm] = usePersistedState<PaymentForm>(
    'payment',
    { ...defaultPaymentForm, header: { ...defaultPaymentForm.header, customer } },
    forceInitial,
  );
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({});
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { post, transform, processing, errors } = useForm({
    customer_id: 0,
    date: new Date(),
    lines: [],
  });

  useEffect(() => {
    if (!receivables) return;

    const { lines, rowSelection } = buildReceivableState(receivables, invoice_uuid);

    setRowSelection(rowSelection);
    setPaymentForm((prev) => ({
      ...prev,
      lines,
    }));
  }, [receivables, invoice_uuid, setPaymentForm]);

  useEffect(() => {
    const searchCustomer = () => {
      router.reload({ only: ['customers'], data: { search: dedbouncedSearch }, preserveUrl: true });
    };

    if (dedbouncedSearch) searchCustomer();
  }, [dedbouncedSearch]);

  const totalPaid = (): number => {
    return paymentForm.lines.reduce((acc, line) => {
      return acc + line.payment;
    }, 0);
  };

  const handleRecordPayment = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    const payload = {
      customer_id: paymentForm.header.customer?.uuid,
      date: paymentForm.header.date,
      amount: totalPaid(),
      notes: paymentForm.header.notes,
      lines: paymentForm.lines
        .filter((line) => line.payment > 0)
        .map((line) => {
          return { uuid: line.uuid, amount_due: line.amount_due, payment: line.payment, discount: line.discount };
        }),
      payment: paymentForm.payment,
    };
    transform((data) => ({
      ...data,
      ...payload,
    }));

    post('/payments', {
      ...headers,
      preserveState: 'errors',
      onSuccess: () => {
        removePaymentForm();
        router.get('/payments');
      },
    });
  };

  const handleCustomerSelection = (customer: Customer | undefined) => {
    setPaymentForm(() => {
      return { ...paymentForm, header: { ...paymentForm.header, customer }, lines: [] };
    });
    setOpen(false);
    if (customer !== undefined) {
      router.reload({
        only: ['receivables'],
        data: { customer_id: customer.uuid },
        preserveUrl: true,
        onStart: () => setLoading(true),
        onSuccess: (page) => {
          const receivables = page.props.receivables as Receivable[];
          // Reset before applying new data
          setRowSelection({});
          setPaymentForm((prev) => ({
            ...prev,
            lines: [],
          }));

          const { lines, rowSelection } = buildReceivableState(receivables, invoice_uuid);

          setRowSelection(rowSelection);
          setPaymentForm((prev) => ({
            ...prev,
            lines,
          }));
        },
        onFinish: () => setLoading(false),
      });
    }
  };

  const handleDateChange = (date: unknown) => {
    setPaymentForm(() => {
      return { ...paymentForm, header: { ...paymentForm.header, date: date as Date } };
    });
  };

  const handleCellChange = (rowId: string, columnId: string, newValue: string | number) => {
    setPaymentForm((prev) => {
      const lines = prev.lines.map((line) => {
        if (line.id.toString() === rowId) {
          const payment = columnId === 'payment' ? Number(newValue) : line.payment || 0;
          const discount = columnId === 'discount' ? Number(newValue) : line.discount || 0;

          return {
            ...line,
            [columnId]: columnId === 'payment' || columnId === 'discount' ? Number(newValue) : newValue,
            remaining: (line.amount_due || 0) - payment - discount,
          };
        }
        return line;
      });

      // recompute totals from updated lines
      const totals = lines.reduce(
        (acc, line) => {
          acc.totalPayment += line.payment || 0;
          acc.totalDiscount += line.discount || 0;
          acc.totalRemaining += line.remaining || 0;
          return acc;
        },
        { totalPayment: 0, totalDiscount: 0, totalRemaining: 0 },
      );

      return { ...prev, lines, totals };
    });

    // auto-select the row when edited
    setRowSelection((prev) => ({
      ...prev,
      [rowId]: true,
    }));
  };

  const onSelectionChange = (selection: RowSelectionState) => {
    paymentForm.lines = paymentForm.lines.map((line) => ({ ...line, payment: 0, balance: line.amount_due }));
    const selectedIds = Object.keys(selection);
    selectedIds.map((id) => {
      const index = paymentForm.lines.findIndex((l: ReceivableInvoiceForm) => l.id === Number(id));
      if (index === -1) return;
      paymentForm.lines[index].payment = paymentForm.lines[index].amount_due;
    });

    setPaymentForm((prev) => ({
      ...prev,
      lines: [...paymentForm.lines],
    }));
  };

  const performPaymentCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    router.get('/payments');
    setTimeout(() => {
      removePaymentForm();
    }, 300);
  };

  const handleCheckout = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    if (totalPaid() === 0) return;

    setCheckout(true);
  };

  const handleCheckoutChange = (method: PaymentMethod, form: CashForm | CheckForm | CardForm | BTForm) => {
    // Recalculate totals if the value is set.
    setPaymentForm(() => {
      return { ...paymentForm, payment: { ...paymentForm.payment, [method]: form } };
    });
  };

  const distributePayment = (amount: number, discount: number) => {
    let remaining = amount;

    setPaymentForm((prev) => {
      const updatedLines = prev.lines.map((line: ReceivableInvoiceForm) => {
        if (remaining <= 0) {
          return { ...line, payment: 0, discount: 0, remaining: line.amount_due };
        }

        const discountAmount = (line.amount_due * discount) / 100;
        const netDue = line.amount_due - discountAmount;

        if (remaining >= netDue) {
          // Fully pay this invoice
          remaining -= netDue;
          return { ...line, payment: netDue, discount: discountAmount, remaining: 0 };
        } else {
          const partialDiscount = (remaining * discount) / 100;
          const partialPayment = remaining;
          remaining = 0;
          return { ...line, payment: partialPayment, discount: partialDiscount, remaining: line.amount_due - (partialPayment + partialDiscount) };
        }
      });

      setRowSelection(buildRowSelection(updatedLines));
      return {
        ...prev,
        lines: updatedLines,
      };
    });
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={createPaymentBreadcrumbs}>
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
          <CustomerSection
            customer={paymentForm.header.customer}
            customers={customers}
            errors={errors}
            handleCustomerSelection={handleCustomerSelection}
            setSearch={setSearch}
            setOpen={setOpen}
            open={open}
            dedbouncedSearch={dedbouncedSearch}
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
                    onChange={(e) =>
                      setPaymentForm(() => {
                        return { ...paymentForm, header: { ...paymentForm.header, notes: e.currentTarget.value } };
                      })
                    }
                  />
                </div>
              </div>
            </div>
            <div className="col-span-6 flex flex-col items-end space-y-4">
              <div className="flex flex-col items-end space-y-2">
                <Label htmlFor="bulkPayment">{t('global.bulkPayment')}</Label>
                <MoneyInput name="bulkPayment" value={bulkPayment} onChange={(value) => setBulkPayment(value)} className="text-end" />
              </div>
              <div className="flex flex-col items-end space-y-2">
                <Label htmlFor="bulkDiscount">{t('global.bulkDiscount')}</Label>
                <Input
                  type="number"
                  name="bulkDiscount"
                  min={0}
                  max={100}
                  value={bulkDiscount}
                  onChange={(event) => {
                    let value = event.target.valueAsNumber;
                    if (value < 0) value = 0;
                    if (value > 100) value = 100; // clamp to max
                    setBulkDiscount(value);
                  }}
                  className="w-45 text-end"
                />
              </div>
              <Button disabled={bulkPayment === 0} onClick={() => distributePayment(bulkPayment, bulkDiscount)}>
                {t('payments.applyBulkPayment')}
              </Button>
            </div>
            <div className="col-span-12 grid place-items-end">
              <div className="flex flex-col gap-x-2">
                <Label className="text-muted-foreground block text-end text-lg">{t('global.totalReceived')}</Label>
                <Label className="block text-end text-4xl">{currency(totalPaid())}</Label>
              </div>
            </div>
          </div>
        </div>
        <div className="col-span-12">
          <List
            loading={loading}
            data={paymentForm.lines}
            totals={paymentForm.totals}
            rowSelection={rowSelection}
            setRowSelection={setRowSelection}
            onSelectPaymentLine={() => {}}
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
