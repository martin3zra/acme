'use client';

import { format } from 'date-fns';
import { es } from 'date-fns/locale';
import { Calendar as CalendarIcon } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import { Label } from '@/components/ui/label';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';
import InputError from './input-error';

type DatePickerFieldProps = {
  id: string;
  label: string;
  placeholder?: string;
  value?: Date;
  onChange: (date: Date | undefined) => void;
  error?: string;
  className?: string;
};

export function DatePickerField({ id, label, placeholder = 'Pick a date', value, onChange, error, className }: DatePickerFieldProps) {
  return (
    <div className="flex flex-col gap-y-2">
      <Label htmlFor={id}>{label}</Label>
      <Popover>
        <PopoverTrigger asChild>
          <Button
            type="button"
            variant="outline"
            className={cn('w-70 justify-start text-left font-normal', !value && 'text-muted-foreground', className)}
          >
            <CalendarIcon className="mr-2 h-4 w-4" />
            {value ? format(value, 'dd-MM-yyyy') : <span>{placeholder}</span>}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0">
          <Calendar
            mode="single"
            defaultMonth={value}
            selected={value}
            onSelect={onChange}
            captionLayout="dropdown"
            className="min-w-56 rounded-md border p-3 shadow-md"
            locale={es}
          />
        </PopoverContent>
      </Popover>
      {error && <InputError className="mt-2" message={error} />}
    </div>
  );
}
