import ActionSection from '@/components/action-section';
import { ConfirmsPassword } from '@/components/confirms-password';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Checkbox } from '@/components/ui/checkbox';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { useHeader } from '@/composables/use-headers';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import { Item, ItemType, ItemTypes, PageProps, Tax, Unit, Verb } from '@/types';
import { Field, Radio, RadioGroup } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';
import { CheckCircleIcon } from 'lucide-react';
import { useEffect, useState } from 'react';

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
  has_variants: boolean;
  attribute_ids: number[];
  variant_combos: VariantComboForm[];
};

type VariantComboForm = {
  attribute_value_ids: Record<number, number>;
  sku?: string;
  price?: number;
  cost_price?: number;
  barcode?: string;
  reference?: string;
  vendor_reference?: string;
  active?: boolean;
};

type VariantComboPreview = {
  key: string;
  attribute_value_ids: Record<number, number>;
  label: string;
  sku?: string;
  price?: number;
  cost_price?: number;
  barcode?: string;
  reference?: string;
  vendor_reference?: string;
  active?: boolean;
};

type ItemAttributeValueOption = {
  id: number;
  value: string;
  display_name: string;
};

export type ItemAttributeOption = {
  id: number;
  name: string;
  display_name: string;
  values?: ItemAttributeValueOption[];
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
    tax_id: params.item?.tax?.id || 0,
    unit_id: params.item?.unit?.id || 0,
    item_type: params.item?.item_type || 'product', // Default to 'product'
    reference: params.item?.identifiers?.reference || '',
    code: params.item?.identifiers?.code || '',
    sku: params.item?.identifiers?.sku || '',
    barcode: params.item?.identifiers?.barcode || '',
    vendor_reference: params.item?.identifiers?.vendor_reference || '',
    has_variants: params.item?.has_variants || false,
    attribute_ids: [],
    variant_combos: [],
  });
  const [selectedAttributeIDs, setSelectedAttributeIDs] = useState<number[]>([]);
  const [selectedValuesByAttribute, setSelectedValuesByAttribute] = useState<Record<number, number[]>>({});
  const [variantPriceOverrides, setVariantPriceOverrides] = useState<Record<string, number | undefined>>({});
  const [variantSKUOverrides, setVariantSKUOverrides] = useState<Record<string, string | undefined>>({});
  const [variantBarcodeOverrides, setVariantBarcodeOverrides] = useState<Record<string, string | undefined>>({});
  const [variantReferenceOverrides, setVariantReferenceOverrides] = useState<Record<string, string | undefined>>({});
  const [variantVendorRefOverrides, setVariantVendorRefOverrides] = useState<Record<string, string | undefined>>({});
  const [variantCostPriceOverrides, setVariantCostPriceOverrides] = useState<Record<string, number | undefined>>({});
  const [variantActiveOverrides, setVariantActiveOverrides] = useState<Record<string, boolean>>({});
  const [variantSetupError, setVariantSetupError] = useState<string>('');

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
  const attributeOptions = Array.isArray(params.attributes) ? params.attributes : [];
  const existingVariantSetup = params.item?.variant_setup;
  const existingHasVariants =
    !!existingVariantSetup?.has_variants ||
    (existingVariantSetup?.attribute_ids?.length || 0) > 0 ||
    Object.keys(existingVariantSetup?.selected_values_by_attribute || {}).length > 0 ||
    (existingVariantSetup?.variants?.length || 0) > 1;

  useEffect(() => {
    if (params.action !== 'edit' || params.item?.item_type !== 'product') {
      return;
    }

    const nextAttributeIDs = Array.isArray(existingVariantSetup?.attribute_ids)
      ? existingVariantSetup!.attribute_ids.map((attributeID) => Number(attributeID)).filter((attributeID) => !Number.isNaN(attributeID))
      : [];

    const nextSelectedValuesByAttribute = Object.entries(existingVariantSetup?.selected_values_by_attribute || {}).reduce<Record<number, number[]>>(
      (current, [attributeID, valueIDs]) => {
        const normalizedValues = Array.isArray(valueIDs)
          ? valueIDs.map((valueID) => Number(valueID)).filter((valueID) => !Number.isNaN(valueID))
          : [];
        current[Number(attributeID)] = normalizedValues;
        return current;
      },
      {},
    );

    setSelectedAttributeIDs(nextAttributeIDs);
    setSelectedValuesByAttribute(nextSelectedValuesByAttribute);
    setVariantPriceOverrides({});
    setVariantSKUOverrides({});
    setData('has_variants', existingHasVariants);
  }, [params.action, params.item?.id]);

  const selectedAttributeLabels = (existingVariantSetup?.attribute_ids || [])
    .map((attributeID) => {
      const selectedAttribute = attributeOptions.find((attribute) => attribute.id === Number(attributeID));
      return selectedAttribute?.display_name || selectedAttribute?.name || '';
    })
    .filter((label) => label.length > 0);

  const selectedAttributeValueLabels = Object.entries(existingVariantSetup?.selected_values_by_attribute || {}).map(([attributeID, valueIDs]) => {
    const selectedAttribute = attributeOptions.find((attribute) => attribute.id === Number(attributeID));

    const normalizedValueIDs = Array.isArray(valueIDs) ? valueIDs.map((valueID) => Number(valueID)).filter((valueID) => !Number.isNaN(valueID)) : [];

    const labels = normalizedValueIDs
      .map((valueID) => selectedAttribute?.values?.find((value) => value.id === valueID))
      .filter((value) => !!value)
      .map((value) => value!.display_name || value!.value);

    return {
      attribute: selectedAttribute?.display_name || selectedAttribute?.name || attributeID,
      labels,
    };
  });

  const existingVariantSignatures = new Set(existingVariantSetup?.existing_signatures || []);

  const buildVariantKey = (selection: Record<number, number>): string => {
    const sortedAttributeIDs = [...selectedAttributeIDs].sort((left, right) => left - right);
    return sortedAttributeIDs.map((attributeID) => `${attributeID}:${selection[attributeID]}`).join('|');
  };

  const buildVariantLabel = (selection: Record<number, number>): string => {
    return selectedAttributeIDs
      .map((attributeID) => {
        const attribute = attributeOptions.find((entry) => entry.id === attributeID);
        const valueID = selection[attributeID];
        const value = attribute?.values?.find((entry) => entry.id === valueID);
        return value?.display_name || value?.value || `${attribute?.display_name || attribute?.name || attributeID}:${valueID}`;
      })
      .join(' / ');
  };

  const buildVariantPreviews = (): VariantComboPreview[] => {
    if (selectedAttributeIDs.length === 0) {
      return [];
    }

    const valueGroups = selectedAttributeIDs.map((attributeID) => selectedValuesByAttribute[attributeID] || []);
    if (valueGroups.some((group) => group.length === 0)) {
      return [];
    }

    const combos: VariantComboPreview[] = [];
    const build = (attributeIndex: number, current: Record<number, number>) => {
      if (attributeIndex === selectedAttributeIDs.length) {
        const key = buildVariantKey(current);
        combos.push({
          key,
          attribute_value_ids: current,
          label: buildVariantLabel(current),
          sku: variantSKUOverrides[key],
          price: variantPriceOverrides[key],
          cost_price: variantCostPriceOverrides[key],
          barcode: variantBarcodeOverrides[key],
          reference: variantReferenceOverrides[key],
          vendor_reference: variantVendorRefOverrides[key],
          active: variantActiveOverrides[key] !== undefined ? variantActiveOverrides[key] : true,
        });
        return;
      }

      const attributeID = selectedAttributeIDs[attributeIndex];
      for (const valueID of valueGroups[attributeIndex]) {
        build(attributeIndex + 1, {
          ...current,
          [attributeID]: valueID,
        });
      }
    };

    build(0, {});
    return combos;
  };

  const generateSuggestedVariantSKU = (combo: VariantComboPreview): string => {
    const itemNameChunk = (data.name || 'ITEM')
      .toUpperCase()
      .replace(/[^A-Z0-9]/g, '')
      .slice(0, 8);

    const seed = `${data.name}-${combo.key}`;
    const hash = Array.from(seed).reduce((acc, char) => ((acc * 31 + char.charCodeAt(0)) >>> 0) % 2176782336, 17);
    const hashChunk = hash.toString(36).toUpperCase().padStart(6, '0').slice(0, 6);

    return `SKU-${itemNameChunk || 'ITEM'}-${hashChunk}`;
  };

  const buildSmartVariantSKU = (combo: VariantComboPreview): string => {
    // Extract first 3-4 chars from item name
    const itemPrefix = (data.name || 'ITEM')
      .toUpperCase()
      .replace(/[^A-Z0-9]/g, '')
      .slice(0, 4);

    // Get abbreviated attribute values
    const attributeParts = selectedAttributeIDs
      .map((attributeID) => {
        const attribute = attributeOptions.find((attr) => attr.id === attributeID);
        const valueID = combo.attribute_value_ids[attributeID];
        const value = attribute?.values?.find((v) => v.id === valueID);
        const displayValue = value?.display_name || value?.value || '';
        // Abbreviate to 3-4 chars, remove special chars
        return displayValue
          .toUpperCase()
          .replace(/[^A-Z0-9]/g, '')
          .slice(0, 4);
      })
      .filter((part) => part);

    // Join with hyphens: ITEM-ATTR1-ATTR2
    const parts = [itemPrefix, ...attributeParts];
    return parts.join('-').slice(0, 50); // Max 50 chars
  };

  const generateEAN13Barcode = (variantIndex: number): string => {
    // Simple EAN-13 compatible format: company prefix (6 digits) + item sequence (5 digits) + check digit
    // For demo purposes, using a placeholder company prefix. In production, this would come from company settings.
    const companyPrefix = '100000'; // 6 digits
    const itemSequence = String(variantIndex + 1).padStart(5, '0'); // 5 digits
    const productCode = companyPrefix + itemSequence; // 11 digits

    // Calculate EAN-13 check digit
    const calculateCheckDigit = (code: string): string => {
      const digits = code.split('').map(Number);
      const sum = digits.reduce((acc, digit, idx) => {
        return acc + digit * (idx % 2 === 0 ? 1 : 3);
      }, 0);
      const checkDigit = (10 - (sum % 10)) % 10;
      return String(checkDigit);
    };

    return productCode + calculateCheckDigit(productCode);
  };

  const handleBulkApplyPrice = (price: number) => {
    const updates: Record<string, number> = {};
    variantComboPreviews.forEach((combo) => {
      if (!existingVariantSignatures.has(combo.key)) {
        updates[combo.key] = price;
      }
    });
    setVariantPriceOverrides((current) => ({ ...current, ...updates }));
  };

  const handleBulkGenerateSKUs = () => {
    const updates: Record<string, string> = {};
    variantComboPreviews.forEach((combo) => {
      if (!existingVariantSignatures.has(combo.key) && !variantSKUOverrides[combo.key]) {
        updates[combo.key] = buildSmartVariantSKU(combo);
      }
    });
    setVariantSKUOverrides((current) => ({ ...current, ...updates }));
  };

  const handleBulkGenerateBarcodes = () => {
    const updates: Record<string, string> = {};
    variantComboPreviews.forEach((combo, index) => {
      if (!existingVariantSignatures.has(combo.key) && !variantBarcodeOverrides[combo.key]) {
        updates[combo.key] = generateEAN13Barcode(index);
      }
    });
    setVariantBarcodeOverrides((current) => ({ ...current, ...updates }));
  };

  const variantComboPreviews = buildVariantPreviews();

  useEffect(() => {
    if (!data.has_variants || variantComboPreviews.length === 0) {
      setVariantPriceOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantSKUOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantBarcodeOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantReferenceOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantVendorRefOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantCostPriceOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantActiveOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      return;
    }

    // Sync overrides with current previews
    const syncOverrides = <T,>(current: Record<string, T>, defaultValue?: (combo: VariantComboPreview) => T): Record<string, T> => {
      const next: Record<string, T> = {};
      for (const combo of variantComboPreviews) {
        if (current[combo.key] !== undefined) {
          next[combo.key] = current[combo.key];
        } else if (defaultValue) {
          next[combo.key] = defaultValue(combo);
        }
      }
      return next;
    };

    setVariantPriceOverrides((current) => syncOverrides(current));
    setVariantSKUOverrides((current) => syncOverrides(current, generateSuggestedVariantSKU));
    setVariantBarcodeOverrides((current) => syncOverrides(current));
    setVariantReferenceOverrides((current) => syncOverrides(current));
    setVariantVendorRefOverrides((current) => syncOverrides(current));
    setVariantCostPriceOverrides((current) => syncOverrides(current));
    setVariantActiveOverrides((current) => syncOverrides(current));
  }, [data.has_variants, data.name, selectedAttributeIDs.join(','), JSON.stringify(selectedValuesByAttribute)]);

  const buildVariantCombos = (): VariantComboForm[] => {
    return variantComboPreviews.map((combo) => ({
      attribute_value_ids: combo.attribute_value_ids,
      price: variantPriceOverrides[combo.key] ?? data.price,
      cost_price: variantCostPriceOverrides[combo.key],
      sku: (variantSKUOverrides[combo.key] || '').trim() || undefined,
      barcode: (variantBarcodeOverrides[combo.key] || '').trim() || undefined,
      reference: (variantReferenceOverrides[combo.key] || '').trim() || undefined,
      vendor_reference: (variantVendorRefOverrides[combo.key] || '').trim() || undefined,
      active: variantActiveOverrides[combo.key] !== undefined ? variantActiveOverrides[combo.key] : true,
    }));
  };

  const toggleAttribute = (attributeID: number, checked: boolean) => {
    setSelectedAttributeIDs((current) => {
      if (checked) {
        return current.includes(attributeID) ? current : [...current, attributeID];
      }

      return current.filter((id) => id !== attributeID);
    });

    if (!checked) {
      setSelectedValuesByAttribute((current) => {
        const next = { ...current };
        delete next[attributeID];
        return next;
      });
    }
  };

  const toggleAttributeValue = (attributeID: number, valueID: number, checked: boolean) => {
    setSelectedValuesByAttribute((current) => {
      const existing = current[attributeID] || [];
      const next = checked ? Array.from(new Set([...existing, valueID])) : existing.filter((id) => id !== valueID);

      return {
        ...current,
        [attributeID]: next,
      };
    });
  };

  const submit = () => {
    setVariantSetupError('');

    const hasVariantSetup = params.action !== 'view' && data.item_type === 'product' && data.has_variants;
    const variantCombos = hasVariantSetup ? buildVariantCombos() : [];

    if (hasVariantSetup) {
      if (selectedAttributeIDs.length === 0) {
        setVariantSetupError(t('items.single.variantSetup.attributeRequired'));
        return;
      }

      if (variantCombos.length === 0) {
        setVariantSetupError(t('items.single.variantSetup.valueRequired'));
        return;
      }
    }

    transform((data) => {
      const { reference, code, sku, barcode, vendor_reference, ...rest } = data;

      return {
        ...rest,
        has_variants: data.item_type === 'product' ? data.has_variants : false,
        attribute_ids: hasVariantSetup ? selectedAttributeIDs : [],
        variant_combos: hasVariantSetup ? variantCombos : [],
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
    if (params.action === 'edit') put(`/items/${params.item!.uuid}`, options);
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
              <RadioGroup
                className="grid grid-cols-3 gap-6"
                value={data.item_type}
                onChange={(type: ItemType) => {
                  setData('item_type', type);

                  if (type !== 'product') {
                    setData('has_variants', false);
                    setSelectedAttributeIDs([]);
                    setSelectedValuesByAttribute({});
                    setVariantSetupError('');
                  }
                }}
              >
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

            {params.action !== 'view' && data.item_type === 'product' && (
              <div className="border-border space-y-4 rounded-lg border p-4">
                <div className="flex items-center justify-between gap-4">
                  <div className="space-y-1">
                    <Label htmlFor="has_variants">{t('items.single.hasVariants')}</Label>
                    <p className="text-muted-foreground text-sm">{t('items.single.hasVariantsHelp')}</p>
                  </div>
                  <Switch
                    id="has_variants"
                    checked={data.has_variants}
                    onCheckedChange={(checked) => {
                      setData('has_variants', checked);
                      if (!checked) {
                        setSelectedAttributeIDs([]);
                        setSelectedValuesByAttribute({});
                        setVariantSetupError('');
                      }
                    }}
                  />
                </div>

                {data.has_variants && (
                  <div className="space-y-4">
                    <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.priceTemplateHelp')}</p>
                    <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.skuAutoHelp')}</p>

                    <div className="space-y-2">
                      <Label>{t('items.single.variantSetup.attributes')}</Label>
                      {attributeOptions.length === 0 && (
                        <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.noAttributes')}</p>
                      )}

                      {attributeOptions.map((attribute) => (
                        <div key={attribute.id} className="flex items-center gap-2">
                          <Checkbox
                            id={`attribute-${attribute.id}`}
                            checked={selectedAttributeIDs.includes(attribute.id)}
                            onCheckedChange={(checked) => toggleAttribute(attribute.id, checked === true)}
                          />
                          <Label htmlFor={`attribute-${attribute.id}`}>{attribute.display_name || attribute.name}</Label>
                        </div>
                      ))}
                    </div>

                    {selectedAttributeIDs.map((attributeID) => {
                      const attribute = attributeOptions.find((entry) => entry.id === attributeID);
                      if (!attribute) {
                        return null;
                      }

                      return (
                        <div key={`values-${attributeID}`} className="space-y-2">
                          <Label>{attribute.display_name || attribute.name}</Label>

                          {(attribute.values || []).length === 0 ? (
                            <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.noValues')}</p>
                          ) : (
                            <div className="grid gap-2 sm:grid-cols-2">
                              {(attribute.values || []).map((value) => (
                                <div key={value.id} className="flex items-center gap-2">
                                  <Checkbox
                                    id={`attribute-${attributeID}-value-${value.id}`}
                                    checked={(selectedValuesByAttribute[attributeID] || []).includes(value.id)}
                                    onCheckedChange={(checked) => toggleAttributeValue(attributeID, value.id, checked === true)}
                                  />
                                  <Label htmlFor={`attribute-${attributeID}-value-${value.id}`}>{value.display_name || value.value}</Label>
                                </div>
                              ))}
                            </div>
                          )}
                        </div>
                      );
                    })}

                    {variantComboPreviews.length > 0 && (
                      <div className="space-y-3">
                        <div className="flex items-center justify-between">
                          <Label>{t('items.single.variantSetup.variants')}</Label>
                          <div className="flex gap-2">
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              onClick={() => {
                                const price = prompt(t('items.single.variantSetup.bulkPricePrompt'));
                                if (price) {
                                  const numPrice = parseFloat(price);
                                  if (!isNaN(numPrice) && numPrice >= 0) {
                                    handleBulkApplyPrice(numPrice);
                                  }
                                }
                              }}
                            >
                              Apply Price to All
                            </Button>
                            <Button type="button" variant="outline" size="sm" onClick={handleBulkGenerateSKUs}>
                              Generate SKUs
                            </Button>
                            <Button type="button" variant="outline" size="sm" onClick={handleBulkGenerateBarcodes}>
                              Generate Barcodes
                            </Button>
                          </div>
                        </div>
                        {params.action === 'edit' && (
                          <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.existingLockedHelp')}</p>
                        )}

                        <div className="overflow-x-auto rounded-md border">
                          <table className="w-full text-sm">
                            <thead className="bg-muted/50 border-b">
                              <tr>
                                <th className="p-2 text-left font-medium">{t('global.variant')}</th>
                                <th className="w-32 p-2 text-left font-medium">{t('global.sku')}</th>
                                <th className="w-32 p-2 text-left font-medium">{t('global.barcode')}</th>
                                <th className="w-28 p-2 text-left font-medium">{t('global.reference')}</th>
                                <th className="w-28 p-2 text-left font-medium">Vendor Ref</th>
                                <th className="w-24 p-2 text-right font-medium">Cost Price</th>
                                <th className="w-24 p-2 text-right font-medium">{t('global.price')}</th>
                                <th className="w-16 p-2 text-center font-medium">Active</th>
                              </tr>
                            </thead>
                            <tbody>
                              {variantComboPreviews.map((combo) => {
                                const isExistingVariant = params.action === 'edit' && existingVariantSignatures.has(combo.key);
                                const isActive = variantActiveOverrides[combo.key] !== undefined ? variantActiveOverrides[combo.key] : true;

                                return (
                                  <tr key={combo.key} className={`border-b last:border-0 ${!isActive ? 'opacity-50' : ''}`}>
                                    <td className="p-2">
                                      <div className="text-sm">
                                        {combo.label || '-'}
                                        {isExistingVariant && (
                                          <span className="text-muted-foreground ml-2 text-xs">
                                            ({t('items.single.variantSetup.existingLockedTag')})
                                          </span>
                                        )}
                                      </div>
                                    </td>
                                    <td className="p-2">
                                      <Input
                                        type="text"
                                        className="h-8 text-xs"
                                        value={variantSKUOverrides[combo.key] ?? generateSuggestedVariantSKU(combo)}
                                        disabled={isExistingVariant}
                                        onChange={(e) => {
                                          setVariantSKUOverrides((current) => ({
                                            ...current,
                                            [combo.key]: e.target.value,
                                          }));
                                        }}
                                      />
                                    </td>
                                    <td className="p-2">
                                      <Input
                                        type="text"
                                        className="h-8 text-xs"
                                        value={variantBarcodeOverrides[combo.key] ?? ''}
                                        disabled={isExistingVariant}
                                        placeholder="Barcode"
                                        onChange={(e) => {
                                          setVariantBarcodeOverrides((current) => ({
                                            ...current,
                                            [combo.key]: e.target.value,
                                          }));
                                        }}
                                      />
                                    </td>
                                    <td className="p-2">
                                      <Input
                                        type="text"
                                        className="h-8 text-xs"
                                        value={variantReferenceOverrides[combo.key] ?? ''}
                                        disabled={isExistingVariant}
                                        placeholder="Ref"
                                        onChange={(e) => {
                                          setVariantReferenceOverrides((current) => ({
                                            ...current,
                                            [combo.key]: e.target.value,
                                          }));
                                        }}
                                      />
                                    </td>
                                    <td className="p-2">
                                      <Input
                                        type="text"
                                        className="h-8 text-xs"
                                        value={variantVendorRefOverrides[combo.key] ?? ''}
                                        disabled={isExistingVariant}
                                        placeholder="Vendor"
                                        onChange={(e) => {
                                          setVariantVendorRefOverrides((current) => ({
                                            ...current,
                                            [combo.key]: e.target.value,
                                          }));
                                        }}
                                      />
                                    </td>
                                    <td className="p-2">
                                      <Input
                                        type="number"
                                        min={0}
                                        step="0.01"
                                        className="h-8 text-right text-xs"
                                        value={variantCostPriceOverrides[combo.key]?.toString() ?? ''}
                                        disabled={isExistingVariant}
                                        placeholder="0.00"
                                        onChange={(e) => {
                                          const value = e.target.valueAsNumber;
                                          setVariantCostPriceOverrides((current) => ({
                                            ...current,
                                            [combo.key]: Number.isNaN(value) ? undefined : value,
                                          }));
                                        }}
                                      />
                                    </td>
                                    <td className="p-2">
                                      <Input
                                        type="number"
                                        min={0}
                                        step="0.01"
                                        className="h-8 text-right text-xs"
                                        value={(variantPriceOverrides[combo.key] ?? data.price).toString()}
                                        disabled={isExistingVariant}
                                        onChange={(e) => {
                                          const value = e.target.valueAsNumber;
                                          setVariantPriceOverrides((current) => ({
                                            ...current,
                                            [combo.key]: Number.isNaN(value) ? undefined : value,
                                          }));
                                        }}
                                      />
                                    </td>
                                    <td className="p-2 text-center">
                                      <Checkbox
                                        checked={isActive}
                                        disabled={isExistingVariant}
                                        onCheckedChange={(checked) => {
                                          setVariantActiveOverrides((current) => ({
                                            ...current,
                                            [combo.key]: checked === true,
                                          }));
                                        }}
                                      />
                                    </td>
                                  </tr>
                                );
                              })}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    )}

                    {params.action === 'edit' && (
                      <div className="space-y-2">
                        <Label>{t('items.single.variantSetup.variants')}</Label>
                        {(existingVariantSetup?.variants || []).length === 0 ? (
                          <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.noVariants')}</p>
                        ) : (
                          <div className="space-y-1">
                            {existingVariantSetup!.variants.map((variant) => (
                              <p key={variant.uuid} className="text-sm">
                                <span className="font-medium">{variant.name}</span> ({variant.sku})
                              </p>
                            ))}
                          </div>
                        )}
                      </div>
                    )}

                    {!!variantSetupError && <InputError className="mt-2" message={variantSetupError} />}
                  </div>
                )}

                {!data.has_variants && params.action === 'edit' && (
                  <div className="rounded-md bg-blue-50 p-3">
                    <p className="text-sm text-blue-800">{t('items.single.variantSetup.noVariantsConfigured')}</p>
                  </div>
                )}
              </div>
            )}

            {params.action === 'view' && data.item_type === 'product' && (
              <div className="border-border space-y-4 rounded-lg border p-4">
                <div className="space-y-1">
                  <Label>{t('items.single.variantSetup.currentTitle')}</Label>
                  <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.currentDescription')}</p>
                </div>

                {existingVariantSetup && existingHasVariants ? (
                  <>
                    <div className="space-y-2">
                      <Label>{t('items.single.variantSetup.attributes')}</Label>
                      <p className="text-sm">{selectedAttributeLabels.join(', ') || '-'}</p>
                    </div>

                    <div className="space-y-2">
                      <Label>{t('items.single.variantSetup.values')}</Label>
                      {selectedAttributeValueLabels.length === 0 ? (
                        <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.noValues')}</p>
                      ) : (
                        <div className="space-y-1">
                          {selectedAttributeValueLabels.map((entry) => (
                            <p key={entry.attribute} className="text-sm">
                              <span className="font-medium">{entry.attribute}:</span> {entry.labels.join(', ') || '-'}
                            </p>
                          ))}
                        </div>
                      )}
                    </div>

                    <div className="space-y-2">
                      <Label>{t('items.single.variantSetup.variants')}</Label>
                      {(existingVariantSetup.variants || []).length === 0 ? (
                        <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.noVariants')}</p>
                      ) : (
                        <div className="space-y-1">
                          {existingVariantSetup.variants.map((variant) => (
                            <p key={variant.uuid} className="text-sm">
                              <span className="font-medium">{variant.name}</span> ({variant.sku})
                            </p>
                          ))}
                        </div>
                      )}
                    </div>
                  </>
                ) : (
                  <div className="rounded-md bg-blue-50 p-3">
                    <p className="text-sm text-blue-800">{t('items.single.variantSetup.noVariantsConfigured')}</p>
                  </div>
                )}
              </div>
            )}
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
                path={`/items/${params.item?.uuid}/change-status`}
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
