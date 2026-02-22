import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useTranslation } from '@/hooks/use-translation';
import { Company } from '@/types';
import { useEffect, useState } from 'react';
import Overview from '../Shared/overview';
import { RedirectPreferences } from '../Shared/redirect-preferences';
import TaxList from '../Shared/tax-list';
import { TaxReceipts } from '../Shared/tax-receipts';
import SequenceView from './Sequences';

export type Props = {
  company: Company;
};

export default function Show({ company }: Props) {
  // Read tab from URL query param
  const t = useTranslation().trans;
  const params = new URLSearchParams(window.location.search);
  const initialTab = params.get('tab') || 'overview';
  const [tab, setTab] = useState(initialTab);
  useEffect(() => {
    // Keep URL in sync when tab changes
    const url = new URL(window.location.href);
    url.searchParams.set('tab', tab);
    window.history.replaceState({}, '', url);
  }, [tab]);

  return (
    <Tabs
      defaultValue={tab}
      onValueChange={setTab}
      className="flex min-h-0 flex-1 flex-col [&_[data-slot=tabs-trigger]:not([data-state=active])]:cursor-pointer"
    >
      <TabsList className="shrink-0">
        <TabsTrigger value="overview">{t('profile.companies.viewCompany.overview.title')}</TabsTrigger>
        <TabsTrigger value="sequences">{t('profile.companies.viewCompany.sequences')}</TabsTrigger>
        <TabsTrigger value="taxes">{t('profile.companies.viewCompany.taxes.title')}</TabsTrigger>
        <TabsTrigger value="taxSequences">{t('profile.companies.viewCompany.taxSequences.title')}</TabsTrigger>
        <TabsTrigger value="redirectPreferences">{t('profile.companies.viewCompany.redirectPreferences.title')}</TabsTrigger>
      </TabsList>
      <TabsContent value="overview">
        <Overview company={company} />
      </TabsContent>
      <TabsContent value="sequences" className="min-h-0 flex-1 overflow-y-auto py-0">
        <SequenceView uuid={company.uuid} sequences={company.sequences} />
      </TabsContent>
      <TabsContent value="taxes">
        <TaxList taxes={company.taxes} />
      </TabsContent>
      <TabsContent value="taxSequences" className="min-h-0 flex-1 overflow-y-auto py-0">
        <TaxReceipts uuid={company.uuid} taxReceipts={company.tax_receipts} />
      </TabsContent>
      <TabsContent value="redirectPreferences">
        <RedirectPreferences uuid={company.uuid} preferences={company.redirect_preferences} />
      </TabsContent>
    </Tabs>
  );
}
