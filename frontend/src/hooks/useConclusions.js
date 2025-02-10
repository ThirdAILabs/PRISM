import React, { useEffect, useMemo, useRef, useState } from "react"
import {
  AUTHOR_AFFIL_EOC,
  COAUTHOR_AFFIL_EOC,
  ACK_EOC,
  FORMAL_RELATIONS
} from "../constants/constants";
import apiService from "../services/apiService";


export function useConclusions(author, flags, threshold) {
  const checkedInstitutions = useRef([]);
  const [formalRelations, setFormalRelations] = useState([]);
  const [numWaiting, setNumWaiting] = useState(0);

  const institutionFrequencies = useMemo(() => {
    console.log("Calculating institution freqs");
    let frequencies = {};
    const flagsWithAffils = [
      ...(flags[AUTHOR_AFFIL_EOC] || []),
      ...(flags[COAUTHOR_AFFIL_EOC] || []),
    ];
    for (const flag of flagsWithAffils) {
      if (!frequencies[flag.metadata.affiliations]) {
        frequencies[flag.metadata.affiliations] = 0;
      }
      frequencies[flag.metadata.affiliations] += 1;
    }
    return frequencies;
  }, [flags]);

  const foreignFundingFrequency = useMemo(() => (flags[ACK_EOC] || []).length, [flags]);

  const funderFrequencies = useMemo(() => {
    console.log("Calculating funder freqs");
    let frequencies = {};
    for (const flag of (flags[ACK_EOC] || [])) {
      for (const entity of flag.metadata.entities) {
        const name = entity["aliases"][0];
        if (!frequencies[name]) {
          frequencies[name] = 0;
        }
        frequencies[name] += 1;
      }
    }
    return frequencies;
  }, [flags]);

  function papers(quantity) {
    if (quantity > 1) {
      return "papers";
    }
    return "paper";
  }

  const summary = useMemo(() => {
    console.log("Calculating summary");
    const messages = [];
    const numUniqueInstitutions = Object.keys(institutionFrequencies).length;
    if (numUniqueInstitutions) {
      messages.push(`${numUniqueInstitutions} unique institutions of concern are listed across ${author}'s ${papers(numUniqueInstitutions)}.`);
    }
    const numUniqueFunders = Object.keys(funderFrequencies).length;
    if (numUniqueFunders) {
      messages.push(`${numUniqueFunders} unique funders of concern are listed across ${author}'s ${papers(numUniqueFunders)}.`);
    }
    if (foreignFundingFrequency) {
      messages.push(`${foreignFundingFrequency} ${papers(foreignFundingFrequency)} received funding from foreign sources.`);
    }
    Object.entries(institutionFrequencies).sort(([_inst1, freq1], [_inst2, freq2]) => (freq2 - freq1)).slice(0, 3).forEach(([inst, freq]) => {
      messages.push(`${inst} appears as an affiliation in ${freq} ${papers(freq)}.`);
    });
    Object.entries(funderFrequencies).sort(([_inst1, freq1], [_inst2, freq2]) => (freq2 - freq1)).slice(0, 3).forEach(([inst, freq]) => {
      messages.push(`${inst} appears in ${freq} ${papers(freq)}.`);
    });
    return messages;
  }, [institutionFrequencies, funderFrequencies, foreignFundingFrequency]);

  return {
    conclusions: summary.concat(formalRelations.map(rel => rel.message)),
    formalRelations,
    loading: !!numWaiting,
  };
}