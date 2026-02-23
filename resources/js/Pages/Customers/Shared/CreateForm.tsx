import ActionSection from '@/components/action-section';
import { ConfirmsPassword } from '@/components/confirms-password';
import { DatePickerField } from '@/components/date-picker';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { paymentTerms } from '@/Pages/Invoices/constants';
import { Customer, CustomerType, CustomerTypes, PageProps, PaymentMethods, TaxReceipt, Verb } from '@/types';
import { Field, Radio, RadioGroup, Switch } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';
import { CheckCircleIcon } from 'lucide-react';
import { useState } from 'react';

export type CreateFormParams = {
  customer: Customer | undefined;
  tax_receipts: TaxReceipt[];
  action: Verb;
};

type CreateFormProps = {
  onFinish: () => void;
  params: CreateFormParams;
};

type CustomerForm = {
  id: number | undefined;
  name: string;
  contact: string;
  email: string;
  phone: string;
  payment_method?: string;
  payment_terms?: string;
  credit_limited: boolean;
  credit_limit?: number;
  customer_type: string;
  tax_receipt: number;
  open_balance: number;
  open_balance_as_of: Date | undefined;
};

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const currency = useNumber().currency;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { data, setData, post, put, errors, reset, processing } = useForm<Required<CustomerForm>>({
    id: params.customer?.id,
    name: params.customer?.name || '',
    contact: params.customer?.contact_name || '',
    email: params.customer?.email || '',
    phone: params.customer?.phone || '',
    payment_method: params.customer?.payment_method || '',
    payment_terms: params.customer?.payment_terms || '',
    credit_limited: params.customer?.credit_limited || false,
    credit_limit: params.customer?.credit_limit || 0,
    customer_type: params.customer?.customer_type || 'business',
    tax_receipt: params.customer?.tax_receipt || 0,
    open_balance: params.customer?.open_balance?.amount || 0,
    open_balance_as_of: params.customer?.open_balance?.date || undefined,
  });

  const viewMode = params.action === 'view';
  const isDisabled = params.customer?.status === 'disabled';

  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => {
      reset();
      onFinish();
    },
  };

  const verbName = useVerb().action(params.action);

  const submit = () => {
    if (params.action === 'create') post('/customers', options);
    if (params.action === 'edit') put(`/customers/${params.customer!.id}`, options);
  };

  return (
    <div className="flex flex-col space-y-6">
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('customers.single.title')}</FormSection.Title>
        <FormSection.Description>{t('customers.single.description')}</FormSection.Description>
        <FormSection.Form>
          {propsErrors.status && <div className="col-span-6 mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}
          <div className="col-span-6">
            <div className="flex flex-col gap-2">
              <Label htmlFor="customer_type">{t('customers.single.type')}</Label>
              <RadioGroup
                id="customer_type"
                className="grid grid-cols-3 gap-6"
                value={data.customer_type}
                onChange={(type: CustomerType) => setData('customer_type', type)}
              >
                {CustomerTypes.map((type: CustomerType) => (
                  <Field key={type}>
                    <Radio
                      value={type}
                      className="group data-checked:bg-primary data-checked:text-primary-foreground bg-primary/5 data-focus:outline-primary relative flex cursor-pointer grid-cols-1 rounded-lg px-5 py-4 shadow-md transition focus:not-data-focus:outline-none data-focus:outline"
                    >
                      <div className="flex w-full capitalize">{t(`customers.single.${type}`)}</div>
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
          <Separator className="col-span-6" />
          <div className="col-span-6">
            <h4 className="font-medium">{t('customers.single.paymentSection')}</h4>
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
          <div className="col-span-2">
            <Label htmlFor="credit_limit">{t('global.credit_limit')}</Label>
            <div className="mt-2 flex space-x-6">
              <Switch
                className={cn(
                  'relative inline-flex h-8 w-20 cursor-pointer items-center rounded-full transition',
                  data.credit_limited ? 'bg-primary' : 'bg-gray-300',
                )}
                disabled={viewMode}
                checked={data.credit_limited}
                onChange={(checked: boolean) => {
                  setData('credit_limited', checked);
                  if (!checked) setData('credit_limit', 0);
                }}
              >
                <span
                  className={cn(
                    'inline-block h-6 w-6 transform rounded-full bg-white transition',
                    data.credit_limited ? 'translate-x-6' : 'translate-x-1',
                  )}
                />
              </Switch>
              <Input
                id="credit_limit"
                type="number"
                className="text-end"
                value={data.credit_limit}
                onChange={(e) => setData('credit_limit', e.target.valueAsNumber)}
                disabled={!data.credit_limited}
                placeholder={t('global.credit_limit')}
                readOnly={viewMode}
              />
              <InputError className="mt-2" message={errors.credit_limit} />
            </div>
          </div>
          <Separator className="col-span-6" />
          <div className="col-span-6">
            <h4 className="font-medium">{t('global.additionalInfo')}</h4>
          </div>
          <div className="col-span-6 grid grid-cols-12 space-x-2">
            <div className="col-span-6">
              <Label>{t('global.taxReceipt')}</Label>
              <Select
                name="taxReceipt"
                onValueChange={(value) => setData('tax_receipt', Number(value))}
                value={String(data.tax_receipt)}
                required
                disabled={viewMode}
              >
                <SelectTrigger className="mt-2 w-full">
                  <SelectValue placeholder={t('customers.taxReceipt')} />
                </SelectTrigger>
                <SelectContent className="">
                  {params.tax_receipts.map((receipt) => (
                    <SelectItem key={receipt.id} value={String(receipt.id)}>
                      {receipt.name}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
              <InputError className="mt-2" message={errors.tax_receipt} />
            </div>
            {(params.customer?.open_balance?.amount || 0) > 0 && (
              <h1 className="col-span-12 mt-2 font-extrabold text-red-500">
                Opening Balance: {currency(params.customer?.open_balance?.amount || 0)} → Converted to INV-{params.customer?.open_balance.invoice_id}
              </h1>
            )}
            {!viewMode && (
              <>
                <div className="col-span-3">
                  <Label htmlFor="open_balance">{t('customers.single.openBalance')}</Label>
                  <Input
                    id="open_balance"
                    type="number"
                    className="mt-2 block w-full text-end"
                    value={data.open_balance}
                    onChange={(e) => setData('open_balance', e.target.valueAsNumber)}
                    required
                    placeholder={t('customers.single.openBalance')}
                    readOnly={viewMode}
                  />
                  <InputError className="mt-2" message={errors.open_balance} />
                </div>
                <div className="col-span-3">
                  <DatePickerField
                    id="date"
                    label={t('customers.single.openBalanceAsOf')}
                    placeholder={t('global.datePlaceholder')}
                    value={data.open_balance_as_of}
                    onChange={(value) => setData('open_balance_as_of', value)}
                    error={errors.open_balance_as_of}
                    className="w-52"
                  />
                </div>
              </>
            )}
          </div>
        </FormSection.Form>
        {!viewMode && (
          <FormSection.Actions>
            <Button disabled={processing} className="uppercase">
              {t(`global.actions.${verbName}`)} {t('global.customer')}
            </Button>
          </FormSection.Actions>
        )}
      </FormSection>

      {viewMode && (
        <ActionSection>
          <ActionSection.Title>{t(`customers.statuses.${params.customer?.status || 'enabled'}.section.title`)}</ActionSection.Title>
          <ActionSection.Description>{t(`customers.statuses.${params.customer?.status || 'enabled'}.section.description`)}</ActionSection.Description>
          <ActionSection.Content>
            <div className={`space-y-4 rounded-lg border ${isDisabled ? 'border-primary-100 bg-primary-50' : 'border-red-100 bg-red-50'} p-4`}>
              <div className={`relative space-y-0.5 ${isDisabled ? 'text-primary' : 'text-red-600'}`}>
                <p className="font-medium">{t('global.warning.title')}</p>
                <p className="text-sm">{t('global.warning.description')}</p>
              </div>
              <Button variant={isDisabled ? 'default' : 'destructive'} onClick={() => setDialogOpen(true)}>
                {t(`customers.statuses.${params.customer?.status || 'enabled'}.section.title`)}
              </Button>

              <ConfirmsPassword
                title={t(`customers.statuses.${params.customer?.status || 'enabled'}.confirmsPassword.title`, {
                  customer: params.customer?.name || '',
                })}
                description={t(`customers.statuses.${params.customer?.status || 'enabled'}.confirmsPassword.description`)}
                action={t(`customers.statuses.${params.customer?.status || 'enabled'}.confirmsPassword.confirm`)}
                verb={'update'}
                path={`/customers/${params.customer?.id}/change-status`}
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
