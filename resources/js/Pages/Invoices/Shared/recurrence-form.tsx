import { DatePickerField } from '@/components/date-picker';
import FormSection from '@/components/form-section';
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { Switch } from '@/components/ui/switch';
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group';
import { usePersistedState } from '@/hooks/use-persisted-state';
import { useTranslation } from '@/hooks/use-translation';
import { recurrenceCaption, validateRecurrence } from '@/lib/utils';
import { Frequency, Months, Recurrent, RecurrentType, WeekDay } from '@/types';
import { RefreshCwIcon } from 'lucide-react';
import React, { JSX, useState } from 'react';
import { defaultReccurence } from '../constants';

const recurrenceIntervals: Record<Frequency, { min: number; max: number }> = {
  daily: { min: 1, max: 7 }, // every 1–7 days
  weekly: { min: 1, max: 52 }, // every 1–52 weeks
  monthly: { min: 1, max: 12 }, // every 1–12 months
  quarterly: { min: 1, max: 4 }, // every 1–4 quarters
  yearly: { min: 1, max: 10 }, // every 1–10 years
};

interface Props {
  name?: string;
  onSubmit: (recurrence: Recurrent) => void;
  onCancel: () => void;
}
export function RecurrenceForm({ name, onSubmit, onCancel }: Props) {
  const [recurrenceForm, setRecurrenceForm, removeRecurrenceForm] = usePersistedState<Recurrent>('recurrence', {
    ...defaultReccurence,
    enabled: true,
    name: name ?? '',
  });
  const [error, setError] = useState<string | null>(null);
  const [openCancelConfirmation, setCancelConfirmation] = useState<boolean>(false);
  const startDate = recurrenceForm.start_date ? new Date(recurrenceForm.start_date) : undefined;
  const until = recurrenceForm.until ? new Date(recurrenceForm.until) : undefined;
  const t = useTranslation().trans;
  const handleSubmit = (): void => {
    try {
      setError(null);
      const recurrence = validateRecurrenceForm(recurrenceForm);
      setRecurrenceForm(() => {
        return { ...recurrence };
      });
      onSubmit(recurrence);
    } catch (err) {
      setError((err as Error).message);
    }
  };
  const performRecurrenceCancelation = (event: React.MouseEvent<HTMLButtonElement>) => {
    removeRecurrenceForm();
    onCancel();
  };
  const handleFrequencyChange = (val: Recurrent['frequency']) => {
    const recurrence = { ...recurrenceForm, interval: 1, frequency: val };

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

    setRecurrenceForm(() => {
      return { ...recurrence };
    });
  };
  const handleInputChange = <K extends keyof Recurrent>(key: K, value: number | string | string[] | Date | boolean | undefined) => {
    setRecurrenceForm(() => {
      return { ...recurrenceForm, [key]: value };
    });
  };
  const validateRecurrenceForm = (data: Recurrent): Recurrent => {
    // Validate name
    if (!data.name || data.name.trim().length === 0) {
      throw new Error('Name is required.');
    }
    if (data.name.length > 100) {
      throw new Error('Name cannot exceed 100 characters.');
    }

    // Validate date range
    if (data.start_date && data.until) {
      const start = new Date(data.start_date);
      const until = new Date(data.until);

      if (start > until) {
        throw new Error('Start date must be before or equal to the until date.');
      }
    }

    // Delegate to recurrence-specific validation
    return validateRecurrence(data);
  };

  return (
    <div className="space-y-4 border-t p-4 pt-4 [&_[data-form]]:md:grid-cols-1">
      <FormSection onSubmit={handleSubmit}>
        <FormSection.Title>Recurrence</FormSection.Title>
        <FormSection.Description>
          <RecurrenceDescriptionView />
        </FormSection.Description>
        <FormSection.Form>
          <div className="recurrence-form">{error && <div className="mt-2 text-red-600">{error}</div>}</div>

          <div className="flex items-center space-x-2">
            <div className="flex basis-3/4 flex-col gap-y-2">
              <Label>Name</Label>
              <Input
                name="name"
                value={recurrenceForm.name}
                onChange={(e: React.ChangeEvent<HTMLInputElement>) => handleInputChange('name', e.target.value)}
              />
            </div>
            <div className="flex flex-col gap-y-2">
              <Label>Type</Label>
              <Select value={recurrenceForm.type} onValueChange={(val: RecurrentType) => handleInputChange('type', val)}>
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
              <Select value={recurrenceForm.frequency} onValueChange={handleFrequencyChange}>
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
                min={recurrenceIntervals[recurrenceForm.frequency].min}
                max={recurrenceIntervals[recurrenceForm.frequency].max}
                value={recurrenceForm.interval}
                onChange={(e) => {
                  const min = recurrenceIntervals[recurrenceForm.frequency].min;
                  const max = recurrenceIntervals[recurrenceForm.frequency].max;
                  let value = e.target.valueAsNumber;
                  if (value <= 0) value = min;
                  if (value > max) value = max; // clamp to max
                  handleInputChange('interval', value);
                }}
                className="w-20"
              />
            </div>

            {recurrenceForm.frequency === 'yearly' && (
              <div className="flex flex-col gap-y-2">
                <Label>Month</Label>
                <Select value={String(recurrenceForm.month ?? 1)} onValueChange={(val) => handleInputChange('month', Number(val))}>
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

            {['monthly', 'quarterly', 'yearly'].includes(recurrenceForm.frequency) && (
              <div className="flex flex-col gap-y-2">
                <Label>Day of Month</Label>
                <Select value={String(recurrenceForm.day_of_month ?? 1)} onValueChange={(val) => handleInputChange('day_of_month', Number(val))}>
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

            {recurrenceForm.frequency === 'weekly' && (
              <div className="flex">
                <ToggleGroup
                  type="multiple"
                  size={'sm'}
                  value={recurrenceForm.weekdays}
                  onValueChange={(weeks: WeekDay[]) => handleInputChange('weekdays', weeks)}
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
              <DatePickerField
                id="startDate"
                label={t('global.date')}
                placeholder={t('global.datePlaceholder')}
                value={startDate}
                onChange={(date: Date | undefined) => handleInputChange('start_date', date)}
              />
            </div>
            <div className="flex flex-col gap-y-2">
              <DatePickerField
                id="until"
                label={t('global.date')}
                placeholder={t('global.datePlaceholder')}
                value={until}
                onChange={(date: Date | undefined) => handleInputChange('until', date)}
              />
            </div>
          </div>

          <div className="flex items-center space-x-2">
            <Switch id="send_email" onCheckedChange={(checked: boolean) => handleInputChange('send_email', checked)} />
            <Label htmlFor="send_email">Enviar correos electrónicos automáticamente</Label>
          </div>
          <div className="flex items-center space-x-2 rounded bg-gray-50 py-2 ps-4">
            <RefreshCwIcon className="size-4 text-gray-400" />
            <div>
              <div>{recurrenceCaption(recurrenceForm)}</div>
            </div>
          </div>
        </FormSection.Form>
        <FormSection.Actions>
          <Button type="button" variant={'ghost'} onClick={() => setCancelConfirmation(true)}>
            {t('global.cancel')}
          </Button>
          <Button>{t('global.save')}</Button>
        </FormSection.Actions>
      </FormSection>

      <AlertDialog open={openCancelConfirmation} onOpenChange={setCancelConfirmation}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('invoices.confirmsCancelation.title')}</AlertDialogTitle>
            <AlertDialogDescription>{t('invoices.confirmsCancelation.description')}</AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('global.cancel')}</AlertDialogCancel>
            <AlertDialogAction onClick={performRecurrenceCancelation}>{t('invoices.confirmsCancelation.confirm')}</AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

const RecurrenceDescriptionView = (): JSX.Element => {
  return (
    <>
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
    </>
  );
};
