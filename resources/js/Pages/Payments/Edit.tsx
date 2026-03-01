import { AlertDestructive } from '@/components/alert-destructive';
import { DatePickerField } from '@/components/date-picker';
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
import {
  BTForm,
  CardForm,
  CashForm,
  CheckForm,
  Customer,
  mapPaymentLineToReceivableInvoice,
  PageProps,
  PaymentForm,
  PaymentMethod,
  PaymentVerb,
  PaymentWithLines,
  Receivable,
  ReceivableInvoiceForm,
} from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { RowSelectionState } from '@tanstack/table-core/build/lib/features/RowSelection';
import React, { useEffect } from 'react';
import CheckoutForm from '../Invoices/Shared/checkout-form';
import { CustomerSection } from '../Invoices/Shared/customer-section';
import { List } from './Shared/lines-payment';
import { editPaymentBreadcrumbs } from './constants';

type FlagSet = {
  [key: string]: boolean;
};

export default function Edit({
  auth,
  payment,
  customers,
  receivables,
  invoice_uuid,
}: PageProps<{
  payment: PaymentWithLines;
  customers: Customer[];
  receivables: Receivable[];
  invoice_uuid: string;
}>) {
  const t = useTranslation().trans;
  const { currency } = useNumber();
  const [openCancelConfirmation, setCancelConfirmation] = React.useState(false);
  const [openCheckout, setCheckout] = React.useState(false);

  const [initialized, setInitialized] = React.useState(false);
  const [open, setOpen] = React.useState(false);
  const [search, setSearch] = React.useState('');
  const dedbouncedSearch = useDebounced(search, 500);

  const initialAsPaymentForm = (): PaymentForm => {
    const _payment: PaymentForm = {
      header: {
        customer: payment.header.customer,
        date: new Date(payment.header.date),
        notes: payment.header.notes,
        discount: 0,
      },
      lines: payment.lines.map((line) => {
        return mapPaymentLineToReceivableInvoice(line);
      }),
      payment: payment.header.payment,
    };
    return _payment;
  };
  const [paymentForm, setPaymentForm, removePaymentForm] = usePersistedState<PaymentForm>('payment', initialAsPaymentForm(), true);
  const [rowSelection, setRowSelection] = React.useState<RowSelectionState>({});
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { put, transform, processing, errors } = useForm({
    customer_id: 0,
    date: new Date(),
    lines: [],
  });

  useEffect(() => {
    const _rowSelection: FlagSet = {};
    paymentForm.lines
      .filter((line) => line.payment > 0)
      .map((line) => {
        _rowSelection[`${line.id.toString()}`] = true;
      });

    if (Object.keys(_rowSelection).length > 0) {
      setRowSelection(_rowSelection);
    }
    if (receivables === undefined || initialized) return;

    const lines: ReceivableInvoiceForm[] = [];
    let selectedRowId = -1;
    receivables.map((receivable) => {
      selectedRowId = invoice_uuid === receivable.invoice.uuid ? receivable.invoice.id : -1;
      const line: ReceivableInvoiceForm = {
        ...receivable.invoice,
        payment: invoice_uuid === receivable.invoice.uuid ? receivable.invoice.amount_due : 0,
        discount: 0,
        balance: 0,
        action: 'unchanged',
      };
      lines.push(line);
    });

    if (selectedRowId > 0) {
      setRowSelection((prev) => ({
        ...prev,
        [`${selectedRowId.toString()}`]: true,
      }));
    }

    setPaymentForm((prev) => ({
      ...prev,
      lines: [...lines],
    }));

    setInitialized(true);
  }, [receivables, paymentForm, setPaymentForm, invoice_uuid, initialized]);

  useEffect(() => {
    const searchCustomer = () => {
      router.reload({ only: ['customers'], data: { search: dedbouncedSearch }, preserveUrl: true });
    };

    if (dedbouncedSearch) searchCustomer();
  }, [dedbouncedSearch]);

  const totalPaid = (): number => {
    return paymentForm.lines
      .filter((line) => line.action !== 'deleted')
      .reduce((acc, line) => {
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
          return { id: line.id, uuid: line.uuid, amount_due: line.amount_due, payment: line.payment, discount: line.discount, action: line.action };
        }),
      payment: paymentForm.payment,
    };
    transform((data) => ({
      ...data,
      ...payload,
    }));

    put(`/payments/${payment.header.uuid}`, {
      ...headers,
      preserveState: 'errors',
      onSuccess: () => {
        removePaymentForm();
        router.get('/payments');
      },
    });
  };

  const onSelectPaymentLine = (line: ReceivableInvoiceForm, action: PaymentVerb) => {
    if (action !== 'trash') return;
    const index = paymentForm.lines.findIndex((l: ReceivableInvoiceForm) => l.uuid === line.uuid);
    if (index === -1) return;

    paymentForm.lines[index].payment = line.payment;
    paymentForm.lines[index].action = 'deleted';
    setPaymentForm((prev) => ({
      ...prev,
      lines: [...paymentForm.lines],
    }));
  };

  const handleCustomerSelection = (customer: Customer | undefined) => {
    setPaymentForm(() => {
      return { ...paymentForm, header: { ...paymentForm.header, customer }, lines: [] };
    });
    setOpen(false);
    if (customer !== undefined) {
      router.reload({ only: ['receivables'], data: { customer_id: customer.uuid }, preserveUrl: true });
    }
  };

  const handleDateChange = (date: unknown) => {
    setPaymentForm(() => {
      return { ...paymentForm, header: { ...paymentForm.header, date: date as Date } };
    });
  };

  const handleCellChange = (inputId: string, newValue: string | number) => {
    const index = paymentForm.lines.findIndex((l: ReceivableInvoiceForm) => l.uuid === inputId);
    if (index === -1) return;

    setRowSelection((prev) => ({
      ...prev,
      [`${paymentForm.lines[index].id.toString()}`]: true,
    }));
    paymentForm.lines[index].payment = Number(newValue);
    paymentForm.lines[index].action = 'updated';
    setPaymentForm((prev) => ({
      ...prev,
      lines: [...paymentForm.lines],
    }));
  };

  const onSelectionChange = (selection: RowSelectionState) => {
    paymentForm.lines = paymentForm.lines.map((line) => ({ ...line, payment: 0, balance: line.amount_due }));
    const selectedIds = Object.keys(selection);
    selectedIds.map((id) => {
      const index = paymentForm.lines.findIndex((l: ReceivableInvoiceForm) => l.id === Number(id));
      if (index === -1) return;
      paymentForm.lines[index].payment = paymentForm.lines[index].amount_due;
      paymentForm.lines[index].action = 'updated';
    });

    setPaymentForm((prev) => ({
      ...prev,
      lines: [...paymentForm.lines],
    }));
  };

  const performPaymentCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    removePaymentForm();
    router.get('/payments');
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

  return (
    <AppLayout user={auth.user} breadcrumbs={editPaymentBreadcrumbs}>
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
                  className="w-52"
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
            <div className="col-span-6 grid place-items-end">
              <div className="flex flex-col gap-x-2">
                <Label className="text-muted-foreground block text-end text-lg">{t('global.totalReceived')}</Label>
                <Label className="block text-end text-4xl">{currency(totalPaid())}</Label>
              </div>
              {/* <div className="flex flex-col gap-y-2">
                <Label htmlFor="paymentTerms">Payment terms</Label>
                <Select
                  name="paymentTerms"
                  onValueChange={handlePaymentTermsChange}
                  defaultValue={'0'}
                  value={String(paymentForm.header.terms)}
                  required
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select terms" />
                  </SelectTrigger>
                  <SelectContent className="">
                    {paymentTerms.map((term, index) => (
                      <SelectItem key={index.toString()} value={term.value.toString()}>
                        {term.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <InputError className="mt-2" message={errors.terms} />
              </div> */}
              {/* <div className="flex flex-col gap-y-2">
                <Label htmlFor="paymentTerms">Tax Receipt</Label>
                <Select
                  name="paymentTerms"
                  onValueChange={handleTaxReceiptChange}
                  defaultValue={'0'}
                  value={String(paymentForm.header.taxReceipt)}
                  required
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select terms" />
                  </SelectTrigger>
                  <SelectContent className="">
                    {tax_receipts.map((receipt) => (
                      <SelectItem key={receipt.id} value={String(receipt.id)} disabled={!receipt.available}>
                        {receipt.name}
                        {!receipt.available && <span className="text-red-500">Limit reached</span>}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <InputError className="mt-2" message={errors.tax_receipt} />
              </div> */}
            </div>
          </div>
        </div>
        <div className="col-span-12">
          <List
            data={paymentForm.lines.filter((line) => line?.action !== 'deleted')}
            rowSelection={rowSelection}
            setRowSelection={setRowSelection}
            onSelectPaymentLine={onSelectPaymentLine}
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
              <AlertDialogCancel>{t('payments.confirmsCancelation.cancel')}</AlertDialogCancel>
              <AlertDialogAction onClick={performPaymentCancelation}>{t('payments.confirmsCancelation.confirm')}</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
        <CheckoutForm
          action={t('global.actions.update')}
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
