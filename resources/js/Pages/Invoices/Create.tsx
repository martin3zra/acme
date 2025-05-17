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
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { usePersistedState } from '@/hooks/use-persisted-state';
import { useTranslation } from '@/hooks/use-translation';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { addDays, cn, isNotEmpty } from '@/lib/utils';
import { BTForm, CardForm, CashForm, CheckForm, Customer, InvoiceForm, Item, LineForm, PageProps, PaymentMethod, TaxReceipt } from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { format } from 'date-fns';
import { CalendarIcon } from 'lucide-react';
import React, { useCallback, useEffect } from 'react';
import { createBreadcrumbs, defaultDiscount, defaultInvoiceForm, paymentTerms } from './constants';
import CheckoutForm from './Shared/checkout-form';
import { CustomerSection } from './Shared/customer-section';
import { Lines } from './Shared/lines';

export default function Create({
  auth,
  customers,
  items,
  item,
  tax_receipts,
}: PageProps<{ customers: Customer[]; items: Item[]; item: Item; tax_receipts: TaxReceipt[] }>) {
  const t = useTranslation().trans;
  const currency = useNumber().currency;
  const [open, setOpen] = React.useState(false);
  const [openCancelConfirmation, setCancelConfirmation] = React.useState(false);
  const [openCheckout, setCheckout] = React.useState(false);
  const [isEditing, setEditing] = React.useState(false);
  const referenceInputRef = React.useRef<HTMLInputElement>(null);
  const qtyInputRef = React.useRef<HTMLInputElement>(null);
  const [search, setSearch] = React.useState('');
  const dedbouncedSearch = useDebounced(search, 500);
  const [amount, setAmount] = React.useState(0);
  const [invoiceForm, setInvoiceForm, removeInvoiceForm] = usePersistedState<InvoiceForm>('invoice', defaultInvoiceForm);
  const [currentItem, setCurrentItem] = React.useState<Item | undefined>(undefined);

  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { post, transform, processing, errors } = useForm({
    customer_id: 0,
    terms: 0,
    tax_receipt: 0,
    lines: [],
    date: new Date(),
    discount: defaultDiscount,
  });

  useEffect(() => setCurrentItem(item), [item]);

  const findCurrentItem = useCallback(() => {
    const exists = (element: LineForm) => element.id === currentItem?.id;
    const index = invoiceForm.lines.findIndex(exists);
    if (index >= 0) {
      setEditing(true);
      const line = invoiceForm.lines[index];
      setCurrentItem(line);
      qtyInputRef.current!.value = line.qty.toString();
      setAmount(line.amount);
    }
  }, [currentItem, invoiceForm.lines]);

  useEffect(() => {
    if (currentItem) {
      findCurrentItem();
      qtyInputRef.current?.focus();
    }
  }, [currentItem, findCurrentItem]);

  useEffect(() => {
    const searchCustomer = () => {
      router.reload({ only: ['customers'], data: { search: dedbouncedSearch }, preserveUrl: true });
    };

    if (dedbouncedSearch) {
      // Perform search operation
      searchCustomer();
    }
  }, [dedbouncedSearch]);

  const searchItem = (search: string) => {
    router.reload({
      only: ['item'],
      data: { search },
      preserveUrl: true,
      onSuccess: () => {
        qtyInputRef.current!.value = '1';
      },
    });
  };

  const handleOnSelectedItem = (item: Item) => {
    setCurrentItem(item);
    referenceInputRef.current!.value = item.name;
    qtyInputRef.current!.value = '1';
  };

  const handleOnKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' || event.key === 'Tab') {
      event.preventDefault(); // Prevent default behavior of Enter key
      if (event.currentTarget.name === 'reference' && isNotEmpty(event.currentTarget.value)) {
        searchItem(event.currentTarget.value);
        return;
      }
      if (event.currentTarget.name === 'qty' && currentItem != undefined) {
        processCurrentItem();
      }
    }
  };

  const processCurrentItem = () => {
    const line = currentItem!;

    if (isEditing) {
      const index = invoiceForm.lines.findIndex((element: LineForm) => element.id === line.id);
      if (index >= 0) {
        invoiceForm.lines[index].qty = qtyInputRef.current?.valueAsNumber || 0;
        invoiceForm.lines[index].amount = amount;
      }
      setEditing(false);
    } else {
      // When searching for the current item, if exists on the invoice, then display current values, and update the qty
      invoiceForm.lines.push({ ...line, qty: qtyInputRef.current?.valueAsNumber || 0, amount, action: 'added' });
    }
    setInvoiceForm(() => {
      return { ...invoiceForm, lines: [...invoiceForm.lines] };
    });

    resetInvoiceFormInput();
  };

  const resetInvoiceFormInput = () => {
    setCurrentItem(undefined);
    setAmount(0);
    referenceInputRef.current!.value = '';
    qtyInputRef.current!.value = '';
    referenceInputRef.current?.focus();
  };

  const handleCustomerSelection = (customer: Customer | undefined) => {
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, customer } };
    });

    setOpen(false);
  };

  const handleDateChange = (date: unknown) => {
    invoiceForm.header.date = date as Date;
    invoiceForm.header.due = undefined;
    if (invoiceForm.header.terms > 1) {
      invoiceForm.header.due = addDays(invoiceForm.header.date, invoiceForm.header.terms);
    }

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, date: date as Date } };
    });
  };

  const handlePaymentTermsChange = (value: string) => {
    invoiceForm.header.terms = Number(value);

    if (invoiceForm.header.terms > 1 && invoiceForm.header.date) {
      invoiceForm.header.due = addDays(invoiceForm.header.date, invoiceForm.header.terms);
    } else {
      invoiceForm.header.due = undefined;
    }

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, terms: Number(value) } };
    });
  };

  const handleTaxReceiptChange = (value: string) => {
    invoiceForm.header.taxReceipt = Number(value);

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, taxReceipt: Number(value) } };
    });
  };

  const handleRemoveLine = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    // add confirmation screen here.
    const index = parseInt(event.currentTarget.dataset.index || '-1');
    if (index < 0) return;
    const newItems = invoiceForm.lines.filter((_, i) => i !== index);
    setInvoiceForm(() => {
      return { ...invoiceForm, lines: newItems };
    });
  };

  const handleDiscountTypeChange = (value: 'fixed' | 'percentage') => {
    invoiceForm.header.discount.type = value;
    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, discount: { ...invoiceForm.header.discount, type: value } } };
    });
  };

  const handleDiscountValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    invoiceForm.header.discount.value = event.target.valueAsNumber;

    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, discount: { ...invoiceForm.header.discount, value: event.target.valueAsNumber } } };
    });
  };

  const performInvoiceCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    removeInvoiceForm();
    router.get('/invoices');
  };

  const handleCheckout = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    if (computeTotalAmount() === 0) return;
    if (invoiceForm.header.terms === 1) {
      Object.keys(propsErrors).forEach((key) => delete propsErrors[key]);
      setCheckout(true);
      return;
    }

    placedInvoice();
  };

  const placedInvoice = () => {
    transform((data) => ({
      ...data,
      customer_id: invoiceForm.header.customer?.id,
      date: invoiceForm.header.date,
      terms: invoiceForm.header.terms,
      tax_receipt: invoiceForm.header.taxReceipt,
      discount: invoiceForm.header.discount,
      notes: invoiceForm.header.notes || '',
      lines: invoiceForm.lines.map((line) => {
        return { id: line.id, qty: line.qty, unit: line.unit.id, price: line.price, rate: line.tax.rate, action: line.action };
      }),
      payment: invoiceForm.payment,
    }));
    post('/invoices', {
      ...headers,
      preserveState: 'errors',
      onSuccess: () => {
        removeInvoiceForm();
        router.get('/invoices');
      },
    });
  };

  const computeDiscount = (): number => {
    const discount = invoiceForm.header.discount;
    // Percentage calculation
    if (discount.type === 'percentage') {
      const total = composeSubTotal;
      return total * (discount.value / 100);
    }

    // Fixed calculation
    return discount.value;
  };

  const composeSubTotal = invoiceForm.lines.reduce((acc, line) => {
    return acc + line.amount;
  }, 0);

  const composeTax = invoiceForm.lines.reduce((acc, line) => {
    let discount = invoiceForm.header.discount.value;
    if (invoiceForm.header.discount.type === 'fixed') {
      discount = (discount / composeSubTotal) * 100;
    }

    const lineAmount = line.price * line.qty;
    const lineDiscount = lineAmount * (discount / 100);
    const tax = (lineAmount - lineDiscount) * (line.tax.rate / 100);

    return acc + tax;
  }, 0);

  const computeTotalAmount = (): number => {
    const discount = computeDiscount();

    return composeSubTotal - discount + composeTax;
  };

  const handleCheckoutChange = (method: PaymentMethod, form: CashForm | CheckForm | CardForm | BTForm) => {
    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return { ...invoiceForm, payment: { ...invoiceForm.payment, [method]: form } };
    });
  };

  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={createBreadcrumbs}>
      <AuthenticatedLayout.Actions>
        <div className="flex justify-end gap-x-6">
          <Button variant={'secondary'} onClick={() => setCancelConfirmation(true)}>
            {t('global.actions.cancel')}
          </Button>
          <Button onClick={handleCheckout} disabled={processing || computeTotalAmount() === 0}>
            {invoiceForm.header.terms === 1 ? t('global.actions.checkout') : t('global.actions.save')}
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
            customer={invoiceForm.header.customer}
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
                <Label htmlFor="date">{t('global.date')}</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant={'outline'}
                      className={cn('w-[280px] justify-start text-left font-normal', !invoiceForm.header.date && 'text-muted-foreground')}
                    >
                      <CalendarIcon />
                      {invoiceForm.header.date ? format(invoiceForm.header.date, 'PPP') : <span>{t('global.datePlaceholder')}</span>}
                    </Button>
                  </PopoverTrigger>
                  <PopoverContent className="w-auto p-0">
                    <Calendar
                      mode="single"
                      defaultMonth={invoiceForm.header.date}
                      selected={invoiceForm.header.date}
                      onSelect={handleDateChange}
                      initialFocus
                    />
                  </PopoverContent>
                </Popover>
                <InputError className="mt-2" message={errors.date} />
              </div>
              <div className="flex flex-col gap-y-2">
                <Label htmlFor="date">{t('global.dueDate')}</Label>
                <Label className="text-muted-foreground w-70 rounded-sm border p-2.5">
                  {invoiceForm.header.due ? format(invoiceForm.header.due, 'PPP') : t('global.noAvailable.default')}
                </Label>
              </div>
            </div>
            <div className="col-span-6 flex flex-col gap-y-6">
              <div className="flex flex-col gap-y-2">
                <Label htmlFor="paymentTerms">{t('invoices.paymentTerms')}</Label>
                <Select
                  name="paymentTerms"
                  onValueChange={handlePaymentTermsChange}
                  defaultValue={'0'}
                  value={String(invoiceForm.header.terms)}
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
              </div>
              <div className="flex flex-col gap-y-2">
                <Label htmlFor="paymentTerms">{t('invoices.taxReceipt')}</Label>
                <Select
                  name="paymentTerms"
                  onValueChange={handleTaxReceiptChange}
                  defaultValue={'0'}
                  value={String(invoiceForm.header.taxReceipt)}
                  required
                >
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select terms" />
                  </SelectTrigger>
                  <SelectContent className="">
                    {tax_receipts.map((receipt) => (
                      <SelectItem key={receipt.id} value={String(receipt.id)} disabled={!receipt.available}>
                        {receipt.name}
                        {!receipt.available && <span className="text-red-500">{t('global.limitReached')}</span>}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <InputError className="mt-2" message={errors.tax_receipt} />
              </div>
            </div>
          </div>
        </div>
        <div className="col-span-12">
          <div className="flex flex-col">
            <Lines
              items={items}
              lines={invoiceForm.lines}
              lineError={errors.lines}
              currentItem={currentItem}
              handleRemoveLine={handleRemoveLine}
              handleKeyDown={handleOnKeyDown}
              handleOnSelected={handleOnSelectedItem}
              amount={amount}
              setAmount={setAmount}
              referenceInputRef={referenceInputRef}
              qtyInputRef={qtyInputRef}
            />
          </div>
        </div>
        <div className="col-span-12 min-h-48">
          <div className="flex flex-col gap-y-2">
            <div className="grid grid-cols-12">
              <div className="col-span-10 flex flex-col gap-y-2 p-2">
                <Label className="text-sm/6 font-medium">{t('global.notes')}</Label>
                <Textarea
                  name="notes"
                  rows={4}
                  className="focus:no-data-focus:outline-none block w-1/2 resize-none rounded-lg border px-3 py-1.5 text-sm/6 data-focus:outline-2 data-focus:-outline-offset-2 data-focus:outline-white/25"
                  defaultValue={invoiceForm.header.notes}
                  onChange={(e) =>
                    setInvoiceForm(() => {
                      return { ...invoiceForm, header: { ...invoiceForm.header, notes: e.currentTarget.value } };
                    })
                  }
                />
              </div>
              <div className="col-span-2 flex flex-col gap-y-2 rounded-lg border border-gray-300/25 bg-gray-100/10">
                <div className="grid place-content-end gap-y-4 p-2">
                  {/* Add red border as the customer card, using data attributes */}
                  <InputError message={errors['discount.value']} />
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-base">{t('global.subTotal')}</span>
                    <span className="block text-base">{currency(composeSubTotal)}</span>
                  </div>
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-base">
                      {t('global.discount')}
                      {invoiceForm.header.discount.type === 'percentage' && (
                        <>
                          : <span className="text-muted-foreground text-xs">{currency(computeDiscount())}</span>
                        </>
                      )}
                    </span>
                    <div className="flex w-40 justify-end">
                      <Input
                        type="number"
                        min={0}
                        defaultValue={invoiceForm.header.discount.value}
                        name="discount"
                        className="w-16 text-end"
                        onChange={handleDiscountValueChange}
                      />
                      <Select
                        name="discountType"
                        onValueChange={handleDiscountTypeChange}
                        defaultValue={'percentage'}
                        value={String(invoiceForm.header.discount.type)}
                        required
                      >
                        <SelectTrigger className="w-16">
                          <SelectValue placeholder={t('global.discount')} />
                        </SelectTrigger>
                        <SelectContent className="">
                          <SelectItem value={'fixed'}>$</SelectItem>
                          <SelectItem value={'percentage'}>%</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-base">{t('global.tax')}</span>
                    <span className="block text-base">{currency(composeTax)}</span>
                  </div>
                  <Separator />
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-xl">{t('global.total')}</span>
                    <span className="block text-xl">{currency(computeTotalAmount())}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
        <AlertDialog open={openCancelConfirmation} onOpenChange={setCancelConfirmation}>
          <AlertDialogContent>
            <AlertDialogHeader>
              <AlertDialogTitle>{t('invoices.confirmsCancelation.title')}</AlertDialogTitle>
              <AlertDialogDescription>{t('invoices.confirmsCancelation.description')}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>{t('global.cancel')}</AlertDialogCancel>
              <AlertDialogAction onClick={performInvoiceCancelation}>{t('invoices.confirmsCancelation.confirm')}</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
        <CheckoutForm
          action={t('global.actions.save')}
          openCheckout={openCheckout}
          setCheckout={setCheckout}
          paymentForm={invoiceForm.payment}
          totalAmount={computeTotalAmount()}
          onCompleteCheckout={placedInvoice}
          processing={processing}
          setCancelConfirmation={setCancelConfirmation}
          errors={propsErrors}
          onCheckoutChange={handleCheckoutChange}
          currency={currency}
          t={t}
        />
      </div>

      {/* Command to search for customers */}
    </AuthenticatedLayout>
  );
}
