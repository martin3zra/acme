import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { useForm } from '@inertiajs/react';
import FormSection from './form-section';
import InputError from './input-error';
import { Button } from './ui/button';
import { Input } from './ui/input';
import { Label } from './ui/label';

interface CreateCompanyForm {
  name: string;
  rnc: string;
  city: string;
  address: string;
}
export default function CreateCompanyForm() {
  const t = useTranslation().trans;
  const { headers } = useHeader();
  const { data, setData, errors, processing, post, reset } = useForm<Required<CreateCompanyForm>>({
    name: '',
    rnc: '',
    city: '',
    address: '',
  });
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setData(event.target.name as keyof CreateCompanyForm, event.target.value);
  };

  const submit = () => {
    post('/companies', { ...headers, onFinish: () => reset() });
  };
  return (
    <div>
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('onboarding.company.title')}</FormSection.Title>
        <FormSection.Description>{t('onboarding.company.description')}</FormSection.Description>
        <FormSection.Form>
          <div className="col-span-6 space-y-2 sm:col-span-4">
            <Label htmlFor="name" className="text-end">
              {t('onboarding.company.name')}
            </Label>
            <Input type="text" name="name" className="h-12 md:text-xl" onChange={handleChange} value={data.name} autoFocus />
            <InputError message={errors.name} />
          </div>
          <div className="col-span-3 space-y-2 sm:col-span-3">
            <Label htmlFor="rnc">{t('onboarding.company.rnc')}</Label>
            <Input type="text" name="rnc" maxLength={11} className="h-12 text-start md:text-xl" value={data.rnc} onChange={handleChange} />
            <InputError message={errors.rnc} />
          </div>
          <div className="col-span-6 space-y-2 sm:col-span-4">
            <Label htmlFor="city">{t('onboarding.company.city')}</Label>
            <Input type="text" name="city" className="h-12 text-start md:text-xl" value={data.city} onChange={handleChange} />
            <InputError message={errors.city} />
          </div>
          <div className="col-span-6 space-y-2 sm:col-span-4">
            <Label htmlFor="address">{t('onboarding.company.address')}</Label>
            <Input type="text" name="address" className="h-12 text-start md:text-xl" value={data.address} onChange={handleChange} />
            <InputError message={errors.address} />
          </div>
        </FormSection.Form>
        <FormSection.Actions>
          <Button type="submit" disabled={processing} className="h-12 md:text-xl">
            {t('onboarding.company.action')}
          </Button>
        </FormSection.Actions>
      </FormSection>
    </div>
  );
}
