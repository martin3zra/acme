import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { Tax } from '@/types';
import { useForm } from '@inertiajs/react';
import { formatDate } from 'date-fns/format';

type Props = {
  taxes: Tax[];
};

type TaxKey = 'name' | 'rate';
export default function TaxList({ taxes }: Props) {
  const { headers } = useHeader();
  const t = useTranslation().trans;
  const { data, setData, post, put, processing, errors, reset } = useForm({
    uuid: undefined,
    name: '',
    rate: 0,
  });

  const handleOnTaxClick = (tax: Tax) => {
    Object.entries(tax).map(([k, v]) => setData(k as TaxKey, v));
  };

  const handleSubmit = () => {
    if (data.uuid) {
      put(`/taxes/${data.uuid}`, { ...headers, preserveState: 'errors' });
      return;
    }
    post('/taxes', { ...headers, preserveState: 'errors' });
  };
  return (
    <div className="flex w-full flex-col space-y-6 py-6 **:data-form:md:grid-cols-1">
      <FormSection onSubmit={handleSubmit}>
        <FormSection.Title>{t('profile.companies.viewCompany.taxes.title')}</FormSection.Title>
        <FormSection.Description>{t('profile.companies.viewCompany.taxes.description')}</FormSection.Description>
        <FormSection.Form>
          <ul className="divide-muted w-full divide-y rounded-md border">
            <li className="bg-muted/50 grid grid-cols-3 px-4 py-2 text-sm font-medium">
              <span>{t('global.name')}</span>
              <span className="text-right">{t('global.taxRate')}</span>
              <span className="text-center">{t('global.addedAt')}</span>
            </li>
            {taxes.map((tax, idx) => (
              <li
                onClick={() => handleOnTaxClick(tax)}
                key={tax.id}
                className={cn('hover:bg-muted/30 grid cursor-pointer grid-cols-3 px-4 py-2 text-sm', idx % 2 === 0 && 'bg-muted/10')}
              >
                <span>{tax.name}</span>
                <span className="text-right">{tax.rate}%</span>
                <span className="text-center">{formatDate(new Date(tax.created_at), 'dd-MM-yyyy')}</span>
              </li>
            ))}

            <li className="bg-muted/20 px-4 py-3">
              <div className="grid grid-cols-2 items-start gap-2">
                <div className="w-56">
                  <Input placeholder={t('global.name')} value={data.name} onChange={(e) => setData('name', e.target.value)} />
                  {errors.name && Array.isArray(errors.name) && (
                    <>
                      {errors.name.map((message, index) => (
                        <InputError key={index} message={message} />
                      ))}
                    </>
                  )}
                </div>
                <div className="flex justify-end gap-3">
                  <div>
                    <Input
                      type="number"
                      placeholder="Rate"
                      value={data.rate}
                      onChange={(event) => {
                        let value = event.target.valueAsNumber;
                        if (value < 0) value = 0;
                        if (value > 100) value = 100; // clamp to max
                        setData('rate', value);
                      }}
                      className="w-16 text-end"
                    />
                    <InputError message={errors.rate} />
                  </div>
                  <div className="flex space-x-2">
                    <Button disabled={processing} type="submit" size="sm" className="mt-1">
                      {processing ? (
                        <>
                          <Spinner />
                          {t('global.saving')}
                        </>
                      ) : (
                        <>{t('global.save')}</>
                      )}
                    </Button>
                    {!processing && (
                      <Button onClick={() => reset('name', 'rate', 'uuid')} type="reset" size="sm" className="mt-1" variant={'ghost'}>
                        {t('global.cancel')}
                      </Button>
                    )}
                  </div>
                </div>
              </div>
            </li>
          </ul>
        </FormSection.Form>
      </FormSection>
    </div>
  );
}
