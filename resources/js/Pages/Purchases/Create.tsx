import { AlertDestructive } from '@/components/alert-destructive';
import { Alert, AlertDescription } from '@/components/ui/alert';
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
import { useDebounced } from '@/hooks/use-debounced';
import { usePersistedState } from '@/hooks/use-persisted-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { addDays, getDaysFromTerm, isNotEmpty } from '@/lib/utils';
import type { DiscountType, Item, LineForm, PageProps, PaymentTermValue, PurchaseForm, PurchaseTransactionKind, Vendor } from '@/types';
import { Textarea } from '@headlessui/react';
import { router, useForm, usePage } from '@inertiajs/react';
import { format } from 'date-fns';
import { Info } from 'lucide-react';
import React, { useCallback, useEffect, useState } from 'react';
import { defaultDiscount, defaultPurchaseForm, makeCreateBreadcrumbs, paymentTerms, purchaseKindMeta } from './constants';
import { Lines } from './Shared/lines';
import { VendorSection } from './Shared/vendor-section';

interface PurchaseRedirectProps {
  redirectTo: string;
}

export interface PurchaseFormData {
  vendor_id: number;
  terms: string;
  lines: any[];
  date: Date;
  discount: DiscountType;
  notes: string;
  kind: PurchaseTransactionKind;
  transaction_kind?: PurchaseTransactionKind;
  code?: string;
  source: any;
  invoice_number?: string;
  [key: string]: any;
}

