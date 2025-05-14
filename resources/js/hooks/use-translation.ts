import { usePage } from '@inertiajs/react';

type Replacements = { [key: string]: string };

export const useTranslation = () => {
  const { props } = usePage<{ translations: { [key: string]: string } }>();
  const translations = props.translations || {};

  return {
    trans: (key: string, replacements: Replacements = {}): string => {
      let translation = translations[key] || key;

      Object.keys(replacements).forEach((k) => {
        translation = translation.replace(`:${k}`, replacements[k]);
      });

      return translation;
    },
  };
};
