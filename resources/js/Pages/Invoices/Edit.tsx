import { AlertDestructive } from '@/components/alert-destructive';
import InputError from '@/components/input-error';
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
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { addDays, cn, isNotEmpty } from '@/lib/utils';
import { Auth, Customer, InvoiceForm, InvoiceWithLines, Item, LineForm, Nameable, PageProps } from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { format } from 'date-fns';
import { CalendarIcon } from 'lucide-react';
import React, { useCallback, useEffect, useState } from 'react';
import { defaultDiscount, editBreadcrumbs, paymentTerms } from './constants';
import { CustomerSection } from './Shared/customer-section';
import { Lines } from './Shared/lines';

export default function Edit({
  auth,
  invoice,
  customers,
  items,
  item,
  tax_receipts,
}: PageProps<{ auth: Auth; invoice: InvoiceWithLines; customers: Customer[]; items: Item[]; item: Item; tax_receipts: Nameable[] }>) {
  const { errors: propsErrors } = usePage<PageProps>().props;
  const [open, setOpen] = React.useState(false);
  const [openCancelConfirmation, setCancelConfirmation] = useState(false);
  const [openCheckout, setCheckout] = useState(false);
  const currency = useNumber().currency;
  const { headers } = useHeader();
  const { put, transform, processing, errors } = useForm({
    customer_id: 0,
    terms: 0,
    tax_receipt: 0,
    lines: [],
    date: new Date(),
    discount: defaultDiscount,
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
        return { ...line, amount: line.qty * line.price };
      }),
      payment: invoice.header.payment,
    };

    return _invoice;
  };

  const [invoiceForm, setInvoiceForm, removeInvoiceForm] = usePersistedState<InvoiceForm>('invoice_edit', initialAsInvoiceForm());

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
    if (invoiceForm.header.terms > 1) {
      invoiceForm.header.due = addDays(invoiceForm.header.date, invoiceForm.header.terms);
    }

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, date: date as Date } };
    });
  };

  const performUpdate = () => {
    transform((data) => ({
      ...data,
      customer_id: invoiceForm.header.customer?.id,
      date: invoiceForm.header.date,
      terms: invoiceForm.header.terms,
      tax_receipt: invoiceForm.header.taxReceipt,
      discount: invoiceForm.header.discount,
      notes: invoiceForm.header.notes || '',
      lines: invoiceForm.lines.map((line) => {
        return { id: line.id, unit: line.unit.id, qty: line.qty, price: line.price, rate: line.tax.rate, action: line.action };
      }),
      payment: invoiceForm.payment,
    }));

    put(`/invoices/${invoice.header.uuid}`, {
      ...headers,
      preserveState: 'errors',
      onSuccess: () => {
        removeInvoiceForm();
        router.get('/invoices');
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
      const index = invoiceForm.lines.findIndex((element: LineForm) => element.id === line.id);
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
    if (invoiceForm.header.terms === 1) {
      Object.keys(propsErrors).forEach((key) => delete propsErrors[key]);
      setCheckout(true);
      return;
    }

    performUpdate();
  };
  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={editBreadcrumbs}>
      <AuthenticatedLayout.Actions>
        <div className="flex justify-end gap-x-6">
          <Button variant={'secondary'} onClick={() => setCancelConfirmation(true)}>
            Cancel
          </Button>
          <Button onClick={handleCheckout} disabled={processing}>
            Update
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
            dedbouncedSearch={debouncedSearch}
          />
          <div className="grid grid-cols-12">
            <div className="col-span-6 flex flex-col gap-y-6">
              <div className="flex flex-col gap-y-2">
                <Label htmlFor="date">Date</Label>
                <Popover>
                  <PopoverTrigger asChild>
                    <Button
                      variant={'outline'}
                      className={cn('w-[280px] justify-start text-left font-normal', !invoiceForm.header.date && 'text-muted-foreground')}
                    >
                      <CalendarIcon />
                      {invoiceForm.header.date ? format(invoiceForm.header.date, 'PPP') : <span>Pick a date</span>}
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
                <Label htmlFor="date">Due Date</Label>
                <Label className="text-muted-foreground w-70 rounded-sm border p-2.5">
                  {invoiceForm.header.due ? format(invoiceForm.header.due, 'PPP') : 'Unknow'}
                </Label>
              </div>
            </div>
            <div className="col-span-6 flex flex-col gap-y-6">
              <div className="flex flex-col gap-y-2">
                <Label htmlFor="paymentTerms">Payment terms</Label>
                <Select
                  name="paymentTerms"
                  // onValueChange={handlePaymentTermsChange}
                  defaultValue={String(invoiceForm.header.terms)}
                  value={String(invoiceForm.header.terms)}
                  disabled={true}
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
                <Label htmlFor="paymentTerms">Tax Receipt</Label>
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
                <Label className="text-sm/6 font-medium">Notes</Label>
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
                    <span className="block text-base">Subtotal</span>
                    <span className="block text-base">{currency(composeSubTotal)}</span>
                  </div>
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-base">Discount</span>
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
                          <SelectValue placeholder="Discount" />
                        </SelectTrigger>
                        <SelectContent className="">
                          <SelectItem value={'fixed'}>$</SelectItem>
                          <SelectItem value={'percentage'}>%</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-base">Tax</span>
                    <span className="block text-base">{currency(composeTax)}</span>
                  </div>
                  <Separator />
                  <div className="flex w-60 items-center justify-between">
                    <span className="block text-xl">Total</span>
                    <span className="block text-xl">{currency(computeTotalAmount())}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </AuthenticatedLayout>
  );
}
