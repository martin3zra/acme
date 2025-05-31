import { useState, useEffect } from "react";
const useDebounced = (value, delay) => {
  const [debouncedValue, setDebouncedValue] = useState(value);
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      setDebouncedValue(value);
    }, delay);
    return () => {
      clearTimeout(timeoutId);
    };
  }, [value, delay]);
  return debouncedValue;
};
export {
  useDebounced as u
};
