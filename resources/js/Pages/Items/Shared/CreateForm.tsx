import ActionSection from '@/components/action-section';
import { ConfirmsPassword } from '@/components/confirms-password';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useHeader } from '@/composables/use-headers';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import { Item, ItemType, ItemTypes, PageProps, Tax, Unit, Verb } from '@/types';
import { Field, Radio, RadioGroup } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';
import { CheckCircleIcon } from 'lucide-react';
import { useState } from 'react';

export type CreateFormParams = {
  item: Item | undefined;
  taxes: Tax[];
  units: Unit[];
  action: Verb;
};

type CreateFormProps = {
  onFinish: () => void;
  params: CreateFormParams;
};

type ItemForm = {
  id: number | undefined;
  name: string;
  description: string;
  price: number;
  tax_id: number;
  unit_id: number;
  item_type: ItemType; // This can be 'product' or 'service'
  reference?: string;
  code?: string;
  sku?: string;
  barcode?: string;
  vendor_reference?: string;
};

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { data, setData, post, put, transform, errors, reset, processing } = useForm<Required<ItemForm>>({
    id: params.item?.id,
    name: params.item?.name || '',
    description: params.item?.description || '',
    price: params.item?.price || 0,
    tax_id: params.item?.tax.id || 0,
    unit_id: params.item?.unit.id || 0,
    item_type: params.item?.item_type || 'product', // Default to 'product'
    reference: params.item?.identifiers?.reference || '',
    code: params.item?.identifiers?.code || '',
    sku: params.item?.identifiers?.sku || '',
    barcode: params.item?.identifiers?.barcode || '',
    vendor_reference: params.item?.identifiers?.vendor_reference || '',
  });

  const viewMode = params.action === 'view';
  const isDisabled = params.item?.status === 'disabled';

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
    transform((data) => {
      const { reference, code, sku, barcode, vendor_reference, ...rest } = data;
      return {
        ...rest,
        identifiers: {
          reference,
          code,
          sku,
          barcode,
          vendor_reference,
        },
      };
    });
    if (params.action === 'create') post('/items', options);
    if (params.action === 'edit') put(`/items/${params.item!.id}`, options);
  };

  return (
    <div className="flex flex-col space-y-6">
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('items.single.title')}</FormSection.Title>
        <FormSection.Description>{t('items.single.description')}</FormSection.Description>
        <FormSection.Form>
          {propsErrors.status && <div className="mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}

          <div className="col-span-4 flex flex-col gap-y-6">
            <div className="flex flex-col gap-2">
              <Label htmlFor="item_type">{t('items.single.type')}</Label>
              <RadioGroup className="grid grid-cols-3 gap-6" value={data.item_type} onChange={(type: ItemType) => setData('item_type', type)}>
                {ItemTypes.map((type: ItemType) => (
                  <Field key={type}>
                    <Radio
                      value={type}
                      className="group data-checked:bg-primary data-checked:text-primary-foreground bg-primary/5 data-focus:outline-primary relative flex cursor-pointer grid-cols-1 rounded-lg px-5 py-4 shadow-md transition focus:not-data-focus:outline-none data-focus:outline"
                    >
                      <div className="flex w-full capitalize">{t(`items.single.${type}`)}</div>
                      <CheckCircleIcon className="size-6 opacity-0 transition group-data-checked:opacity-100" />
                    </Radio>
                  </Field>
                ))}
              </RadioGroup>
            </div>
            <div className="flex flex-col gap-2">
              <div className="gap-2">
                <Label htmlFor="name">{t('global.name')}</Label>
                <Input
                  id="name"
                  className="mt-1 block w-full"
                  value={data.name}
                  onChange={(e) => setData('name', e.target.value)}
                  required
                  autoComplete="name"
                  placeholder="Item name"
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.name} />
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <div className="flex items-center justify-between gap-x-2">
                <div className="flex flex-col gap-2">
                  <Label>{t('global.unit')}</Label>
                  <Select
                    onValueChange={(value) => setData('unit_id', Number(value))}
                    disabled={viewMode}
                    defaultValue={data.unit_id.toString()}
                    value={data.unit_id.toString()}
                    required
                  >
                    <SelectTrigger className="w-44">
                      <SelectValue placeholder="Select item unit" />
                    </SelectTrigger>
                    <SelectContent className="">
                      {params.units.map((unit) => (
                        <SelectItem key={unit.id} value={unit.id.toString()}>
                          {unit.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <InputError className="mt-2" message={errors.unit_id} />
                </div>
                <div className="flex flex-col gap-2">
                  <Label>{t('global.tax')}</Label>
                  <Select
                    onValueChange={(value) => setData('tax_id', Number(value))}
                    disabled={viewMode}
                    defaultValue={data.tax_id.toString()}
                    value={data.tax_id.toString()}
                    required
                  >
                    <SelectTrigger className="w-44">
                      <SelectValue placeholder="Select item tax" />
                    </SelectTrigger>
                    <SelectContent className="">
                      {params.taxes.map((tax) => (
                        <SelectItem key={tax.id} value={tax.id.toString()}>
                          {tax.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <InputError className="mt-2" message={errors.tax_id} />
                </div>
                <div className="flex flex-col gap-2">
                  <Label htmlFor="price">{t('global.price')}</Label>
                  <Input
                    id="price"
                    type="number"
                    className="mt-0 block w-full max-w-40 text-right"
                    value={data.price}
                    onChange={(e) => setData('price', e.target.valueAsNumber)}
                    placeholder="Jane Doe"
                    readOnly={viewMode}
                  />
                  <InputError className="mt-2" message={errors.price} />
                </div>
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <div className="grid gap-2">
                <Label htmlFor="description">{t('global.description')}</Label>
                <textarea
                  id="description"
                  className="mt-1 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500"
                  value={data.description}
                  onChange={(e) => setData('description', e.target.value)}
                  placeholder="Wrire some description here..."
                  rows={3}
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.description} />
              </div>
            </div>
          </div>
          <div className="col-span-2 space-y-6">
            <div className="flex flex-col gap-2">
              <div className="gap-2">
                <Label htmlFor="reference">{t('global.reference')}</Label>
                <Input
                  id="reference"
                  className="mt-1 block w-full"
                  value={data.reference}
                  onChange={(e) => setData('reference', e.target.value)}
                  autoComplete="reference"
                  placeholder="Item reference"
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.reference} />
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <div className="gap-2">
                <Label htmlFor="code">{t('global.code')}</Label>
                <Input
                  id="code"
                  className="mt-1 block w-full"
                  value={data.code}
                  onChange={(e) => setData('code', e.target.value)}
                  autoComplete="code"
                  placeholder="Item code"
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.code} />
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <div className="gap-2">
                <Label htmlFor="sku">{t('global.sku')}</Label>
                <Input
                  id="sku"
                  className="mt-1 block w-full"
                  value={data.sku}
                  onChange={(e) => setData('sku', e.target.value)}
                  autoComplete="sku"
                  placeholder="Item sku"
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.sku} />
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <div className="gap-2">
                <Label htmlFor="barcode">{t('global.barcode')}</Label>
                <Input
                  id="barcode"
                  className="mt-1 block w-full"
                  value={data.barcode}
                  onChange={(e) => setData('barcode', e.target.value)}
                  autoComplete="barcode"
                  placeholder="Item barcode"
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.barcode} />
              </div>
            </div>
            <div className="flex flex-col gap-2">
              <div className="gap-2">
                <Label htmlFor="vendor_reference">{t('items.single.vendor_reference')}</Label>
                <Input
                  id="vendor_reference"
                  className="mt-1 block w-full"
                  value={data.vendor_reference}
                  onChange={(e) => setData('vendor_reference', e.target.value)}
                  placeholder="Item vendor reference"
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.vendor_reference} />
              </div>
            </div>
          </div>
        </FormSection.Form>
        {!viewMode && (
          <FormSection.Actions>
            <div className="flex items-center gap-4">
              <Button disabled={processing} className="uppercase">
                {t(`global.actions.${verbName}`)} {t('global.item')}
              </Button>
            </div>
          </FormSection.Actions>
        )}
      </FormSection>

      {viewMode && (
        <ActionSection>
          <ActionSection.Title>{t(`items.statuses.${params.item?.status || 'enabled'}.section.title`)}</ActionSection.Title>
          <ActionSection.Description>{t(`items.statuses.${params.item?.status || 'enabled'}.section.description`)}</ActionSection.Description>
          <ActionSection.Content>
            <div className={`space-y-2 rounded-lg border ${isDisabled ? 'border-primary-100 bg-primary-50' : 'border-red-100 bg-red-50'} p-4`}>
              <div className={`relative space-y-0.5 ${isDisabled ? 'text-primary' : 'text-red-600'}`}>
                <p className="font-medium">{t('global.warning.title')}</p>
                <p className="text-sm">{t('global.warning.description')}</p>
              </div>
              <Button variant={isDisabled ? 'default' : 'destructive'} onClick={() => setDialogOpen(true)}>
                {t(`items.statuses.${params.item?.status || 'enabled'}.section.title`)}
              </Button>

              <ConfirmsPassword
                title={t(`items.statuses.${params.item?.status || 'enabled'}.confirmsPassword.title`, { item: params.item?.name || '' })}
                description={t(`items.statuses.${params.item?.status || 'enabled'}.confirmsPassword.description`)}
                action={t(`items.statuses.${params.item?.status || 'enabled'}.confirmsPassword.confirm`)}
                verb={'update'}
                path={`/items/${params.item?.id}/change-status`}
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
