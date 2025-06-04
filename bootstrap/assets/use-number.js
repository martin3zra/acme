function useNumber() {
  return { currency };
}
function currency(value, precision = 2, inCent = false) {
  const formatter = new Intl.NumberFormat("en-US", {
    minimumFractionDigits: precision,
    maximumFractionDigits: precision
  });
  if (typeof value === "string") {
    return `$${formatter.format(inCent ? parseInt(value) / 100 : Number(value))}`;
  }
  return `$${formatter.format(inCent ? value / 100 : value)}`;
}
export {
  useNumber as u
};
