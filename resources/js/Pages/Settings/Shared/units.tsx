import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { Unit } from '@/types';
import { useForm } from '@inertiajs/react';
import { formatDate } from 'date-fns/format';

type Props = {
  units: Unit[];
};

type UnitKey = 'id' | 'name' | 'base_qty';

export default function UnitList({ units }: Props) {
  const { headers } = useHeader();
  const t = useTranslation().trans;
  const { data, setData, post, put, processing, errors, reset } = useForm({
    id: 0,
    name: '',
    base_qty: 1,
  });
  const handleSubmit = () => {
    if (data.id > 0) {
      put(`/units/${data.id}`, { ...headers, preserveState: 'errors' });
      return;
    }
    post('/units', { ...headers, preserveState: 'errors' });
  };
  const handleOnUnitClick = (unit: Unit) => {
    Object.entries(unit).map(([k, v]) => setData(k as UnitKey, v));
  };
  return (
    <div className="flex w-full flex-col space-y-6 py-6 **:data-form:md:grid-cols-1">
      <FormSection onSubmit={handleSubmit}>
        <FormSection.Title>{t('profile.companies.viewCompany.units.title')}</FormSection.Title>
        <FormSection.Description>{t('profile.companies.viewCompany.units.description')}</FormSection.Description>
        <FormSection.Form>
          <ul className="divide-muted w-full divide-y rounded-md border">
            <li className="bg-muted/50 grid grid-cols-3 px-4 py-2 text-sm font-medium">
              <span>{t('global.name')}</span>
              <span>{t('global.description')}</span>
              <span className="text-end">{t('global.addedAt')}</span>
            </li>
            {units.map((unit, idx) => (
              <li
                onClick={() => handleOnUnitClick(unit)}
                key={unit.id}
                className={cn('hover:bg-muted/30 grid cursor-pointer grid-cols-3 px-4 py-2 text-sm', idx % 2 === 0 && 'bg-muted/10')}
              >
                <span>{unit.name}</span>
                <span>{unit.base_qty}</span>
                <span className="text-end">{formatDate(new Date(unit.created_at), 'dd-MM-yyyy')}</span>
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
                      value={data.base_qty}
                      onChange={(event) => {
                        let value = event.target.valueAsNumber;
                        if (value < 0) value = 0;
                        if (value > 100) value = 100; // clamp to max
                        setData('base_qty', value);
                      }}
                      className="w-16 text-end"
                    />
                    <InputError message={errors.base_qty} />
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
                      <Button onClick={() => reset('name', 'base_qty', 'id')} type="reset" size="sm" className="mt-1" variant={'ghost'}>
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
