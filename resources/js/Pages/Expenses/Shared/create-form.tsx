import ActionSection from '@/components/action-section';
import { ConfirmsPassword } from '@/components/confirms-password';
import { DatePickerField } from '@/components/date-picker';
import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import { useHeader } from '@/composables/use-headers';
import { useNumber } from '@/composables/use-number';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import { Expense, ExpenseCategory, PageProps, Verb } from '@/types';
import { Textarea } from '@headlessui/react';
import { useForm, usePage } from '@inertiajs/react';
import { useState } from 'react';
export type CreateFormParams = {
  expense: Expense | undefined;
  categories: ExpenseCategory[];
  action: Verb;
};

type CreateFormProps = {
  onFinish: () => void;
  params: CreateFormParams;
};

type ExpenseForm = {
  id: number | undefined;
  date: Date | undefined;
  amount: number;
  category: string;
  notes: string;
};

export default function CreateForm({ onFinish, params }: CreateFormProps) {
  const t = useTranslation().trans;
  const currency = useNumber().currency;
  const [dialogOpen, setDialogOpen] = useState(false);
  const { headers } = useHeader();
  const { errors: propsErrors } = usePage<PageProps>().props;
  const { data, setData, post, put, errors, reset, processing } = useForm<Required<ExpenseForm>>({
    id: params.expense?.id,
    date: params.expense?.date || new Date(),
    amount: params.expense?.amount || 0,
    category: params.expense?.category?.uuid || '',
    notes: params.expense?.notes || '',
  });

  const viewMode = params.action === 'view';
  const isDisabled = params.expense?.deleted_at !== undefined;
  const verbName = useVerb().action(params.action);

  const options = {
    ...headers,
    preserveScroll: true,
    onSuccess: () => {
      reset();
      onFinish();
    },
  };

  const submit = () => {
    if (params.action === 'create') post('/expenses', options);
    if (params.action === 'edit' && params.expense) put(`/expenses/${params.expense.uuid}`, options);
  };

  const handleDateChange = (date: unknown) => {
    setData('date', date as Date);
  };

  return (
    <div className="flex flex-col space-y-6">
      <FormSection onSubmit={submit}>
        <FormSection.Title>{t('expenses.single.title')}</FormSection.Title>
        <FormSection.Description>{t('expenses.single.description')}</FormSection.Description>
        <FormSection.Form>
          {propsErrors.status && <div className="col-span-6 mb-4 text-center text-sm font-medium text-red-600">{propsErrors.status}</div>}
          <div className="col-span-12">
            <div className="grid grid-cols-12 gap-x-2">
              <div className="col-span-4">
                <Label>{t('global.category')}</Label>
                <Select name="category" onValueChange={(value) => setData('category', value)} value={data.category} required disabled={viewMode}>
                  <SelectTrigger className="mt-2 w-full">
                    <SelectValue placeholder={t('global.category')} />
                  </SelectTrigger>
                  <SelectContent className="">
                    {params.categories.map((category) => (
                      <SelectItem key={category.id} value={category.uuid}>
                        {category.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <InputError className="mt-2" message={errors.category} />
              </div>
              <div className="col-span-4">
                <Label htmlFor="amount">{t('global.amount')}</Label>
                <Input
                  id="amount"
                  type="number"
                  className="mt-2 block w-40 text-end"
                  value={data.amount}
                  onChange={(e) => setData('amount', e.target.valueAsNumber)}
                  required
                  readOnly={viewMode}
                />
                <InputError className="mt-2" message={errors.amount} />
              </div>
              <div className="col-span-4">
                <DatePickerField
                  id="date"
                  label={t('global.date')}
                  value={data.date}
                  placeholder={t('global.datePlaceholder')}
                  onChange={handleDateChange}
                  disabled={viewMode}
                  error={errors.date}
                />
              </div>
            </div>
          </div>
          <div className="col-span-12">
            <div className="grid grid-cols-12">
              <div className="col-span-12 flex flex-col gap-y-2 py-2">
                <Label className="text-sm/6 font-medium">{t('global.notes')}</Label>
                <Textarea
                  name="notes"
                  rows={4}
                  className="focus:no-data-focus:outline-none block resize-none rounded-lg border px-3 py-1.5 text-sm/6 data-focus:outline-2 data-focus:-outline-offset-2 data-focus:outline-white/25"
                  defaultValue={data.notes}
                  onChange={(e) => setData('notes', e.target.value)}
                  readOnly={viewMode}
                />
              </div>
            </div>
          </div>
        </FormSection.Form>
        {!viewMode && (
          <FormSection.Actions>
            <Button disabled={processing} className="uppercase">
              {processing ? (
                <>
                  <Spinner />
                  {t('global.saving')}
                </>
              ) : (
                <>
                  {t(`global.actions.${verbName}`)} {t('global.expense')}
                </>
              )}
            </Button>
          </FormSection.Actions>
        )}
      </FormSection>

      {viewMode && (
        <ActionSection>
          <ActionSection.Title>{t(`expenses.${isDisabled ? 'unarchive' : 'archive'}.title`)}</ActionSection.Title>
          <ActionSection.Description>{t(`expenses.${isDisabled ? 'unarchive' : 'archive'}.description`)}</ActionSection.Description>
          <ActionSection.Content>
            <div className={`space-y-4 rounded-lg border ${isDisabled ? 'border-primary-100 bg-primary-50' : 'border-red-100 bg-red-50'} p-4`}>
              <div className={`relative space-y-0.5 ${isDisabled ? 'text-primary' : 'text-red-600'}`}>
                <p className="font-medium">{t('global.warning.title')}</p>
                <p className="text-sm">{t('global.warning.description')}</p>
              </div>
              <Button variant={isDisabled ? 'default' : 'destructive'} onClick={() => setDialogOpen(true)}>
                {t(`expenses.${isDisabled ? 'unarchive' : 'archive'}.title`)}
              </Button>

              <ConfirmsPassword
                title={t(`expenses.confirmsPassword.title`, {
                  expense: params.expense?.category.name || '',
                })}
                description={t(`expenses.confirmsPassword.description`, { total: currency(params.expense?.amount || 0) })}
                action={t(`expenses.confirmsPassword.confirm`)}
                verb={params.expense?.deleted_at === undefined ? 'destroy' : 'update'}
                path={`/expenses/${params.expense?.uuid}`}
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
