import { AlertDestructive } from '@/components/alert-destructive';
import InputError from '@/components/input-error';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog';
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
import { useLocalStorage } from '@/hooks/use-local-storage';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { addDays, cn, isNotEmpty } from '@/lib/utils';
import { BTForm, CardForm, CashForm, CheckForm, Customer, DiscountType, Item, LineForm, Nameable, PageProps, PaymentForm, PaymentMethod } from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { format } from 'date-fns';
import { CalendarIcon } from 'lucide-react';
import React, { useCallback, useEffect } from 'react';
import { breadcrumbs, defaultBTForm, defaultCardForm, defaultCashForm, defaultCheckForm, paymentTerms } from './constants';
import { CustomerSection } from './Shared/customer-section';
import { Lines } from './Shared/lines';
import CheckoutForm from './Shared/checkout-form';

type HeaderForm = {
  customer: Customer | undefined;
  date: Date | undefined;
  due: Date | undefined;
  terms: number;
  taxReceipt: number;
  notes: string | undefined;
  discount: DiscountType;
};

type InvoiceForm = {
  header: HeaderForm;
  lines: LineForm[];
  payment: PaymentForm
};

const defaultPaymentForm: PaymentForm = {cash: defaultCashForm, ck: defaultCheckForm, card: defaultCardForm, bt: defaultBTForm}
const defaultDiscount: DiscountType = {value: 0, type: "fixed"}
const defaultHeaderForm: HeaderForm =  { customer: undefined, date: undefined, due: undefined, terms: 0, taxReceipt: 0, notes: undefined, discount: defaultDiscount}
const defaultInvoiceForm: InvoiceForm = { header: defaultHeaderForm, lines: [], payment: defaultPaymentForm }

const { setItem: storageInvoiceForm, getItem: getStorageInvoiceForm, removeItem: removeStorageIvoinceForm } = useLocalStorage('invoice');
const getInvoiceFromStorage = () => {
  return getStorageInvoiceForm() || defaultInvoiceForm;
}

