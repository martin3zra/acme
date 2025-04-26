import { AlertDestructive } from '@/components/alert-destructive';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle } from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetFooter, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { useLocalStorage } from '@/hooks/use-local-storage';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { cn } from '@/lib/utils';
import { BreadcrumbItem, Customer, DiscountType, Item, PageProps } from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { format } from 'date-fns';
import { CalendarIcon, User, UserPlus, XCircleIcon } from 'lucide-react';
import React, { JSX, useCallback, useEffect } from 'react';

interface PaymentFormType {
  amount: number;
  reference: string;
}
// On focus display element
type InputViewProps = {
  value: number;
  method: paymentMethod;
  autoFocus?: boolean;
  onChange: (method: paymentMethod, value: number) => void;
  onFocus: (method: paymentMethod) => void;
}

const InputView = ({ value, method, autoFocus, onChange, onFocus }: InputViewProps): JSX.Element => {

  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    onChange(method, event.currentTarget.valueAsNumber)
  }

  const handleOnFocus = (event: React.FocusEvent<HTMLInputElement>) => {
    if (method !== "cash") onFocus(method)
  }

  return (
    <div className='p-0'>
      <Input
        key={method}
        type="number"
        min={0}
        className="text-end"
        value={value}
        autoFocus={autoFocus}
        onFocus={handleOnFocus}
        onChange={handleChange}
      />
    </div>
  )
}
const CashFormView = () => <></>

type CheckFormProps = Partial<PaymentFormType> & {
  onChange: (value: number|string) => void;
}

type CheckForm = PaymentFormType & {}

const defaultCheckForm: CheckForm = {amount: 0, reference: ""}

const CheckFormView = ({amount, reference, onChange}: CheckFormProps) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if(event.currentTarget.name === "ck") {
      onChange(event.currentTarget.valueAsNumber)
      return
    }

    onChange(event.currentTarget.value)
  }
  return (
    <div>
      <FormSection onSubmit={() => {}}>
        <FormSection.Title>Check payment</FormSection.Title>
        <FormSection.Description>Specify the amount of the Check and the number for future reference.</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 sm:col-span-4 space-y-2">
            <Label  htmlFor='ck' className='text-end'>Amount</Label>
            <Input
              type="number"
              min={0}
              name="ck"
              className="text-end h-12 md:text-xl"
              onChange={handleChange}
              autoFocus
              value={amount}
            />
          </div>
          <div className="col-span-6 sm:col-span-4 space-y-2">
            <Label  htmlFor='ck'>CK Number</Label>
            <Input
              type="text"
              name="reference"
              className="text-start h-12 md:text-xl"
              onChange={handleChange}
              value={reference}
            />
          </div>
        </FormSection.Form>
      </FormSection>
    </div>
  )
}

type CardBrand = {
  value: string;
  name: string;
}

const defaultCardBrands: CardBrand[] = [
  {value:"visa", name: "Visa"},
  {value:"mastercard", name: "MasterCard"},
  {value:"ae", name: "American Express"},
  {value:"unknown", name: "Unknown"},
]

type CardFormInput = "last4" | "brand" | "reference" | "amount"

type CardForm = PaymentFormType & {
  last4: number;
  brand: string;
}
const defaultCardForm: CardForm = {last4: 0, brand: "unknow", amount: 0, reference: ""}

type CardFormProps = PaymentFormType & {
  last4: number;
  brand: string;
  onChange: (value: number|string, key: CardFormInput) => void
 }

