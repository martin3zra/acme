import FormSection from '@/components/form-section';
import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { useTranslation } from '@/hooks/use-translation';
import { cn, recurrenceCaption, validateRecurrence } from '@/lib/utils';
import { Frequency, Months, Recurrent, RecurrentType, WeekDay } from '@/types';
import { format } from 'date-fns';
import { CalendarIcon, RefreshCwIcon } from 'lucide-react';
import { useState } from 'react';
import { InvoiceFormData } from '../Create';

const recurrenceIntervals: Record<Frequency, { min: number; max: number }> = {
  daily: { min: 1, max: 7 }, // every 1–7 days
  weekly: { min: 1, max: 52 }, // every 1–52 weeks
  monthly: { min: 1, max: 12 }, // every 1–12 months
  quarterly: { min: 1, max: 4 }, // every 1–4 quarters
  yearly: { min: 1, max: 10 }, // every 1–10 years
};

interface Props {
  data: InvoiceFormData;
  setData: <K extends keyof InvoiceFormData>(key: K, value: InvoiceFormData[K]) => void;
}
export function RecurrenceForm({ data, setData }: Props) {
  const [error, setError] = useState<string | null>(null);
  const [openStartDate, setOpenStartDate] = useState(false);
  const [openUntilDate, setOpenUntilDate] = useState(false);
  const startDate = data.recurrence.start_date ? new Date(data.recurrence.start_date) : undefined;
  const until = data.recurrence.until ? new Date(data.recurrence.until) : undefined;
  const t = useTranslation().trans;
  const handleSubmit = (): void => {
    try {
      setError(null);
      const payload = validateRecurrenceForm(data);
      console.log('Validated recurrence:', payload);
    } catch (err) {
      setError((err as Error).message);
    }
  };
  const handleStartDateChange = (date: unknown) => {
    setData('recurrence', { ...data.recurrence, start_date: date as Date });
  };
  const handleUntilDateChange = (date: unknown) => {
    setData('recurrence', { ...data.recurrence, until: date as Date });
  };

  const handleFrequencyChange = (val: Recurrent['frequency']) => {
    const recurrence = { ...data.recurrence, interval: 1, frequency: val };

    switch (val) {
      case 'yearly':
        recurrence.month = recurrence.month ?? 1; // default January
        recurrence.day_of_month = recurrence.day_of_month ?? 1;
        break;
      case 'monthly':
      case 'quarterly':
        recurrence.day_of_month = recurrence.day_of_month ?? 1;
        break;
      case 'weekly':
        recurrence.weekdays = recurrence.weekdays && recurrence.weekdays.length > 0 ? recurrence.weekdays : ['monday']; // default Monday
        break;
      case 'daily':
        // no extra defaults needed
        break;
    }

    setData('recurrence', recurrence);
  };
  const validateRecurrenceForm = (data: InvoiceFormData): Recurrent => {
    // Validate name
    if (!data.recurrence.name || data.recurrence.name.trim().length === 0) {
      throw new Error('Name is required.');
    }
    if (data.recurrence.name.length > 100) {
      throw new Error('Name cannot exceed 100 characters.');
    }

    // Validate date range
    if (data.recurrence.start_date && data.recurrence.until) {
      const start = new Date(data.recurrence.start_date);
      const until = new Date(data.recurrence.until);

      if (start > until) {
        throw new Error('Start date must be before or equal to the until date.');
      }
    }

    // Delegate to recurrence-specific validation
    return validateRecurrence(data.recurrence);
  };

  return (
    <div className="space-y-4 border-t p-4 pt-4 [&_[data-form]]:md:grid-cols-1">
      <FormSection onSubmit={handleSubmit}>
        <FormSection.Title>Recurrence</FormSection.Title>
        <FormSection.Description>
          <p>Manage your recurrent invoices by setting a template name, and recurrency type with the frequency, interval and optional end date.</p>
          <div className="my-4 space-y-2 rounded bg-gray-50 px-2 py-4 [&_li]:mb-1 [&_span]:font-bold">
            <ul>
              <li>
                <span>Daily</span> → “Every X days”
                <ul className="list-circle text-muted-foreground mt-2 space-y-1 pl-6 text-sm">
                  <li>“Every 1 day”</li>
                  <li>“Every 3 days”</li>
                </ul>
              </li>
              <li>
                <span>Weekly</span> → “Every X weeks on [weekday(s)]”
                <ul className="list-circle text-muted-foreground mt-2 space-y-1 pl-6 text-sm">
                  <li>“Every 1 week on Monday”</li>
                  <li>“Every 2 weeks on Tuesday and Thursday”</li>
                </ul>
              </li>
              <li>
                <span>Monthly</span> → “Every X months on day [N]”
                <ul className="list-circle text-muted-foreground mt-2 space-y-1 pl-6 text-sm">
                  <li>“Every 1 month on day 15”</li>
                  <li>“Every 2 month on day 31 (clamped if shorter month)”</li>
                </ul>
              </li>
              <li>
                <span>Quarterly</span> → “Every X quarters on day [N]”
                <ul className="list-circle text-muted-foreground mt-2 space-y-1 pl-6 text-sm">
                  <li>“Every 1 quarter on day 15” → generates invoices every 3 months on the 15th</li>
                  <li>“Every 2 quarters on day 1” → generates invoices every 6 months on the 1st</li>
                </ul>
              </li>
              <li>
                <span>Yearly</span> → “Every X years on [Month] [Day]”
                <ul className="list-circle text-muted-foreground mt-2 space-y-1 pl-6 text-sm">
                  <li>“Every 1 year on February 15”</li>
                  <li>“Every 2 years on December 31”</li>
                </ul>
              </li>
            </ul>
          </div>
        </FormSection.Description>
        <FormSection.Form>
          <div className="recurrence-form">{error && <div className="mt-2 text-red-600">{error}</div>}</div>

          <div className="flex items-center space-x-2">
            <div className="flex basis-3/4 flex-col gap-y-2">
              <Label>Name</Label>
              <Input
                name="name"
                value={data.recurrence.name}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => setData('recurrence', { ...data.recurrence, name: e.target.value })}
              />
            </div>
            <div className="flex flex-col gap-y-2">
              <Label>Type</Label>
              <Select value={data.recurrence.type} onValueChange={(val: RecurrentType) => setData('recurrence', { ...data.recurrence, type: val })}>
                <SelectTrigger className="w-40">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="schedule">Schedule</SelectItem>
                  <SelectItem value="reminder">Reminder</SelectItem>
                </SelectContent>
              </Select>
            </div>
          </div>
          <div className="flex items-end space-x-2">
            <div className="flex flex-col gap-y-2">
              <Label>Frequency</Label>
              <Select value={data.recurrence.frequency} onValueChange={handleFrequencyChange}>
                <SelectTrigger className="w-32">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="daily">Daily</SelectItem>
                  <SelectItem value="weekly">Weekly</SelectItem>
                  <SelectItem value="monthly">Monthly</SelectItem>
                  <SelectItem value="quarterly">Quarterly</SelectItem>
                  <SelectItem value="yearly">Yearly</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="flex flex-col gap-y-2">
              <Label>Interval</Label>
              <Input
                type="number"
                min={recurrenceIntervals[data.recurrence.frequency].min}
                max={recurrenceIntervals[data.recurrence.frequency].max}
                value={data.recurrence.interval}
                onChange={(e) => {
                  const min = recurrenceIntervals[data.recurrence.frequency].min;
                  const max = recurrenceIntervals[data.recurrence.frequency].max;
                  let value = e.target.valueAsNumber;
                  if (value <= 0) value = min;
                  if (value > max) value = max; // clamp to max
                  setData('recurrence', { ...data.recurrence, interval: value });
                }}
                className="w-20"
              />
            </div>

            {data.recurrence.frequency === 'yearly' && (
              <div className="flex flex-col gap-y-2">
                <Label>Month</Label>
                <Select
                  value={String(data.recurrence.month ?? 1)}
                  onValueChange={(val) => setData('recurrence', { ...data.recurrence, month: Number(val) })}
                >
                  <SelectTrigger className="w-36">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {Months.map((month, index) => (
                      <SelectItem key={month} value={`${index + 1}`}>
                        {month}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {['monthly', 'quarterly', 'yearly'].includes(data.recurrence.frequency) && (
              <div className="flex flex-col gap-y-2">
                <Label>Day of Month</Label>
                <Select
                  value={String(data.recurrence.day_of_month ?? 1)}
                  onValueChange={(val) => setData('recurrence', { ...data.recurrence, day_of_month: Number(val) })}
                >
                  <SelectTrigger className="w-22">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent className="w-16">
                    {Array.from({ length: 31 }, (_, i) => (
                      <SelectItem key={i + 1} value={String(i + 1)}>
                        {i + 1}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
            )}

            {data.recurrence.frequency === 'weekly' && (
              <div className="flex">
                <ToggleGroup
                  type="multiple"
                  size={'sm'}
                  value={data.recurrence.weekdays}
                  onValueChange={(weeks: WeekDay[]) => setData('recurrence', { ...data.recurrence, weekdays: weeks })}
                >
                  {['sunday', 'monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday'].map((w, index) => (
                    <ToggleGroupItem value={w} key={index} className="data-[state=on]:bg-primary data-[state=on]:text-primary-foreground capitalize">
                      <span className="block px-1">{w}</span>
                    </ToggleGroupItem>
                  ))}
                </ToggleGroup>
              </div>
            )}
          </div>

          <div className="flex space-x-2">
            <div className="flex flex-col gap-y-2">
              <Label>Start Date</Label>
              <Button
                variant="outline"
                className={cn('w-[260px] justify-start text-left font-normal', !startDate && 'text-muted-foreground')}
                onClick={() => setOpenStartDate((v) => !v)}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {startDate ? format(startDate, 'PPP') : <span>{t('global.datePlaceholder')}</span>}
              </Button>

              <div className="relative">
                {openStartDate && (
                  <div className="bg-popover absolute z-50 mt-2 rounded-md border shadow-md">
                    <Calendar
                      mode="single"
                      selected={startDate}
                      onSelect={(date) => {
                        if (!date) return;
                        handleStartDateChange(date);
                        setOpenStartDate(false); // close after select
                      }}
                    />
                  </div>
                )}
              </div>
            </div>
            <div className="flex flex-col gap-y-2">
              <Label>Until Date</Label>
              <Button
                variant="outline"
                className={cn('w-[260px] justify-start text-left font-normal', !until && 'text-muted-foreground')}
                onClick={() => setOpenUntilDate((v) => !v)}
              >
                <CalendarIcon className="mr-2 h-4 w-4" />
                {until ? format(until, 'PPP') : <span>{t('global.datePlaceholder')}</span>}
              </Button>

              <div className="relative">
                {openUntilDate && (
                  <div className="bg-popover absolute z-50 mt-2 rounded-md border shadow-md">
                    <Calendar
                      mode="single"
                      selected={until}
                      onSelect={(date) => {
                        if (!date) return;
                        handleUntilDateChange(date);
                        setOpenUntilDate(false); // close after select
                      }}
                    />
                  </div>
                )}
              </div>
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <Switch id="send_email" />
            <Label htmlFor="send_email">Enviar correos electrónicos automáticamente</Label>
          </div>
          <div className="flex items-center space-x-2 rounded bg-gray-50 py-2 ps-4">
            <RefreshCwIcon className="size-4 text-gray-400" />
            <div>
              <div>{recurrenceCaption(data.recurrence)}</div>
            </div>
          </div>
        </FormSection.Form>
        <FormSection.Actions>
          <Button>{t('global.save')}</Button>
        </FormSection.Actions>
      </FormSection>
    </div>
  );
}
