import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps } from '@/types';
import { useForm, usePage } from '@inertiajs/react';
import { StockAdjustmentForm, Warehouse } from '../types';

type CreateFormProps = {
  warehouses: Warehouse[];
  onFinish: () => void;
};

export default function CreateForm({ warehouses, onFinish }: CreateFormProps) {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;

  const { data, setData, post, errors, processing, reset } = useForm<StockAdjustmentForm>({
    warehouse_id: '',
    variant_id: '',
    quantity: '',
    reason: '',
  });

  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => {
      reset();
      onFinish();
    },
  };

  const submit = () => {
    post('/stock-levels/adjust', options);
  };

  return (
    <FormSection onSubmit={submit}>
      <FormSection.Form>
        {propsErrors.status && <div className="col-span-6 mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}

        <div className="col-span-6 gap-2">
          <Label htmlFor="warehouse_id">{t('@global.warehouse')}</Label>
          <select
            id="warehouse_id"
            value={data.warehouse_id}
            onChange={(e) => setData('warehouse_id', e.target.value)}
            className="mt-1 block w-full rounded-md border p-2"
            required
          >
            <option value="">Select warehouse</option>
            {warehouses.map((warehouse) => (
              <option key={warehouse.id} value={warehouse.id}>
                {warehouse.name}
              </option>
            ))}
          </select>
          <InputError className="mt-2" message={errors.warehouse_id} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="variant_id">{t('@global.variant')}</Label>
          <Input
            id="variant_id"
            type="number"
            className="mt-1 block w-full"
            value={data.variant_id}
            onChange={(e) => setData('variant_id', e.target.value)}
            placeholder="Variant ID"
            required
          />
          <InputError className="mt-2" message={errors.variant_id} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="quantity">{t('@global.quantity')}</Label>
          <Input
            id="quantity"
            type="number"
            className="mt-1 block w-full"
            value={data.quantity}
            onChange={(e) => setData('quantity', e.target.value)}
            placeholder="Enter quantity"
            required
          />
          <InputError className="mt-2" message={errors.quantity} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="reason">{t('@global.reason')}</Label>
          <textarea
            id="reason"
            className="mt-1 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900"
            value={data.reason}
            onChange={(e) => setData('reason', e.target.value)}
            rows={3}
            placeholder="Optional reason for adjustment"
          />
          <InputError className="mt-2" message={errors.reason} />
        </div>
      </FormSection.Form>

      <FormSection.Actions>
        <Button type="submit" disabled={processing} className="uppercase">
          {t('@global.adjustStock')}
        </Button>
      </FormSection.Actions>
    </FormSection>
  );
}
