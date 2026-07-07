type AttributeValueMap = Record<string, unknown[]>;

type VariantCombination<T extends AttributeValueMap> = {
  [K in keyof T]: T[K][number];
};

export const generateVariantCombinations = <T extends AttributeValueMap>(attributes: T): Array<VariantCombination<T>> => {
  const entries = Object.entries(attributes) as Array<[keyof T, T[keyof T]]>;

  if (entries.length === 0) {
    return [];
  }

  if (entries.some(([, values]) => values.length === 0)) {
    return [];
  }

  let combinations: Array<Partial<VariantCombination<T>>> = [{}];

  for (const [attribute, values] of entries) {
    const nextCombinations: Array<Partial<VariantCombination<T>>> = [];

    for (const combination of combinations) {
      for (const value of values) {
        nextCombinations.push({
          ...combination,
          [attribute]: value as VariantCombination<T>[typeof attribute],
        });
      }
    }

    combinations = nextCombinations;
  }

  return combinations as Array<VariantCombination<T>>;
};
