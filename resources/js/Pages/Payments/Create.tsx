import { AlertDestructive } from '@/components/alert-destructive';
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
import { Calendar } from '@/components/ui/calendar';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { usePersistedState } from '@/hooks/use-persisted-state';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { cn } from '@/lib/utils';
import { Customer, PageProps, PaymentForm, Receivable, ReceivableInvoiceForm } from '@/types';
import { router, useForm, usePage } from '@inertiajs/react';
import { format, formatDate } from 'date-fns/format';
import { CalendarIcon } from 'lucide-react';
import React, { useEffect } from 'react';
import { createPaymentBreadcrumbs } from '../Invoices/constants';
import { CustomerSection } from '../Invoices/Shared/customer-section';

const defaultPaymentForm: PaymentForm = { header: { customer: undefined, date: new Date() }, lines: [] };
export default function Create({ auth, customers, receivables }: PageProps<{ customers: Customer[]; receivables: Receivable[] }>) {
  const { currency } = useNumber();
  const [openCancelConfirmation, setCancelConfirmation] = React.useState(false);
  const [openCheckout, setCheckout] = React.useState(false);
  const [open, setOpen] = React.useState(false);
  const [search, setSearch] = React.useState('');
  const dedbouncedSearch = useDebounced(search, 500);
  const [paymentForm, setPaymentForm, removePaymentForm] = usePersistedState<PaymentForm>('payment', defaultPaymentForm);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { post, transform, processing, errors } = useForm({
    customer_id: 0,
    date: new Date(),
    lines: [],
  });

  useEffect(() => {
    if (receivables === undefined) return;
    const lines: ReceivableInvoiceForm[] = [];

    receivables.map((receivable) => {
      const line: ReceivableInvoiceForm = {
        ...receivable.invoice,
        payment: 0,
        discount: 0,
        balance: 0,
      };
      lines.push(line);
    });

    setPaymentForm((prev) => ({
      ...prev,
      lines,
    }));
  }, [receivables, setPaymentForm]);

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
      lines: paymentForm.lines
        .filter((line) => line.payment > 0)
        .map((line) => {
          return { uuid: line.uuid, amount_due: line.amount_due, payment: line.payment, discount: line.discount };
        }),
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
      router.reload({ only: ['receivables'], data: { customer_id: customer.uuid }, preserveUrl: true });
    }
  };

  const handleDateChange = (date: unknown) => {
    setPaymentForm(() => {
      return { ...paymentForm, header: { ...paymentForm.header, date: date as Date } };
    });
  };

  const handlePaymentAmountChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    event.preventDefault();
    const current = paymentForm.lines.filter((l) => l.uuid === event.currentTarget.id);
    if (current.length > 0) {
      current[0].payment = event.currentTarget.valueAsNumber;
      setPaymentForm((prev) => ({
        ...prev,
        lines: paymentForm.lines,
      }));
    }
  };

  const performPaymentCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    removePaymentForm();
    router.get('/payments');
  };

  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={createPaymentBreadcrumbs}>
      <AuthenticatedLayout.Actions>
        <div className="flex justify-end gap-x-6">
          <Button variant={'secondary'} onClick={() => setCancelConfirmation(true)}>
            Cancel
          </Button>
          <Button onClick={handleRecordPayment} disabled={totalPaid() === 0 || processing}>
            Record Payment
          </Button>
        </div>
      </AuthenticatedLayout.Actions>
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
                <Label htmlFor="date">Date</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant={'outline'}
                      className={cn('w-[280px] justify-start text-left font-normal', !paymentForm.header.date && 'text-muted-foreground')}
                    >
                      <CalendarIcon />
                      {paymentForm.header.date ? format(paymentForm.header.date, 'PPP') : <span>Pick a date</span>}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0">
                    <Calendar
                      mode="single"
                      defaultMonth={paymentForm.header.date}
                      selected={paymentForm.header.date}
                      onSelect={handleDateChange}
                      initialFocus
                    />
                  </PopoverContent>
                </Popover>
                <InputError className="mt-2" message={errors.date} />
              </div>
              {/* <div className="flex flex-col gap-y-2">
                <Label htmlFor="date">Due Date</Label>
                <Label className="text-muted-foreground w-70 rounded-sm border p-2.5">
                  {paymentForm.header.due ? format(paymentForm.header.due, 'PPP') : 'Unknow'}
                </Label>
              </div> */}
            </div>
            <div className="col-span-6 flex flex-col gap-y-6">
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
          <table
            className={cn(
              'w-full table-auto',
              '[&_th]:border-none [&_th]:border-gray-200 [&_th]:bg-gray-50/25 [&_th]:text-sm [&_th]:font-semibold [&_th]:uppercase',
              '[&_th]:p-2 [&_th]:text-start [&_th[data-format=number]]:text-end',
              '[&_td]:border-y [&_td]:p-2 [&_td]:text-start [&_td[data-format=number]]:w-36 [&_td[data-format=number]]:text-end',
            )}
          >
            <thead>
              <tr>
                <th>Invoice</th>
                <th>NCF</th>
                <th>Date</th>
                <th>Due On</th>
                <th data-format="number">Amount</th>
                <th data-format="number">Balance</th>
                <th data-format="number">Payment</th>
                <th data-format="number">Discount</th>
                <th data-format="number">Remaining</th>
              </tr>
            </thead>
            <tbody>
              {paymentForm.lines.map((line: ReceivableInvoiceForm) => (
                <tr key={line.id}>
                  <td>{line.number}</td>
                  <td>{line.ncf}</td>
                  <td>{formatDate(line.date, 'dd-MM-yyyy')}</td>
                  <td>{formatDate(line.due_on, 'dd-MM-yyyy')}</td>
                  <td data-format="number">{currency(line.total)}</td>
                  <td data-format="number">{currency(line.total - line.amount_due)}</td>
                  <td data-format="number">
                    <Input
                      className="border-none p-0 text-end text-lg"
                      type="number"
                      id={line.uuid}
                      value={line.payment}
                      max={line.amount_due}
                      min={0}
                      onChange={handlePaymentAmountChange}
                    />
                  </td>
                  <td data-format="number">{currency(line.discount)}</td>
                  <td data-format="number" className="font-medium text-red-600">
                    {currency(line.amount_due - line.payment)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <AlertDialog open={openCancelConfirmation} onOpenChange={setCancelConfirmation}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
              <AlertDialogDescription>
                This action cannot be undone. This will permanently delete this invoice and remove the data from our servers.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={performPaymentCancelation}>Yes, Continue</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
      </div>
    </AuthenticatedLayout>
  );
}
