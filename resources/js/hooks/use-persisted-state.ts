import { useEffect, useState } from 'react';
import { useLocalStorage } from './use-local-storage';

export function usePersistedState<T>(key: string, initialValue: T) {
  const { getItem, setItem, removeItem: deleteItem } = useLocalStorage(key);
  const [value, setValue] = useState<T>(() => {
    const item = getItem();
    return (item as T) || initialValue;
  });

  const removeItem = () => {
    deleteItem();
  };

  useEffect(() => {
    setItem(value);
  }, [value, setItem]);

  return [value, setValue, removeItem] as const;
}
