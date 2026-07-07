import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useTranslation } from '@/hooks/use-translation';
import { Company } from '@/types';
import { useEffect, useState } from 'react';
import ExpenseCategoryList from '../Shared/expense-categories';
import Overview from '../Shared/overview';
import { RedirectPreferences } from '../Shared/redirect-preferences';
import TaxList from '../Shared/tax-list';
import { TaxReceipts } from '../Shared/tax-receipts';
import UnitList from '../Shared/units';
import { VariantsSetting } from '../Shared/variants-setting';
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
      orientation="vertical"
      defaultValue={tab}
      onValueChange={setTab}
      className="flex min-h-0 flex-1 flex-row gap-6 [&_[data-slot=tabs-trigger]:not([data-state=active])]:cursor-pointer"
    >
      <TabsList className="h-auto w-48 shrink-0 flex-col items-stretch justify-start gap-1 self-start bg-transparent p-0 [&_[data-slot=tabs-trigger]]:w-full [&_[data-slot=tabs-trigger]]:justify-start">
        <TabsTrigger value="overview">{t('profile.companies.viewCompany.overview.title')}</TabsTrigger>
        <TabsTrigger value="sequences">{t('profile.companies.viewCompany.sequences')}</TabsTrigger>
        <TabsTrigger value="taxes">{t('profile.companies.viewCompany.taxes.title')}</TabsTrigger>
        <TabsTrigger value="taxSequences">{t('profile.companies.viewCompany.taxSequences.title')}</TabsTrigger>
        <TabsTrigger value="expenseCategories">{t('profile.companies.viewCompany.expenseCategories.title')}</TabsTrigger>
        <TabsTrigger value="units">{t('profile.companies.viewCompany.units.title')}</TabsTrigger>
        <TabsTrigger value="redirectPreferences">{t('profile.companies.viewCompany.redirectPreferences.title')}</TabsTrigger>
        <TabsTrigger value="variants">{t('profile.companies.viewCompany.variants.title')}</TabsTrigger>
      </TabsList>
      <div className="min-h-0 flex-1 overflow-y-auto [&_[data-form-aside]]:hidden [&_[data-form-section]]:md:grid-cols-1">
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
        <TabsContent value="expenseCategories" className="min-h-0 flex-1 overflow-y-auto py-0">
          <ExpenseCategoryList categories={company.expense_categories} />
        </TabsContent>
        <TabsContent value="units" className="min-h-0 flex-1 overflow-y-auto py-0">
          <UnitList units={company.units} />
        </TabsContent>
        <TabsContent value="redirectPreferences">
          <RedirectPreferences uuid={company.uuid} preferences={company.redirect_preferences} />
        </TabsContent>
        <TabsContent value="variants">
          <VariantsSetting uuid={company.uuid} enabled={company.handles_variants} />
        </TabsContent>
      </div>
    </Tabs>
  );
}
