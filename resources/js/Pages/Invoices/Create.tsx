import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { useNumber } from '@/composables/use-number';
import { useDebounced } from '@/hooks/use-debounced';
import { useLocalStorage } from '@/hooks/use-local-storage';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { cn } from '@/lib/utils';
import { BreadcrumbItem, Customer, Item, PageProps } from '@/types';
import { router } from '@inertiajs/react';
import { format } from 'date-fns';
import { CalendarIcon, User, UserPlus, XCircleIcon } from 'lucide-react';
import React, { useCallback, useEffect } from 'react';

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

type HeaderForm = {
  customer: Customer | undefined;
  date: Date | undefined;
  terms: string | undefined;
  notes: string | undefined;
};

interface LineForm extends Item {
  quantity: number;
  amount: number;
}

type InvoiceForm = {
  header: HeaderForm;
  lines: LineForm[];
};

// const InvoiceItems: InvoiceItemForm[] = [];

export default function Create({ auth, customers, item }: PageProps<{ customers: Customer[]; item: Item }>) {
  const currency = useNumber().currency;
  const [open, setOpen] = React.useState(false);
  const [isEditing, setEditing] = React.useState(false);

  const referenceInputRef = React.useRef<HTMLInputElement>(null);
  const qtyInputRef = React.useRef<HTMLInputElement>(null);
  const addButtonRef = React.useRef<HTMLButtonElement>(null);
  // const [date, setDate] = React.useState<Date>();
  const [search, setSearch] = React.useState('');
  const dedbouncedSearch = useDebounced(search, 500);
  const [amount, setAmount] = React.useState(0);
  const { setItem: storageInvoiceForm, getItem: getStorageInvoiceForm, removeItem } = useLocalStorage('invoice');
  const [invoiceForm, setInvoiceForm] = React.useState<InvoiceForm>(() => {
    return getStorageInvoiceForm() || { header: { customer: undefined, date: undefined, terms: undefined, notes: undefined }, lines: [] };
  });
  const [currenItem, setCurrentItem] = React.useState<Item | undefined>(undefined);

  // const { headers } = useHeader();

  // const { data, setData, post, errors, reset, processing } = useForm<Required<IvoiceItemForm>>({
  //   id: item?.id || 0,
  //   quantity: 1,
  // });

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

  useEffect(() => {
    storageInvoiceForm(invoiceForm);
  }, [invoiceForm, storageInvoiceForm]);

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

  const handleDateChange = (date: unknown) => {
    invoiceForm.header.date = date as Date;
    setInvoiceForm(() => {
      return { ...invoiceForm, header: { ...invoiceForm.header, date: date as Date } };
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

  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4">
        <div className="z-50 col-span-12 grid h-60 grid-cols-2 gap-x-6">
          <div className="rounded-lg bg-white shadow">
            {!open && !invoiceForm.header.customer && (
              <button onClick={() => setOpen(!open)} className="flex h-full w-full cursor-pointer items-center justify-center gap-2">
                <div className="flex size-10 items-center justify-center rounded-full bg-gray-200">
                  <User className="size-6" color="white" />
                </div>
                <div className="text-lg">Customer</div>
              </button>
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
            <div className="col-span-6 flex flex-col gap-y-2">
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
            </div>
            <div className="col-span-6"></div>
          </div>
        </div>
        <div className="col-span-12">
          <div className="flex flex-col">
            <table className="w-full table-auto">
              <thead>
                <tr>
                  <th scope="col" className="w-60 border border-gray-300">
                    <Input
                      name="reference"
                      ref={referenceInputRef}
                      data-reset={false}
                      placeholder="Item reference"
                      onKeyDown={handleKeyDown}
                      className=""
                      tabIndex={0}
                    />
                  </th>
                  <th scope="col" className="w-auto border border-gray-300">
                    <Label>{currenItem?.description}</Label>
                  </th>
                  <th scope="col" className="w-36 border border-gray-300">
                    <Input
                      type="number"
                      min={1}
                      // defaultValue={currenItem?.quantity || 0}
                      name="quantity"
                      className=""
                      tabIndex={1}
                      ref={qtyInputRef}
                      onFocus={(e) => computedCurrentItemAmount(e.currentTarget.valueAsNumber)}
                      onChange={handleQtyChange}
                      onKeyDown={handleKeyDown}
                    />
                  </th>
                  <th scope="col" className="w-36 border border-gray-300 bg-red-100 text-end">
                    <Label className="block">{currency(currenItem?.price || 0)}</Label>
                  </th>
                  <th scope="col" className="w-36 border border-gray-300 text-end">
                    {amount > 0 ? currency(amount) : ''}
                  </th>
                  <th scope="col" className="w-6 border border-gray-300 text-end">
                    <Button
                      tabIndex={2}
                      ref={addButtonRef}
                      onKeyDown={handleDoneButtonKeyPress}
                      onClick={handleDoneButtonClick}
                      // disabled={processing}
                      className="h-8 w-8 rounded-full p-0"
                    >
                      +
                    </Button>
                  </th>
                </tr>
              </thead>
            </table>
            <table className="w-full table-auto">
              <thead>
                <tr>
                  <th scope="col" className="w-60 border border-gray-300 text-start">
                    Reference
                  </th>
                  <th scope="col" className="w-auto border border-gray-300 text-start">
                    Description
                  </th>
                  <th scope="col" className="w-36 border border-gray-300 text-end">
                    Quantity
                  </th>
                  <th scope="col" className="w-36 border border-gray-300 text-end">
                    Price
                  </th>
                  <th scope="col" className="w-36 border border-gray-300 text-end">
                    Amount
                  </th>
                  <th scope="col" className="w-6 gap-2 border border-gray-300 px-5 text-end whitespace-nowrap"></th>
                </tr>
              </thead>
              <tbody>
                {invoiceForm.lines &&
                  invoiceForm.lines.map((line, index) => (
                    <tr key={index}>
                      <td className="border border-gray-300 text-start">{line.name}</td>
                      <td className="border border-gray-300 text-start">{line.description}</td>
                      <td className="border border-gray-300 text-end">{line.quantity}</td>
                      <td className="border border-gray-300 text-end">{currency(line.price || 0)}</td>
                      <td className="border border-gray-300 text-end">{currency(line.amount || 0)}</td>
                      <td className="border border-gray-300 text-end">
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
        <div className="col-span-12 min-h-48 bg-yellow-500">Footer</div>
      </div>

      {/* Command to search for customers */}
    </AuthenticatedLayout>
  );
}