export default function Create({
  auth,
  vendors,
  vendor,
  items,
  item,
  kind,
}: PageProps<{
  vendors: Vendor[];
  vendor: Vendor;
  items: Item[];
  item: Item;
  kind: PurchaseTransactionKind;
}>) {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const meta = purchaseKindMeta(kind);

  const [openVendor, setOpenVendor] = useState(false);
  const [openCancelConfirmation, setCancelConfirmation] = useState(false);
  const [currentItem, setCurrentItem] = useState<Item | undefined>(undefined);
  const [isEditing, setEditing] = useState(false);
  const [searchVendor, setSearchVendor] = useState('');
  const debouncedVendorSearch = useDebounced(searchVendor, 500);
  const [amount, setAmount] = useState(0);
  const [codeError, setCodeError] = useState<string | null>(null);

  const referenceInputRef = React.useRef<HTMLInputElement>(null);
  const qtyInputRef = React.useRef<HTMLInputElement>(null);

  const initialForm = (): PurchaseForm => {
    const f = defaultPurchaseForm(kind);
    if (vendor) {
      f.header.vendor = vendor;
      f.header.terms = vendor.payment_terms || 'pia';
    }
    return f;
  };

  const [purchaseForm, setPurchaseForm, removePurchaseForm] = usePersistedState<PurchaseForm>(`purchase_${kind}`, initialForm());

  const { setData, post, transform, processing, errors } = useForm<PurchaseFormData>({
    vendor_id: 0,
    terms: 'pia',
    lines: [],
    date: new Date(),
    discount: defaultDiscount,
    notes: '',
    kind,
    source: { type: kind, id: '' },
  });

  useEffect(() => setCurrentItem(item), [item]);

  useEffect(() => {
    const searchVendors = () => {
      router.reload({ only: ['vendors'], data: { search: debouncedVendorSearch }, preserveUrl: true });
    };

    if (debouncedVendorSearch) {
      searchVendors();
    }
  }, [debouncedVendorSearch]);

  useEffect(() => {
    // Keep derived due date in sync
    if (!purchaseForm?.header?.date) return;
    if (purchaseForm.header.terms !== 'pia') {
      const days = getDaysFromTerm(purchaseForm.header.terms);
      purchaseForm.header.due = addDays(purchaseForm.header.date, days);
    } else {
      purchaseForm.header.due = undefined;
    }

    setPurchaseForm(() => ({ ...purchaseForm, header: { ...purchaseForm.header } }));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [purchaseForm.header.terms]);

  const findCurrentItem = useCallback(() => {
    const exists = (element: LineForm) => element.id === currentItem?.id;
    const index = purchaseForm.lines.findIndex(exists);
    if (index >= 0) {
      setEditing(true);
      const line = purchaseForm.lines[index];
      setCurrentItem(line);
      qtyInputRef.current!.value = line.qty.toString();
      setAmount(line.amount);
    }
  }, [currentItem, purchaseForm.lines]);

  useEffect(() => {
    if (currentItem) {
      findCurrentItem();
      qtyInputRef.current!.value = '1';
      qtyInputRef.current?.focus();
    }
  }, [currentItem, findCurrentItem]);

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

  const processCurrentItem = () => {
    const line = currentItem!;

    if (isEditing) {
      const index = purchaseForm.lines.findIndex((element: LineForm) => element.id === line.id);
      if (index >= 0) {
        purchaseForm.lines[index].qty = qtyInputRef.current?.valueAsNumber || 0;
        purchaseForm.lines[index].amount = amount;
        purchaseForm.lines[index].action = 'updated';
      }
      setEditing(false);
    } else {
      purchaseForm.lines.push({ ...line, qty: qtyInputRef.current?.valueAsNumber || 0, amount, action: 'added' });
    }

    setPurchaseForm(() => ({ ...purchaseForm, lines: [...purchaseForm.lines] }));
    resetPurchaseFormInput();
  };

  const handleOnKeyDown = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === 'Enter' || event.key === 'Tab') {
      event.preventDefault();
      if (event.currentTarget.name === 'reference' && isNotEmpty(event.currentTarget.value)) {
        searchItem(event.currentTarget.value);
        return;
      }
      if (event.currentTarget.name === 'qty' && currentItem != undefined) {
        processCurrentItem();
      }
    }
  };

  const resetPurchaseFormInput = () => {
    setCurrentItem(undefined);
    setAmount(0);
    referenceInputRef.current!.value = '';
    qtyInputRef.current!.value = '';
    referenceInputRef.current?.focus();
  };

  const handleVendorSelection = (vendor: Vendor | undefined) => {
    setPurchaseForm(() => {
      return {
        ...purchaseForm,
        header: {
          ...purchaseForm.header,
          vendor,
          terms: vendor?.payment_terms || 'pia',
        },
      };
    });

    setOpenVendor(false);
  };

  const handleDateChange = (date: unknown) => {
    purchaseForm.header.date = date as Date;
    purchaseForm.header.due = undefined;
    if (purchaseForm.header.terms !== 'pia') {
      const days = getDaysFromTerm(purchaseForm.header.terms);
      purchaseForm.header.due = addDays(purchaseForm.header.date, days);
    }

    setPurchaseForm(() => ({ ...purchaseForm, header: { ...purchaseForm.header, date: date as Date } }));
  };

  const handlePaymentTermsChange = (value: PaymentTermValue) => {
    purchaseForm.header.terms = value;
    if (purchaseForm.header.terms !== 'pia' && purchaseForm.header.date) {
      const days = getDaysFromTerm(purchaseForm.header.terms);
      purchaseForm.header.due = addDays(purchaseForm.header.date, days);
    } else {
      purchaseForm.header.due = undefined;
    }

    setPurchaseForm(() => ({ ...purchaseForm, header: { ...purchaseForm.header, terms: value } }));
  };

  const handleDiscountValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    purchaseForm.header.discount.value = event.target.valueAsNumber;
    setPurchaseForm(() => ({ ...purchaseForm, header: { ...purchaseForm.header, discount: { ...purchaseForm.header.discount, value: event.target.valueAsNumber } } }));
  };

  const handleDiscountTypeChange = (value: 'fixed' | 'percentage') => {
    purchaseForm.header.discount.type = value;
    setPurchaseForm(() => ({ ...purchaseForm, header: { ...purchaseForm.header, discount: { ...purchaseForm.header.discount, type: value } } }));
  };

  const composeSubTotal = purchaseForm.lines.filter((l) => l.action !== 'deleted').reduce((acc, line) => acc + line.amount, 0);

  const computeDiscount = (): number => {
    const discount = purchaseForm.header.discount;
    if (discount.type === 'percentage') {
      return composeSubTotal * (discount.value / 100);
    }
    return discount.value;
  };

  const composeTax = purchaseForm.lines
    .filter((l) => l.action !== 'deleted')
    .reduce((acc, line) => {
      let discount = purchaseForm.header.discount.value;
      if (purchaseForm.header.discount.type === 'fixed') {
        discount = composeSubTotal > 0 ? (discount / composeSubTotal) * 100 : 0;
      }
      const lineAmount = line.price * line.qty;
      const lineDiscount = lineAmount * (discount / 100);
      const tax = (lineAmount - lineDiscount) * (line.tax.rate / 100);
      return acc + tax;
    }, 0);

  const computeTotalAmount = (): number => {
    return composeSubTotal - computeDiscount() + composeTax;
  };

  const performSave = () => {
    if (kind === 'purchase_receipt' && !isNotEmpty(purchaseForm.code ?? '')) {
      setCodeError(t('purchases.receipts.form.vendorInvoiceNumberRequired'));
      return;
    }

    setCodeError(null);

    transform((data) => {
      const payload: Record<string, any> = {
        ...data,
        vendor_id: purchaseForm.header.vendor?.id,
        date: purchaseForm.header.date,
        terms: purchaseForm.header.terms,
        discount: purchaseForm.header.discount,
        notes: purchaseForm.header.notes || '',
        kind,
        source: purchaseForm.source,
        lines: purchaseForm.lines
          .filter((l) => l.action !== 'deleted')
          .map((line) => ({ id: line.id, qty: line.qty, unit: line.unit.id, price: line.price, rate: line.tax.rate, action: 'added' })),
      };

      if (kind === 'purchase_receipt') {
        payload.transaction_kind = 'purchase_receipt';
        payload.invoice_number = purchaseForm.code ?? '';
      }

      if (kind === 'vendor_bill') {
        payload.transaction_kind = 'vendor_bill';
        payload.invoice_number = purchaseForm.header.invoice_number ?? '';
      }

      return payload;
    });

    post('/purchases', {
      ...headers,
      preserveState: 'errors',
      onSuccess: (event) => {
        const page = event as unknown as { props: PurchaseRedirectProps };
        removePurchaseForm();
        if (page.props.redirectTo) {
          router.visit(page.props.redirectTo);
          return;
        }
        router.visit(meta.listUrl);
      },
    });
  };

  const performCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    router.get(meta.listUrl);
    setTimeout(() => removePurchaseForm(), 200);
  };

  return (
    <AppLayout user={auth.user} breadcrumbs={makeCreateBreadcrumbs(kind)}>
      <AppLayout.Actions>
        <div className="flex justify-end gap-x-6">
          <Button variant={'secondary'} onClick={() => setCancelConfirmation(true)}>
            {t('global.actions.cancel')}
          </Button>
          <Button onClick={performSave} disabled={processing}>
            {t('global.actions.save')}
          </Button>
        </div>
      </AppLayout.Actions>

      <div className="grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4 bg-gray-50/10">
        {propsErrors.status && (
          <div className="col-span-12">
            <AlertDestructive description={propsErrors.status} onDestroy={() => delete propsErrors.status} />
          </div>
        )}

        {kind === 'purchase_receipt' && purchaseForm.source?.type === 'purchase_order' && (
          <div className="col-span-12">
            <Alert>
              <Info className="size-4" />
              <AlertDescription>{t('purchases.receipts.partialReceiptNotice')}</AlertDescription>
            </Alert>
          </div>
        )}

        <div className="z-50 col-span-12 grid min-h-42 grid-cols-2 gap-x-6">
          <VendorSection
            vendor={purchaseForm.header.vendor}
            vendors={vendors}
            errors={errors}
            handleVendorSelection={handleVendorSelection}
            setSearch={setSearchVendor}
            setOpen={setOpenVendor}
            open={openVendor}
            debouncedSearch={debouncedVendorSearch}
          />

          <div className="grid grid-cols-12">
            <div className="col-span-6 flex flex-col gap-y-6">
              <DatePickerField
                id="date"
                label={t('global.date')}
                placeholder={t('global.datePlaceholder')}
                value={purchaseForm.header.date}
                onChange={handleDateChange}
                error={errors.date}
                className="w-52"
              />

              <div className="flex flex-col gap-y-2">
                <Label htmlFor="date">{t('global.dueDate')}</Label>
                <Label className="text-muted-foreground w-70 rounded-sm border p-2.5">
                  {purchaseForm.header.due ? format(purchaseForm.header.due, 'PPP') : t('global.noAvailable.default')}
                </Label>
              </div>
            </div>

            <div className="col-span-6 flex flex-col gap-y-6">
              {kind === 'purchase_receipt' && (
                <div className="flex flex-col gap-y-2">
                  <Label htmlFor="code">{t('purchases.receipts.form.vendorInvoiceNumber')}</Label>
                  <Input
                    id="code"
                    name="code"
                    value={purchaseForm.code ?? ''}
                    placeholder={t('purchases.receipts.form.vendorInvoiceNumberPlaceholder')}
                    onChange={(e) => {
                      setCodeError(null);
                      setPurchaseForm(() => ({ ...purchaseForm, code: e.target.value }));
                    }}
                    required
                  />
                  <InputError className="mt-2" message={codeError || (errors as any).code} />
                </div>
              )}

              {kind === 'vendor_bill' && (
                <div className="flex flex-col gap-y-2">
                  <Label htmlFor="invoice_number">{t('purchases.vendorBills.form.vendorInvoiceNumber')}</Label>
                  <Input
                    id="invoice_number"
                    name="invoice_number"
                    value={purchaseForm.header.invoice_number ?? ''}
                    placeholder={t('purchases.vendorBills.form.vendorInvoiceNumberPlaceholder')}
                    onChange={(e) => {
                      setPurchaseForm(() => ({ ...purchaseForm, header: { ...purchaseForm.header, invoice_number: e.target.value } }));
                    }}
                  />
                  <InputError className="mt-2" message={(errors as any).invoice_number} />
                </div>
              )}

              <div className="flex flex-col gap-y-2">
                <Label htmlFor="paymentTerms">{t('global.paymentTerms')}</Label>
                <Select name="paymentTerms" onValueChange={handlePaymentTermsChange} value={purchaseForm.header.terms} required>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select terms" />
                  </SelectTrigger>
                  <SelectContent>
                    {paymentTerms.map((term) => (
                      <SelectItem key={term.value} value={term.value}>
                        {term.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <InputError className="mt-2" message={errors.terms} />
              </div>

              <div className="flex flex-col gap-y-2">
                <span className="block text-base">{t('global.discount')}</span>
                <div className="flex items-center gap-x-2">
                  <Input
                    type="number"
                    min={0}
                    step={0.01}
                    defaultValue={purchaseForm.header.discount.value}
                    name="discount"
                    onChange={handleDiscountValueChange}
                  />
                  <Select name="discountType" onValueChange={handleDiscountTypeChange} value={String(purchaseForm.header.discount.type)}>
                    <SelectTrigger className="w-40">
                      <SelectValue placeholder={t('global.discount')} />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="fixed">Fixed</SelectItem>
                      <SelectItem value="percentage">%</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <InputError message={errors['discount']} />
              </div>
            </div>
          </div>
        </div>

        <div className="col-span-12 rounded-lg bg-white shadow">
          <Lines
            items={items}
            lines={purchaseForm.lines}
            lineError={errors['lines']}
            currentItem={currentItem}
            handleRemoveLine={(e) => {
              e.preventDefault();
              const index = Number(e.currentTarget.getAttribute('data-index'));
              const newItems = [...purchaseForm.lines];
              newItems[index].action = 'deleted';
              setPurchaseForm(() => ({ ...purchaseForm, lines: newItems }));
            }}
            handleKeyDown={handleOnKeyDown}
            handleOnSelected={handleOnSelectedItem}
            amount={amount}
            setAmount={setAmount}
            referenceInputRef={referenceInputRef}
            qtyInputRef={qtyInputRef}
          />
        </div>

        <div className="col-span-12 grid grid-cols-12 gap-x-6">
          <div className="col-span-8 rounded-lg bg-white p-4 shadow">
            <Label>{t('global.notes')}</Label>
            <Textarea
              name="notes"
              className="mt-2 w-full rounded-md border p-2"
              value={purchaseForm.header.notes || ''}
              onChange={(e) => setPurchaseForm(() => ({ ...purchaseForm, header: { ...purchaseForm.header, notes: e.target.value } }))}
            />
          </div>
          <div className="col-span-4 rounded-lg bg-white p-4 shadow">
            <Label>{t('global.total')}</Label>
            <Separator className="my-2" />
            <div className="flex items-center justify-between">
              <Label>{t('global.subTotal')}</Label>
              <Label>{composeSubTotal.toFixed(2)}</Label>
            </div>
            <div className="flex items-center justify-between">
              <Label>{t('global.discount')}</Label>
              <Label>{computeDiscount().toFixed(2)}</Label>
            </div>
            <div className="flex items-center justify-between">
              <Label>{t('global.tax')}</Label>
              <Label>{composeTax.toFixed(2)}</Label>
            </div>
            <Separator className="my-2" />
            <div className="flex items-center justify-between">
              <Label>{t('global.total')}</Label>
              <Label>{computeTotalAmount().toFixed(2)}</Label>
            </div>
          </div>
        </div>
      </div>

      <AlertDialog open={openCancelConfirmation} onOpenChange={setCancelConfirmation}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('global.actions.cancel')}</AlertDialogTitle>
            <AlertDialogDescription>{t('global.warning.description')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('global.cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={performCancelation}>{t('global.ok')}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </AppLayout>
  );
}
