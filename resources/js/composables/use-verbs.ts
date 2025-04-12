import { Verb } from "@/types";

export function useVerb() {
    return {action}
}

function action(verb: Verb): string {
    return {
        view: 'View',
        edit: 'Update',
        trash: 'Trash',
        create: 'Create',
      }[verb];
}