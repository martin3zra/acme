import React, { JSX } from 'react';

function Title({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

function Description({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

function Aside({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

interface Props {
  children: React.ReactNode;
}

class SectionTitle extends React.PureComponent<Props> {
  static Title = Title;
  static Description = Description;
  static Aside = Aside;
  render() {
    const { children } = this.props;
    const title = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Title);
    const description = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Description);
    const aside = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Aside);
    return (
      <div className="flex justify-between md:col-span-1">
        <div className="px-4 sm:px-0">
          <h3 className="text-lg font-medium text-gray-900 dark:text-gray-100">{title ? (title as JSX.Element) : null}</h3>
          <p className="mt-1 text-sm text-gray-600 dark:text-gray-400">{description ? (description as JSX.Element) : null}</p>
        </div>
        <div className="px-4 sm:px-0">{aside ? (aside as JSX.Element) : null}</div>
      </div>
    );
  }
}

export default SectionTitle;
