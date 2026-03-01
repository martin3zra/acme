import FormSection from '@/components/form-section';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Spinner } from '@/components/ui/spinner';
import { useHeader } from '@/composables/use-headers';
import { useTranslation } from '@/hooks/use-translation';
import { PageProps, RedirectPreference, RedirectPreferenceValue } from '@/types';
import { useForm, usePage } from '@inertiajs/react';

type Props = {
  uuid: string;
  preferences: RedirectPreference;
};

export function RedirectPreferences({ uuid, preferences }: Props) {
  const t = useTranslation().trans;
  const { auth } = usePage<PageProps>().props;
  const { headers } = useHeader();
  const { data, setData, put, processing } = useForm({
    invoice: preferences.invoice || 'detail',
    estimate: preferences.estimate || 'list',
    customer: preferences.customer || 'list',
    order: preferences.order || 'list',
    payment: preferences.payment || 'list',
    item: preferences.item || 'list',
  });

  const handleSubmit = () => {
    put(`/settings/${auth.account.uuid}/companies/${uuid}/redirect-preferences`, { ...headers, preserveState: 'errors' });
  };

  return (
    <div className="flex w-full flex-col space-y-6 py-6 **:data-form:md:grid-cols-1">
      <FormSection onSubmit={handleSubmit}>
        <FormSection.Title>{t('profile.companies.viewCompany.redirectPreferences.title')}</FormSection.Title>
        <FormSection.Description>{t('profile.companies.viewCompany.redirectPreferences.description')}</FormSection.Description>
        <FormSection.Form>
          <div className="flex justify-between space-x-2">
            <div className="basis-1/2 space-y-2">
              <Label>{t('profile.companies.viewCompany.redirectPreferences.afterCreatingInvoice')}</Label>
              <Select value={data.invoice} onValueChange={(val: RedirectPreferenceValue) => setData('invoice', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToList', { resource: t('global.invoice') })}
                  </SelectItem>
                  <SelectItem value="detail">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToDetail', { resource: t('global.invoice') })}
                  </SelectItem>
                  <SelectItem value="stay">{t('profile.companies.viewCompany.redirectPreferences.stayOnForm')}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="basis-1/2 space-y-2">
              <Label>{t('profile.companies.viewCompany.redirectPreferences.afterCreatingEstimate')}</Label>
              <Select value={data.estimate} onValueChange={(val: RedirectPreferenceValue) => setData('estimate', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToList', { resource: t('global.estimate') })}
                  </SelectItem>
                  <SelectItem value="detail">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToDetail', { resource: t('global.estimate') })}
                  </SelectItem>
                  <SelectItem value="stay">{t('profile.companies.viewCompany.redirectPreferences.stayOnForm')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex justify-between space-x-2">
            <div className="basis-1/2 space-y-2">
              <Label>{t('profile.companies.viewCompany.redirectPreferences.afterCreatingCustomer')}</Label>
              <Select value={data.customer} onValueChange={(val: RedirectPreferenceValue) => setData('customer', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToList', { resource: t('global.customer') })}
                  </SelectItem>
                  <SelectItem value="detail">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToDetail', { resource: t('global.customer') })}
                  </SelectItem>
                  <SelectItem value="stay">{t('profile.companies.viewCompany.redirectPreferences.stayOnForm')}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="basis-1/2 space-y-2">
              <Label>{t('profile.companies.viewCompany.redirectPreferences.afterCreatingOrder')}</Label>
              <Select value={data.order} onValueChange={(val: RedirectPreferenceValue) => setData('order', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToList', { resource: t('global.order') })}
                  </SelectItem>
                  <SelectItem value="detail">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToDetail', { resource: t('global.order') })}
                  </SelectItem>
                  <SelectItem value="stay">{t('profile.companies.viewCompany.redirectPreferences.stayOnForm')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex justify-between space-x-2">
            <div className="basis-1/2 space-y-2">
              <Label>{t('profile.companies.viewCompany.redirectPreferences.afterCreatingItem')}</Label>
              <Select value={data.item} onValueChange={(val: RedirectPreferenceValue) => setData('item', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToList', { resource: t('global.item') })}
                  </SelectItem>
                  <SelectItem value="detail">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToDetail', { resource: t('global.item') })}
                  </SelectItem>
                  <SelectItem value="stay">{t('profile.companies.viewCompany.redirectPreferences.stayOnForm')}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="basis-1/2 space-y-2">
              <Label>{t('profile.companies.viewCompany.redirectPreferences.afterCreatingPayment')}</Label>
              <Select value={data.payment} onValueChange={(val: RedirectPreferenceValue) => setData('payment', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToList', { resource: t('global.payment') })}
                  </SelectItem>
                  <SelectItem value="detail">
                    {t('profile.companies.viewCompany.redirectPreferences.redirectToDetail', { resource: t('global.payment') })}
                  </SelectItem>
                  <SelectItem value="stay">{t('profile.companies.viewCompany.redirectPreferences.stayOnForm')}</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </FormSection.Form>
        <FormSection.Actions>
          <Button type="submit" disabled={processing}>
            {processing ? (
              <>
                <Spinner />
                {t('global.saving')}
              </>
            ) : (
              t('global.save')
            )}
          </Button>
        </FormSection.Actions>
      </FormSection>
    </div>
  );
}
