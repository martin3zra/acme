import { usePage } from '@inertiajs/react';

type Resource = {
  [key: string]: boolean;
};
export function useGate() {
  const { props } = usePage();
  const usingWildcard = (ability: string = ''): boolean => {
    const abilities = props.abilities as Resource;
    if (
      (typeof window === 'undefined' || typeof window._abilities !== 'object' || window._abilities === null) &&
      abilities &&
      Object.keys(abilities).length === 0
    )
      return false;

    if (abilities && Object.keys(abilities).length === 0) return false;

    if (abilities['*'] || abilities[ability]) {
      return true;
    }

    // Split the requested ability into parts
    const parts = ability.split(':');

    // Test prefix wildcard: e.g., `view:invoice:*`
    for (let i = 1; i < parts.length; i++) {
      const prefixWildcard = `${parts.slice(0, i).join(':')}:*`;
      if (abilities[prefixWildcard]) {
        return true;
      }
    }

    // Check suffix wildcards: e.g. *:invoice:summary
    for (let i = 1; i < parts.length; i++) {
      const suffixWildcard = `*:${parts.slice(i).join(':')}`;
      if (abilities[suffixWildcard]) {
        return true;
      }
    }

    return false;
  };

  return {
    can: (ability: string = '', resource: Resource = {}): boolean => {
      if (Object.keys(resource).length > 0 && resource[ability]) {
        return resource[ability];
      }
      return usingWildcard(ability);
    },
    canAny: (abilities: string[] = [], resource: Resource = {}): boolean => {
      if (Object.keys(resource).length > 0) {
        const values: boolean[] = [];
        abilities.forEach((ability) => {
          if (resource[ability]) {
            values.push(resource[ability]);
          }
        });

        return Object.values(values).every((value: boolean) => value);
      }

      const result: boolean[] = [];
      abilities.forEach((ability) => {
        result.push(usingWildcard(ability));
      });

      return Object.values(result).some((value: boolean) => value);
    },
  };
}
