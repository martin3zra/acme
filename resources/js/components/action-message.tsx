import { Transition } from '@headlessui/react';
import React from 'react';

interface Props extends React.ComponentProps<'div'> {
  show: boolean;
  children: React.ReactNode;
}
export default function ActionMessage({ show, className, children }: Props) {
  return (
    <div className={className}>
      <Transition show={show} enter="transition  ease-in-out" enterFrom="opacity-0" leave="transition ease-in-out" leaveTo="opacity-0">
        <p className="text-sm text-neutral-600">{children}</p>
      </Transition>
    </div>
  );
}
