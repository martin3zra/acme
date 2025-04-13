import React, { FormEvent, JSX } from 'react';
import SectionTitle from './section-title';

function Title({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

function Description({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

function Form({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

function Actions({ children }: React.ReactNode): JSX.Element {
  return <>{children}</>;
}

interface Props {
  onSubmit: () => void;
  children: React.ReactNode;
}

class FormSection extends React.Component<Props> {
  static Title = Title;
  static Description = Description;
  static Form = Form;
  static Actions = Actions;

  onSubmit = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    this.props.onSubmit();
  };
  render() {
    const { children } = this.props;

    const title = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Title);
    const description = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Description);
    const form = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Form);
    const actions = (React.Children.toArray(children) as React.ReactNode).find((children: React.ReactNode) => children.type === Actions);

    return (
      <div className="md:grid md:grid-cols-3 md:gap-6">
        <SectionTitle>
          <SectionTitle.Title>{title}</SectionTitle.Title>
          <SectionTitle.Description>{description}</SectionTitle.Description>
        </SectionTitle>
        <div className="mt-5 md:col-span-2 md:mt-0">
          <form onSubmit={this.onSubmit}>
            <div className={`bg-white px-4 py-5 shadow sm:p-6 dark:bg-gray-800 ${actions ? 'sm:rounded-tl-md sm:rounded-tr-md' : 'sm:rounded-md'}`}>
              <div className="grid grid-cols-6 gap-6">{form}</div>
            </div>
            {actions && (
              <div className="flex items-center justify-end bg-gray-50 px-4 py-3 text-end shadow sm:rounded-br-md sm:rounded-bl-md sm:px-6 dark:bg-gray-800">
                {actions}
              </div>
            )}
          </form>
        </div>
      </div>
    );
  }
}

export default FormSection;