const CardFormView = ({last4, brand, amount, reference, onChange}: CardFormProps) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if(event.currentTarget.name === "last4") {
      event.currentTarget.value = event.currentTarget.value.replace(/\D/g, "")
      if (event.currentTarget.value.length > event.currentTarget.maxLength) {
        event.currentTarget.value = event.currentTarget.value.slice(0, event.currentTarget.maxLength);
      }
      onChange(event.currentTarget.valueAsNumber, event.currentTarget.name)
      return
    }

    if(event.currentTarget.name === "amount") {
      onChange(event.currentTarget.valueAsNumber, "amount")
    }

    onChange(event.currentTarget.value, event.currentTarget.name as CardFormInput)
  }

  return (
    <div>
      <FormSection onSubmit={() => {}}>
        <FormSection.Title>Debit/Credit payment</FormSection.Title>
        <FormSection.Description>Specify the amount of the Debit/Credit Card and the last 4 digits for future reference.</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 sm:col-span-3 space-y-2">
            <Label  htmlFor='last4' className='text-end'>Last 4 Digits</Label>
            <Input
              type="number"
              inputMode="numeric"
              name="last4"
              pattern="[0-9]*"
              maxLength={4}
              className="text-end h-12 md:text-xl"
              onChange={handleChange}
              autoFocus
              value={last4}
            />
          </div>
          <div className="col-span-6 sm:col-span-3 space-y-2">
            <Label  htmlFor='brand' className='text-end'>Brand</Label>
            <Select
              name='brand'
              onValueChange={(value) => onChange(value, "brand")}
              value={brand}
              required
            >
              <SelectTrigger className="w-full" size={"lg"}>
                <SelectValue placeholder="Select brand" />
              </SelectTrigger>
              <SelectContent className="w-full">
                {defaultCardBrands.map((brand, index) => (
                  <SelectItem key={index.toString()} value={brand.value.toString()}>
                    {brand.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="col-span-6 sm:col-span-3 space-y-2">
            <Label  htmlFor='reference'>Authorization</Label>
            <Input
              type="text"
              name="reference"
              className="text-start h-12 md:text-xl"
              onChange={handleChange}
              value={reference}
            />
          </div>
          <div className="col-span-6 sm:col-span-3 space-y-2">
            <Label  htmlFor='amount' className='text-end'>Amount</Label>
            <Input
              type="number"
              inputMode="numeric"
              name="amount"
              pattern="[0-9]*"
              className="text-end h-12 md:text-xl"
              onChange={handleChange}
              value={amount}
            />
          </div>
        </FormSection.Form>
      </FormSection>
    </div>
  )
}

type BTFormProps = CheckFormProps & {}

type BTForm = PaymentFormType & {}

const defaultBTForm: BTForm = {amount: 0, reference: ""}

const BankTransferFormView = ({amount, reference, onChange}: BTFormProps) => {
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if(event.currentTarget.name === "amount") {
      onChange(event.currentTarget.valueAsNumber)
      return
    }

    onChange(event.currentTarget.value)
  }
  return (
    <div>
      <FormSection onSubmit={() => {}}>
        <FormSection.Title>Bank Transfer payment</FormSection.Title>
        <FormSection.Description>Specify the amount of the Bank Transfer and the number for future reference.</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 sm:col-span-4 space-y-2">
            <Label  htmlFor='amount' className='text-end'>Amount</Label>
            <Input
              type="number"
              min={0}
              name="amount"
              className="text-end h-12 md:text-xl"
              onChange={handleChange}
              autoFocus
              value={amount}
            />
          </div>
          <div className="col-span-6 sm:col-span-4 space-y-2">
            <Label  htmlFor='reference'>Reference</Label>
            <Input
              type="text"
              name="reference"
              className="text-start h-12 md:text-xl"
              onChange={handleChange}
              value={reference}
            />
          </div>
        </FormSection.Form>
      </FormSection>
    </div>
  )
}

type paymentMethod = "cash" | "ck" | "card" | "bt";

type paymentMethodType = {
  value: paymentMethod
  name: string;
  amount: number
  autoFocus?: boolean
}

const defaultPaymentMethods: paymentMethodType[] = [
  {value: "cash", name: "Cash", amount: 0, autoFocus: true},
  {value: "ck", name: "CK", amount: 0},
  {value: "card", name: "Debit/Credit Card", amount: 0},
  {value: "bt", name: "Bank Transfer", amount: 0},
]

const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Home',
    href: '/home',
  },
  {
    title: 'Invoices',
    href: '/invoices',
  },
  {
    title: 'New Invoice',
    href: '/invoices/create',
  },
];

type paymentTerm = {
  value: number
  label: string
}

