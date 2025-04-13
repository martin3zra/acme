import React, { JSX } from 'react';
import SectionTitle from './section-title';

function Title({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

function Description({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

function Content({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

interface Props {
  children: React.ReactNode;
}

class ActionSection extends React.Component<Props> {
  static Title = Title;
  static Description = Description;
  static Content = Content;

  render() {
    const { children } = this.props;

    const title = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Title);
    const description = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Description);
    const content = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Content);

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
