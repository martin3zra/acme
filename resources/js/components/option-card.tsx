import ChipSelector from '@/components/chip-selector';
import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';

type OptionCardValue = {
  value: number;
  label: string;
};

type OptionCardProps = {
  title: string;
  checked: boolean;
  values: OptionCardValue[];
  selectedValueIDs: number[];
  onCheckedChange: (checked: boolean) => void;
  onToggleValue: (valueID: number, checked: boolean) => void;
  description?: string;
  disabled?: boolean;
};

export default function OptionCard({
  title,
  checked,
  values,
  selectedValueIDs,
  onCheckedChange,
  onToggleValue,
  description,
  disabled = false,
}: OptionCardProps) {
  const cardID = `option-card-${title
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')}`;

  return (
    <section className={`space-y-3 rounded-xl border p-3 sm:p-4 ${checked ? 'border-primary/40 bg-primary/5' : 'border-border bg-background'}`}>
      <div className="flex items-start justify-between gap-3">
        <div className="space-y-1">
          <Label htmlFor={cardID} className="text-sm font-semibold tracking-tight">
            {title}
          </Label>
          {description && <p className="text-muted-foreground text-xs">{description}</p>}
        </div>

        <Checkbox
          id={cardID}
          checked={checked}
          disabled={disabled}
          onCheckedChange={(next) => onCheckedChange(next === true)}
        />
      </div>

      {!checked ? (
        <p className="text-muted-foreground text-xs">Enable this option to choose values.</p>
      ) : values.length === 0 ? (
        <p className="text-muted-foreground text-xs">No values configured for this option yet.</p>
      ) : (
        <div className="space-y-2">
          <p className="text-muted-foreground text-xs">{selectedValueIDs.length} selected</p>
          <ChipSelector
            options={values}
            selectedValues={selectedValueIDs}
            onToggle={(valueID) => onToggleValue(valueID, !selectedValueIDs.includes(valueID))}
            chipClassName="text-xs"
          />
        </div>
      )}
    </section>
  );
}
