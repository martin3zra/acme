import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { Company, Verb } from '@/types';
import { useForm } from '@inertiajs/react';
import FormSection from './form-section';
import InputError from './input-error';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';

export interface CompanyForm {
  name: string;
  identifier: string;
  city: string;
  address: string;
}

export type CreateFormParams = {
  company: Company | undefined;
  action: Verb;
};

export type CreateFormProps = {
  onFinish: () => void;
  params: CreateFormParams;
};

export default function CreateCompanyForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { data, setData, errors, processing, post, transform, reset } = useForm<Required<CompanyForm>>({
    name: params.company?.name || '',
    identifier: params.company?.identifier || '',
    city: params.company?.city || '',
    address: params.company?.address || '',
  });
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setData(event.target.name as keyof CompanyForm, event.target.value);
  };

  const submit = () => {
    transform((data) => ({
      ...data,
      rnc: data.identifier,
    }));
    post('/companies', {
      ...headers,
      onFinish: () => {
        reset();
        onFinish();
      },
    });
  };
  return (
    <div>
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('companies.single.title')}</FormSection.Title>
        <FormSection.Description>{t('companies.single.description')}</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 space-y-2">
            <Label htmlFor="name" className="text-end">
              {t('companies.single.name')}
            </Label>
            <Input type="text" name="name" className="h-12 md:text-xl" onChange={handleChange} value={data.name} autoFocus />
            <InputError message={errors.name} />
          </div>
          <div className="col-span-3 space-y-2 sm:col-span-3">
            <Label htmlFor="identifier">{t('companies.single.rnc')}</Label>
            <Input
              type="text"
              name="identifier"
              maxLength={11}
              className="h-12 text-start md:text-xl"
              value={data.identifier}
              onChange={handleChange}
            />
            <InputError message={errors.identifier} />
          </div>
          <div className="col-span-6 space-y-2">
            <Label htmlFor="city">{t('companies.single.city')}</Label>
            <Input type="text" name="city" className="h-12 text-start md:text-xl" value={data.city} onChange={handleChange} />
            <InputError message={errors.city} />
          </div>
          <div className="col-span-6 space-y-2">
            <Label htmlFor="address">{t('companies.single.address')}</Label>
            <Input type="text" name="address" className="h-12 text-start md:text-xl" value={data.address} onChange={handleChange} />
            <InputError message={errors.address} />
          </div>
        </FormSection.Form>
        <FormSection.Actions>
          <Button type="submit" disabled={processing} className="h-12 md:text-xl">
            {t('global.save')}
          </Button>
        </FormSection.Actions>
      </FormSection>
    </div>
  );
}
