import { cn } from '@/lib/utils';
import { SlotProps } from '@/types';
import { PopoverClose } from '@radix-ui/react-popover';
import React, { JSX } from 'react';
import { Command, CommandGroup, CommandInput, CommandItem, CommandList } from './ui/command';
import { Popover, PopoverContent, PopoverTrigger } from './ui/popover';

interface Props<T extends object> {
  title?: string;
  value?: T;
  valueKey: keyof T;
  search: string;
  renderEmptyText: string;
  options: T[];
  open: boolean;
  renderText: (value: T) => string;
  onSelected?: (value: T) => void;
  onChange?: (value: string) => void;
  onOpenChange?: ((open: boolean) => void) | undefined;
  children: React.ReactNode;
}

interface InputSearchableState {
  search: string;
}

function Actions({ children }: SlotProps): JSX.Element {
  return <>{children}</>;
}

function Trigger({ children }: SlotProps): JSX.Element {
  return <>{children}</>;
}

class InputSearchable<T extends object> extends React.Component<Props<T>, InputSearchableState> {
  static Actions = Actions;
  static Trigger = Trigger;
  constructor(props: Props<T>) {
    super(props);
    this.state = {
      search: props.search,
    };
  }

  handleOnValueChange = (value: string) => {
    this.setState({ search: value });
    this.props.onChange?.(value);
  };

  handleOnTrigger = (event: React.MouseEvent<HTMLButtonElement>) => {
    if (this.props.value) {
      event.preventDefault();
    }
  };

  render() {
    const { children, onOpenChange, title, options, value, valueKey, renderText, renderEmptyText, onSelected, open } = this.props;
    const { search } = this.state;

    const actions = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Actions);
    const trigger = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Trigger);

    return (
      <Popover modal={true} onOpenChange={onOpenChange}>
        <PopoverTrigger asChild onClick={this.handleOnTrigger}>
          <div className={cn('h-fit', value === undefined && !open && 'h-full', value !== undefined && open && 'h-fit', !open && 'h-full')}>
            {trigger}
          </div>
        </PopoverTrigger>
        <PopoverContent className="w-[var(--radix-popover-trigger-width)] p-0" onOpenAutoFocus={(e) => e.preventDefault()}>
          <Command shouldFilter={false} className="w-full">
            <CommandInput placeholder={`${title}`} value={typeof search === 'string' ? search : ''} onValueChange={this.handleOnValueChange} />
            <CommandList>
              <CommandGroup className="max-h-60 overflow-y-auto">
                <PopoverClose asChild>
                  <div>
                    {options &&
                      options.map((option) => (
                        <CommandItem value={option[valueKey] as string} key={option[valueKey] as string} onSelect={() => onSelected?.(option)}>
                          {renderText(option)}
                        </CommandItem>
                      ))}
                  </div>
                </PopoverClose>
                {options && options.length === 0 && search && <CommandItem>{renderEmptyText}</CommandItem>}
                {actions && <CommandItem>{actions}</CommandItem>}
              </CommandGroup>
            </CommandList>
          </Command>
        </PopoverContent>
      </Popover>
    );
  }
}

export default InputSearchable;
