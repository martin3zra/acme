import InputError from '@/components/input-error';
import InputSearchable from '@/components/input-searchable';
import { cn } from '@/lib/utils';
import { Customer } from '@/types';
import { User, UserPlus, XCircleIcon } from 'lucide-react';
import React, { JSX } from 'react';

type CustomerSectionProps = {
  customer: Customer | undefined;
  customers: Customer[];
  errors: Partial<Record<'customer_id' | 'terms' | 'discount' | 'lines' | 'date', string>>;
  handleCustomerSelection: (customer: Customer | undefined) => void;
  setSearch: React.Dispatch<React.SetStateAction<string>>;
  setOpen: React.Dispatch<React.SetStateAction<boolean>>;
  open: boolean;
  dedbouncedSearch: string;
};

export const CustomerSection = ({
  customer,
  customers,
  errors,
  handleCustomerSelection,
  setSearch,
  open,
  setOpen,
  dedbouncedSearch,
}: CustomerSectionProps) => {
  const handleOnCloseClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    handleCustomerSelection(undefined);
    setOpen(false);
  };

  const EmptyCard = (): JSX.Element => {
    return (
      <div
        data-slot={`${errors?.customer_id ? 'customer-error' : 'default'}`}
        className={cn(
          'flex h-full w-full flex-col items-center justify-center px-2 pb-2 [&_svg]:text-white',
          'data-[slot=customer-error]:rounded-lg data-[slot=customer-error]:border data-[slot=customer-error]:bg-red-100/50',
          'data-[slot=customer-error]:border-red-500 data-[slot=customer-error]:[&_[data-label=true]]:text-red-500 data-[slot=customer-error]:[&_svg]:text-red-500',
        )}
      >
        <button onClick={() => setOpen(true)} className="flex h-full w-full cursor-pointer items-center justify-center gap-2">
          <div className="flex size-10 items-center justify-center rounded-full bg-gray-200">
            <User className="size-6 *:data-[slot=customer-error]:text-red-500" />
          </div>
          <div data-label="true" className="text-lg">
            Customer
          </div>
        </button>
        <InputError className="mt-2" message={errors.customer_id} />
      </div>
    );
  };

  const CustomerCard = (): JSX.Element => {
    return (
      <div className="flex h-full flex-col overflow-y-hidden p-2">
        <div className="flex w-full items-center justify-between">
          <div>{customer?.name}</div>
          <button onClick={handleOnCloseClick} className="cursor-pointer p-1">
            <XCircleIcon />
          </button>
        </div>
        <div>{customer?.email}</div>
        <div>{customer?.phone}</div>
        <div>{customer?.address}</div>
      </div>
    );
  };

  return (
    <div className="rounded-lg bg-white shadow">
      <InputSearchable
        open={open}
        title="customer"
        valueKey={'id'}
        renderText={(customer: Customer) => `${customer.name}`}
        renderEmptyText="No customers found"
        onSelected={handleCustomerSelection}
        onChange={(value) => setSearch(value)}
        onOpenChange={setOpen}
        value={customer}
        options={customers}
        search={dedbouncedSearch}
      >
        <InputSearchable.Trigger>
          <div className="h-full">
            {!open && !customer && <EmptyCard />}
            {customer && <CustomerCard />}
          </div>
        </InputSearchable.Trigger>
        <InputSearchable.Actions>
          <div className="flex w-full items-center justify-center rounded-b-lg border bg-gray-100/25 py-2">
            <button className="flex cursor-pointer items-center justify-center gap-x-2 text-indigo-400" onClick={() => alert('Create new customer')}>
              <UserPlus className="size-4" /> Add New Customer
            </button>
          </div>
        </InputSearchable.Actions>
      </InputSearchable>
    </div>
  );
};