const paymentTerms: paymentTerm[] = [
  {value: 1, label: "Cash"},
  {value: 7, label: "7 Days"},
  {value: 10, label: "10 Days"},
  {value: 15, label: "15 Days"},
  {value: 30, label: "30 Days"},
  {value: 60, label: "60 Days"},
  {value: 90, label: "90 Days"},
]

type HeaderForm = {
  customer: Customer | undefined;
  date: Date | undefined;
  due: Date | undefined;
  terms: number;
  notes: string | undefined;
  discount: DiscountType;
};

interface LineForm extends Item {
  quantity: number;
  amount: number;
}

type InvoiceForm = {
  header: HeaderForm;
  lines: LineForm[];
};

const defaultDiscount: DiscountType = {value: 0, type: "fixed"}
const defaultHeaderForm: HeaderForm =  { customer: undefined, date: undefined, due: undefined, terms: 0, notes: undefined, discount: defaultDiscount}
const defaultInvoiceForm: InvoiceForm = { header: defaultHeaderForm, lines: [] }

export default function Create({ auth, customers, item }: PageProps<{ customers: Customer[]; item: Item }>) {
  const currency = useNumber().currency;
  const [open, setOpen] = React.useState(false);
  const [openCancelConfirmation, setCancelConfirmation] = React.useState(false);
  const [openCheckout, setCheckout] = React.useState(false);
  const [isEditing, setEditing] = React.useState(false);
  const [activePaymentForm, setActivePaymentForm] = React.useState<paymentMethod>("cash");
  // Payment methods
  const [paymentMethods, setPaymentMethods] = React.useState<paymentMethodType[]>(defaultPaymentMethods)
  const [cashAmount, setCashAmount] = React.useState(0);
  const [ckForm, setCkForm] = React.useState<CheckForm>(defaultCheckForm);
  const [cardForm, setCardForm] = React.useState<CardForm>(defaultCardForm);
  const [btForm, setBTForm] = React.useState<BTForm>(defaultBTForm);

  const referenceInputRef = React.useRef<HTMLInputElement>(null);
  const qtyInputRef = React.useRef<HTMLInputElement>(null);
  const addButtonRef = React.useRef<HTMLButtonElement>(null);
  const [search, setSearch] = React.useState('');
  const [notes, setNotes] = React.useState('');
  const dedbouncedSearch = useDebounced(search, 500);
  const dedbouncedNotes = useDebounced(notes, 500);
  const [amount, setAmount] = React.useState(0);
  const { setItem: storageInvoiceForm, getItem: getStorageInvoiceForm, removeItem: removeStorageIvoinceForm } = useLocalStorage('invoice');
  const [invoiceForm, setInvoiceForm] = React.useState<InvoiceForm>(() => {
    return getStorageInvoiceForm() || defaultInvoiceForm;
  });
  const [currenItem, setCurrentItem] = React.useState<Item | undefined>(undefined);

  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { post, transform, processing, errors } = useForm({
    customer_id: 0,
    terms: 0,
    lines: [],
    date: new Date(),
  });

  const computedCurrentItemAmount = (qty: number) => {
    setAmount(qty * (currenItem?.price || 0));
  };

  useEffect(() => setCurrentItem(item), [item]);

  const findCurrentItem = useCallback(() => {
    const exists = (element: LineForm) => element.id === currenItem?.id;
    const index = invoiceForm.lines.findIndex(exists);
    if (index >= 0) {
      setEditing(true);
      const line = invoiceForm.lines[index];
      setCurrentItem(line);
      qtyInputRef.current!.value = line.quantity.toString();
      setAmount(line.amount);
    }
  }, [currenItem, invoiceForm.lines]);

  useEffect(() => {
    if (currenItem) {
      findCurrentItem();
      qtyInputRef.current?.focus();
    }
  }, [currenItem, findCurrentItem]);

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

  const handleKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' || event.key === 'Tab') {
      event.preventDefault(); // Prevent default behavior of Enter key
      if (event.currentTarget.name === 'reference') {
        searchItem(event.currentTarget.value);
      }
      if (event.currentTarget.name === 'quantity') {
        addButtonRef.current?.focus();

        addButtonRef.current?.click(); // Simulate button click
      }
    }
  };

  const handleDoneButtonKeyPress = (event: React.KeyboardEvent<HTMLButtonElement>) => {
    console.log('Done button pressed', event.key);
    if (event.key === 'Enter') {
      console.log('Done button pressed');
    }
  };

  const handleDoneButtonClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    const line = currenItem!;

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

  const handleQtyChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    computedCurrentItemAmount(event.target.valueAsNumber);
  };

  const handleCustomerSelection = (event: React.MouseEvent<HTMLButtonElement>, customer: Customer | undefined) => {
    event.preventDefault();

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, customer } };
    });

    setOpen(false);
  };

  const addDays = (dateValue: Date, days: number): Date => {
    const date: Date = dateValue instanceof Date ? dateValue : new Date(dateValue);
    date.setDate(date.getDate() + days)
    return date
  }

  const handleDateChange = (date: unknown) => {
    invoiceForm.header.date = date as Date;
    invoiceForm.header.due = undefined
    if (invoiceForm.header.terms > 1) {
      invoiceForm.header.due = addDays(date as Date, invoiceForm.header.terms)
    }

    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, date: date as Date } };
    });
  };

  const handlePaymentTermsChange = (value: string) => {
    invoiceForm.header.terms = Number(value)

    if (invoiceForm.header.terms > 1 && invoiceForm.header.date) {
      invoiceForm.header.due = addDays(new Date(invoiceForm.header.date.getDate()), invoiceForm.header.terms)
    } else {
      invoiceForm.header.due = undefined
    }

    setInvoiceForm(() => {
      return {...invoiceForm, header: {...invoiceForm.header, terms: Number(value)}}
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
    if (invoiceForm.header.terms === 1) {
      setCheckout(true)
      return
    }

    placedInvoice()
  }

  const composePaymentMethods = () => {
    return {
      cash: {amount: cashAmount},
      check: ckForm,
      card: cardForm,
      bt: btForm,
    }
  }

  const placedInvoice = () => {
    // Check if the total and payment match on cash terms.
    transform((data) => ({
      ...data,
      customer_id: invoiceForm.header.customer?.id,
      date: invoiceForm.header.date,
      terms: invoiceForm.header.terms,
      discount: invoiceForm.header.discount,
      notes: invoiceForm.header.notes || '',
      lines: invoiceForm.lines.map((line) => {
        return { id: line.id, quantity: line.quantity, unit: line.unit.id, price: line.price, rate: line.tax.rate}
      }),
      payment: composePaymentMethods()
    }))
    post('/invoices', {...headers, onSuccess: () => {
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

  const composeTotalAmount = (): number => {
    return (composeSubTotal + composeTax) - computeDiscount()
  }

  const handleOnChangeInputView = (method: paymentMethod, value: number) => {
    setActivePaymentForm(method)

    if (typeof value  === "number" && method === "cash") {
      const givenValue =  isNaN(value) ? 0 : value
      setCashAmount(givenValue)
      paymentMethods.filter((p) => p.value === "cash")[0].amount = givenValue
      return
    }
  }

  const handleOnChangeCheckFormView = (value: number|string) => {
    if (typeof value  === "number") {
      const givenValue =  isNaN(value) ? 0 : value
      setCkForm(() => { return {...ckForm, amount: givenValue} })
      paymentMethods.filter((p) => p.value === "ck")[0].amount = givenValue
      return
    }

    setCkForm(() => { return {...ckForm, reference: value} })
  }

  const handleOnChangeCardFormView = (value: number | string, key: CardFormInput) => {
    if (typeof value  === "number" && key === "last4") {
      const givenValue =  isNaN(value) ? 0 : value
      setCardForm(() => { return {...cardForm, last4: givenValue} })
      return
    }
    if (key === "amount"){
      setCardForm(() => { return {...cardForm, [key]: Number(value)} })
      paymentMethods.filter((p) => p.value === "card")[0].amount = Number(value)
      return
    }

    setCardForm(() => { return {...cardForm, [key]: value} })
  }

  const handleOnChangeBTFormView = (value: number|string) => {
    if (typeof value  === "number") {
      const givenValue =  isNaN(value) ? 0 : value
      setBTForm(() => { return {...btForm, amount: givenValue} })
      paymentMethods.filter((p) => p.value === "bt")[0].amount = givenValue
      return
    }

    setBTForm(() => { return {...btForm, reference: value} })
  }

  const renderPaymentMethodForm = () => {
    if (activePaymentForm === "ck") return <CheckFormView amount={ckForm.amount} reference={ckForm.reference} onChange={handleOnChangeCheckFormView}/>
    if (activePaymentForm === "card") return <CardFormView last4={cardForm.last4} brand={cardForm.brand} amount={cardForm.amount} reference={cardForm.reference} onChange={handleOnChangeCardFormView} />
    if (activePaymentForm === "bt") return <BankTransferFormView amount={btForm.amount} reference={btForm.reference} onChange={handleOnChangeBTFormView}  />
    return <CashFormView />
  }

  const computeReceivedAmount = (): number => {
    return (Number(cashAmount) + Number(ckForm.amount) + Number(cardForm.amount) + Number(btForm.amount))
  }

  const computeRemainingBalance = (): number => {
    return composeTotalAmount() - computeReceivedAmount()
  }
  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <AuthenticatedLayout.Actions>
        <div className='flex justify-end gap-x-6'>
          <Button variant={"secondary"} onClick={() => setCancelConfirmation(true)}>Cancel</Button>
          <Button onClick={handleCheckout} disabled={processing}>Checkout</Button>
        </div>
      </AuthenticatedLayout.Actions>
      <div className="grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4 bg-gray-50/10">
        {!openCheckout && propsErrors.status && <div className="col-span-12"><AlertDestructive description={propsErrors.status} onDestroy={() => delete propsErrors.status }/></div>}
        <div className="z-50 col-span-12 grid h-42 grid-cols-2 gap-x-6">
          <div className="rounded-lg bg-white shadow">
            {!open && !invoiceForm.header.customer && (
              <div data-slot={`${errors.customer_id ? 'customer-error' : 'default'}`} className='flex flex-col h-full w-full items-center justify-center [&_svg]:text-white px-2 pb-2 data-[slot=customer-error]:rounded-lg data-[slot=customer-error]:bg-red-100/50 data-[slot=customer-error]:border data-[slot=customer-error]:border-red-500 data-[slot=customer-error]:[&_svg]:text-red-500 data-[slot=customer-error]:[&_[data-label=true]]:text-red-500'>
                <button onClick={() => setOpen(!open)} className="flex h-full w-full cursor-pointer items-center justify-center gap-2">
                  <div className="flex size-10 items-center justify-center rounded-full bg-gray-200">
                    <User className="size-6 *:data-[slot=customer-error]:text-red-500" />
                  </div>
                  <div data-label="true" className="text-lg">Customer</div>
                </button>
                <InputError className="mt-2" message={errors.customer_id} />
              </div>
            )}
            {invoiceForm.header.customer && (
              <div className="flex h-full flex-col overflow-y-hidden p-2">
                <div className="flex w-full items-center justify-between">
                  <div>{invoiceForm.header.customer?.name}</div>
                  <button onClick={(event) => handleCustomerSelection(event, undefined)} className="cursor-pointer p-1">
                    <XCircleIcon />
                  </button>
                </div>
                <div>{invoiceForm.header.customer?.email}</div>
                <div>{invoiceForm.header.customer?.phone}</div>
                <div>Address here!!!</div>
              </div>
            )}
            {open && !invoiceForm.header.customer && (
              <div className="flex h-full min-h-48 grow flex-col justify-start shadow">
                <div className="w-full border-b border-gray-200 p-2">
                  <Input
                    type="search"
                    placeholder="Search for a customer"
                    className="h-11 w-full rounded-t-lg"
                    onChange={(e) => setSearch(e.currentTarget.value)}
                  />
                </div>
                {/* Search result */}
                <div className="bg-gray-50">
                  {customers && customers.length > 0 ? (
                    customers.map((customer) => (
                      <button
                        key={customer.id}
                        className="flex w-full cursor-pointer items-center justify-start gap-2 rounded-lg p-2 hover:bg-gray-100"
                        onClick={(event) => handleCustomerSelection(event, customer)}
                      >
                        <div className="flex size-10 items-center justify-center rounded-full bg-gray-200">
                          <User className="size-6" color="white" />
                        </div>
                        <div className="text-lg">{customer.name}</div>
                      </button>
                    ))
                  ) : (
                    <div className="flex w-full items-center justify-center p-4 text-sm text-gray-500">
                      {dedbouncedSearch ? <p>No customers found</p> : null}
                    </div>
                  )}
                </div>
                {/* Create new action */}
                <div className="flex w-full items-center justify-center rounded-b-lg border bg-gray-100 p-2">
                  <button
                    className="flex cursor-pointer items-center justify-center gap-x-2 text-indigo-400"
                    onClick={() => alert('Create new customer')}
                  >
                    <UserPlus className="size-4" /> Add New Customer
                  </button>
                </div>
              </div>
            )}
          </div>
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
                <Input type='date' value={invoiceForm.header.due ? format(invoiceForm.header.due, 'yyyy-MM-dd') : ''} disabled className='w-70'/>
              </div>
            </div>
            <div className="col-span-6 flex flex-col gap-y-2">
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
          </div>
        </div>
        <div className="col-span-12">
          <div className="flex flex-col">
            <InputError message={errors.lines} />
            <table className="w-full table-auto">
              <thead>
                <tr>
                  <th scope="col" className="w-60 pe-1 border border-gray-300">
                    <Input
                      name="reference"
                      ref={referenceInputRef}
                      data-reset={false}
                      placeholder="Item reference"
                      onKeyDown={handleKeyDown}
                      className="border-none focus-visible:border-none focus-visible:ring-[2px] rounded-none"
                      tabIndex={0}
                    />
                  </th>
                  <th scope="col" className="w-auto px-1 border bg-gray-50 border-gray-300">
                    <Label>{currenItem?.description}</Label>
                  </th>
                  <th scope="col" className="w-36 px-1 border bg-gray-50 border-gray-300">
                    <Label>{currenItem?.unit.name}</Label>
                  </th>
                  <th scope="col" className="w-36 border border-gray-300">
                    <Input
                      type="number"
                      min={1}
                      name="quantity"
                      className="text-end border-none focus-visible:border-none focus-visible:ring-[2px] rounded-none"
                      tabIndex={1}
                      ref={qtyInputRef}
                      onFocus={(e) => computedCurrentItemAmount(e.currentTarget.valueAsNumber)}
                      onChange={handleQtyChange}
                      onKeyDown={handleKeyDown}
                    />
                  </th>
                  <th scope="col" className="w-36 px-1 border border-gray-300 bg-gray-50 text-end">
                    <Label className="block">{currency(currenItem?.price || 0)}</Label>
                  </th>
                  <th scope="col" className="w-36 px-1 border border-gray-300 text-end">
                    {amount > 0 ? currency(amount) : ''}
                  </th>
                  <th scope="col" className="w-6 border border-gray-300 text-end">
                    <Button
                      tabIndex={2}
                      ref={addButtonRef}
                      onKeyDown={handleDoneButtonKeyPress}
                      onClick={handleDoneButtonClick}
                      className="h-8 w-8 rounded-full p-0"
                    >
                      +
                    </Button>
                  </th>
                </tr>
                <tr>
                  <th scope="col" className="w-60 px-1 border border-gray-300 text-start">
                    Reference
                  </th>
                  <th scope="col" className="w-auto px-1 border border-gray-300 text-start">
                    Description
                  </th>
                  <th scope="col" className="w-36 px-1 border border-gray-300 text-start">
                    Unit
                  </th>
                  <th scope="col" className="w-36 px-1 border border-gray-300 text-end">
                    Quantity
                  </th>
                  <th scope="col" className="w-36 px-1 border border-gray-300 text-end">
                    Price
                  </th>
                  <th scope="col" className="w-36 px-1 border border-gray-300 text-end">
                    Amount
                  </th>
                  <th scope="col" className="w-6 gap-2 border border-gray-300 px-5 text-end whitespace-nowrap"></th>
                </tr>
              </thead>
              <tbody>
                {invoiceForm.lines &&
                  invoiceForm.lines.map((line, index) => (
                    <tr key={index}>
                      <td className="border px-1 border-gray-300 text-start">{line.name}</td>
                      <td className="border px-1 border-gray-300 text-start">{line.description}</td>
                      <td className="border px-1 border-gray-300 text-start">{line.unit.name}</td>
                      <td className="border px-1 border-gray-300 text-end">{line.quantity}</td>
                      <td className="border px-1 border-gray-300 text-end">{currency(line.price || 0)}</td>
                      <td className="border px-1 border-gray-300 text-end">{currency(line.amount || 0)}</td>
                      <td className="border px-1 border-gray-300 text-end">
                        <Button variant={'link'} size={'icon'} className="h-8 w-8 rounded-full p-0" data-index={index} onClick={handleRemoveLine}>
                          <XCircleIcon />
                        </Button>
                      </td>
                    </tr>
                  ))}
              </tbody>
            </table>
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
                  ></Textarea>
              </div>
              <div className="col-span-2 flex flex-col gap-y-2 rounded-lg border border-gray-300/25 bg-gray-100/10">
                <div className='grid place-content-end p-2 gap-y-4'>
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
                    <span className="block text-xl">{currency(composeTotalAmount())}</span>
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
              <AlertDialogAction onClick={performInvoiceCancelation}>Continue</AlertDialogAction>
            </AlertDialogFooter>
          </AlertDialogContent>
        </AlertDialog>

        <Sheet open={openCheckout} onOpenChange={setCheckout}>
          <SheetContent side='right' className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl">
            <SheetHeader>
              <SheetTitle>Checkout</SheetTitle>
              <SheetDescription className="text-[12px]">Checkout process</SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 px-4">
              {propsErrors.status && <AlertDestructive description={propsErrors.status} onDestroy={() => delete propsErrors.status }/>}
              <h4>Payment detail</h4>
              <div className='flex justify-between items-center w-full'>
                 <table className="w-full table-auto">
                  <thead>
                    <tr>
                    {paymentMethods.map((method) =>
                      <th scope="col" key={method.value} className="w-60 border border-gray-300">{method.name}</th>
                    )}
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                    {paymentMethods.map((method) =>
                      <td key={method.value} className="border px-1 border-gray-300 text-start">
                        <InputView
                          key={method.value}
                          value={method.amount}
                          method={method.value}
                          onChange={handleOnChangeInputView}
                          onFocus={(methodType) => setActivePaymentForm(methodType)}
                          />
                      </td>
                    )}
                    </tr>
                  </tbody>
                </table>

              </div>
              <div className='pb-6'>
                {renderPaymentMethodForm()}
              </div>
              <Separator className='' />
              <div>
                <div className='flex justify-between items-center w-60'>
                  <span className="block text-2xl">To collect</span>
                  <span className="block text-2xl">{currency(composeTotalAmount())}</span>
                </div>
                <div className='flex justify-between items-center w-60'>
                  <span className="block text-2xl">Received</span>
                  <span className="block text-2xl">{currency(computeReceivedAmount())}</span>
                </div>
                <div className='flex justify-between items-center w-60'>
                  <span className="block text-2xl">Remaining</span>
                  <span className="block text-2xl text-red-600 font-medium">{currency(computeRemainingBalance())}</span>
                </div>
              </div>
            </div>
            <SheetFooter>
              {computeRemainingBalance() !== 0 && <AlertDestructive description="The amount collected must be equals to the Invoice total amount." destroyable={false} />}
              <div className='flex justify-end gap-x-6'>
                <Button variant={"secondary"} onClick={() => setCancelConfirmation(true)}>Cancel</Button>
                <Button onClick={placedInvoice} disabled={processing || computeRemainingBalance() !== 0}>Complete Invoice</Button>
              </div>
            </SheetFooter>
          </SheetContent>
        </Sheet>
      </div>

      {/* Command to search for customers */}
    </AuthenticatedLayout>
  );
}
