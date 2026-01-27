import FormSection from '@/components/form-section';
import { Button } from '@/components/ui/button';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { useHeader } from '@/composables/use-headers';
import { PageProps, RedirectPreference, RedirectPreferenceValue } from '@/types';
import { useForm, usePage } from '@inertiajs/react';

type Props = {
  uuid: string;
  preferences: RedirectPreference;
};

export function RedirectPreferences({ uuid, preferences }: Props) {
  const { auth } = usePage<PageProps>().props;
  const { headers } = useHeader();
  const { data, setData, put, processing } = useForm({
    invoice: preferences.invoice || 'detail',
    estimate: preferences.estimate || 'list',
    customer: preferences.customer || 'list',
    order: preferences.order || 'list',
    payment: preferences.payment || 'list',
    item: preferences.item || 'list',
  });

  const handleSubmit = () => {
    put(`/settings/${auth.account.uuid}/companies/${uuid}/redirect-preferences`, { ...headers, preserveState: 'errors' });
  };

  return (
    <div className="flex w-full flex-col space-y-6 py-6 [&_[data-form]]:md:grid-cols-1">
      <FormSection onSubmit={handleSubmit}>
        <FormSection.Title>Redirect Preferences</FormSection.Title>
        <FormSection.Description>
          Manage were you want to be redirect after the creation of an invoice, customer, estimate or an order.
        </FormSection.Description>
        <FormSection.Form>
          <div className="flex justify-between space-x-2">
            <div className="basis-1/2 space-y-2">
              <Label>After creating an Invoice</Label>
              <Select value={data.invoice} onValueChange={(val: RedirectPreferenceValue) => setData('invoice', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">Redirect to Invoices List</SelectItem>
                  <SelectItem value="detail">Redirect to Invoice Detail</SelectItem>
                  <SelectItem value="stay">Stay on Create Form</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="basis-1/2 space-y-2">
              <Label>After creating an Estimate</Label>
              <Select value={data.estimate} onValueChange={(val: RedirectPreferenceValue) => setData('estimate', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">Redirect to Estimates List</SelectItem>
                  <SelectItem value="detail">Redirect to Estimate Detail</SelectItem>
                  <SelectItem value="stay">Stay on Create Form</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex justify-between space-x-2">
            <div className="basis-1/2 space-y-2">
              <Label>After creating a Customer</Label>
              <Select value={data.customer} onValueChange={(val: RedirectPreferenceValue) => setData('customer', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">Redirect to Customers List</SelectItem>
                  <SelectItem value="detail">Redirect to Customer Detail</SelectItem>
                  <SelectItem value="stay">Stay on Create Form</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="basis-1/2 space-y-2">
              <Label>After creating an Order</Label>
              <Select value={data.order} onValueChange={(val: RedirectPreferenceValue) => setData('order', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">Redirect to Orders List</SelectItem>
                  <SelectItem value="detail">Redirect to Order Detail</SelectItem>
                  <SelectItem value="stay">Stay on Create Form</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex justify-between space-x-2">
            <div className="basis-1/2 space-y-2">
              <Label>After creating a Item</Label>
              <Select value={data.item} onValueChange={(val: RedirectPreferenceValue) => setData('item', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">Redirect to Items List</SelectItem>
                  <SelectItem value="detail">Redirect to Item Detail</SelectItem>
                  <SelectItem value="stay">Stay on Create Form</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="basis-1/2 space-y-2">
              <Label>After creating a Payment</Label>
              <Select value={data.payment} onValueChange={(val: RedirectPreferenceValue) => setData('payment', val)}>
                <SelectTrigger className="w-full">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="list">Redirect to Payments List</SelectItem>
                  <SelectItem value="detail">Redirect to Payment Detail</SelectItem>
                  <SelectItem value="stay">Stay on Create Form</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
        </FormSection.Form>
        <FormSection.Actions>
          <Button type="submit" disabled={processing}>
            Save Preferences
          </Button>
        </FormSection.Actions>
      </FormSection>
    </div>
  );
}
