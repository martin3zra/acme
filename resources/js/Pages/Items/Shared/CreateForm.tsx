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
import { useCallback, useEffect, useMemo, useRef, useState } from 'react';

export type CreateFormParams = {
  item: Item | undefined;
  taxes: Tax[];
  units: Unit[];
  attributes: ItemAttributeOption[];
  warehouses: ItemWarehouseOption[];
  action: Verb;
};

type ItemWarehouseOption = {
  id: number;
  code: string;
  name: string;
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
  cost_price?: number;
  tax_id: number;
  unit_id: number;
  item_type: ItemType; // This can be 'product' or 'service'
  reference?: string;
  code?: string;
  sku?: string;
  barcode?: string;
  vendor_reference?: string;
  track_inventory: boolean;
  has_variants: boolean;
  attribute_ids: number[];
  variant_combos: VariantComboForm[];
};

type VariantComboForm = {
  variant_id?: number;
  attribute_value_ids: Record<number, number>;
  sku?: string;
  price?: number;
  cost_price?: number;
  track_inventory?: boolean;
  stock_by_warehouse?: Record<number, number>;
  barcode?: string;
  reference?: string;
  vendor_reference?: string;
  active?: boolean;
};

type VariantComboPreview = {
  key: string;
  variant_id?: number;
  attribute_value_ids: Record<number, number>;
  label: string;
  sku?: string;
  price?: number;
  cost_price?: number;
  track_inventory?: boolean;
  stock_by_warehouse?: Record<number, number>;
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

const normalizeStockByWarehouse = (stockByWarehouse?: Record<number, number> | Record<string, number>): Record<number, number> => {
  if (!stockByWarehouse) {
    return {};
  }

  return Object.entries(stockByWarehouse).reduce<Record<number, number>>((current, [warehouseID, quantity]) => {
    const normalizedWarehouseID = Number(warehouseID);
    if (Number.isNaN(normalizedWarehouseID)) {
      return current;
    }

    const normalizedQuantity = Number(quantity);
    current[normalizedWarehouseID] = Number.isNaN(normalizedQuantity) ? 0 : Math.max(0, normalizedQuantity);
    return current;
  }, {});
};

const areNumberMapsEqual = (left: Record<number, number>, right: Record<number, number>): boolean => {
  const leftEntries = Object.entries(left);
  const rightEntries = Object.entries(right);

  if (leftEntries.length !== rightEntries.length) {
    return false;
  }

  for (const [key, value] of leftEntries) {
    if (right[Number(key)] !== value) {
      return false;
    }
  }

  return true;
};

const areRecordEntriesEqual = <T,>(left: Record<string, T>, right: Record<string, T>, equal: (a: T, b: T) => boolean): boolean => {
  const leftKeys = Object.keys(left);
  const rightKeys = Object.keys(right);

  if (leftKeys.length !== rightKeys.length) {
    return false;
  }

  for (const key of leftKeys) {
    if (!(key in right)) {
      return false;
    }
    if (!equal(left[key], right[key])) {
      return false;
    }
  }

  return true;
};

const VARIANT_TABLE_PAGE_SIZE = 50;

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const existingVariantSetup = params.item?.variant_setup;
  const existingHasVariants =
    !!existingVariantSetup?.has_variants ||
    (existingVariantSetup?.attribute_ids?.length || 0) > 0 ||
    Object.keys(existingVariantSetup?.selected_values_by_attribute || {}).length > 0 ||
    (existingVariantSetup?.variants?.length || 0) > 1;
  const defaultExistingVariant = existingVariantSetup?.variants?.find((variant) => variant.is_default);
  const { data, setData, post, put, transform, errors, reset, processing } = useForm<Required<ItemForm>>({
    id: params.item?.id,
    name: params.item?.name || '',
    description: params.item?.description || '',
    price: defaultExistingVariant?.price ?? params.item?.price ?? 0,
    cost_price: defaultExistingVariant?.cost_price,
    tax_id: params.item?.tax?.id || 0,
    unit_id: params.item?.unit?.id || 0,
    item_type: params.item?.item_type || 'product', // Default to 'product'
    reference: defaultExistingVariant?.reference || params.item?.identifiers?.reference || '',
    code: params.item?.identifiers?.code || '',
    sku: defaultExistingVariant?.sku || params.item?.identifiers?.sku || '',
    barcode: defaultExistingVariant?.barcode || params.item?.identifiers?.barcode || '',
    vendor_reference: defaultExistingVariant?.vendor_reference || params.item?.identifiers?.vendor_reference || '',
    track_inventory: defaultExistingVariant?.track_inventory ?? params.item?.item_type === 'product',
    has_variants: params.item?.has_variants || existingHasVariants || false,
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
  const [variantTrackInventoryOverrides, setVariantTrackInventoryOverrides] = useState<Record<string, boolean>>({});
  const [variantStockByWarehouseOverrides, setVariantStockByWarehouseOverrides] = useState<Record<string, Record<number, number>>>({});
  const [singleVariantStockByWarehouse, setSingleVariantStockByWarehouse] = useState<Record<number, number>>(
    normalizeStockByWarehouse(defaultExistingVariant?.stock_by_warehouse),
  );
  const [variantSetupError, setVariantSetupError] = useState<string>('');
  const [bulkPriceInput, setBulkPriceInput] = useState<string>('');
  const [variantSKUFilter, setVariantSKUFilter] = useState<string>('');
  const [variantBarcodeFilter, setVariantBarcodeFilter] = useState<string>('');
  const [variantAttributeValueFilters, setVariantAttributeValueFilters] = useState<Record<number, number[]>>({});
  const [variantTablePage, setVariantTablePage] = useState<number>(1);
  const variantTableContainerRef = useRef<HTMLDivElement | null>(null);
  const variantTableLastScrollTopRef = useRef<number>(0);
  const variantTableScrollAnchorRef = useRef<'top' | 'bottom' | null>(null);
  const suppressVariantTableScrollHandlerRef = useRef<boolean>(false);

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
  const attributeOptions = useMemo(() => (Array.isArray(params.attributes) ? params.attributes : []), [params.attributes]);
  const warehouseOptions = useMemo(() => (Array.isArray(params.warehouses) ? params.warehouses : []), [params.warehouses]);
  const existingVariantsBySignature = useMemo(
    () =>
      new Map(
        (existingVariantSetup?.variants || [])
          .filter((variant) => typeof variant.combination_signature === 'string' && variant.combination_signature.length > 0)
          .map((variant) => [variant.combination_signature!, variant] as const),
      ),
    [existingVariantSetup],
  );
  const variantSetupHydrationKey = useMemo(
    () =>
      JSON.stringify({
        item_id: params.item?.id || 0,
        item_type: params.item?.item_type || '',
        has_variants: existingVariantSetup?.has_variants || false,
        attribute_ids: existingVariantSetup?.attribute_ids || [],
        selected_values_by_attribute: existingVariantSetup?.selected_values_by_attribute || {},
        variants: (existingVariantSetup?.variants || []).map((variant) => ({
          id: variant.id,
          signature: variant.combination_signature || '',
          sku: variant.sku,
          barcode: variant.barcode || '',
          reference: variant.reference || '',
          vendor_reference: variant.vendor_reference || '',
          price: variant.price,
          cost_price: variant.cost_price,
          active: variant.active !== false,
          track_inventory: variant.track_inventory !== false,
          stock_by_warehouse: variant.stock_by_warehouse || {},
        })),
      }),
    [params.item?.id, params.item?.item_type, existingVariantSetup],
  );

  useEffect(() => {
    if (params.action !== 'edit' || params.item?.item_type !== 'product') {
      return;
    }

    const nextAttributeIDs = Array.isArray(existingVariantSetup?.attribute_ids)
      ? Array.from(
          new Set(existingVariantSetup!.attribute_ids.map((attributeID) => Number(attributeID)).filter((attributeID) => !Number.isNaN(attributeID))),
        )
      : [];

    const nextSelectedValuesByAttribute = Object.entries(existingVariantSetup?.selected_values_by_attribute || {}).reduce<Record<number, number[]>>(
      (current, [attributeID, valueIDs]) => {
        const normalizedValues = Array.isArray(valueIDs)
          ? valueIDs.map((valueID) => Number(valueID)).filter((valueID) => !Number.isNaN(valueID))
          : [];
        current[Number(attributeID)] = Array.from(new Set(normalizedValues));
        return current;
      },
      {},
    );

    setSelectedAttributeIDs(nextAttributeIDs);
    setSelectedValuesByAttribute(nextSelectedValuesByAttribute);

    const nextSKUOverrides: Record<string, string> = {};
    const nextBarcodeOverrides: Record<string, string> = {};
    const nextReferenceOverrides: Record<string, string> = {};
    const nextVendorRefOverrides: Record<string, string> = {};
    const nextPriceOverrides: Record<string, number> = {};
    const nextCostPriceOverrides: Record<string, number> = {};
    const nextActiveOverrides: Record<string, boolean> = {};
    const nextTrackInventoryOverrides: Record<string, boolean> = {};
    const nextStockByWarehouseOverrides: Record<string, Record<number, number>> = {};

    for (const variant of existingVariantSetup?.variants || []) {
      const signature = variant.combination_signature || '';
      if (!signature) {
        continue;
      }

      nextSKUOverrides[signature] = variant.sku;

      if (variant.barcode) {
        nextBarcodeOverrides[signature] = variant.barcode;
      }
      if (variant.reference) {
        nextReferenceOverrides[signature] = variant.reference;
      }
      if (variant.vendor_reference) {
        nextVendorRefOverrides[signature] = variant.vendor_reference;
      }
      if (typeof variant.price === 'number') {
        nextPriceOverrides[signature] = variant.price;
      }
      if (typeof variant.cost_price === 'number') {
        nextCostPriceOverrides[signature] = variant.cost_price;
      }
      nextActiveOverrides[signature] = variant.active !== false;
      nextTrackInventoryOverrides[signature] = variant.track_inventory !== false;
      nextStockByWarehouseOverrides[signature] = normalizeStockByWarehouse(variant.stock_by_warehouse);
    }

    setVariantPriceOverrides(nextPriceOverrides);
    setVariantSKUOverrides(nextSKUOverrides);
    setVariantBarcodeOverrides(nextBarcodeOverrides);
    setVariantReferenceOverrides(nextReferenceOverrides);
    setVariantVendorRefOverrides(nextVendorRefOverrides);
    setVariantCostPriceOverrides(nextCostPriceOverrides);
    setVariantActiveOverrides(nextActiveOverrides);
    setVariantTrackInventoryOverrides(nextTrackInventoryOverrides);
    setVariantStockByWarehouseOverrides(nextStockByWarehouseOverrides);
    setData('has_variants', existingHasVariants);
    setSingleVariantStockByWarehouse(normalizeStockByWarehouse(defaultExistingVariant?.stock_by_warehouse));
  }, [
    params.action,
    params.item?.item_type,
    defaultExistingVariant?.stock_by_warehouse,
    existingVariantSetup,
    existingHasVariants,
    setData,
    variantSetupHydrationKey,
  ]);

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

  const buildVariantKey = useCallback(
    (selection: Record<number, number>): string => {
      const sortedAttributeIDs = [...selectedAttributeIDs].sort((left, right) => left - right);
      return sortedAttributeIDs.map((attributeID) => `${attributeID}:${selection[attributeID]}`).join('|');
    },
    [selectedAttributeIDs],
  );

  const buildVariantLabel = useCallback(
    (selection: Record<number, number>): string => {
      return selectedAttributeIDs
        .map((attributeID) => {
          const attribute = attributeOptions.find((entry) => entry.id === attributeID);
          const valueID = selection[attributeID];
          const value = attribute?.values?.find((entry) => entry.id === valueID);
          return value?.display_name || value?.value || `${attribute?.display_name || attribute?.name || attributeID}:${valueID}`;
        })
        .join(' / ');
    },
    [attributeOptions, selectedAttributeIDs],
  );

  const variantComboPreviews = useMemo((): VariantComboPreview[] => {
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
        const existingVariant = existingVariantsBySignature.get(key);

        combos.push({
          key,
          variant_id: existingVariant?.id,
          attribute_value_ids: current,
          label: buildVariantLabel(current),
          sku: variantSKUOverrides[key] ?? existingVariant?.sku,
          price: variantPriceOverrides[key] ?? existingVariant?.price,
          cost_price: variantCostPriceOverrides[key] ?? existingVariant?.cost_price,
          track_inventory:
            variantTrackInventoryOverrides[key] !== undefined
              ? variantTrackInventoryOverrides[key]
              : (existingVariant?.track_inventory ?? data.track_inventory),
          stock_by_warehouse: variantStockByWarehouseOverrides[key] ?? normalizeStockByWarehouse(existingVariant?.stock_by_warehouse),
          barcode: variantBarcodeOverrides[key] ?? existingVariant?.barcode,
          reference: variantReferenceOverrides[key] ?? existingVariant?.reference,
          vendor_reference: variantVendorRefOverrides[key] ?? existingVariant?.vendor_reference,
          active: variantActiveOverrides[key] !== undefined ? variantActiveOverrides[key] : existingVariant?.active !== false,
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
  }, [
    buildVariantKey,
    buildVariantLabel,
    data.track_inventory,
    existingVariantsBySignature,
    selectedAttributeIDs,
    selectedValuesByAttribute,
    variantActiveOverrides,
    variantBarcodeOverrides,
    variantCostPriceOverrides,
    variantPriceOverrides,
    variantReferenceOverrides,
    variantSKUOverrides,
    variantStockByWarehouseOverrides,
    variantTrackInventoryOverrides,
    variantVendorRefOverrides,
  ]);

  const toAlphaNumericUpper = (value: string): string => {
    return value
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .toUpperCase()
      .replace(/[^A-Z0-9]/g, '');
  };

  const currentVariantPrice = (combo: VariantComboPreview): number | undefined => {
    return variantPriceOverrides[combo.key] ?? combo.price;
  };

  const currentVariantSKU = (combo: VariantComboPreview): string => {
    return (variantSKUOverrides[combo.key] ?? combo.sku ?? '').trim();
  };

  const currentVariantBarcode = (combo: VariantComboPreview): string => {
    return (variantBarcodeOverrides[combo.key] ?? combo.barcode ?? '').trim();
  };

  const currentVariantTrackInventory = (combo: VariantComboPreview): boolean => {
    if (variantTrackInventoryOverrides[combo.key] !== undefined) {
      return variantTrackInventoryOverrides[combo.key];
    }

    return combo.track_inventory ?? data.track_inventory;
  };

  const currentVariantStockByWarehouse = (combo: VariantComboPreview): Record<number, number> => {
    return variantStockByWarehouseOverrides[combo.key] ?? normalizeStockByWarehouse(combo.stock_by_warehouse);
  };

  const isPriceEmpty = (price: number | undefined): boolean => {
    return price === undefined || price === null || Number.isNaN(price);
  };

  const buildSkuPrefix = (): string => {
    const fromCode = toAlphaNumericUpper(data.code || '');
    if (fromCode) {
      return fromCode.slice(0, 16);
    }

    const fromName = toAlphaNumericUpper(data.name || '');
    return (fromName || 'ITEM').slice(0, 16);
  };

  const buildAttributeToken = (attributeName: string, rawValue: string): string => {
    const normalizedName = attributeName
      .normalize('NFD')
      .replace(/[\u0300-\u036f]/g, '')
      .toLowerCase();

    const valueToken = toAlphaNumericUpper(rawValue);
    if (!valueToken) {
      return 'NA';
    }

    const isSizeAttribute = normalizedName.includes('size') || normalizedName.includes('talla') || normalizedName.includes('tamano');
    return isSizeAttribute ? valueToken.slice(0, 1) : valueToken.slice(0, 3);
  };

  const buildSmartVariantSKU = (combo: VariantComboPreview): string => {
    const itemPrefix = buildSkuPrefix();

    const attributeParts = selectedAttributeIDs
      .map((attributeID) => {
        const attribute = attributeOptions.find((attr) => attr.id === attributeID);
        const valueID = combo.attribute_value_ids[attributeID];
        const value = attribute?.values?.find((v) => v.id === valueID);
        const attributeName = attribute?.display_name || attribute?.name || '';
        const displayValue = value?.display_name || value?.value || '';
        return buildAttributeToken(attributeName, displayValue);
      })
      .filter((part) => part);

    const parts = [itemPrefix, ...attributeParts];
    return parts.join('-').slice(0, 50);
  };

  const calculateEAN13CheckDigit = (digits12: string): string => {
    const sum = digits12
      .split('')
      .map(Number)
      .reduce((acc, digit, idx) => acc + digit * (idx % 2 === 0 ? 1 : 3), 0);

    return String((10 - (sum % 10)) % 10);
  };

  const generateEAN13Barcode = (combo: VariantComboPreview, attempt: number = 0): string => {
    const seed = `${buildSkuPrefix()}-${combo.key}-${attempt}`;
    const hash = Array.from(seed).reduce((acc, char) => ((acc * 33 + char.charCodeAt(0)) >>> 0) % 10000000000, 7);
    const digits12 = `20${String(hash).padStart(10, '0')}`;
    const checkDigit = calculateEAN13CheckDigit(digits12);
    return `${digits12}${checkDigit}`;
  };

  const withNumericSuffix = (base: string, suffix: number): string => {
    const suffixPart = `-${suffix}`;
    const maxBaseLength = Math.max(1, 50 - suffixPart.length);
    return `${base.slice(0, maxBaseLength)}${suffixPart}`;
  };

  const handleBulkApplyPrice = (price: number) => {
    const updates: Record<string, number> = {};
    variantComboPreviews.forEach((combo) => {
      if (isPriceEmpty(currentVariantPrice(combo))) {
        updates[combo.key] = price;
      }
    });
    setVariantPriceOverrides((current) => ({ ...current, ...updates }));
  };

  const applyBulkPriceFromInput = () => {
    const value = Number(bulkPriceInput);
    if (bulkPriceInput.trim() === '' || Number.isNaN(value) || value < 0) {
      setVariantSetupError('Enter a valid non-negative price before applying bulk price.');
      return;
    }

    setVariantSetupError('');
    handleBulkApplyPrice(value);
  };

  const handleBulkGenerateSKUs = () => {
    const updates: Record<string, string> = {};
    const usedSKUs = new Set(variantComboPreviews.map((combo) => currentVariantSKU(combo).toUpperCase()).filter((sku) => sku.length > 0));

    variantComboPreviews.forEach((combo) => {
      if (currentVariantSKU(combo).length > 0) {
        return;
      }

      const baseSKU = buildSmartVariantSKU(combo) || 'ITEM';
      let candidate = baseSKU;
      let suffix = 2;

      for (; usedSKUs.has(candidate.toUpperCase()); suffix++) {
        candidate = withNumericSuffix(baseSKU, suffix);
      }

      updates[combo.key] = candidate;
      usedSKUs.add(candidate.toUpperCase());
    });

    setVariantSKUOverrides((current) => ({ ...current, ...updates }));
  };

  const handleBulkGenerateBarcodes = () => {
    const updates: Record<string, string> = {};
    const usedBarcodes = new Set(variantComboPreviews.map((combo) => currentVariantBarcode(combo)).filter((barcode) => barcode.length > 0));

    variantComboPreviews.forEach((combo) => {
      if (currentVariantBarcode(combo).length > 0) {
        return;
      }

      let attempt = 0;
      let candidate = generateEAN13Barcode(combo, attempt);
      for (; usedBarcodes.has(candidate); attempt++) {
        candidate = generateEAN13Barcode(combo, attempt + 1);
      }

      updates[combo.key] = candidate;
      usedBarcodes.add(candidate);
    });

    setVariantBarcodeOverrides((current) => ({ ...current, ...updates }));
  };

  const selectedAttributeIDsKey = selectedAttributeIDs.join(',');
  const selectedValuesByAttributeKey = JSON.stringify(selectedValuesByAttribute);

  useEffect(() => {
    setVariantAttributeValueFilters((current) => {
      const next: Record<number, number[]> = {};

      for (const attributeID of selectedAttributeIDs) {
        const activeFilters = current[attributeID] || [];
        if (activeFilters.length === 0) {
          continue;
        }

        const allowedValues = new Set(selectedValuesByAttribute[attributeID] || []);
        const validFilters = activeFilters.filter((valueID) => allowedValues.has(valueID));
        if (validFilters.length > 0) {
          next[attributeID] = Array.from(new Set(validFilters));
        }
      }

      const normalizeRecord = (record: Record<number, number[]>) =>
        Object.entries(record)
          .sort(([left], [right]) => Number(left) - Number(right))
          .map(([attributeID, valueIDs]) => [attributeID, [...new Set(valueIDs)].sort((left, right) => left - right)]);

      return JSON.stringify(normalizeRecord(current)) === JSON.stringify(normalizeRecord(next)) ? current : next;
    });
  }, [selectedAttributeIDsKey, selectedValuesByAttributeKey]);

  useEffect(() => {
    if (!data.has_variants || variantComboPreviews.length === 0) {
      setVariantPriceOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantSKUOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantBarcodeOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantReferenceOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantVendorRefOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantCostPriceOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantActiveOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantTrackInventoryOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
      setVariantStockByWarehouseOverrides((current) => (Object.keys(current).length === 0 ? current : {}));
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

    const syncOverridesIfChanged = <T,>(
      current: Record<string, T>,
      defaultValue: ((combo: VariantComboPreview) => T) | undefined,
      equal: (a: T, b: T) => boolean,
    ): Record<string, T> => {
      const next = syncOverrides(current, defaultValue);
      return areRecordEntriesEqual(current, next, equal) ? current : next;
    };

    setVariantPriceOverrides((current) => syncOverridesIfChanged(current, undefined, Object.is));
    setVariantSKUOverrides((current) => syncOverridesIfChanged(current, undefined, Object.is));
    setVariantBarcodeOverrides((current) => syncOverridesIfChanged(current, undefined, Object.is));
    setVariantReferenceOverrides((current) => syncOverridesIfChanged(current, undefined, Object.is));
    setVariantVendorRefOverrides((current) => syncOverridesIfChanged(current, undefined, Object.is));
    setVariantCostPriceOverrides((current) => syncOverridesIfChanged(current, undefined, Object.is));
    setVariantActiveOverrides((current) => syncOverridesIfChanged(current, undefined, Object.is));
    setVariantTrackInventoryOverrides((current) =>
      syncOverridesIfChanged(current, (combo) => combo.track_inventory ?? data.track_inventory, Object.is),
    );
    setVariantStockByWarehouseOverrides((current) =>
      syncOverridesIfChanged(current, (combo) => normalizeStockByWarehouse(combo.stock_by_warehouse), areNumberMapsEqual),
    );
  }, [data.has_variants, data.name, data.track_inventory, selectedAttributeIDsKey, selectedValuesByAttributeKey, variantComboPreviews]);

  const variantFilterAttributes = useMemo(() => {
    return selectedAttributeIDs
      .map((attributeID) => {
        const attribute = attributeOptions.find((entry) => entry.id === attributeID);
        if (!attribute) {
          return null;
        }

        const selectedValueIDs = new Set(selectedValuesByAttribute[attributeID] || []);
        const values = (attribute.values || []).filter((value) => selectedValueIDs.has(value.id));
        if (values.length === 0) {
          return null;
        }

        return {
          id: attributeID,
          name: attribute.display_name || attribute.name,
          values,
        };
      })
      .filter((entry): entry is { id: number; name: string; values: ItemAttributeValueOption[] } => entry !== null);
  }, [attributeOptions, selectedAttributeIDs, selectedValuesByAttribute]);

  const filteredVariantComboPreviews = useMemo(() => {
    const skuQuery = variantSKUFilter.trim().toLowerCase();
    const barcodeQuery = variantBarcodeFilter.trim().toLowerCase();

    return variantComboPreviews.filter((combo) => {
      const sku = (variantSKUOverrides[combo.key] ?? combo.sku ?? '').toLowerCase();
      const barcode = (variantBarcodeOverrides[combo.key] ?? combo.barcode ?? '').toLowerCase();

      if (skuQuery.length > 0 && !sku.includes(skuQuery)) {
        return false;
      }

      if (barcodeQuery.length > 0 && !barcode.includes(barcodeQuery)) {
        return false;
      }

      for (const [attributeID, valueIDs] of Object.entries(variantAttributeValueFilters)) {
        if (valueIDs.length === 0) {
          continue;
        }

        const selectedValueID = combo.attribute_value_ids[Number(attributeID)];
        if (!valueIDs.includes(selectedValueID)) {
          return false;
        }
      }

      return true;
    });
  }, [variantBarcodeFilter, variantBarcodeOverrides, variantSKUFilter, variantSKUOverrides, variantAttributeValueFilters, variantComboPreviews]);

  const variantAttributeFiltersKey = useMemo(() => JSON.stringify(variantAttributeValueFilters), [variantAttributeValueFilters]);

  useEffect(() => {
    setVariantTablePage(1);
  }, [variantSKUFilter, variantBarcodeFilter, variantAttributeFiltersKey]);

  const totalVariantTablePages = Math.max(1, Math.ceil(filteredVariantComboPreviews.length / VARIANT_TABLE_PAGE_SIZE));
  const currentVariantTablePage = Math.min(variantTablePage, totalVariantTablePages);

  useEffect(() => {
    if (variantTablePage > totalVariantTablePages) {
      setVariantTablePage(totalVariantTablePages);
    }
  }, [variantTablePage, totalVariantTablePages]);

  const visibleVariantComboPreviews = useMemo(() => {
    const startIndex = (currentVariantTablePage - 1) * VARIANT_TABLE_PAGE_SIZE;
    const endIndex = startIndex + VARIANT_TABLE_PAGE_SIZE;
    return filteredVariantComboPreviews.slice(startIndex, endIndex);
  }, [currentVariantTablePage, filteredVariantComboPreviews]);

  const visibleVariantStart = filteredVariantComboPreviews.length === 0 ? 0 : (currentVariantTablePage - 1) * VARIANT_TABLE_PAGE_SIZE + 1;
  const visibleVariantEnd = Math.min(currentVariantTablePage * VARIANT_TABLE_PAGE_SIZE, filteredVariantComboPreviews.length);

  useEffect(() => {
    const container = variantTableContainerRef.current;
    if (!container) {
      return;
    }

    if (variantTableScrollAnchorRef.current === null) {
      return;
    }

    suppressVariantTableScrollHandlerRef.current = true;

    if (variantTableScrollAnchorRef.current === 'top') {
      container.scrollTop = 4;
      variantTableLastScrollTopRef.current = 4;
    } else {
      const nextScrollTop = Math.max(0, container.scrollHeight - container.clientHeight - 4);
      container.scrollTop = nextScrollTop;
      variantTableLastScrollTopRef.current = nextScrollTop;
    }

    variantTableScrollAnchorRef.current = null;

    window.setTimeout(() => {
      suppressVariantTableScrollHandlerRef.current = false;
    }, 0);
  }, [currentVariantTablePage]);

  const handleVariantTableScroll = (event: React.UIEvent<HTMLDivElement>) => {
    if (suppressVariantTableScrollHandlerRef.current) {
      return;
    }

    const container = event.currentTarget;
    const currentScrollTop = container.scrollTop;
    const scrollingDown = currentScrollTop > variantTableLastScrollTopRef.current;
    variantTableLastScrollTopRef.current = currentScrollTop;

    const edgeThreshold = 28;
    const reachedBottom = currentScrollTop + container.clientHeight >= container.scrollHeight - edgeThreshold;
    const reachedTop = currentScrollTop <= edgeThreshold;

    if (reachedBottom && scrollingDown && currentVariantTablePage < totalVariantTablePages) {
      variantTableScrollAnchorRef.current = 'top';
      setVariantTablePage((current) => Math.min(totalVariantTablePages, current + 1));
      return;
    }

    if (reachedTop && !scrollingDown && currentVariantTablePage > 1) {
      variantTableScrollAnchorRef.current = 'bottom';
      setVariantTablePage((current) => Math.max(1, current - 1));
    }
  };

  const hasActiveVariantFilters =
    variantSKUFilter.trim().length > 0 ||
    variantBarcodeFilter.trim().length > 0 ||
    Object.values(variantAttributeValueFilters).some((valueIDs) => valueIDs.length > 0);

  const toggleVariantFilterValue = (attributeID: number, valueID: number, checked: boolean) => {
    setVariantAttributeValueFilters((current) => {
      const currentValues = current[attributeID] || [];
      const nextValues = checked ? Array.from(new Set([...currentValues, valueID])) : currentValues.filter((entry) => entry !== valueID);

      if (nextValues.length === 0) {
        const next = { ...current };
        delete next[attributeID];
        return next;
      }

      return {
        ...current,
        [attributeID]: nextValues,
      };
    });
  };

  const clearVariantFilters = () => {
    setVariantSKUFilter('');
    setVariantBarcodeFilter('');
    setVariantAttributeValueFilters({});
  };

  const buildVariantCombos = (): VariantComboForm[] => {
    return variantComboPreviews.map((combo) => ({
      variant_id: combo.variant_id,
      attribute_value_ids: combo.attribute_value_ids,
      price: variantPriceOverrides[combo.key] ?? combo.price ?? data.price,
      cost_price: variantCostPriceOverrides[combo.key] ?? combo.cost_price,
      track_inventory: currentVariantTrackInventory(combo),
      stock_by_warehouse: currentVariantTrackInventory(combo) ? currentVariantStockByWarehouse(combo) : undefined,
      sku: (variantSKUOverrides[combo.key] ?? combo.sku ?? '').trim() || undefined,
      barcode: (variantBarcodeOverrides[combo.key] ?? combo.barcode ?? '').trim() || undefined,
      reference: (variantReferenceOverrides[combo.key] ?? combo.reference ?? '').trim() || undefined,
      vendor_reference: (variantVendorRefOverrides[combo.key] ?? combo.vendor_reference ?? '').trim() || undefined,
      active: variantActiveOverrides[combo.key] !== undefined ? variantActiveOverrides[combo.key] : combo.active !== false,
    }));
  };

  const buildDefaultCombo = (): VariantComboForm => ({
    attribute_value_ids: {},
    price: data.price,
    cost_price: data.cost_price,
    track_inventory: data.track_inventory,
    stock_by_warehouse: data.track_inventory ? singleVariantStockByWarehouse : undefined,
    sku: (data.sku || '').trim() || undefined,
    barcode: (data.barcode || '').trim() || undefined,
    reference: (data.reference || '').trim() || undefined,
    vendor_reference: (data.vendor_reference || '').trim() || undefined,
    active: true,
  });

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

  const updateSingleVariantWarehouseStock = (warehouseID: number, nextQuantity: number) => {
    setSingleVariantStockByWarehouse((current) => ({
      ...current,
      [warehouseID]: Math.max(0, nextQuantity),
    }));
  };

  const updateVariantWarehouseStock = (comboKey: string, warehouseID: number, nextQuantity: number) => {
    setVariantStockByWarehouseOverrides((current) => {
      const currentStock = current[comboKey] || {};
      return {
        ...current,
        [comboKey]: {
          ...currentStock,
          [warehouseID]: Math.max(0, nextQuantity),
        },
      };
    });
  };

  const submit = () => {
    setVariantSetupError('');

    const isProduct = data.item_type === 'product';
    const hasVariantSetup = params.action !== 'view' && isProduct && data.has_variants;
    const usesSingleProductSetup = params.action !== 'view' && isProduct && !data.has_variants;
    const uniqueAttributeIDs = Array.from(new Set(selectedAttributeIDs));
    const variantCombos = hasVariantSetup ? buildVariantCombos() : usesSingleProductSetup ? [buildDefaultCombo()] : [];

    if (hasVariantSetup) {
      if (selectedAttributeIDs.length === 0) {
        setVariantSetupError(t('items.single.variantSetup.attributeRequired'));
        return;
      }

      if (uniqueAttributeIDs.length !== selectedAttributeIDs.length) {
        setVariantSetupError('Duplicate attributes are not allowed.');
        return;
      }

      const hasDuplicateAttributeValues = uniqueAttributeIDs.some((attributeID) => {
        const valueIDs = selectedValuesByAttribute[attributeID] || [];
        return new Set(valueIDs).size !== valueIDs.length;
      });

      if (hasDuplicateAttributeValues) {
        setVariantSetupError('Duplicate values are not allowed for the same attribute.');
        return;
      }

      if (variantCombos.length === 0) {
        setVariantSetupError(t('items.single.variantSetup.valueRequired'));
        return;
      }

      const seenSKUs = new Set<string>();
      const seenBarcodes = new Set<string>();
      for (const combo of variantCombos) {
        const sku = (combo.sku || '').trim();
        const barcode = (combo.barcode || '').trim();

        if (sku) {
          if (seenSKUs.has(sku)) {
            setVariantSetupError('SKU values must be unique across variants.');
            return;
          }
          seenSKUs.add(sku);
        }

        if (barcode) {
          if (seenBarcodes.has(barcode)) {
            setVariantSetupError('Barcode values must be unique across variants when provided.');
            return;
          }
          seenBarcodes.add(barcode);
        }
      }
    }

    transform((data) => {
      const { reference, code, sku, barcode, vendor_reference, ...rest } = data;

      return {
        ...rest,
        has_variants: isProduct ? data.has_variants : false,
        attribute_ids: hasVariantSetup ? uniqueAttributeIDs : [],
        variant_combos: isProduct ? variantCombos : [],
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

  const showVariantMatrixEditor = params.action !== 'view' && data.item_type === 'product' && data.has_variants;
  const showSingleVariantFields =
    data.item_type !== 'product' || (params.action !== 'view' && !data.has_variants) || (params.action === 'view' && !existingHasVariants);
  const showExpandedVariantSection = showVariantMatrixEditor || (params.action === 'view' && data.item_type === 'product' && existingHasVariants);
  const variantSectionLabel = showExpandedVariantSection ? 'VARIANTS' : 'PRICING & INVENTORY';

  return (
    <div className="flex flex-col space-y-6">
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('items.single.title')}</FormSection.Title>
        <FormSection.Description>{t('items.single.description')}</FormSection.Description>
        <FormSection.Form>
          {propsErrors.status && <div className="mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}

          <div className="col-span-6 grid gap-6 lg:grid-cols-12">
            <div className="border-border space-y-4 rounded-lg border p-4 lg:col-span-7">
              <p className="text-muted-foreground text-xs font-semibold tracking-[0.2em]">GENERAL</p>

              <div className="space-y-2">
                <Label htmlFor="item_type">{t('items.single.type')}</Label>
                <RadioGroup
                  className="grid grid-cols-1 gap-3 sm:grid-cols-3"
                  value={data.item_type}
                  onChange={(type: ItemType) => {
                    setData('item_type', type);

                    if (type !== 'product') {
                      setData('has_variants', false);
                      setData('track_inventory', false);
                      setSelectedAttributeIDs([]);
                      setSelectedValuesByAttribute({});
                      setSingleVariantStockByWarehouse({});
                      setVariantTrackInventoryOverrides({});
                      setVariantStockByWarehouseOverrides({});
                      setVariantSetupError('');
                    } else {
                      setData('track_inventory', true);
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

              <div className="space-y-2">
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

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="flex flex-col gap-2">
                  <Label>{t('global.unit')}</Label>
                  <Select
                    onValueChange={(value) => setData('unit_id', Number(value))}
                    disabled={viewMode}
                    defaultValue={data.unit_id.toString()}
                    value={data.unit_id.toString()}
                    required
                  >
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select item unit" />
                    </SelectTrigger>
                    <SelectContent>
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
                    <SelectTrigger className="w-full">
                      <SelectValue placeholder="Select item tax" />
                    </SelectTrigger>
                    <SelectContent>
                      {params.taxes.map((tax) => (
                        <SelectItem key={tax.id} value={tax.id.toString()}>
                          {tax.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                  <InputError className="mt-2" message={errors.tax_id} />
                </div>
              </div>
            </div>

            <div className={`border-border space-y-4 rounded-lg border p-4 ${showExpandedVariantSection ? 'lg:col-span-12' : 'lg:col-span-5'}`}>
              <p className="text-muted-foreground text-xs font-semibold tracking-[0.2em]">{variantSectionLabel}</p>

              {params.action !== 'view' && data.item_type === 'product' && (
                <div className="flex flex-col gap-3 rounded-md border p-3 sm:flex-row sm:items-center sm:justify-between">
                  <div className="space-y-1">
                    <Label htmlFor="has_variants">This product has variants</Label>
                    <p className="text-muted-foreground text-sm">
                      {data.has_variants ? t('items.single.hasVariantsHelp') : 'When disabled, this item uses a single default variant setup.'}
                    </p>
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
              )}

              {showSingleVariantFields && (
                <div className="space-y-4">
                  <div className="grid gap-4 sm:grid-cols-2">
                    <div className="space-y-2">
                      <Label htmlFor="price">{t('global.price')}</Label>
                      <Input
                        id="price"
                        type="number"
                        min={0}
                        step="0.01"
                        className="text-right"
                        value={data.price}
                        onChange={(e) => setData('price', e.target.valueAsNumber)}
                        placeholder="0.00"
                        readOnly={viewMode}
                      />
                      <InputError className="mt-2" message={errors.price} />
                    </div>

                    <div className="space-y-2">
                      <Label htmlFor="cost_price">Cost</Label>
                      <Input
                        id="cost_price"
                        type="number"
                        min={0}
                        step="0.01"
                        className="text-right"
                        value={data.cost_price ?? ''}
                        onChange={(e) => {
                          const value = e.target.valueAsNumber;
                          setData('cost_price', Number.isNaN(value) ? undefined : value);
                        }}
                        placeholder="0.00"
                        readOnly={viewMode}
                      />
                      <InputError className="mt-2" message={errors.cost_price} />
                    </div>
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="sku">{t('global.sku')}</Label>
                    <Input
                      id="sku"
                      value={data.sku}
                      onChange={(e) => setData('sku', e.target.value)}
                      autoComplete="sku"
                      placeholder="Item sku"
                      readOnly={viewMode}
                    />
                    <InputError className="mt-2" message={errors.sku} />
                  </div>

                  <div className="space-y-2">
                    <Label htmlFor="barcode">{t('global.barcode')}</Label>
                    <Input
                      id="barcode"
                      value={data.barcode}
                      onChange={(e) => setData('barcode', e.target.value)}
                      autoComplete="barcode"
                      placeholder="Item barcode"
                      readOnly={viewMode}
                    />
                    <InputError className="mt-2" message={errors.barcode} />
                  </div>

                  {data.item_type === 'product' && (
                    <div className="flex flex-col gap-3 rounded-md border p-3 sm:flex-row sm:items-center sm:justify-between">
                      <div className="space-y-1">
                        <Label htmlFor="track_inventory">Track Inventory</Label>
                        <p className="text-muted-foreground text-sm">Use stock levels and movements for this product.</p>
                      </div>
                      <Switch
                        id="track_inventory"
                        checked={data.track_inventory}
                        disabled={viewMode}
                        onCheckedChange={(checked) => setData('track_inventory', checked)}
                      />
                    </div>
                  )}

                  {data.item_type === 'product' && data.track_inventory && (
                    <div className="space-y-3 rounded-md border p-3">
                      <div className="space-y-1">
                        <Label>Warehouse quantities</Label>
                        <p className="text-muted-foreground text-sm">Inventory belongs to the variant and warehouse.</p>
                      </div>

                      {warehouseOptions.length === 0 ? (
                        <p className="text-muted-foreground text-sm">No warehouses configured yet.</p>
                      ) : (
                        <div className="space-y-2">
                          {warehouseOptions.map((warehouse) => (
                            <div key={`single-stock-${warehouse.id}`} className="grid grid-cols-[1fr_120px] items-center gap-3">
                              <p className="text-sm font-medium">{warehouse.name} -&gt; quantity</p>
                              <Input
                                type="number"
                                min={0}
                                step={1}
                                className="h-8 text-right"
                                value={(singleVariantStockByWarehouse[warehouse.id] ?? 0).toString()}
                                readOnly={viewMode}
                                onChange={(e) => {
                                  const nextQuantity = e.target.valueAsNumber;
                                  updateSingleVariantWarehouseStock(warehouse.id, Number.isNaN(nextQuantity) ? 0 : nextQuantity);
                                }}
                              />
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  )}
                </div>
              )}

              {showVariantMatrixEditor && (
                <div className="space-y-4">
                  <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.priceTemplateHelp')}</p>
                  <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.skuAutoHelp')}</p>

                  <div className="space-y-2">
                    <Label>{t('items.single.variantSetup.attributes')}</Label>
                    {attributeOptions.length === 0 && <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.noAttributes')}</p>}

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
                      <div className="flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
                        <Label>{t('items.single.variantSetup.variants')}</Label>
                        <div className="flex flex-wrap gap-2">
                          <Input
                            type="number"
                            min={0}
                            step="0.01"
                            className="h-8 w-28 text-right text-xs"
                            value={bulkPriceInput}
                            onChange={(e) => setBulkPriceInput(e.target.value)}
                            placeholder="0.00"
                          />
                          <Button type="button" variant="outline" size="sm" onClick={applyBulkPriceFromInput}>
                            Apply Price to Empty
                          </Button>
                          <Button type="button" variant="outline" size="sm" onClick={handleBulkGenerateSKUs}>
                            Generate Empty SKUs
                          </Button>
                          <Button type="button" variant="outline" size="sm" onClick={handleBulkGenerateBarcodes}>
                            Generate Empty Barcodes
                          </Button>
                        </div>
                      </div>

                      <div className="space-y-3 rounded-md border p-3">
                        <div className="grid gap-3 md:grid-cols-2">
                          <div className="space-y-2">
                            <Label htmlFor="variant-sku-search">Search by SKU</Label>
                            <Input
                              id="variant-sku-search"
                              type="text"
                              className="h-8"
                              placeholder="Type SKU..."
                              value={variantSKUFilter}
                              onChange={(e) => setVariantSKUFilter(e.target.value)}
                            />
                          </div>

                          <div className="space-y-2">
                            <Label htmlFor="variant-barcode-search">Search by barcode</Label>
                            <Input
                              id="variant-barcode-search"
                              type="text"
                              className="h-8"
                              placeholder="Type barcode..."
                              value={variantBarcodeFilter}
                              onChange={(e) => setVariantBarcodeFilter(e.target.value)}
                            />
                          </div>
                        </div>

                        {variantFilterAttributes.length > 0 && (
                          <div className="space-y-2">
                            <Label>Filter by attribute values</Label>
                            <div className="space-y-2">
                              {variantFilterAttributes.map((attribute) => (
                                <div key={`variant-filter-${attribute.id}`} className="space-y-1">
                                  <p className="text-muted-foreground text-xs font-medium">{attribute.name}</p>
                                  <div className="grid gap-2 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4">
                                    {attribute.values.map((value) => (
                                      <div key={`variant-filter-${attribute.id}-${value.id}`} className="flex items-center gap-2">
                                        <Checkbox
                                          id={`variant-filter-${attribute.id}-value-${value.id}`}
                                          checked={(variantAttributeValueFilters[attribute.id] || []).includes(value.id)}
                                          onCheckedChange={(checked) => toggleVariantFilterValue(attribute.id, value.id, checked === true)}
                                        />
                                        <Label htmlFor={`variant-filter-${attribute.id}-value-${value.id}`}>
                                          {value.display_name || value.value}
                                        </Label>
                                      </div>
                                    ))}
                                  </div>
                                </div>
                              ))}
                            </div>
                          </div>
                        )}

                        <div className="flex items-center justify-between gap-3">
                          <p className="text-muted-foreground text-xs">
                            Showing {visibleVariantStart}-{visibleVariantEnd} of {filteredVariantComboPreviews.length} filtered variants (
                            {variantComboPreviews.length} total)
                          </p>
                          <Button type="button" variant="ghost" size="sm" onClick={clearVariantFilters} disabled={!hasActiveVariantFilters}>
                            Clear filters
                          </Button>
                        </div>
                        {filteredVariantComboPreviews.length > 0 && (
                          <p className="text-muted-foreground text-xs">Variants are loaded on demand by page. Each page renders up to 50 rows.</p>
                        )}
                      </div>

                      <div ref={variantTableContainerRef} className="max-h-136 overflow-auto rounded-md border" onScroll={handleVariantTableScroll}>
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
                              <th className="w-20 p-2 text-center font-medium">Track</th>
                              <th className="p-2 text-left font-medium">Warehouse quantities</th>
                              <th className="w-16 p-2 text-center font-medium">Active</th>
                            </tr>
                          </thead>
                          <tbody>
                            {visibleVariantComboPreviews.length === 0 ? (
                              <tr>
                                <td colSpan={10} className="text-muted-foreground p-4 text-center text-sm">
                                  No variants match the current filters.
                                </td>
                              </tr>
                            ) : (
                              visibleVariantComboPreviews.map((combo) => {
                                const isActive =
                                  variantActiveOverrides[combo.key] !== undefined ? variantActiveOverrides[combo.key] : combo.active !== false;
                                const isTrackingInventory = currentVariantTrackInventory(combo);
                                const stockByWarehouse = currentVariantStockByWarehouse(combo);

                                return (
                                  <tr key={combo.key} className={`border-b last:border-0 ${!isActive ? 'opacity-50' : ''}`}>
                                    <td className="p-2">
                                      <div className="text-sm">{combo.label || '-'}</div>
                                    </td>
                                    <td className="p-2">
                                      <Input
                                        type="text"
                                        className="h-8 text-xs"
                                        value={variantSKUOverrides[combo.key] ?? combo.sku ?? ''}
                                        placeholder="SKU"
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
                                        value={variantBarcodeOverrides[combo.key] ?? combo.barcode ?? ''}
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
                                        value={variantReferenceOverrides[combo.key] ?? combo.reference ?? ''}
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
                                        value={variantVendorRefOverrides[combo.key] ?? combo.vendor_reference ?? ''}
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
                                        value={(variantCostPriceOverrides[combo.key] ?? combo.cost_price)?.toString() ?? ''}
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
                                        value={(variantPriceOverrides[combo.key] ?? combo.price ?? data.price).toString()}
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
                                      <Switch
                                        checked={isTrackingInventory}
                                        onCheckedChange={(checked) => {
                                          setVariantTrackInventoryOverrides((current) => ({
                                            ...current,
                                            [combo.key]: checked,
                                          }));
                                        }}
                                      />
                                    </td>
                                    <td className="p-2 align-top">
                                      {!isTrackingInventory ? (
                                        <p className="text-muted-foreground text-xs">Tracking disabled for this variant.</p>
                                      ) : warehouseOptions.length === 0 ? (
                                        <p className="text-muted-foreground text-xs">No warehouses configured yet.</p>
                                      ) : (
                                        <div className="space-y-2">
                                          {warehouseOptions.map((warehouse) => (
                                            <div key={`${combo.key}-stock-${warehouse.id}`} className="grid grid-cols-[1fr_120px] items-center gap-3">
                                              <p className="text-xs font-medium whitespace-nowrap">{warehouse.name} -&gt; quantity</p>
                                              <Input
                                                type="number"
                                                min={0}
                                                step={1}
                                                className="h-8 text-right text-xs"
                                                value={(stockByWarehouse[warehouse.id] ?? 0).toString()}
                                                onChange={(e) => {
                                                  const value = e.target.valueAsNumber;
                                                  updateVariantWarehouseStock(combo.key, warehouse.id, Number.isNaN(value) ? 0 : value);
                                                }}
                                              />
                                            </div>
                                          ))}
                                        </div>
                                      )}
                                    </td>
                                    <td className="p-2 text-center">
                                      <Checkbox
                                        checked={isActive}
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
                              })
                            )}
                          </tbody>
                        </table>
                      </div>

                      {filteredVariantComboPreviews.length > 0 && (
                        <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:justify-between">
                          <p className="text-muted-foreground text-xs">
                            Page {currentVariantTablePage} of {totalVariantTablePages}. Scroll to the bottom/top edge to lazy-load next/previous 50
                            variants.
                          </p>
                          <div className="flex items-center gap-2">
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              disabled={currentVariantTablePage <= 1}
                              onClick={() => setVariantTablePage((current) => Math.max(1, current - 1))}
                            >
                              Previous 50
                            </Button>
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              disabled={currentVariantTablePage >= totalVariantTablePages}
                              onClick={() => setVariantTablePage((current) => Math.min(totalVariantTablePages, current + 1))}
                            >
                              Next 50
                            </Button>
                          </div>
                        </div>
                      )}
                    </div>
                  )}

                  {!!variantSetupError && <InputError className="mt-2" message={variantSetupError} />}
                </div>
              )}

              {params.action !== 'view' && data.item_type === 'product' && !data.has_variants && (
                <div className="rounded-md bg-blue-50 p-3">
                  <p className="text-sm text-blue-800">Using a single product setup.</p>
                </div>
              )}

              {params.action === 'view' && data.item_type === 'product' && existingHasVariants && (
                <div className="space-y-4 rounded-md border p-3">
                  <div className="space-y-1">
                    <Label>{t('items.single.variantSetup.currentTitle')}</Label>
                    <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.currentDescription')}</p>
                  </div>

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
                      {(existingVariantSetup?.variants || []).length === 0 ? (
                        <p className="text-muted-foreground text-sm">{t('items.single.variantSetup.noVariants')}</p>
                      ) : (
                        <div className="space-y-1">
                          {existingVariantSetup?.variants.map((variant) => (
                            <div key={variant.uuid} className="rounded-md border p-2">
                              <p className="text-sm">
                                <span className="font-medium">{variant.name}</span> ({variant.sku})
                              </p>
                              {variant.track_inventory === false ? (
                                <p className="text-muted-foreground text-xs">Tracking disabled for this variant.</p>
                              ) : (
                                <div className="mt-1 space-y-1">
                                  {warehouseOptions.map((warehouse) => (
                                    <p key={`${variant.uuid}-warehouse-${warehouse.id}`} className="text-xs">
                                      <span className="font-medium">{warehouse.name}</span> -&gt; quantity{' '}
                                      {variant.stock_by_warehouse?.[warehouse.id] ?? 0}
                                    </p>
                                  ))}
                                  {warehouseOptions.length === 0 && <p className="text-muted-foreground text-xs">No warehouses configured yet.</p>}
                                </div>
                              )}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  </>
                </div>
              )}
            </div>

            <div className="border-border space-y-4 rounded-lg border p-4 lg:col-span-6">
              <p className="text-muted-foreground text-xs font-semibold tracking-[0.2em]">INVENTORY</p>

              <div className="grid gap-4 sm:grid-cols-2">
                <div className="space-y-2">
                  <Label htmlFor="reference">{t('global.reference')}</Label>
                  <Input
                    id="reference"
                    value={data.reference}
                    onChange={(e) => setData('reference', e.target.value)}
                    autoComplete="reference"
                    placeholder="Item reference"
                    readOnly={viewMode}
                  />
                  <InputError className="mt-2" message={errors.reference} />
                </div>

                <div className="space-y-2">
                  <Label htmlFor="code">{t('global.code')}</Label>
                  <Input
                    id="code"
                    value={data.code}
                    onChange={(e) => setData('code', e.target.value)}
                    autoComplete="code"
                    placeholder="Item code"
                    readOnly={viewMode}
                  />
                  <InputError className="mt-2" message={errors.code} />
                </div>
              </div>
            </div>

            <div className="border-border space-y-4 rounded-lg border p-4 lg:col-span-6">
              <p className="text-muted-foreground text-xs font-semibold tracking-[0.2em]">PURCHASING</p>

              <div className="space-y-2">
                <Label htmlFor="vendor_reference">{t('items.single.vendor_reference')}</Label>
                <Input
                  id="vendor_reference"
                  value={data.vendor_reference}
                  onChange={(e) => setData('vendor_reference', e.target.value)}
                  placeholder="Item vendor reference"
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.vendor_reference} />
              </div>
            </div>

            <div className="border-border space-y-4 rounded-lg border p-4 lg:col-span-12">
              <p className="text-muted-foreground text-xs font-semibold tracking-[0.2em]">DESCRIPTION</p>

              <div className="grid gap-2">
                <Label htmlFor="description">{t('global.description')}</Label>
                <textarea
                  id="description"
                  className="mt-1 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900 focus:border-blue-500 focus:ring-blue-500"
                  value={data.description}
                  onChange={(e) => setData('description', e.target.value)}
                  placeholder="Write some description here..."
                  rows={3}
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.description} />
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
