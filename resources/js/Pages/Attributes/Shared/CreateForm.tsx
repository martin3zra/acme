import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps } from '@/types';
import { useForm, usePage } from '@inertiajs/react';
import { Attribute, AttributeForm } from '../types';

type CreateFormProps = {
  attribute?: Attribute;
  onFinish: () => void;
};

export default function CreateForm({ attribute, onFinish }: CreateFormProps) {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;

  const { data, setData, post, put, errors, processing, reset } = useForm<AttributeForm>({
    name: attribute?.name || '',
    type: attribute?.type || 'select',
    display_name: attribute?.display_name || '',
    description: attribute?.description || '',
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
    if (attribute) {
      put(`/attributes/${attribute.id}`, options);
      return;
    }

    post('/attributes', options);
  };

  return (
    <FormSection onSubmit={submit}>
      <FormSection.Form>
        {propsErrors.status && <div className="col-span-6 mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}

        <div className="col-span-6 gap-2">
          <Label htmlFor="name">{t('global.name')}</Label>
          <Input
            id="name"
            className="mt-1 block w-full"
            value={data.name}
            onChange={(e) => setData('name', e.target.value)}
            placeholder="e.g., color, size, length"
            required
          />
          <InputError className="mt-2" message={errors.name} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="display_name">{t('global.displayName')}</Label>
          <Input
            id="display_name"
            className="mt-1 block w-full"
            value={data.display_name}
            onChange={(e) => setData('display_name', e.target.value)}
            placeholder="e.g., Color, Shirt Size, Length"
            required
          />
          <InputError className="mt-2" message={errors.display_name} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="type">{t('global.type')}</Label>
          <select
            id="type"
            value={data.type}
            onChange={(e) => setData('type', e.target.value)}
            className="mt-1 block w-full rounded-md border p-2"
            required
          >
            <option value="select">Select (Dropdown)</option>
            <option value="text">Text</option>
            <option value="numeric">Numeric</option>
          </select>
          <InputError className="mt-2" message={errors.type} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="description">{t('global.description')}</Label>
          <textarea
            id="description"
            className="mt-1 block w-full rounded-lg border border-gray-300 bg-gray-50 p-2.5 text-sm text-gray-900"
            value={data.description}
            onChange={(e) => setData('description', e.target.value)}
            rows={3}
            placeholder="Optional description"
          />
          <InputError className="mt-2" message={errors.description} />
        </div>
      </FormSection.Form>

      <FormSection.Actions>
        <Button type="submit" disabled={processing} className="uppercase">
          {t('global.save')}
        </Button>
      </FormSection.Actions>
    </FormSection>
  );
}
