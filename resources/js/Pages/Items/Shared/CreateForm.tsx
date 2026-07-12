import ActionSection from '@/components/action-section';
import { ConfirmsPassword } from '@/components/confirms-password';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useHeader } from '@/composables/use-headers';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import { generateVariantCombinations } from '@/lib/variants';
import { Item, ItemType, ItemTypes, PageProps, Tax, Unit, Verb } from '@/types';
import { Field, Radio, RadioGroup } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';
import { CheckCircleIcon } from 'lucide-react';
import { useMemo, useState } from 'react';

// Attribute options offered for building the variant matrix. Provided by the
// Items page (findAttributesWithValues).
export type ItemAttributeValueOption = { id: number; value: string; display_name: string };
export type ItemAttributeOption = { id: number; name: string; display_name: string; values?: ItemAttributeValueOption[] };

// One variant to create: which attribute value it maps to per attribute, plus
// optional per-variant overrides the backend understands.
type VariantComboForm = {
  attribute_value_ids: Record<number, number>;
  sku?: string;
  price?: number;
  cost_price?: number;
};

export type CreateFormParams = {
  item: Item | undefined;
  taxes: Tax[];
  units: Unit[];
  attributes: ItemAttributeOption[];
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

type ComboOverride = { price?: number; cost_price?: number; sku?: string };

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors, variantsEnabled, defaultTaxId } = usePage<
    PageProps & { variantsEnabled?: boolean; defaultTaxId?: number | null }
  >().props;
  const { data, setData, post, put, transform, errors, reset, processing } = useForm<Required<ItemForm>>({
    id: params.item?.id,
    name: params.item?.name || '',
    description: params.item?.description || '',
    price: params.item?.price || 0,
    tax_id: params.item?.tax.id || defaultTaxId || 0,
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

  // ── Variant matrix state (create-only; products only) ─────────────────────
  const attributes = params.attributes || [];
  const [hasVariants, setHasVariants] = useState(false);
  // Selected attribute value ids, keyed by attribute id.
  const [valuesByAttribute, setValuesByAttribute] = useState<Record<number, number[]>>({});
  // Per-combo overrides, keyed by a stable combo signature.
  const [overrides, setOverrides] = useState<Record<string, ComboOverride>>({});

  const canUseVariants = variantsEnabled === true && params.action === 'create' && data.item_type === 'product' && attributes.length > 0;

  const selectedAttributeIds = useMemo(
    () =>
      Object.keys(valuesByAttribute)
        .map(Number)
        .filter((id) => (valuesByAttribute[id] || []).length > 0),
    [valuesByAttribute],
  );

  // Cartesian product of the selected values per attribute -> attribute_value_ids maps.
  const combos = useMemo<Record<number, number>[]>(() => {
    const active: Record<number, number[]> = {};
    for (const attrId of selectedAttributeIds) {
      active[attrId] = valuesByAttribute[attrId];
    }
    if (Object.keys(active).length === 0) return [];
    return generateVariantCombinations(active) as Record<number, number>[];
  }, [selectedAttributeIds, valuesByAttribute]);

  const comboKey = (combo: Record<number, number>) =>
    Object.keys(combo)
      .map(Number)
      .sort((a, b) => a - b)
      .map((attrId) => `${attrId}:${combo[attrId]}`)
      .join('|');

  const valueLabel = (attrId: number, valueId: number) => {
    const attr = attributes.find((a) => a.id === attrId);
    const value = attr?.values?.find((v) => v.id === valueId);
    return value?.display_name || value?.value || String(valueId);
  };

  const comboLabel = (combo: Record<number, number>) =>
    Object.keys(combo)
      .map(Number)
      .map((attrId) => valueLabel(attrId, combo[attrId]))
      .join(' / ');

  const toggleAttribute = (attrId: number, on: boolean) => {
    setValuesByAttribute((current) => {
      const next = { ...current };
      if (on) next[attrId] = next[attrId] || [];
      else delete next[attrId];
      return next;
    });
  };

  const toggleValue = (attrId: number, valueId: number, on: boolean) => {
    setValuesByAttribute((current) => {
      const existing = current[attrId] || [];
      const next = on ? [...new Set([...existing, valueId])] : existing.filter((id) => id !== valueId);
      return { ...current, [attrId]: next };
    });
  };

  const setOverride = (key: string, patch: ComboOverride) => setOverrides((current) => ({ ...current, [key]: { ...current[key], ...patch } }));

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
    const identifiers = {
      reference: data.reference,
      code: data.code,
      sku: data.sku,
      barcode: data.barcode,
      vendor_reference: data.vendor_reference,
    };

    // Variant-bearing product: post the matrix to the dedicated endpoint.
    if (hasVariants && combos.length > 0) {
      const variant_combos: VariantComboForm[] = combos.map((combo) => {
        const key = comboKey(combo);
        const override = overrides[key] || {};
        return {
          attribute_value_ids: combo,
          sku: override.sku,
          price: override.price,
          cost_price: override.cost_price,
        };
      });

      transform(() => ({
        name: data.name,
        description: data.description,
        price: data.price,
        tax_id: data.tax_id,
        unit_id: data.unit_id,
        item_type: data.item_type,
        identifiers,
        attribute_ids: selectedAttributeIds,
        variant_combos,
      }));
      post('/items/variants', options);
      return;
    }

    transform((form) => {
      const { reference, code, sku, barcode, vendor_reference, ...rest } = form;
      void reference;
      void code;
      void sku;
      void barcode;
      void vendor_reference;
      return { ...rest, identifiers };
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
                    placeholder="0.00"
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

            {canUseVariants && (
              <div className="flex flex-col gap-4 rounded-lg border border-gray-200 p-4">
                <label className="flex items-center gap-2">
                  <Checkbox checked={hasVariants} onCheckedChange={(checked) => setHasVariants(checked === true)} />
                  <span className="text-sm font-medium">{t('items.single.hasVariants')}</span>
                </label>

                {hasVariants && (
                  <div className="flex flex-col gap-4">
                    {attributes.map((attr) => {
                      const selectedValues = valuesByAttribute[attr.id] || [];
                      const included = attr.id in valuesByAttribute;
                      return (
                        <div key={attr.id} className="flex flex-col gap-2">
                          <label className="flex items-center gap-2">
                            <Checkbox checked={included} onCheckedChange={(checked) => toggleAttribute(attr.id, checked === true)} />
                            <span className="text-sm font-medium">{attr.display_name}</span>
                          </label>
                          {included && (
                            <div className="flex flex-wrap gap-2 pl-6">
                              {(attr.values || []).map((value) => {
                                const on = selectedValues.includes(value.id);
                                return (
                                  <button
                                    type="button"
                                    key={value.id}
                                    onClick={() => toggleValue(attr.id, value.id, !on)}
                                    className={`rounded-full border px-3 py-1 text-sm transition ${
                                      on ? 'bg-primary text-primary-foreground border-primary' : 'border-gray-300 bg-white'
                                    }`}
                                  >
                                    {value.display_name || value.value}
                                  </button>
                                );
                              })}
                            </div>
                          )}
                        </div>
                      );
                    })}

                    {combos.length > 0 && (
                      <div className="overflow-x-auto">
                        <table className="w-full text-sm">
                          <thead>
                            <tr className="text-left text-gray-500">
                              <th className="py-1 pr-4">{t('items.single.variant')}</th>
                              <th className="py-1 pr-4">SKU</th>
                              <th className="py-1 pr-4">{t('global.price')}</th>
                              <th className="py-1 pr-4">{t('items.single.cost')}</th>
                            </tr>
                          </thead>
                          <tbody>
                            {combos.map((combo) => {
                              const key = comboKey(combo);
                              const override = overrides[key] || {};
                              return (
                                <tr key={key} className="border-t border-gray-100">
                                  <td className="py-1 pr-4">{comboLabel(combo)}</td>
                                  <td className="py-1 pr-4">
                                    <Input
                                      className="h-8 w-32"
                                      value={override.sku ?? ''}
                                      onChange={(e) => setOverride(key, { sku: e.target.value })}
                                      placeholder="auto"
                                    />
                                  </td>
                                  <td className="py-1 pr-4">
                                    <Input
                                      type="number"
                                      className="h-8 w-24 text-right"
                                      value={override.price ?? ''}
                                      onChange={(e) => setOverride(key, { price: e.target.valueAsNumber })}
                                      placeholder="0.00"
                                    />
                                  </td>
                                  <td className="py-1 pr-4">
                                    <Input
                                      type="number"
                                      className="h-8 w-24 text-right"
                                      value={override.cost_price ?? ''}
                                      onChange={(e) => setOverride(key, { cost_price: e.target.valueAsNumber })}
                                      placeholder="0.00"
                                    />
                                  </td>
                                </tr>
                              );
                            })}
                          </tbody>
                        </table>
                      </div>
                    )}
                  </div>
                )}
              </div>
            )}
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
