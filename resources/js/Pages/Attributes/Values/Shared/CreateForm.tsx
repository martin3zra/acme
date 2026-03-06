import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps } from '@/types';
import { useForm, usePage } from '@inertiajs/react';
import { AttributeValue, AttributeValueForm } from '../../types';

type CreateFormProps = {
  attributeId: number;
  value?: AttributeValue;
  onFinish: () => void;
};

export default function CreateForm({ attributeId, value, onFinish }: CreateFormProps) {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;

  const { data, setData, post, put, errors, processing, reset } = useForm<AttributeValueForm>({
    attribute_id: attributeId,
    value: value?.value || '',
    display_name: value?.display_name || '',
    sort_order: value?.sort_order || 0,
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
    if (value) {
      put(`/attribute-values/${value.id}`, options);
      return;
    }

    post(`/attributes/${attributeId}/values`, options);
  };

  return (
    <FormSection onSubmit={submit}>
      <FormSection.Form>
        {propsErrors.status && <div className="col-span-6 mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}

        <div className="col-span-6 gap-2">
          <Label htmlFor="value">{t('global.code')}</Label>
          <Input
            id="value"
            className="mt-1 block w-full"
            value={data.value}
            onChange={(e) => setData('value', e.target.value)}
            placeholder="e.g., red, blue, s, m, l"
            required
          />
          <InputError className="mt-2" message={errors.value} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="display_name">{t('global.displayName')}</Label>
          <Input
            id="display_name"
            className="mt-1 block w-full"
            value={data.display_name}
            onChange={(e) => setData('display_name', e.target.value)}
            placeholder="e.g., Red, Blue, Small, Medium, Large"
            required
          />
          <InputError className="mt-2" message={errors.display_name} />
        </div>

        <div className="col-span-6 gap-2">
          <Label htmlFor="sort_order">{t('attributes.values.sortOrder')}</Label>
          <Input
            id="sort_order"
            type="number"
            className="mt-1 block w-full"
            value={data.sort_order}
            onChange={(e) => setData('sort_order', parseInt(e.target.value) || 0)}
            placeholder="0"
          />
          <InputError className="mt-2" message={errors.sort_order} />
          <p className="mt-1 text-sm text-gray-500">{t('attributes.values.sortOrderHelp')}</p>
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
