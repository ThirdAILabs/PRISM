export const TALENT_CONTRACTS = 'TalentContracts';
export const ASSOCIATIONS_WITH_DENIED_ENTITIES = 'AssociationsWithDeniedEntities';
export const HIGH_RISK_FUNDERS = 'HighRiskFunders';
export const AUTHOR_AFFILIATIONS = 'AuthorAffiliations';
export const POTENTIAL_AUTHOR_AFFILIATIONS = 'PotentialAuthorAffiliations';
export const MISC_HIGH_RISK_AFFILIATIONS = 'MiscHighRiskAssociations';
export const COAUTHOR_AFFILIATIONS = 'CoauthorAffiliations';

// Deprecated
export const MULTI_AFFIL = 'oa_multi_affil';
export const FUNDER_EOC = 'oa_funder_eoc';
export const PUBLISHER_EOC = 'oa_publisher_eoc'; // isn't used
export const COAUTHOR_EOC = 'oa_coauthor_eoc';
export const COAUTHOR_AFFIL_EOC = 'oa_coauthor_affiliation_eoc';
export const AUTHOR_AFFIL_EOC = 'oa_author_affiliation_eoc';
export const ACK_EOC = 'oa_acknowledgement_eoc';
export const FORMAL_RELATIONS = 'formal_relations';
export const UNI_FACULTY_EOC = 'uni_faculty_eoc';
export const DOJ_PRESS_RELEASES_EOC = 'doj_press_release_eoc';

export const TitlesAndDescriptions = {
  [TALENT_CONTRACTS]: {
    title: 'Talent Contracts',
    desc: 'These papers are funded by talent programs run by foreign governents that intend to recruit high-performing researchers.',
  },
  [ASSOCIATIONS_WITH_DENIED_ENTITIES]: {
    title: 'Funding from Denied Entities',
    desc: 'Some of the parties involved in these papers are in denied entity lists of government agencies.',
  },
  [HIGH_RISK_FUNDERS]: {
    title: 'High Risk Funding Sources',
    desc: 'These papers are funded by sources that have close ties to high-risk foreign governments.',
  },
  [AUTHOR_AFFILIATIONS]: {
    title: 'Affiliations with High Risk Foreign Institutes',
    desc: 'These papers list the queried author as being affiliated with a high-risk foreign institute.',
  },
  [POTENTIAL_AUTHOR_AFFILIATIONS]: {
    title: 'Appointments at High Risk Foreign Institutes*',
    desc: 'The author may have an appointment at a high-risk foreign institute.\n\n*Collated information from the web, might contain false positives.',
  },
  [MISC_HIGH_RISK_AFFILIATIONS]: {
    title: 'Miscellaneous High Risk Connections*',
    desc: 'The author or an associate may be mentioned in a press release.\n\n*Collated information from the web, might contain false positives.',
  },
  [COAUTHOR_AFFILIATIONS]: {
    title: "Co-authors' affiliations with High Risk Foreign Institutes",
    desc: 'Co-authors in these papers are affiliated with high-risk foreign institutes.',
  },
};
