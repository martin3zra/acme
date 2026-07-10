import InputError from '@/components/input-error';
import InputSearchable from '@/components/input-searchable';
import { useTranslation } from '@/hooks/use-translation';
import { cn } from '@/lib/utils';
import type { Vendor } from '@/types';
import { Link } from '@inertiajs/react';
import { Eye, UserPlus, XCircleIcon } from 'lucide-react';
import React, { JSX } from 'react';

type VendorSectionProps = {
  vendor: Vendor | undefined;
  vendors: Vendor[];
  errors: Partial<Record<'vendor_id' | 'terms' | 'discount' | 'lines' | 'date', string>>;
  handleVendorSelection: (vendor: Vendor | undefined) => void;
  setSearch: React.Dispatch<React.SetStateAction<string>>;
  setOpen: React.Dispatch<React.SetStateAction<boolean>>;
  open: boolean;
  debouncedSearch: string;
};

export const VendorSection = ({ vendor, vendors, errors, handleVendorSelection, setSearch, open, setOpen, debouncedSearch }: VendorSectionProps) => {
  const t = useTranslation().trans;

  const handleOnCloseClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    handleVendorSelection(undefined);
    setOpen(false);
  };

  const handleNewVendorClick = (event: React.MouseEvent<HTMLButtonElement>) => {
    event.preventDefault();
    alert('This feature is not implemented yet. Please navigate to the vendors page to create a new vendor.');
  };

  const EmptyCard = (): JSX.Element => {
    return (
      <div
        data-slot={`${errors?.vendor_id ? 'vendor-error' : 'default'}`}
        className={cn(
          'flex h-full w-full flex-col items-center justify-center px-2 pb-2',
          'data-[slot=vendor-error]:rounded-lg data-[slot=vendor-error]:border data-[slot=vendor-error]:bg-red-100/50',
          'data-[slot=vendor-error]:border-red-500 data-[slot=vendor-error]:**:data-[label=true]:text-red-500 data-[slot=vendor-error]:[&_svg]:text-red-500',
        )}
      >
        <button onClick={() => setOpen(true)} className="flex h-full w-full cursor-pointer items-center justify-center gap-2">
          <div className="flex size-10 items-center justify-center rounded-full bg-gray-200" />
          <div data-label="true" className="text-lg">
            {t('global.vendor')}
          </div>
        </button>
        <InputError className="mt-2" message={errors.vendor_id} />
      </div>
    );
  };

  const VendorCard = (): JSX.Element => {
    return (
      <div className="flex h-full flex-col overflow-y-hidden p-2">
        <div className="flex w-full items-center justify-between">
          <div>{vendor?.name}</div>
          <div className="flex items-center gap-x-1.5 p-1">
            <Link href={`/vendors?id=${vendor?.uuid}`}>
              <Eye className="text-muted-foreground size-8 stroke-1" />
            </Link>
            <button onClick={handleOnCloseClick} className="cursor-pointer">
              <XCircleIcon />
            </button>
          </div>
        </div>
        <div>{vendor?.email}</div>
        <div>{vendor?.phone}</div>
        <div>{vendor?.address}</div>
      </div>
    );
  };

  return (
    <div className="rounded-lg bg-white shadow">
      <InputSearchable
        open={open}
        title={t('global.vendorSection.search')}
        valueKey={'id'}
        renderText={(vendor: Vendor) => `${vendor.name}`}
        renderEmptyText={t('global.vendorSection.emptyText')}
        onSelected={handleVendorSelection}
        onChange={(value) => setSearch(value)}
        onOpenChange={setOpen}
        value={vendor}
        options={vendors}
        search={debouncedSearch}
      >
        <InputSearchable.Trigger>
          <div className="h-full">
            {!open && !vendor && <EmptyCard />}
            {vendor && <VendorCard />}
          </div>
        </InputSearchable.Trigger>
        <InputSearchable.Actions>
          <div className="flex w-full items-center justify-center rounded-b-lg border bg-gray-100/25 py-2">
            <button className="flex cursor-pointer items-center justify-center gap-x-2 text-indigo-400" onClick={handleNewVendorClick}>
              <UserPlus className="size-4" /> {t('global.vendorSection.addNew')}
            </button>
          </div>
        </InputSearchable.Actions>
      </InputSearchable>
    </div>
  );
};