export default function Create({ auth, customers, items, item, tax_receipts }: PageProps<{ customers: Customer[]; items: Item[]; item: Item, tax_receipts: Nameable[] }>) {
  const currency = useNumber().currency;
  const [open, setOpen] = React.useState(false);
  const [openCancelConfirmation, setCancelConfirmation] = React.useState(false);
  const [openCheckout, setCheckout] = React.useState(false);
  const [isEditing, setEditing] = React.useState(false);
  const referenceInputRef = React.useRef<HTMLInputElement>(null);
  const qtyInputRef = React.useRef<HTMLInputElement>(null);
  const [search, setSearch] = React.useState('');
  const [notes, setNotes] = React.useState('');
  const dedbouncedSearch = useDebounced(search, 500);
  const dedbouncedNotes = useDebounced(notes, 500);
  const [amount, setAmount] = React.useState(0);
  const [invoiceForm, setInvoiceForm] = React.useState<InvoiceForm>(getInvoiceFromStorage);
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
      qtyInputRef.current!.value = line.quantity.toString();
      setAmount(line.amount);
    }
  }, [currentItem, invoiceForm.lines]);

  useEffect(() => {
    if (currentItem) {
      findCurrentItem();
      qtyInputRef.current?.focus();
    }
  }, [currentItem, findCurrentItem]);

  const synInvoiceForm = useCallback(() => {
    storageInvoiceForm(invoiceForm)
  }, [invoiceForm])

  useEffect(() => synInvoiceForm(), [invoiceForm]);

  useEffect(() => {
    const searchCustomer = () => {
      router.reload({ only: ['customers'], data: { search: dedbouncedSearch }, preserveUrl: true });
    };

    if (dedbouncedSearch) {
      // Perform search operation
      searchCustomer();
    }
  }, [dedbouncedSearch]);

  useEffect(() => {
    if (dedbouncedNotes) {
      setInvoiceForm(() => {
        return { ...invoiceForm, header: { ...invoiceForm.header, notes: dedbouncedNotes } };
      });
    }
  }, [dedbouncedNotes])

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
    setCurrentItem(item)
    referenceInputRef.current!.value = item.name
    qtyInputRef.current!.value = '1';
  }

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' || event.key === 'Tab') {
      event.preventDefault(); // Prevent default behavior of Enter key
      if (event.currentTarget.name === 'reference' && isNotEmpty(event.currentTarget.value)) {
        searchItem(event.currentTarget.value);
        return
      }
      if (event.currentTarget.name === 'quantity' && currentItem != undefined) {
        processCurrentItem()
      }
    }
  };

  const processCurrentItem = () => {

    const line = currentItem!;

    if (isEditing) {
      const index = invoiceForm.lines.findIndex((element: LineForm) => element.id === line.id);
      if (index >= 0) {
        invoiceForm.lines[index].quantity = qtyInputRef.current?.valueAsNumber || 0;
        invoiceForm.lines[index].amount = amount;
      }
      setEditing(false);
    } else {
      // When searching for the current item, if exists on the invoice, then display current values, and update the quantity
      invoiceForm.lines.push({ ...line, quantity: qtyInputRef.current?.valueAsNumber || 0, amount });
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
    invoiceForm.header.due = undefined
    if (invoiceForm.header.terms > 1) {
      invoiceForm.header.due = addDays(invoiceForm.header.date, invoiceForm.header.terms)
    }

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, date: date as Date } };
    });
  };

  const handlePaymentTermsChange = (value: string) => {
    invoiceForm.header.terms = Number(value)

    if (invoiceForm.header.terms > 1 && invoiceForm.header.date) {
      invoiceForm.header.due = addDays(invoiceForm.header.date, invoiceForm.header.terms)
    } else {
      invoiceForm.header.due = undefined
    }

    setInvoiceForm(() => {
      return {...invoiceForm, header: {...invoiceForm.header, terms: Number(value)}}
    })
  }

  const handleTaxReceiptChange = (value: string) => {
    invoiceForm.header.taxReceipt = Number(value)

    setInvoiceForm(() => {
      return {...invoiceForm, header: {...invoiceForm.header, taxReceipt: Number(value)}}
    })
  }

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

  const handleDiscountTypeChange = (value: "fixed" | "percentage") => {
    invoiceForm.header.discount.type = value
    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return {...invoiceForm, header: {...invoiceForm.header, discount: {...invoiceForm.header.discount, type: value}}}
    })
  }

  const handleDiscountValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    invoiceForm.header.discount.value = event.target.valueAsNumber

    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return {...invoiceForm, header: {...invoiceForm.header, discount: {...invoiceForm.header.discount, value: event.target.valueAsNumber}}}
    })
  };

  const performInvoiceCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    removeStorageIvoinceForm();
    router.get('/invoices')
  }

  const handleCheckout = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault()
    if (computeTotalAmount()=== 0) return
    if (invoiceForm.header.terms === 1 ) {
      Object.keys(propsErrors).forEach(key => delete propsErrors[key]);
      setCheckout(true)
      return
    }

    placedInvoice()
  }

  const placedInvoice = () => {
    // Check if the total and payment match on cash terms.
    transform((data) => ({
      ...data,
      customer_id: invoiceForm.header.customer?.id,
      date: invoiceForm.header.date,
      terms: invoiceForm.header.terms,
      tax_receipt: invoiceForm.header.taxReceipt,
      discount: invoiceForm.header.discount,
      notes: invoiceForm.header.notes || '',
      lines: invoiceForm.lines.map((line) => {
        return { id: line.id, quantity: line.quantity, unit: line.unit.id, price: line.price, rate: line.tax.rate}
      }),
      payment: invoiceForm.payment,
    }))
    post('/invoices', {...headers, preserveState: "errors", onSuccess: () => {
      removeStorageIvoinceForm();
      router.get('/invoices')
    }})
  }

  const computeDiscount = (): number => {
    const discount = invoiceForm.header.discount
    // Percentage calculation
    if (discount.type === "percentage") {
      const total = composeSubTotal + composeTax
      return (total * (discount.value / 100))
    }

    // Fixed calculation
    return discount.value
  }

  const composeSubTotal = invoiceForm.lines.reduce((acc, line) => {
      return acc + line.amount;
    }, 0);

  const composeTax = invoiceForm.lines.reduce((acc, line) => {
    const tax = line.price * (line.tax.rate / 100);
    return acc + (tax * line.quantity);
  }, 0);

  const computeTotalAmount = (): number => {
    return (composeSubTotal + composeTax) - computeDiscount()
  }

  const handleCheckoutChange = (method: PaymentMethod, form: CashForm | CheckForm | CardForm | BTForm) => {
    // Recalculate totals if the value is set.
    setInvoiceForm(() => {
      return {...invoiceForm, payment: {...invoiceForm.payment, [method]:form}}
    })
  }

  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <AuthenticatedLayout.Actions>
        <div className='flex justify-end gap-x-6'>
          <Button variant={"secondary"} onClick={() => setCancelConfirmation(true)}>Cancel</Button>
          <Button onClick={handleCheckout} disabled={processing || (computeTotalAmount()=== 0)}>Checkout</Button>
        </div>
      </AuthenticatedLayout.Actions>
      <div className="grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4 bg-gray-50/10">
        {!openCheckout && propsErrors.status && <div className="col-span-12"><AlertDestructive description={propsErrors.status} onDestroy={() => delete propsErrors.status }/></div>}
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
                <Label className='w-70 border p-2.5 rounded-sm text-muted-foreground'>
                  {invoiceForm.header.due ? format(invoiceForm.header.due, 'PPP') : 'Unknow'}
                </Label>
              </div>
            </div>
            <div className="col-span-6 flex flex-col gap-y-6">
              <div className="flex flex-col gap-y-2">
                <Label htmlFor='paymentTerms'>Payment terms</Label>
                <Select
                  name='paymentTerms'
                  onValueChange={handlePaymentTermsChange}
                  defaultValue={"0"}
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
                <Label htmlFor='paymentTerms'>Tax Receipt</Label>
                <Select
                  name='paymentTerms'
                  onValueChange={handleTaxReceiptChange}
                  defaultValue={"0"}
                  value={String(invoiceForm.header.taxReceipt)}
                  required
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
              handleKeyDown={handleKeyDown}
              handleOnSelected={handleOnSelectedItem}
              amount={amount}
              setAmount={setAmount}
              referenceInputRef={referenceInputRef}
              qtyInputRef={qtyInputRef}
            />
          </div>
        </div>
        <div className="col-span-12 min-h-48">
          <div className='flex flex-col gap-y-2'>
            <div className='grid grid-cols-12'>
              <div className="col-span-10 flex flex-col gap-y-2 p-2">
                <Label className='text-sm/6 font-medium'>Notes</Label>
                <Textarea
                  name='notes'
                  rows={4}
                  className="block w-1/2 rounded-lg text-sm/6 resize-none border px-3 py-1.5 focus:no-data-focus:outline-none data-focus:outline-2 data-focus:-outline-offset-2 data-focus:outline-white/25"
                  defaultValue={invoiceForm.header.notes}
                  onChange={(e) => setNotes(e.currentTarget.value)}
                  />
              </div>
              <div className="col-span-2 flex flex-col gap-y-2 rounded-lg border border-gray-300/25 bg-gray-100/10">
                <div className='grid place-content-end p-2 gap-y-4'>
                  {/* Add red border as the customer card, using data attributes */}
                  <InputError message={errors["discount.value"]} />
                  <div className='flex justify-between items-center w-60'>
                    <span className="block text-base">Subtotal</span>
                    <span className="block text-base">{currency(composeSubTotal)}</span>
                  </div>
                  <div className='flex justify-between items-center w-60'>
                    <span className="block text-base">Discount</span>
                    <div className='flex justify-end w-40'>
                      <Input
                        type="number"
                        min={0}
                        defaultValue={invoiceForm.header.discount.value}
                        name="discount"
                        className="text-end w-20"
                        onChange={handleDiscountValueChange}
                      />
                      <Select
                        name='discountType'
                        onValueChange={handleDiscountTypeChange}
                        defaultValue={"percentage"}
                        value={String(invoiceForm.header.discount.type)}
                        required
                      >
                        <SelectTrigger className="w-16">
                          <SelectValue placeholder="Discount" />
                        </SelectTrigger>
                        <SelectContent className="">
                            <SelectItem value={"fixed"}>$</SelectItem>
                            <SelectItem value={"percentage"}>%</SelectItem>
                        </SelectContent>
                      </Select>
                    </div>
                  </div>
                  <div className='flex justify-between items-center w-60'>
                    <span className="block text-base">Tax</span>
                    <span className="block text-base">{currency(composeTax)}</span>
                  </div>
                  <Separator />
                  <div className='flex justify-between items-center w-60'>
                    <span className="block text-xl">Total</span>
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
              <AlertDialogTitle>Are you absolutely sure?</AlertDialogTitle>
              <AlertDialogDescription>
                This action cannot be undone. This will permanently delete this
                invoice and remove the data from our servers.
              </AlertDialogDescription>
            </AlertDialogHeader>
            <AlertDialogFooter>
              <AlertDialogCancel>Cancel</AlertDialogCancel>
              <AlertDialogAction onClick={performInvoiceCancelation}>Yes, Continue</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>
        <CheckoutForm
          openCheckout={openCheckout}
          setCheckout={setCheckout}
          paymentForm={invoiceForm.payment}
          totalAmount={computeTotalAmount()}
          onPlacedInvoice={placedInvoice}
          processing={processing}
          setCancelConfirmation={setCancelConfirmation}
          errors={propsErrors}
          onCheckoutChange={handleCheckoutChange}
        />
      </div>

      {/* Command to search for customers */}
    </AuthenticatedLayout>
  );
}
