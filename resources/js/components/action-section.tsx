import { SlotProps } from '@/types';
import React, { JSX } from 'react';
import SectionTitle from './section-title';

function Title({ children }: SlotProps): JSX.Element {
  return <>{children}</>;
}

function Description({ children }: SlotProps): JSX.Element {
  return <>{children}</>;
}

function Content({ children }: SlotProps): JSX.Element {
  return <>{children}</>;
}

class ActionSection extends React.Component<SlotProps> {
  static Title = Title;
  static Description = Description;
  static Content = Content;

  render() {
    const { children } = this.props;

    const array = React.Children.toArray(children);
    const title = array.find((child): child is React.ReactElement => React.isValidElement(child) && child.type === Title);
    const description = array.find((child): child is React.ReactElement => React.isValidElement(child) && child.type === Description);
    const content = array.find((child): child is React.ReactElement => React.isValidElement(child) && child.type === Content);

    return (
      <div className="md:grid md:grid-cols-3 md:gap-6">
        <SectionTitle>
          <SectionTitle.Title>{title}</SectionTitle.Title>
          <SectionTitle.Description>{description}</SectionTitle.Description>
        </SectionTitle>

        <div className="mt-5 md:col-span-2 md:mt-0">
          <div className="bg-white px-4 py-5 shadow sm:rounded-lg sm:p-6 dark:bg-gray-800">{content}</div>
        </div>
      </div>
    );
  }
}

export default ActionSection;
