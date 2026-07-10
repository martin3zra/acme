import CreateCompanyForm, { CreateFormParams } from '@/components/create-company-form';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import SettingsLayout from '@/layouts/settings/layout';
import { BreadcrumbItem, Company, PageProps, Verb } from '@/types';
import { Head, usePage } from '@inertiajs/react';
import { useEffect, useState } from 'react';
import { CompanyList } from './List/Index';
import Show from './Show';
const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Companies Settings',
    href: '',
  },
];
export default function Index({ companies, company }: PageProps<{ companies: Company[]; company: Company }>) {
  const { auth } = usePage<PageProps>().props;
  const t = useTranslation().trans;
  const [open, setOpen] = useState(company !== undefined);
  // No delete dialog is rendered yet; only the setter is wired up.
  const [, setDeleteDialogOpen] = useState(false);
  const [selectedCompany, setSelectedCompany] = useState<CreateFormParams>({
    company: company,
    action: company !== undefined ? 'view' : 'create',
  });
  const verbName = useVerb().action(selectedCompany.action);
  const hasCompanies = companies.length > 0;

  const onSelectCompany = (company: Company, action: Verb): void => {
    if (action === 'trash') {
      return;
    }

    setSelectedCompany({ company, action });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) setSelectedCompany({ company: undefined, action: 'create' });
  };

  useEffect(() => {
    if (selectedCompany && selectedCompany.company !== undefined) {
      if (selectedCompany.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedCompany]);

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AppLayout breadcrumbs={breadcrumbs} user={auth.user}>
      <Head title="Companies Settings" />
      <SettingsLayout>
        {hasCompanies && <CompanyList data={companies} onSelectCompany={onSelectCompany} />}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-[1600px]">
            <SheetHeader>
              <SheetTitle>
                {t(`global.actions.${verbName}`)} {t(`global.company`).toLocaleLowerCase()}
              </SheetTitle>
              <SheetDescription className="text-[12px]">{t(`companies.newCompany.description`)}</SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 px-4">
              {selectedCompany.action === 'view' && selectedCompany.company !== undefined && <Show company={selectedCompany.company} />}
              {selectedCompany.action === 'create' ||
                (selectedCompany.action === 'edit' && <CreateCompanyForm params={selectedCompany} onFinish={modalHandler} />)}
            </div>
          </SheetContent>
        </Sheet>
      </SettingsLayout>
    </AppLayout>
  );
}
