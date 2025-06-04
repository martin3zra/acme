import { usePage } from "@inertiajs/react";
const useTranslation = () => {
  const { props } = usePage();
  const translations = props.translations || {};
  return {
    trans: (key, replacements = {}) => {
      let translation = translations[key] || key;
      Object.keys(replacements).forEach((k) => {
        translation = translation.replace(`:${k}`, `${replacements[k]}`);
      });
      return translation;
    }
  };
};
export {
  useTranslation as u
};
