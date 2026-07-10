import FormSection from '@/components/form-section';
import InputError from '@/components/input-error';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Spinner } from '@/components/ui/spinner';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import { ExpenseCategory } from '@/types';
import { Textarea } from '@headlessui/react';
import { useForm } from '@inertiajs/react';
import { formatDate } from 'date-fns/format';

type Props = {
  categories: ExpenseCategory[];
};

export default function ExpenseCategoryList({ categories }: Props) {
  const { headers } = useHeader();
  const t = useTranslation().trans;
  const { data, setData, post, put, processing, errors, reset } = useForm<{ uuid?: string; name: string; description: string }>({
    uuid: undefined,
    name: '',
    description: '',
  });
  const handleSubmit = () => {
    if (data.uuid) {
      put(`/expense-categories/${data.uuid}`, { ...headers, preserveState: 'errors' });
      return;
    }
    post('/expense-categories', { ...headers, preserveState: 'errors' });
  };
  const handleOnCategoryClick = (category: ExpenseCategory) => {
    setData({ uuid: category.uuid, name: category.name, description: category.description });
  };
  return (
    <div className="flex w-full flex-col space-y-6 py-6 **:data-form:md:grid-cols-1">
      <FormSection onSubmit={handleSubmit}>
        <FormSection.Title>{t('profile.companies.viewCompany.expenseCategories.title')}</FormSection.Title>
        <FormSection.Description>{t('profile.companies.viewCompany.expenseCategories.description')}</FormSection.Description>
        <FormSection.Form>
          <ul className="divide-muted w-full divide-y rounded-md border">
            <li className="bg-muted/50 grid grid-cols-3 px-4 py-2 text-sm font-medium">
              <span>{t('global.name')}</span>
              <span>{t('global.description')}</span>
              <span className="text-end">{t('global.addedAt')}</span>
            </li>
            {categories.map((category, idx) => (
              <li
                onClick={() => handleOnCategoryClick(category)}
                key={category.id}
                className={cn('hover:bg-muted/30 grid cursor-pointer grid-cols-3 px-4 py-2 text-sm', idx % 2 === 0 && 'bg-muted/10')}
              >
                <span>{category.name}</span>
                <span>{category.description}</span>
                <span className="text-end">{formatDate(new Date(category.created_at), 'dd-MM-yyyy')}</span>
              </li>
            ))}

            <li className="bg-muted/20 px-4 py-3">
              <div className="grid grid-cols-3 items-start gap-2">
                <div className="col-span-2 flex w-full flex-col space-y-3">
                  <div>
                    <Input placeholder={t('global.name')} value={data.name} onChange={(e) => setData('name', e.target.value)} />
                    {errors.name && Array.isArray(errors.name) && (
                      <>
                        {errors.name.map((message, index) => (
                          <InputError key={index} message={message} />
                        ))}
                      </>
                    )}
                  </div>
                  <div className="w-full">
                    <Textarea
                      className="focus:no-data-focus:outline-none block w-full resize-none rounded-lg border px-3 py-1.5 text-sm/6 data-focus:outline-2 data-focus:-outline-offset-2 data-focus:outline-white/25"
                      placeholder={t('global.description')}
                      value={data.description}
                      onChange={(e) => setData('description', e.target.value)}
                    />
                    <InputError message={errors.description} />
                  </div>
                </div>
                <div className="col-span-1 flex gap-3">
                  <div className="flex w-full justify-end space-x-2">
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
                      <Button onClick={() => reset('name', 'description', 'uuid')} type="reset" size="sm" className="mt-1" variant={'ghost'}>
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
