import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Button } from '@/components/ui/button';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
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

export default function Index({ auth, customers }: PageProps<{ customers: Customer[] }>) {
  const [open, setOpen] = useState(false);
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [selectedCustomer, setSelectedCustomer] = useState<CreateFormParams>({ customer: undefined, action: 'create' });

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
            title="Customers"
            description="All created customers are shown here."
            rightPanel={
              <Button onClick={onCreateNewCustomer}>
                <Plus /> Add Customers
              </Button>
            }
          />
        )}

        {!hasCustomers && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">Create your first customer</h4>
              <p className="text-sm text-gray-400">Once you create your customer, it will appear here.</p>
              <Button onClick={onCreateNewCustomer}>+ Create Customer</Button>
            </div>
          </>
        )}

        {hasCustomers && <List data={customers} onSelectCustomer={onSelectCustomer} />}

        <Sheet open={open} onOpenChange={onOpenChange}>
          <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-4xl">
            <SheetHeader>
              <SheetTitle>{verbName} Customer</SheetTitle>
              <SheetDescription className="text-[12px]">Create a new customer</SheetDescription>
            </SheetHeader>
            <div className="grid gap-4 px-4">
              <CreateForm params={selectedCustomer} onFinish={() => modalHandler(false)} />
            </div>
          </SheetContent>
        </Sheet>

        {selectedCustomer.customer && (
          <ConfirmsPassword
            title={`Are you sure you want to delete ${selectedCustomer?.customer?.name}?`}
            description={`Once ${selectedCustomer?.customer?.name} is deleted, all of its resources will continue to be available, but no new operation can be performed.`}
            action={`Delete it`}
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
