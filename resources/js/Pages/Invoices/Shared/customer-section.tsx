import InputError from "@/components/input-error";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import { Customer } from "@/types"
import { User, UserPlus, XCircleIcon } from "lucide-react";
import React from "react";

type CustomerSectionProps = {
  customer: Customer | undefined
  customers: Customer[]
  errors: Partial<Record<"customer_id" | "terms" | "discount" | "lines" | "date", string>>
  handleCustomerSelection: (event: React.MouseEvent<HTMLButtonElement>, customer: Customer | undefined) => void
  setSearch: React.Dispatch<React.SetStateAction<string>>
  setOpen: React.Dispatch<React.SetStateAction<boolean>>
  open: boolean
  dedbouncedSearch: string
}

export const CustomerSection = ({customer, customers, errors, handleCustomerSelection, setSearch, open, setOpen, dedbouncedSearch }: CustomerSectionProps) => {
  return (
    <div className="rounded-lg bg-white shadow">
      {!open && !customer && (
        <div
          data-slot={`${errors?.customer_id ? 'customer-error' : 'default'}`}
          className={cn(
            "flex flex-col h-full w-full items-center justify-center [&_svg]:text-white px-2 pb-2",
            "data-[slot=customer-error]:rounded-lg data-[slot=customer-error]:bg-red-100/50 data-[slot=customer-error]:border",
            "data-[slot=customer-error]:border-red-500 data-[slot=customer-error]:[&_svg]:text-red-500 data-[slot=customer-error]:[&_[data-label=true]]:text-red-500"
            )}>
          <button onClick={() => setOpen(!open)} className="flex h-full w-full cursor-pointer items-center justify-center gap-2">
            <div className="flex size-10 items-center justify-center rounded-full bg-gray-200">
              <User className="size-6 *:data-[slot=customer-error]:text-red-500" />
            </div>
            <div data-label="true" className="text-lg">Customer</div>
          </button>
          <InputError className="mt-2" message={errors.customer_id} />
        </div>
      )}
      {customer && (
        <div className="flex h-full flex-col overflow-y-hidden p-2">
          <div className="flex w-full items-center justify-between">
            <div>{customer?.name}</div>
            <button onClick={(event) => handleCustomerSelection(event, undefined)} className="cursor-pointer p-1">
              <XCircleIcon />
            </button>
          </div>
          <div>{customer?.email}</div>
          <div>{customer?.phone}</div>
          <div>Address here!!!</div>
        </div>
      )}
      {open && !customer && (
        <div className="flex h-full min-h-48 grow flex-col justify-start shadow">
          <div className="w-full border-b border-gray-200 p-2">
            <Input
              type="search"
              placeholder="Search for a customer"
              className="h-11 w-full rounded-t-lg"
              onChange={(e) => setSearch(e.currentTarget.value)}
            />
          </div>
          {/* Search result */}
          <div className="bg-gray-50">
            {customers && customers.length > 0 ? (
              customers.map((customer) => (
                <button
                  key={customer.id}
                  className="flex w-full cursor-pointer items-center justify-start gap-2 rounded-lg p-2 hover:bg-gray-100"
                  onClick={(event) => handleCustomerSelection(event, customer)}
                >
                  <div className="flex size-10 items-center justify-center rounded-full bg-gray-200">
                    <User className="size-6" color="white" />
                  </div>
                  <div className="text-lg">{customer.name}</div>
                </button>
              ))
            ) : (
              <div className="flex w-full items-center justify-center p-4 text-sm text-gray-500">
                {dedbouncedSearch ? <p>No customers found</p> : null}
              </div>
            )}
          </div>
          {/* Create new action */}
          <div className="flex w-full items-center justify-center rounded-b-lg border bg-gray-100 p-2">
            <button
              className="flex cursor-pointer items-center justify-center gap-x-2 text-indigo-400"
              onClick={() => alert('Create new customer')}
            >
              <UserPlus className="size-4" /> Add New Customer
            </button>
          </div>
        </div>
      )}
    </div>
  )
}