import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps } from '@/types';
import { useForm, usePage } from '@inertiajs/react';
import { Warehouse } from '../types';

type CreateFormProps = {
  warehouse?: Warehouse;
  onFinish: () => void;
};

type WarehouseForm = {
  name: string;
  address: string;
  description: string;
};

export default function CreateForm({ warehouse, onFinish }: CreateFormProps) {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;

  const { data, setData, post, put, errors, processing, reset } = useForm<WarehouseForm>({
    name: warehouse?.name || '',
    address: warehouse?.address || '',
    description: warehouse?.description || '',
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
    if (warehouse) {
      put(`/warehouses/${warehouse.id}`, options);
      return;
    }

    post('/warehouses', options);
  };

  return (
    <FormSection onSubmit={submit}>
      <FormSection.Form>
        {propsErrors.status && <div className="col-span-6 mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}

        <div className="col-span-6 gap-2">
          <Label htmlFor="name">{t('global.name')}</Label>
          <Input id="name" className="mt-1 block w-full" value={data.name} onChange={(e) => setData('name', e.target.value)} required />
          <InputError className="mt-2" message={errors.name} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="address">{t('global.address')}</Label>
          <Input id="address" className="mt-1 block w-full" value={data.address} onChange={(e) => setData('address', e.target.value)} />
          <InputError className="mt-2" message={errors.address} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="description">{t('global.description')}</Label>
          <textarea
            id="description"
            className="mt-1 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900"
            value={data.description}
            onChange={(e) => setData('description', e.target.value)}
            rows={3}
          />
          <InputError className="mt-2" message={errors.description} />
        </div>
      </FormSection.Form>

      <FormSection.Actions>
        <Button disabled={processing} className="uppercase">
          {t('global.save')}
        </Button>
      </FormSection.Actions>
    </FormSection>
  );
}
