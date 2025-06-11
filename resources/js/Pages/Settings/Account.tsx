import CreateCompanyForm, { CreateFormParams } from '@/components/create-company-form';
import { StatusBadge } from '@/components/status-badge';
import { Avatar, AvatarFallback } from '@/components/ui/avatar';
import { Button } from '@/components/ui/button';
import { Separator } from '@/components/ui/separator';
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle } from '@/components/ui/sheet';
import useCallbackState from '@/hooks/use-callback-state';
import { useInitials } from '@/hooks/use-initials';
import { useTranslation } from '@/hooks/use-translation';
import AppLayout from '@/layouts/app-layout';
import { BreadcrumbItem, Company, PageProps, Role, User, UserVerb, Verb } from '@/types';
import { router, usePage } from '@inertiajs/react';
import { format } from 'date-fns';
import { BadgeCheck } from 'lucide-react';
import { useState } from 'react';
import { CompanyList } from './Companies/List/Index';
import Show from './Companies/Show';
import AccountForm from './Shared/account-form';
import { UserList } from './Users/List/Index';
import UserForm, { UserFormParams } from './Users/UserForm';

export const breadcrumbs: BreadcrumbItem[] = [
  {
    title: 'Account Settings',
    href: '',
  },
];

type SheetContentType = 'profile' | 'company:view' | 'company:form' | 'user:view' | 'user:form';

type State = {
  sheetState: boolean;
  sheetContent: SheetContentType;
};

export default function Account({
  auth,
  companies,
  company,
  users,
  currentUser,
  roles,
  initialState = false,
  subject = 'profile',
}: PageProps<{
  companies: Company[];
  company: Company;
  users: User[];
  currentUser: User;
  roles: Role[];
  initialState: boolean;
  subject: SheetContentType;
}>) {
  const t = useTranslation().trans;
  const getInitials = useInitials();
  const page = usePage<PageProps>();
  const [state, setState] = useCallbackState<State>({ sheetState: initialState, sheetContent: subject });
  const [selectedCompany, setSelectedCompany] = useState<CreateFormParams>({
    company: company,
    action: company !== undefined ? 'view' : 'create',
  });
  const [selectedUser, setSelectedUser] = useState<UserFormParams>({
    user: currentUser,
    action: currentUser !== undefined ? 'view' : 'create',
  });
  const hasCompanies = companies.length > 0;
  const hasUsers = users.length > 0;

  const onSelectCompany = (company: Company, action: Verb): void => {
    if (action === 'trash') {
      return;
    }

    setSelectedCompany({ company, action });

    setState({ sheetState: true, sheetContent: action === 'view' ? 'company:view' : 'company:form' });
  };

  const onSelectUser = (user: User, action: UserVerb): void => {
    if (action === 'trash') {
      return;
    }

    if (action !== 'view') return;

    setState(
      (current) => ({ sheetState: !current.sheetState, sheetContent: action === 'view' ? 'user:view' : 'user:form' }),
      (newVal) => {
        if (newVal.sheetState) findSelectedUser(user.uuid);
      },
    );
  };

  const findSelectedUser = (uuid: string) => {
    router.visit(page.url, {
      except: ['companies', 'users'],
      data: { user_id: uuid },
      preserveScroll: true,
      preserveState: 'errors',
    });
  };

  const onEditProfile = () => {
    setState({ sheetState: true, sheetContent: 'profile' });
  };

  const onOpenChange = (open: boolean) => {
    setSelectedUser({ user: undefined, action: 'create' });
    setState({ sheetState: open, sheetContent: 'profile' });

    if (open) return;
    // Remove query string from URL
    router.replace({
      url: window.location.pathname,
      preserveScroll: true,
      preserveState: true,
    });
  };

  const modalHandler = (open: boolean = false) => {
    onOpenChange(open);
    // setDeleteDialogOpen(open);
  };

  const onAddNewUser = () => {
    setState({ sheetState: true, sheetContent: 'user:form' });
  };
  return (
    <AppLayout breadcrumbs={breadcrumbs} user={auth.user}>
      <div className="flex">
        <div className="flex basis-[30vw] flex-col gap-y-6 py-6">
          <div className="flex items-end gap-6">
            <div className="relative flex size-22 items-center">
              <Avatar className="bg-muted flex h-22 w-22 items-center justify-center rounded-full">
                <AvatarFallback className="rounded-lg text-4xl">{getInitials(auth.user.name)}</AvatarFallback>
              </Avatar>

              {auth.user.email_verified_at !== null && <BadgeCheck size={22} className="absolute right-0 bottom-2" />}
            </div>
            <div className="mb-2">
              <div className="flex items-end gap-2">
                <h1 className="text-2xl">{auth.user.name}</h1>
                <StatusBadge type="status" status={auth.user.status} />
              </div>
              <h4 className="text-foreground text-sm">{auth.user.email}</h4>
            </div>
          </div>
          <div>
            <h4>Member since</h4>
            {format(auth.user.created_at, 'PPP')}
          </div>
          <div className="space-y-6">
            <Button onClick={onEditProfile}>Edit Profile</Button>
          </div>
        </div>
        <div className="basis-[70vw] space-y-6">
          {hasCompanies && <CompanyList data={companies} onSelectCompany={onSelectCompany} />}
          {hasUsers && (
            <>
              <Separator /> <UserList data={users} onSelectUser={onSelectUser} onAddNewUser={onAddNewUser} />
            </>
          )}
        </div>
      </div>

      <Sheet open={state.sheetState} onOpenChange={onOpenChange}>
        <SheetContent side="right" className="m-4 flex h-[calc(~'(100%-var(--spacing)*4)/3')] w-full flex-col rounded-md sm:max-w-[1380px]">
          <SheetHeader>
            <SheetTitle>{t(`global.profile`)}</SheetTitle>
            <SheetDescription className="text-[12px]">Manage your subscription and billing details</SheetDescription>
          </SheetHeader>
          <div className="grid gap-4 overflow-y-scroll px-4">
            {state.sheetContent === 'company:view' && selectedCompany.company !== undefined && <Show company={selectedCompany.company} />}
            {state.sheetContent === 'company:form' && selectedCompany.company !== undefined && (
              <CreateCompanyForm params={selectedCompany} onFinish={modalHandler} />
            )}
            {['user:view', 'user:form'].includes(state.sheetContent) && (
              <UserForm params={selectedUser} companies={companies} roles={roles} onFinish={modalHandler} />
            )}
            {state.sheetContent === 'profile' && <AccountForm />}
          </div>
        </SheetContent>
      </Sheet>
    </AppLayout>
  );
}
