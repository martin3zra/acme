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
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { usePersistedState } from '@/hooks/use-persisted-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { addDays, getDaysFromTerm, isNotEmpty } from '@/lib/utils';
import {
  Auth,
  BTForm,
  CardForm,
  CashForm,
  CheckForm,
  Customer,
  InvoiceForm,
  InvoiceWithLines,
  Item,
  LineForm,
  Nameable,
  PageProps,
  PaymentMethod,
  PaymentTermValue,
  TransactionKind,
} from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { format } from 'date-fns';
import React, { useCallback, useEffect, useState } from 'react';
import { defaultDiscount, makeEditBreadcrumbs, paymentTerms } from './constants';
import CheckoutForm from './Shared/checkout-form';
import { CustomerSection } from './Shared/customer-section';
import { Lines } from './Shared/lines';

export default function Edit({
  auth,
  invoice,
  customers,
  items,
  item,
  tax_receipts,
  kind,
}: PageProps<{
  auth: Auth;
  invoice: InvoiceWithLines;
  customers: Customer[];
  items: Item[];
  item: Item;
  tax_receipts: Nameable[];
  kind: TransactionKind;
}>) {
  const isInvoice = kind === 'invoice';
  const t = useTranslation().trans;
  const { errors: propsErrors } = usePage<PageProps>().props;
  const [open, setOpen] = React.useState(false);
  const [openCancelConfirmation, setCancelConfirmation] = useState(false);
  const [openCheckout, setCheckout] = useState(false);
  const currency = useNumber().currency;
  const { headers } = useHeader();
  const { put, transform, processing, errors } = useForm({
    customer_id: 0,
    terms: 'pia',
    tax_receipt: 0,
    lines: [],
    date: new Date(),
    discount: defaultDiscount,
    kind,
  });

  const initialAsInvoiceForm = (): InvoiceForm => {
    const _invoice: InvoiceForm = {
      header: {
        customer: invoice.header.customer,
        date: new Date(invoice.header.date),
        due: invoice.header.due_on ? new Date(invoice.header.due_on) : undefined,
        terms: invoice.header.terms,
        taxReceipt: invoice.header.tax_receipt_id,
        notes: invoice.header.notes,
        discount: invoice.header.discount,
      },
      lines: invoice.lines.map((line) => {
        return { ...line };
      }),
      payment: invoice.header.payment,
      kind: kind,
      source: { type: kind, id: '' },
    };

    return _invoice;
  };

  const [invoiceForm, setInvoiceForm, removeInvoiceForm] = usePersistedState<InvoiceForm>(`${kind}_edit`, initialAsInvoiceForm(), true);
  const [currentItem, setCurrentItem] = React.useState<Item | undefined>(undefined);
  const [isEditing, setEditing] = React.useState(false);
  const [search, setSearch] = React.useState('');
  const [amount, setAmount] = React.useState(0);
  const debouncedSearch = useDebounced(search, 500);
  const referenceInputRef = React.useRef<HTMLInputElement>(null);
  const qtyInputRef = React.useRef<HTMLInputElement>(null);

  useEffect(() => {
    const searchCustomer = () => {
      router.reload({ only: ['customers'], data: { search: debouncedSearch }, preserveUrl: true });
    };

    if (debouncedSearch) {
      // Perform search operation
      searchCustomer();
    }
  }, [debouncedSearch]);

  useEffect(() => setCurrentItem(item), [item]);

  const findCurrentItem = useCallback(() => {
    const exists = (element: LineForm) => element.id === currentItem?.id && element.variant_id === currentItem?.variant_id;
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

  const handleOnSelectedItem = (item: Item) => {
    setCurrentItem(item);
    referenceInputRef.current!.value = item.name;
    qtyInputRef.current!.value = '1';
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
    if (invoiceForm.header.terms !== 'pia') {
      const days = getDaysFromTerm(invoiceForm.header.terms);
      invoiceForm.header.due = addDays(invoiceForm.header.date, days);
    }

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, date: date as Date } };
    });
  };

  const handlePaymentTermsChange = (value: PaymentTermValue) => {
    invoiceForm.header.terms = value;

    if (invoiceForm.header.terms !== 'pia' && invoiceForm.header.date) {
      const days = getDaysFromTerm(invoiceForm.header.terms);
      invoiceForm.header.due = addDays(invoiceForm.header.date, days);
    } else {
      invoiceForm.header.due = undefined;
    }

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, terms: value } };
    });
  };

  const performUpdate = () => {
    transform((data) => {
      const payload: Record<string, any> = {
        ...data,
        customer_id: invoiceForm.header.customer?.id,
        date: invoiceForm.header.date,
        terms: invoiceForm.header.terms,
        // tax_receipt: invoiceForm.header.taxReceipt,
        discount: invoiceForm.header.discount,
        notes: invoiceForm.header.notes || '',
        kind: kind,
        lines: invoiceForm.lines.map((line) => {
          return { id: line.id, variant_id: line.variant_id, unit: line.unit.id, qty: line.qty, price: line.price, rate: line.tax.rate, action: line.action };
        }),
        // payment: invoiceForm.payment,
      };

      if (isInvoice) {
        // payload.terms = invoiceForm.header.terms;
        payload.tax_receipt = invoiceForm.header.taxReceipt;
        payload.discount = invoiceForm.header.discount;
        payload.payment = invoiceForm.payment;
        payload.source = invoiceForm.source;
      } else {
        payload.source = { type: 'template', id: '' };
      }

      if (invoiceForm.kind === 'template') {
        payload.terms = invoiceForm.header.terms;
        payload.tax_receipt = invoiceForm.header.taxReceipt;
        payload.discount = invoiceForm.header.discount;
      }

      if (invoiceForm.kind !== 'template') {
        payload.recurrence = null;
      }

      return payload;
    });

    put(`/${kind}s/${invoice.header.uuid}`, {
      ...headers,
      preserveState: 'errors',
      onSuccess: () => {
        removeInvoiceForm();
        router.get(`/${kind}s`);
      },
    });
  };

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

  const handleRemoveLine = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    // add confirmation screen here.
    const index = parseInt(event.currentTarget.dataset.index || '-1');
    if (index < 0) return;
    // const newItems = invoiceForm.lines.filter((_, i) => i !== index);
    const newItems = invoiceForm.lines;
    newItems[index].action = 'deleted';
    setInvoiceForm(() => {
      return { ...invoiceForm, lines: newItems };
    });
  };

  const processCurrentItem = () => {
    const line = currentItem!;

    if (isEditing) {
      const index = invoiceForm.lines.findIndex((element: LineForm) => element.id === line.id && element.variant_id === line.variant_id);
      if (index >= 0) {
        invoiceForm.lines[index].qty = qtyInputRef.current?.valueAsNumber || 0;
        invoiceForm.lines[index].amount = amount;
        invoiceForm.lines[index].action = 'updated';
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

  const resetInvoiceFormInput = () => {
    setCurrentItem(undefined);
    setAmount(0);
    referenceInputRef.current!.value = '';
    qtyInputRef.current!.value = '';
    referenceInputRef.current?.focus();
  };

  const computeDiscount = (): number => {
    const discount = invoiceForm.header.discount;
    // Percentage calculation
    if (discount.type === 'percentage') {
      const total = composeSubTotal + composeTax;
      return total * (discount.value / 100);
    }

    // Fixed calculation
    return discount.value;
  };

  const composeSubTotal = invoiceForm.lines
    .filter((line) => line.action !== 'deleted')
    .reduce((acc, line) => {
      return acc + line.amount;
    }, 0);

  const composeTax = invoiceForm.lines
    .filter((line) => line.action !== 'deleted')
    .reduce((acc, line) => {
      const tax = line.price * (line.tax.rate / 100);
      return acc + tax * line.qty;
    }, 0);

  const computeTotalAmount = (): number => {
    return composeSubTotal + composeTax - computeDiscount();
  };

  const handleDiscountValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    invoiceForm.header.discount.value = event.target.valueAsNumber;

    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, discount: { ...invoiceForm.header.discount, value: event.target.valueAsNumber } } };
    });
  };

  const handleDiscountTypeChange = (value: 'fixed' | 'percentage') => {
    invoiceForm.header.discount.type = value;
    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, discount: { ...invoiceForm.header.discount, type: value } } };
    });
  };

  const handleCheckout = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    if (computeTotalAmount() === 0) return;
    if (invoiceForm.header.terms === 'pia') {
      Object.keys(propsErrors).forEach((key) => delete propsErrors[key]);
      setCheckout(true);
      return;
    }

    performUpdate();
  };
  const handleCheckoutChange = (method: PaymentMethod, form: CashForm | CheckForm | CardForm | BTForm) => {
    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return { ...invoiceForm, payment: { ...invoiceForm.payment, [method]: form } };
    });
  };

  const performInvoiceCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    router.get(`/${kind}s`);
    setTimeout(() => {
      removeInvoiceForm();
    }, 200);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={makeEditBreadcrumbs(kind)}>
      <AppLayout.Actions>
        <div className="flex justify-end gap-x-6">
          <Button variant={'secondary'} onClick={() => setCancelConfirmation(true)}>
            {t('global.actions.cancel')}
          </Button>
          {isInvoice && (
            <Button onClick={handleCheckout} disabled={processing}>
              {invoiceForm.header.terms === 'pia' ? t('global.actions.checkout') : t('global.actions.update')}
            </Button>
          )}
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
            customer={invoiceForm.header.customer}
            customers={customers}
            errors={errors}
            handleCustomerSelection={handleCustomerSelection}
            setSearch={setSearch}
            setOpen={setOpen}
            open={open}
            dedbouncedSearch={debouncedSearch}
          />
          <div className="grid grid-cols-12">
            <div className="col-span-6 flex flex-col gap-y-6">
              <div className="flex flex-col gap-y-2">
                <DatePickerField
                  id="date"
                  label={t('global.date')}
                  placeholder={t('global.datePlaceholder')}
                  value={invoiceForm.header.date}
                  onChange={handleDateChange}
                  error={errors.date}
                  className="w-52"
                />
              </div>
              {isInvoice && (
                <div className="flex flex-col gap-y-2">
                  <Label htmlFor="date">{t('global.dueDate')}</Label>
                  <Label className="text-muted-foreground w-70 rounded-sm border p-2.5">
                    {invoiceForm.header.due ? format(invoiceForm.header.due, 'PPP') : t('global.noAvailable.default')}
                  </Label>
                </div>
              )}
            </div>
            {/* Check here, what data do we need to display */}
            {isInvoice ||
              (kind === 'order' && (
                <div className="col-span-6 flex flex-col gap-y-6">
                  <div className="flex flex-col gap-y-2">
                    <Label htmlFor="paymentTerms">{t(`${kind}s.paymentTerms`)}</Label>
                    <Select
                      name="paymentTerms"
                      onValueChange={handlePaymentTermsChange}
                      defaultValue={String(invoiceForm.header.terms)}
                      value={String(invoiceForm.header.terms)}
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
                  {isInvoice && (
                    <div className="flex flex-col gap-y-2">
                      <Label htmlFor="paymentTerms">{t('invoices.taxReceipt')}</Label>
                      <Select
                        name="paymentTerms"
                        // onValueChange={handleTaxReceiptChange}
                        defaultValue={String(invoiceForm.header.taxReceipt)}
                        value={String(invoiceForm.header.taxReceipt)}
                        disabled={true}
                      >
                        <SelectTrigger className="w-full">
                          <SelectValue placeholder="Select terms" />
                        </SelectTrigger>
                        <SelectContent className="">
                          {tax_receipts.map((receipt) => (
                            <SelectItem key={receipt.id} value={String(receipt.id)}>
                              {receipt.name}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                      <InputError className="mt-2" message={errors.tax_receipt} />
                    </div>
                  )}
                </div>
              ))}
            {!isInvoice && (
              <div className="col-span-12 flex flex-col place-items-end gap-y-6">
                <Button disabled={invoiceForm.lines.length === 0} onClick={performUpdate}>
                  {t('global.actions.save')}
                </Button>
              </div>
            )}
          </div>
        </div>

        <div className="col-span-12">
          <div className="flex flex-col">
            <Lines
              kind={kind}
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
              <div className="col-span-10 flex flex-col gap-y-2 py-2">
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
                  <InputError message={errors['discount']} />
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-base">{t('global.subTotal')}</span>
                    <span className="block text-base">{currency(composeSubTotal)}</span>
                  </div>
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-base">{t('global.discount')}</span>
                    <div className="flex w-40 justify-end">
                      <Input
                        type="number"
                        min={0}
                        defaultValue={invoiceForm.header.discount.value}
                        name="discount"
                        className="w-20 text-end"
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
              <AlertDialogTitle>{t(`${kind}s.confirmsCancelation.title`)}</AlertDialogTitle>
              <AlertDialogDescription>{t(`${kind}s.confirmsCancelation.description`)}</AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>{t(`${kind}s.confirmsCancelation.cancel`)}</AlertDialogCancel>
              <AlertDialogAction onClick={performInvoiceCancelation}>{t(`${kind}s.confirmsCancelation.confirm`)}</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
        {isInvoice && (
          <CheckoutForm
            action={t('global.actions.update')}
            openCheckout={openCheckout}
            setCheckout={setCheckout}
            paymentForm={invoiceForm.payment}
            totalAmount={computeTotalAmount()}
            onCompleteCheckout={performUpdate}
            processing={processing}
            setCancelConfirmation={setCancelConfirmation}
            errors={propsErrors}
            onCheckoutChange={handleCheckoutChange}
            currency={currency}
            t={t}
          />
        )}
      </div>
    </AppLayout>
  );
}
