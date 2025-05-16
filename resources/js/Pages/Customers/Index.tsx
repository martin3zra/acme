import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import { useTranslation } from '@/hooks/use-translation';
import AuthenticatedLayout from '@/layouts/authenticated-layout';
import { BreadcrumbItem, Customer, CustomerVerb, PageProps } from '@/types';
import { router } from '@inertiajs/react';
import { Plus } from 'lucide-react';
import { useEffect, useState } from 'react';
import { List } from './List/Index';
import CreateForm, { CreateFormParams } from './Shared/CreateForm';

const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Home',
    href: '/home',
  },
  {
    title: 'Customers',
    href: '/customers',
  },
];

export default function Index({ auth, customers, customer }: PageProps<{ customers: Customer[]; customer: Customer }>) {
  const t = useTranslation().trans;
  const [open, setOpen] = useState(customer !== undefined);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState<CreateFormParams>({
    customer: customer,
    action: customer !== undefined ? 'view' : 'create',
  });

  const verbName = useVerb().action(selectedCustomer.action);
  const hasCustomers = customers.length > 0;

  const onCreateNewCustomer = () => {
    setSelectedCustomer({ customer: undefined, action: 'create' });
    setOpen(!open);
  };

  const onSelectCustomer = (customer: Customer, action: CustomerVerb): void => {
    if (action === 'record-payment') {
      router.visit(`/payments/create`, { data: { customer_id: customer.uuid } });
      return;
    }
    setSelectedCustomer({ customer, action });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) setSelectedCustomer({ customer: undefined, action: 'create' });
  };

  useEffect(() => {
    if (selectedCustomer && selectedCustomer.customer !== undefined) {
      if (selectedCustomer.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedCustomer]);

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };

  return (
    <AuthenticatedLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasCustomers && (
          <HeadingSmall
            title={t('customers.title')}
            description={t('customers.description')}
            rightPanel={
              <Button onClick={onCreateNewCustomer}>
                <Plus /> {t('customers.newCustomer.title')}
              </Button>
            }
          />
        )}

        {!hasCustomers && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t('customers.emptyState.title')}</h4>
              <p className="text-sm text-gray-400">{t('customers.emptyState.description')}</p>
              <Button onClick={onCreateNewCustomer}>
                <Plus /> {t('customers.newCustomer.title')}
              </Button>
            </div>
          </>
        )}

        {hasCustomers && <List data={customers} onSelectCustomer={onSelectCustomer} />}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl">
            <SheetHeader>
              <SheetTitle>
                {t(`global.actions.${verbName}`)} {t(`global.customer`).toLocaleLowerCase()}
              </SheetTitle>
              <SheetDescription className="text-[12px]">{t(`customers.newCustomer.description`)}</SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 px-4">
              <CreateForm params={selectedCustomer} onFinish={() => modalHandler(false)} />
            </div>
          </SheetContent>
        </Sheet>

        {selectedCustomer.customer && (
          <ConfirmsPassword
            title={t(`customers.confirmsPassword.title`, { customer: selectedCustomer?.customer?.name })}
            description={t(`customers.confirmsPassword.description`, { customer: selectedCustomer?.customer?.name })}
            action={t(`customers.confirmsPassword.confirm`)}
            verb={'destroy'}
            path={`/customers/${selectedCustomer?.customer?.id}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AuthenticatedLayout>
  );
}
