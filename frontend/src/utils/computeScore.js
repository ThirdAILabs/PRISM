const MULTI_AFFIL = 'oa_multi_affil';
const COAUTHOR_AFFIL_EOC = 'oa_coauthor_affiliation_eoc';
const COAUTHOR_EOC = 'oa_coauthor_eoc';
const FUNDER_EOC = 'oa_funder_eoc';
const ACK_EOC = 'oa_acknowledgement_eoc';
const AUTHOR_AFFIL_EOC = 'oa_author_affiliation_eoc';

// TODO: doj_press_release_eoc

function score(flags) {
  const counts = {};
  for (const flag of flags) {
    if (!counts[flag.flagger_id]) {
      counts[flag.flagger_id] = 0;
    }
    counts[flag.flagger_id] += 1;
  }
  let score = 0;
  if (counts[MULTI_AFFIL]) {
    score = 2;
  }
  if (counts[COAUTHOR_AFFIL_EOC] || counts[COAUTHOR_EOC]) {
    score = 3;
  }
  if (counts[FUNDER_EOC] || counts[ACK_EOC]) {
    score = 4;
  }
  if (counts[AUTHOR_AFFIL_EOC]) {
    score = 5;
  }
  return [score, counts];
}
