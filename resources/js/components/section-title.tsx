import { SlotProps } from '@/types';
import React, { JSX } from 'react';

function Title({ children }: SlotProps): JSX.Element {
  return <>{children}</>;
}

function Description({ children }: SlotProps): JSX.Element {
  return <>{children}</>;
}

function Aside({ children }: SlotProps): JSX.Element {
  return <>{children}</>;
}

class SectionTitle extends React.PureComponent<SlotProps> {
  static Title = Title;
  static Description = Description;
  static Aside = Aside;
  render() {
    const { children } = this.props;
    const array = React.Children.toArray(children);
    const title = array.find((child): child is React.ReactElement => React.isValidElement(child) && child.type === Title);
    const description = array.find((child): child is React.ReactElement => React.isValidElement(child) && child.type === Description);
    const aside = array.find((child): child is React.ReactElement => React.isValidElement(child) && child.type === Aside);
    return (
      <div className="flex justify-between md:col-span-1">
        <div className="px-4 sm:px-0">
          <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">{title ? (title as JSX.Element) : null}</h3>
          <div className="mt-1 text-sm text-gray-600 dark:text-gray-400">{description ? (description as JSX.Element) : null}</div>
        </div>
        <div className="px-4 sm:px-0">{aside ? (aside as JSX.Element) : null}</div>
      </div>
    );
  }
}

export default SectionTitle;
