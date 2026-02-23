import { ConfirmsPassword } from '@/components/confirms-password';
import HeadingSmall from '@/components/heading-small';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import { useVerb } from '@/composables/use-verbs';
import useCallbackState from '@/hooks/use-callback-state';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { Expense, ExpenseCategory, PageProps, Verb } from '@/types';
import { router, usePage } from '@inertiajs/react';
import { useEffect, useState } from 'react';
import { breadcrumbs } from './constants';
import { List } from './List/Index';
import { AddNewExpense } from './Shared/add-new-expense';
import CreateForm, { CreateFormParams } from './Shared/create-form';

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
  const page = usePage<PageProps>();
  const [loadingExpense, setLoadingExpense] = useCallbackState<boolean>(false);
  const [open, setOpen] = useCallbackState<boolean>(openState);
  const [deleteDialogOpen, setDeleteDialogOpen] = useCallbackState<boolean>(false);
  const hasExpenses = expenses.length > 0;
  const [selectedExpense, setSelectedExpense] = useState<CreateFormParams>({
    expense: expense,
    action: expense !== undefined ? 'view' : 'create',
    categories: categories,
  });

  const verbName = useVerb().action(selectedExpense.action);

  useEffect(() => {
    if (expense === undefined) return;
    setSelectedExpense((val) => ({ ...val, expense }));
  }, [expense, setSelectedExpense]);

  const onCreateNewExpense = () => {
    setSelectedExpense({ expense: undefined, action: 'create', categories });
    setOpen(true);
  };

  const onSelectExpense = (expense: Expense, action: Verb): void => {
    setSelectedExpense({ expense, action, categories });
    if (action === 'trash') {
      setDeleteDialogOpen(true);
      return;
    }

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
      setSelectedExpense({ expense: undefined, action: 'create', categories });
      // Remove query string from URL
      router.replace({
        url: window.location.pathname,
        preserveScroll: true,
        preserveState: true,
      });
    }
  };

  useEffect(() => {
    if (selectedExpense && selectedExpense.expense !== undefined) {
      if (selectedExpense.action !== 'trash') {
        setOpen(true);
      } else {
        setDeleteDialogOpen(true);
      }
    }
  }, [selectedExpense, setOpen]);

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
            <div className="absolute top-1/2 left-1/2 flex h-61 min-w-3xl -translate-x-1/2 -translate-y-1/2 flex-col items-center gap-4 rounded-[16px] bg-white p-[40px] shadow-[0px_8px_12px_-4px_rgba(16,12,12,0.08),0px_0px_2px_rgba(16,12,12,0.1),0px_1px_2px_rgba(16,12,12,0.1)]">
              <h4 className="text-2xl">{t('expenses.emptyState.title')}</h4>
              <p className="text-sm text-gray-400">{t('expenses.emptyState.description')}</p>
              <AddNewExpense onCreateNewExpense={onCreateNewExpense} />
            </div>
          </>
        )}

        {hasExpenses && <List data={expenses} onSelectExpense={onSelectExpense} />}

        {!loadingExpense && (
          <Sheet open={open} onOpenChange={onOpenChange}>
            <SheetContent side="right" className="m-4 flex h-[calc(~'(100%_-_var(--spacing)_*_4)_/_3')] w-full flex-col rounded-md sm:max-w-[1380px]">
              <SheetHeader>
                <SheetTitle>
                  {t(`global.actions.${verbName}`)} {t(`global.expense`).toLocaleLowerCase()}
                </SheetTitle>
                <SheetDescription className="text-[12px]">{t('expenses.viewExpense.description')}</SheetDescription>
              </SheetHeader>
              <div className="relative grid gap-4 overflow-y-scroll px-4 pb-4">
                <CreateForm params={selectedExpense} onFinish={() => modalHandler(false)} />
              </div>
            </SheetContent>
          </Sheet>
        )}

        {selectedExpense.expense !== undefined && (
          <ConfirmsPassword
            title={t('expenses.confirmsPassword.title', { expense: selectedExpense.expense.category.name })}
            description={t('expenses.confirmsPassword.description', { total: selectedExpense.expense.amount })}
            action={t('expenses.confirmsPassword.confirm')}
            verb={'destroy'}
            path={`/expenses/${selectedExpense.expense.uuid}`}
            open={deleteDialogOpen}
            onOpenChange={modalHandler}
          />
        )}
      </div>
    </AppLayout>
  );
}
