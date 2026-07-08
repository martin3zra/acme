import ActionSection from '@/components/action-section';
import { ConfirmsPassword } from '@/components/confirms-password';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useTranslation } from '@/hooks/use-translation';
import { paymentTerms } from '@/Pages/Invoices/constants';
import { PageProps, PaymentMethods, TaxReceipt, Vendor, VendorType, VendorTypes, Verb } from '@/types';
import { Field, Radio, RadioGroup } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';
import { CheckCircleIcon } from 'lucide-react';
import { useState } from 'react';

export type CreateFormParams = {
  vendor: Vendor | undefined;
  tax_receipts: TaxReceipt[];
  action: Verb;
};

type CreateFormProps = {
  onFinish: () => void;
  params: CreateFormParams;
};

type VendorForm = {
  id: number | undefined;
  name: string;
  contact: string;
  email: string;
  phone: string;
  address: string;
  purchase_note: string;
  lead_time_days: number;
  payment_method?: string;
  payment_terms?: string;
  vendor_type: string;
  // tax_receipt: number;
  open_balance: number;
  open_balance_as_of: Date | undefined;
};

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const currency = useNumber().currency;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { data, setData, post, put, errors, reset, processing } = useForm<Required<VendorForm>>({
    id: params.vendor?.id,
    name: params.vendor?.name || '',
    contact: params.vendor?.contact_name || '',
    email: params.vendor?.email || '',
    phone: params.vendor?.phone || '',
    address: params.vendor?.address || '',
    purchase_note: params.vendor?.purchase_note || '',
    lead_time_days: params.vendor?.lead_time_days ?? 0,
    payment_method: params.vendor?.payment_method || '',
    payment_terms: params.vendor?.payment_terms || '',
    vendor_type: params.vendor?.vendor_type || 'business',
    open_balance: params.vendor?.open_balance?.amount || 0,
    open_balance_as_of: params.vendor?.open_balance?.date ? new Date(params.vendor.open_balance.date) : undefined,
  });

  const viewMode = params.action === 'view';
  const isDisabled = params.vendor?.status === 'disabled';

  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => {
      reset();
      onFinish();
    },
  };

  const submit = () => {
    if (params.action === 'create') post('/vendors', options);
    else if (params.action === 'edit') put(`/vendors/${params.vendor!.id}`, options);
  };

  return (
    <div className="flex flex-col space-y-6">
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('vendors.single.title')}</FormSection.Title>
        <FormSection.Description>{t('vendors.single.description')}</FormSection.Description>
        <FormSection.Form>
          {propsErrors.status && <div className="col-span-6 mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}
          <div className="col-span-6">
            <div className="flex flex-col gap-2">
              <Label htmlFor="vendor_type">{t('vendors.single.type')}</Label>
              <RadioGroup
                disabled={viewMode}
                id="vendor_type"
                className="grid grid-cols-3 gap-6"
                value={data.vendor_type as VendorType}
                onChange={(type: VendorType) => setData('vendor_type', type)}
              >
                {VendorTypes.map((type: VendorType) => (
                  <Field key={type}>
                    <Radio
                      value={type}
                      className="group data-checked:bg-primary data-checked:text-primary-foreground bg-primary/5 data-focus:outline-primary relative flex cursor-pointer grid-cols-1 rounded-lg px-5 py-4 shadow-md transition focus:not-data-focus:outline-none data-focus:outline"
                    >
                      <div className="flex w-full capitalize">{t(`vendors.single.${type}`)}</div>
                      <CheckCircleIcon className="size-6 opacity-0 transition group-data-checked:opacity-100" />
                    </Radio>
                  </Field>
                ))}
              </RadioGroup>
            </div>
          </div>
          <div className="col-span-6 gap-2">
            <Label htmlFor="name">{t('global.name')}</Label>
            <Input
              id="name"
              className="mt-1 block w-full"
              value={data.name}
              onChange={(e) => setData('name', e.target.value)}
              required
              autoComplete="name"
              placeholder={t('global.name')}
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.name} />
          </div>
          <div className="col-span-6 gap-2">
            <Label htmlFor="contact">{t('global.contact')}</Label>
            <Input
              id="contact"
              className="mt-1 block w-full"
              value={data.contact}
              onChange={(e) => setData('contact', e.target.value)}
              placeholder="Jane Doe"
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.contact} />
          </div>
          <div className="col-span-4 gap-2">
            <Label htmlFor="email">{t('global.email')}</Label>
            <Input
              id="email"
              type="email"
              className="mt-1 block w-full"
              value={data.email}
              onChange={(e) => setData('email', e.target.value)}
              required
              placeholder="jane.doe@example.com"
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.email} />
          </div>
          <div className="col-span-2 gap-2">
            <Label htmlFor="phone">{t('global.phone')}</Label>
            <Input
              id="phone"
              className="mt-1 block w-full"
              value={data.phone}
              onChange={(e) => setData('phone', e.target.value)}
              placeholder="809-983-3897"
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.phone} />
          </div>
          <div className="col-span-6 gap-2">
            <Label htmlFor="address">{t('global.address')}</Label>
            <Input
              id="address"
              className="mt-1 block w-full"
              value={data.address}
              onChange={(e) => setData('address', e.target.value)}
              placeholder={t('global.address')}
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.address} />
          </div>
          <Separator className="col-span-6" />
          <div className="col-span-6">
            <h4 className="font-medium">{t('vendors.single.paymentSection')}</h4>
          </div>
          <div className="col-span-2">
            <Label>{t('global.paymentMethod')}</Label>
            <Select
              onValueChange={(value) => setData('payment_method', value)}
              disabled={viewMode}
              defaultValue={data.payment_method}
              value={data.payment_method}
              required
            >
              <SelectTrigger className="mt-2 w-64">
                <SelectValue placeholder={t('global.paymentMethod')} />
              </SelectTrigger>
              <SelectContent className="">
                {PaymentMethods.map((method) => (
                  <SelectItem key={method} value={method}>
                    {t(`global.paymentMethods.${method}.title`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <InputError className="mt-2" message={errors.payment_method} />
          </div>
          <div className="col-span-2">
            <Label>{t('global.paymentTerms')}</Label>
            <Select
              name="paymentTerms"
              disabled={viewMode}
              onValueChange={(value) => setData('payment_terms', value)}
              defaultValue={'net30'}
              value={data.payment_terms}
              required
            >
              <SelectTrigger className="mt-2 w-full">
                <SelectValue placeholder={t('global.paymentTerms')} />
              </SelectTrigger>
              <SelectContent className="">
                {paymentTerms.map((term, index) => (
                  <SelectItem key={index.toString()} value={term.value.toString()}>
                    {t(`global.paymentTermsOptions.${term.value}`)}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <InputError className="mt-2" message={errors.payment_terms} />
          </div>

          <Separator className="col-span-6" />
          <div className="col-span-6">
            <h4 className="font-medium">{t('global.additionalInfo')}</h4>
          </div>
          <div className="col-span-6 grid grid-cols-12 gap-4">
            <div className="col-span-3">
              <Label htmlFor="lead_time_days">{t('vendors.single.leadTimeDays')}</Label>
              <Input
                id="lead_time_days"
                type="number"
                min={0}
                className="mt-2 block w-full"
                value={data.lead_time_days}
                onChange={(e) => {
                  const value = e.target.valueAsNumber;
                  setData('lead_time_days', Number.isNaN(value) ? 0 : value);
                }}
                placeholder="0"
                readOnly={viewMode}
              />
              <InputError className="mt-2" message={errors.lead_time_days} />
            </div>

            <div className="col-span-9">
              <Label htmlFor="purchase_note">{t('vendors.single.purchaseNote')}</Label>
              <textarea
                id="purchase_note"
                className="mt-2 block w-full resize-none rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500"
                value={data.purchase_note}
                onChange={(e) => setData('purchase_note', e.target.value)}
                placeholder={t('vendors.single.purchaseNote')}
                rows={3}
                readOnly={viewMode}
              />
              <InputError className="mt-2" message={errors.purchase_note} />
            </div>

            {(params.vendor?.open_balance?.amount || 0) > 0 && (
              <h1 className="col-span-12 mt-2 font-extrabold text-red-500">
                Opening Balance: {currency(params.vendor?.open_balance?.amount || 0)} → Converted to INV-{params.vendor?.open_balance.invoice_id}
              </h1>
            )}

            {!viewMode && (
              <>
                <div className="col-span-3">
                  <Label htmlFor="open_balance">{t('vendors.single.openBalance')}</Label>
                  <Input
                    id="open_balance"
                    type="number"
                    className="mt-2 block w-full text-end"
                    value={data.open_balance}
                    onChange={(e) => setData('open_balance', e.target.valueAsNumber)}
                    required
                    placeholder={t('vendors.single.openBalance')}
                    readOnly={viewMode || (data.open_balance > 0 && !!params.vendor)}
                  />
                  <InputError className="mt-2" message={errors.open_balance} />
                </div>
                <div className="col-span-3">
                  <Label htmlFor="open_balance_as_of">{t('vendors.single.openBalanceDate')}</Label>
                  <Input
                    type="date"
                    className="mt-2 block w-full"
                    value={data.open_balance_as_of ? new Date(data.open_balance_as_of).toISOString().split('T')[0] : ''}
                    onChange={(e) => setData('open_balance_as_of', new Date(e.target.value))}
                    disabled={viewMode || data.open_balance === 0}
                  />
                </div>
              </>
            )}
          </div>
        </FormSection.Form>
        {!viewMode && (
          <FormSection.Actions>
            <Button disabled={processing} className="uppercase">
              {t(`global.actions.${params.action}`)} {t('global.vendor')}
            </Button>
          </FormSection.Actions>
        )}
      </FormSection>

      {viewMode && (
        <ActionSection>
          <ActionSection.Title>{t(`vendors.statuses.${params.vendor?.status || 'enabled'}.section.title`)}</ActionSection.Title>
          <ActionSection.Description>{t(`vendors.statuses.${params.vendor?.status || 'enabled'}.section.description`)}</ActionSection.Description>
          <ActionSection.Content>
            <div className={`space-y-4 rounded-lg border ${isDisabled ? 'border-primary-100 bg-primary-50' : 'border-red-100 bg-red-50'} p-4`}>
              <div className={`relative space-y-0.5 ${isDisabled ? 'text-primary' : 'text-red-600'}`}>
                <p className="font-medium">{t('global.warning.title')}</p>
                <p className="text-sm">{t('global.warning.description')}</p>
              </div>
              <Button variant={isDisabled ? 'default' : 'destructive'} onClick={() => setDialogOpen(true)}>
                {t(`vendors.statuses.${params.vendor?.status || 'enabled'}.section.title`)}
              </Button>

              <ConfirmsPassword
                title={t(`vendors.statuses.${params.vendor?.status || 'enabled'}.confirmsPassword.title`, {
                  vendor: params.vendor?.name || '',
                })}
                description={t(`vendors.statuses.${params.vendor?.status || 'enabled'}.confirmsPassword.description`)}
                action={t(`vendors.statuses.${params.vendor?.status || 'enabled'}.confirmsPassword.confirm`)}
                verb={'update'}
                path={`/vendors/${params.vendor?.id}/change-status`}
                open={dialogOpen}
                onOpenChange={setDialogOpen}
              />
            </div>
          </ActionSection.Content>
        </ActionSection>
      )}
    </div>
  );
}
