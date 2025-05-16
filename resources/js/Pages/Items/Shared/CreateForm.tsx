import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useHeader } from '@/composables/use-headers';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import { Item, PageProps, Tax, Unit, Verb } from '@/types';
import { useForm, usePage } from '@inertiajs/react';
import { FormEventHandler, useState } from 'react';

export type CreateFormParams = {
  item: Item | undefined;
  taxes: Tax[];
  units: Unit[];
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
};

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { data, setData, post, put, errors, reset, processing } = useForm<Required<ItemForm>>({
    id: params.item?.id,
    name: params.item?.name || '',
    description: params.item?.description || '',
    price: params.item?.price || 0,
    tax_id: params.item?.tax.id || 0,
    unit_id: params.item?.unit.id || 0,
  });

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

  const submit: FormEventHandler = (e) => {
    e.preventDefault();

    if (params.action === 'create') post('/items', options);
    if (params.action === 'edit') put(`/items/${params.item!.id}`, options);
  };

  return (
    <div>
      {propsErrors.status && <div className="mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}
      <form onSubmit={submit} className="space-y-6">
        <div className="grid grid-cols-2 gap-2">
          <div className="grid gap-2">
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
          <div className="grid gap-2">
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
        </div>
        <div className="grid grid-cols-2 gap-2">
          <div className="flex items-center justify-between">
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
        <div className="grid grid-cols-2 gap-2">
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
          {/* <div className="grid gap-2">
            <Label htmlFor="phone">Phone</Label>
            <Input
              id="phone"
              className="mt-1 block w-full"
              value={data.phone}
              onChange={(e) => setData('phone', e.target.value)}
              placeholder="809-983-3897"
              readOnly={viewMode}
            />
            <InputError className="mt-2" message={errors.phone} />
          </div> */}
        </div>
        {!viewMode && (
          <div className="flex items-center gap-4">
            <Button disabled={processing}>
              {t(`global.actions.${verbName}`)} {t('global.item')}
            </Button>
          </div>
        )}
      </form>

      {viewMode && (
        <div className="space-y-6 pt-12">
          <HeadingSmall
            title={t(`items.statuses.${params.item?.status || 'enabled'}.section.title`)}
            description={t(`items.statuses.${params.item?.status || 'enabled'}.section.description`)}
          />
          <div className={`space-y-4 rounded-lg border ${isDisabled ? 'border-primary-100 bg-primary-50' : 'border-red-100 bg-red-50'} p-4`}>
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
              path={`/items/${params.item?.id}/change-status`}
              open={dialogOpen}
              onOpenChange={setDialogOpen}
            />
          </div>
        </div>
      )}
    </div>
  );
}
