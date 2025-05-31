import { clsx } from "clsx";
import { twMerge } from "tailwind-merge";
function cn(...inputs) {
  return twMerge(clsx(inputs));
}
const addDays = (dateValue, days) => {
  const date = new Date(dateValue.valueOf());
  date.setDate(date.getDate() + days);
  return date;
};
const isNotEmpty = (value) => {
  return !!value && value.trim().length > 0;
};
const subtractFloats = (a, b, precision = 2) => {
  const result = a - b;
  const factor = Math.pow(10, precision);
  return Math.round(result * factor) / factor;
};
export {
  addDays as a,
  cn as c,
  isNotEmpty as i,
  subtractFloats as s
};
