import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import useCallbackState from '@/hooks/use-callback-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { Expense, ExpenseCategory, PageProps, PaymentVerb } from '@/types';
import { router, usePage } from '@inertiajs/react';
import { breadcrumbs } from './constants';
import { List } from './List/Index';
import { AddNewExpense } from './Shared/add-new-expense';
import CreateForm from './Shared/create-form';

export default function Index({
  auth,
  expenses,
  categories,
  expense,
  openState,
}: PageProps<{
  openState: boolean;
  categories: ExpenseCategory[];
  expenses: Expense[];
  expense: Expense;
}>) {
  const t = useTranslation().trans;
  const [loadingExpense, setLoadingExpense] = useCallbackState<boolean>(false);
  const [open, setOpen] = useCallbackState<boolean>(openState);
  const [selectedExpense, setSelectedExpense] = useCallbackState<Expense | undefined>(undefined);
  const [deleteDialogOpen, setDeleteDialogOpen] = useCallbackState<boolean>(false);
  const page = usePage<PageProps>();
  const hasExpenses = expenses.length > 0;

  const onCreateNewExpense = () => {
    // setSelectedCustomer({ customer: undefined, action: 'create', tax_receipts });
    setOpen(true);
  };

  const onSelectExpense = (expense: Expense, action: PaymentVerb): void => {
    setSelectedExpense(expense);
    if (action === 'void') {
      setDeleteDialogOpen(true);
      return;
    }
    if (action === 'edit') {
      router.visit(`/expenses/${expense.uuid}/edit`);
      return;
    }
    if (action !== 'view') return;

    setOpen(
      (open) => !open,
      (newVal) => {
        if (newVal) findSelectedExpense(expense.uuid);
      },
    );
  };

  const findSelectedExpense = (uuid: string) => {
    router.visit(page.url, {
      except: ['expenses'],
      data: { id: uuid },
      preserveScroll: true,
      preserveState: true,
      onStart: () => setLoadingExpense(true),
      onFinish: () => setLoadingExpense(false),
    });
  };

  const onOpenChange = (open: boolean) => {
    setOpen(open);
    if (!open) {
      // Remove query string from URL
      router.replace({
        url: window.location.pathname,
        preserveScroll: true,
        preserveState: true,
      });
    }
  };

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    setDeleteDialogOpen(open);
  };
  return (
    <AppLayout user={auth.user} breadcrumbs={breadcrumbs}>
      <div className="space-y-6">
        {hasExpenses && (
          <HeadingSmall
            title={t('expenses.title')}
            description={t('expenses.description')}
            rightPanel={<AddNewExpense onCreateNewExpense={onCreateNewExpense} />}
          />
        )}

        {!hasExpenses && (
          <>
            <div className="absolute top-1/2 left-1/2 flex h-[244px] min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t('expenses.emptyState.title')}</h4>
              <p className="text-sm text-gray-400">{t('expenses.emptyState.description')}</p>
              <AddNewExpense onCreateNewExpense={onCreateNewExpense} />
            </div>
          </>
        )}

        {hasExpenses && <List data={expenses} onSelectExpense={onSelectExpense} />}

        {!loadingExpense && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-[1380px]">
              <SheetHeader>
                <div className="mr-6 flex items-start justify-between">
                  <div className="flex flex-col">
                    <SheetTitle>{page.props.auth.company.name}</SheetTitle>
                    <SheetDescription className="text-[12px]">{t('expenses.viewExpense.description')}</SheetDescription>
                  </div>
                  {/* <div className="mx-4 flex gap-x-3">
                    {expense.deleted_at !== null && (
                      <>
                        <Button variant={'destructive'} onClick={() => onSelectExpense(expense, 'void')}>
                          <Ban /> {t('global.actions.void')}
                        </Button>
                        <Separator orientation="vertical" />
                        <Button asChild disabled={expense.deleted_at === null}>
                          <Link href={`/expenses/${expense.uuid}/edit`} as="button">
                            <NotebookPen /> {t('global.actions.edit')}
                          </Link>
                        </Button>
                      </>
                    )}

                    <a
                      href={expense.uuid}
                      className="bg-primary flex items-center gap-x-3 rounded-sm px-4 text-white"
                      target="_blank"
                      rel="noreferrer"
                    >
                      <Printer /> {t('global.actions.print')}
                    </a>
                  </div> */}
                </div>
              </SheetHeader>
              <div className="relative grid gap-4 overflow-y-scroll px-4 pb-4">
                {/* {expense.deleted_at === null && (
                  <div className="absolute inset-0 flex w-full items-center justify-center overflow-y-hidden bg-transparent">
                    <h1 className="-rotate-45 border-8 border-red-500/25 p-8 text-8xl font-extrabold text-red-500/25">VOID</h1>
                  </div>
                )} */}
                <CreateForm params={{ expense: undefined, categories: categories, action: 'create' }} onFinish={() => {}} />
              </div>
            </SheetContent>
          </Sheet>
        )}

        {selectedExpense && (
          <ConfirmsPassword
            title={t('expenses.confirmsPassword.title', { expense: selectedExpense.id })}
            description={t('expenses.confirmsPassword.description', { total: selectedExpense.amount })}
            action={t('expenses.confirmsPassword.confirm')}
            verb={'update'}
            path={`/expenses/${selectedExpense.uuid}/void`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AppLayout>
  );
}
