import { Verb } from '@/types';

export function useVerb() {
  return { action };
}

function action(verb: Verb): string {
  return {
    view: 'view',
    edit: 'update',
    trash: 'trash',
    create: 'create',
  }[verb];
}
