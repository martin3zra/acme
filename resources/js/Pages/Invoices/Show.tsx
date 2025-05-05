import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Separator } from '@/components/ui/separator';
import { useNumber } from '@/composables/use-number';
import { cn, isNotEmpty } from '@/lib/utils';
import { Auth, InvoiceWithLines, PaidStatuses } from '@/types';
import { format } from 'date-fns';
import { Calendar1, CircleCheckIcon, CircleDollarSignIcon, CreditCardIcon, InfoIcon, UserPen } from 'lucide-react';
import PaymentSummary from './Shared/payment-summary';

type Props = {
  invoice: InvoiceWithLines;
  auth: Auth;
};

export default function Show({ invoice, auth }: Props) {
  const { currency } = useNumber();

  return (
    <div className="grid grid-cols-12 gap-x-4">
      <div className="col-span-12 pb-6">
        <Separator />
        <div className="flex justify-between py-6 [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-medium">
          <div className="col-span-6 flex items-center gap-x-6 [&>div]:flex [&>div]:gap-x-2">
            <div>
              <Label>Invoice</Label>
              <Label>#{invoice.header.number}</Label>
            </div>
            <div>
              <Label>NCF</Label>
              <Label>{invoice.header.ncf}</Label>
            </div>
          </div>
          <div className="col-span-6 flex items-center gap-x-2 [&_[data-slot=label]]:font-normal">
            <Label>Date</Label>
            <Label className="">{format(invoice.header.date, 'PPP')}</Label>
          </div>
        </div>
        <Separator />
      </div>
      <div className="col-span-9 grid h-full w-full grid-cols-12 grid-rows-[auto_1fr_auto] gap-y-4">
        {/* header */}
        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-6">
            <div className="w-56">
              <Label className="font-bold">Pay to:</Label>
              <div className="pt-2">
                <span className="text-sm font-semibold">{auth.company.name}</span>
                <address className="text-muted-foreground text-sm font-normal">{auth.company.address}</address>
              </div>
              <div className="text-sm font-medium">{invoice.header.customer.email}</div>
            </div>
          </div>
          <div className="col-span-6 place-items-end">
            <div className="w-56">
              <Label className="font-bold">Invoice to:</Label>
              <div className="pt-2">
                <span className="text-sm font-semibold">{invoice.header.customer.name}</span>
                <address className="text-muted-foreground text-sm font-normal">{invoice.header.customer.address}</address>
              </div>
              <div className="text-sm font-medium">{invoice.header.customer.email}</div>
            </div>
          </div>
        </div>
        {/* lines */}
        <div className="col-span-12">
          <table
            className={cn(
              'w-full table-auto',
              '[&_th]:border-none [&_th]:border-gray-200 [&_th]:bg-gray-50/25 [&_th]:text-sm [&_th]:font-semibold [&_th]:uppercase',
              '[&_th]:p-2 [&_th]:text-start [&_th[data-format=number]]:text-end',
              '[&_td]:border-y [&_td]:p-2 [&_td]:text-start [&_td[data-format=number]]:w-36 [&_td[data-format=number]]:text-end',
            )}
          >
            <thead>
              <tr className="bg-gray-50/50">
                <th scope="col" className="text-start">
                  ID
                </th>
                <th scope="col" className="text-start">
                  Name
                </th>
                <th scope="col" className="w-40 text-start">
                  Unit
                </th>
                <th scope="col" data-format="number" className="!w-20">
                  Qty
                </th>
                <th scope="col" data-format="number">
                  Price
                </th>
                <th scope="col" data-format="number">
                  Tax
                </th>
                <th scope="col" data-format="number">
                  Amount
                </th>
              </tr>
            </thead>
            <tbody>
              {invoice.lines.map((line) => (
                <tr key={line.id}>
                  <td>{line.id}</td>
                  <td>{line.name}</td>
                  <td>{line.unit.name}</td>
                  <td data-format="number" className="!w-20">
                    {line.qty}
                  </td>
                  <td data-format="number">{currency(line.price)}</td>
                  <td data-format="number">{currency(line.tax.amount)}</td>
                  <td data-format="number">{currency(line.amount)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        {/* footer */}
        <div className="col-span-12 grid grid-cols-12">
          <div className="col-span-8">
            <div className="max-w-sm rounded-md border p-4">
              <Label className="text-sm/6 font-medium">Notes:</Label>
              <div className="text-muted-foreground text-sm">{isNotEmpty(invoice.header.notes) ? invoice.header.notes : 'No notes left'}</div>
            </div>
          </div>
          <div className="col-span-4 rounded-md border p-4">
            <Label>Order summary</Label>
            <div
              className={cn(
                'flex flex-col gap-y-3 py-4',
                '[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-base/2 [&_[data-slot=label]]:font-normal',
                '[&_[data-slot=label-value]]:text-lg [&_[data-slot=label-value]]:font-medium',
              )}
            >
              <div className="flex items-center justify-between">
                <Label>Amount</Label>
                <Label data-slot="label-value">{currency(invoice.header.amount)}</Label>
              </div>
              <div className="flex items-center justify-between">
                <Label>Tax</Label>
                <Label data-slot="label-value">{currency(invoice.header.tax)}</Label>
              </div>
              <div className="flex items-center justify-between">
                <Label>Discount</Label>
                <Label data-slot="label-value">
                  {invoice.header.discount.type === 'fixed' ? currency(invoice.header.discount.value) : `${invoice.header.discount.value}%`}
                </Label>
              </div>
              <Separator />
              <div className="flex items-center justify-between">
                <Label>Total</Label>
                <Label data-slot="label-value">{currency(invoice.header.total)}</Label>
              </div>
            </div>
          </div>
        </div>
      </div>
      {/* side panel with summary */}
      <div className="col-span-3 flex flex-col gap-y-3 rounded-lg border p-3">
        {/* add composable/hook to decorate the paid status, to get color and proper labels */}
        <div className="flex h-fit items-center gap-x-1 rounded-md border border-green-300 bg-green-50 px-4 py-3 text-green-800">
          <InfoIcon size={16} />
          Paid status: {invoice.header.paid_status}
        </div>
        <Label>Details:</Label>
        <Separator />
        <div className="flex items-center justify-between">
          <Label className="text-lg">{currency(invoice.header.total)}</Label>
          <Select name="paid_status" defaultValue={'0'} value={invoice.header.paid_status} required disabled={invoice.header.status === 'void'}>
            <SelectTrigger className="w-28">
              <SelectValue placeholder="Paid status" />
            </SelectTrigger>
            <SelectContent className="w-12">
              {PaidStatuses.map((status) => (
                <SelectItem key={status} value={status}>
                  {status}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <Separator />
        <div className="flex items-center gap-x-1 text-sm">
          <UserPen size={14} />
          <span className="font-medium">Created by:</span>
          <span className="text-muted-foreground">Jane Doe</span>
        </div>
        {invoice.header.due_on && (
          <div className="flex items-center gap-x-1 text-sm">
            <Calendar1 size={14} />
            <span className="font-medium">Due date:</span>
            <span className="text-muted-foreground">{format(invoice.header.due_on, 'PPP')}</span>
          </div>
        )}
        <div className="flex items-center gap-x-1 text-sm">
          <CircleDollarSignIcon size={14} />
          <span className="font-medium">Currency:</span>
          <span className="text-muted-foreground">Domincan Peso</span>
        </div>
        <div className="flex items-center gap-x-1 text-sm">
          <CreditCardIcon size={14} />
          <span className="font-medium">Payment summary</span>
        </div>
        {invoice.header.due_on !== null && <span className="text-muted-foreground -m-1.5 block px-1.5 text-sm">No available yet</span>}
        <PaymentSummary paymentData={invoice.header.payment} />
        <Separator />
        <div className="px-1 py-4">
          <ol
            className={cn(
              'relative flex flex-col justify-between gap-y-2 border-s border-gray-200',
              '[&_li]:ms-4 [&_li]:flex [&_li]:items-center [&_li]:justify-between [&_li]:text-sm',
              '[&_li>div]:flex [&_li>div]:items-center [&_li>div]:gap-x-1 [&_svg]:absolute [&_svg]:-start-2.5 [&_svg]:size-5',
              '[&_svg]:fill-green-400 [&_svg]:stroke-white',
              '[&_[data-status=pending]]:text-muted-foreground [&_[data-status=pending]]:fill-gray-200',
              '[&_[data-slot=label]]:text-muted-foreground [&_[data-slot=label]]:text-sm/1 [&_[data-slot=label]]:font-normal',
            )}
          >
            <li>
              <div>
                <CircleCheckIcon />
                <div>Invoice created</div>
              </div>
              <Label>{format(invoice.header.date, 'P')}</Label>
            </li>
            <li>
              <div>
                <CircleCheckIcon />
                <div>Invoice sent</div>
              </div>
              <Label>{format(invoice.header.date, 'P')}</Label>
            </li>
            <li>
              <div>
                <CircleCheckIcon data-status="pending" />
                <div>Invoice paid</div>
              </div>
              <Label>{format(invoice.header.date, 'P')}</Label>
            </li>
          </ol>
        </div>
      </div>
    </div>
  );
}
