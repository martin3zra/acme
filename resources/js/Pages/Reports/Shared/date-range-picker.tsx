import { Button } from '@/components/ui/button';
import { Calendar } from '@/components/ui/calendar';
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover';
import { cn } from '@/lib/utils';
import { es } from 'date-fns/locale';
import { CalendarIcon, X } from 'lucide-react';
import { DateRange } from 'react-day-picker';

interface Props {
  dateRange: DateRange | undefined;
  setDateRange: (range: DateRange | undefined) => void;
}

export default function DateRangePicker({ dateRange, setDateRange }: Props) {
  return (
    <div className="relative flex w-fit items-center">
      <Popover>
        <PopoverTrigger asChild>
          <Button variant={'outline'} className={cn('w-[280px] justify-start text-left font-normal')}>
            <CalendarIcon />
            {dateRange?.from ? (
              dateRange.to ? (
                <>
                  {dateRange.from.toLocaleDateString()} - {dateRange.to.toLocaleDateString()}
                </>
              ) : (
                dateRange.from.toLocaleDateString()
              )
            ) : (
              <span>Select date range</span>
            )}
          </Button>
        </PopoverTrigger>
        <PopoverContent className="w-auto p-0">
          <Calendar
            mode="range"
            selected={dateRange}
            onSelect={setDateRange}
            captionLayout="dropdown"
            className="w-[280px] rounded border p-3 shadow-md"
            locale={es}
          />
        </PopoverContent>
      </Popover>
      {dateRange && (
        <X
          className="text-muted-foreground hover:text-foreground absolute top-1/2 right-2 h-4 w-4 -translate-y-1/2 cursor-pointer"
          onClick={() => setDateRange(undefined)}
        />
      )}
    </div>
  );
}
