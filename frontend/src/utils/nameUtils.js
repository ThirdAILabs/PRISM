function givenLastSplits(name) {
  const names = name.split(' ');
  return Array.from({ length: names.length }, (_, i) => [
    names.slice(0, -i - 1),
    names.slice(-i - 1),
  ]);
}

function withAndWithoutEachGivenName(names) {
  let variations = [names];
  for (let i = 0; i < names.length; i++) {
    const newVariations = [];
    for (const variation of variations) {
      const newVariation = [...variation];
      newVariation[i] = null;
      newVariations.push(newVariation);
    }
    variations = variations.concat(newVariations);
  }
  variations = variations.map((variation) => variation.filter(Boolean));
  variations = variations.filter((variation) => variation.length > 0);
  return variations;
}

function useInitials(name) {
  return name[0] + '.';
}

function joinNames(names) {
  let last = names[0];
  const parts = [names[0]];
  for (let i = 1; i < names.length; i++) {
    const name = names[i];
    const lastIsInitial = last.length === 1 || (last.length === 2 && last[1] === '.');
    const nameIsInitial = name.length === 1 || (name.length === 2 && name[1] === '.');
    if (!(lastIsInitial && nameIsInitial)) {
      parts.push(' ');
    }
    parts.push(name);
    last = name;
  }
  return parts.join('');
}

function initialsVariations(names) {
  const beginningInitials = Array.from({ length: names.length - 1 }, (_, i) =>
    joinNames([...names.slice(0, i).map(useInitials), ...names.slice(i)])
  );
  const endingInitials = Array.from({ length: names.length - 1 }, (_, i) =>
    joinNames([...names.slice(0, -i - 1), ...names.slice(-i - 1).map(useInitials)])
  );
  const allInitials = names.map(useInitials).join('');
  return [...beginningInitials, ...endingInitials, allInitials, joinNames(names)];
}

export function makeVariations(authorName) {
  return givenLastSplits(authorName).flatMap(([givenName, lastName]) =>
    withAndWithoutEachGivenName(givenName).flatMap((givenNameSet) =>
      initialsVariations(givenNameSet).map((givenNameVariation) => [
        givenNameVariation,
        lastName.join(' '),
      ])
    )
  );
}

export const levenshteinDistance = (str1, str2) => {
  const track = Array(str2.length + 1)
    .fill(null)
    .map(() => Array(str1.length + 1).fill(null));
  for (let i = 0; i <= str1.length; i += 1) {
    track[0][i] = i;
  }
  for (let j = 0; j <= str2.length; j += 1) {
    track[j][0] = j;
  }
  for (let j = 1; j <= str2.length; j += 1) {
    for (let i = 1; i <= str1.length; i += 1) {
      const indicator = str1[i - 1] === str2[j - 1] ? 0 : 1;
      track[j][i] = Math.min(
        track[j][i - 1] + 1, // deletion
        track[j - 1][i] + 1, // insertion
        track[j - 1][i - 1] + indicator // substitution
      );
    }
  }
  return track[str2.length][str1.length];
};
