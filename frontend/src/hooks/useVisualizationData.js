import { useEffect, useMemo, useState } from "react";
import {
    AUTHOR_AFFIL_EOC,
    FUNDER_EOC,
    PUBLISHER_EOC,
    ACK_EOC,
    COAUTHOR_AFFIL_EOC,
    COAUTHOR_EOC,
    MULTI_AFFIL,
    UNI_FACULTY_EOC,
    DOJ_PRESS_RELEASES_EOC
} from "../constants/constants";

import { talentPrograms, deniedEntities, fundingSources, institutions } from "../constants/eoc_categories";

const FOREIGN_TALENT_PROGRAMS = "foreign_talent_programs";
const DENIED_ENTITIES = "denied_entities";
const HIGH_RISK_FUNDING_SOURCES = "high_risk_funding_sources";
const HIGH_RISK_FOREIGN_INSTITUTIONS = "high_risk_foreign_institutions";
const HIGH_RISK_APPOINTMENTS = "high_risk_appointments";
const UNIVERSITY_FACULTY_APPOINTMENTS = "university_faculty_appointments";
const DOJ_PRESS_RELEASES = "doj_press_releases";

export function useVisualizationData(name, idToFlags, formalRelations, worksCount, startYear, endYear) {
    const [affiliations, setAffiliations] = useState({});

    const [weights, setWeights] = useState({
        [FOREIGN_TALENT_PROGRAMS]: 1.0,
        [DENIED_ENTITIES]: 1.0,
        [HIGH_RISK_FUNDING_SOURCES]: 1.0,
        [HIGH_RISK_FOREIGN_INSTITUTIONS]: 1.0,
        [HIGH_RISK_APPOINTMENTS]: 1.0,
        [UNIVERSITY_FACULTY_APPOINTMENTS]: 1.0,
        [DOJ_PRESS_RELEASES]: 1.0
    })

    function addAffiliations(workAffiliations) {
        // console.log(`workAffiliations -> ${workAffiliations}`);
        // If keys exist in prev, keep value from prev.
        setAffiliations(prev => {
            const newAffils = { ...prev };
            for (const affil of workAffiliations) {
                if (newAffils[affil]) {
                    const [count, show] = newAffils[affil];
                    newAffils[affil] = [count, show];
                } else {
                    newAffils[affil] = [1, true];
                }
            }
            return newAffils;
        });
    }

    const numHighRiskAppointments = useMemo(() => {
        const appointments = new Set();
        for (const rel of formalRelations) {
            appointments.add(rel.institution);
        }
        for (const flag of (idToFlags[AUTHOR_AFFIL_EOC] || [])) {
            for (const inst of flag.Institutions) {
                appointments.add(inst);
            }
        }
        return appointments.length;
    }, [idToFlags, formalRelations]);

    const flagByType = useMemo(() => {
        const foreignTalentProgramFlags = [];
        const deniedEntityFlags = [];
        const highRiskFundingSourceFlags = [];
        const universityFacultyFlags = [];
        const dojPRFlags = [];

        function fromList(list, flag) {
            return (
                flag.Entities
                    .flatMap((entity) => entity['lists'])
                    .reduce((prev, curr) => (list.includes(curr) || prev), false)
            );
        }

        for (const flag of (idToFlags[ACK_EOC] || [])) {
            if (flag.FlagMessage.toLowerCase().includes("talent") || fromList(talentPrograms, flag)) {
                foreignTalentProgramFlags.push(flag);
            } else if (fromList(deniedEntities, flag)) {
                deniedEntityFlags.push(flag);
            } else {
                highRiskFundingSourceFlags.push(flag);
            }
        }

        const seenUniURLs = new Set();
        for (const flag of (idToFlags[UNI_FACULTY_EOC] || [])) {
            if (flag.FlagMessage.toLowerCase().includes("concerning entity")) {
                const uniURL = flag.UniversityUrl;

                if (!seenUniURLs.has(uniURL)) {
                    universityFacultyFlags.push(flag);
                    seenUniURLs.add(uniURL);
                }
            }
        }

        const seenDoJArticles = new Set();
        for (const flag of (idToFlags[DOJ_PRESS_RELEASES_EOC] || [])) {
            if (flag.FlagMessage.toLowerCase().includes("press release")) {
                const dojArticle = flag.DocTitle;

                if (!seenDoJArticles.has(dojArticle)) {
                    dojPRFlags.push(flag);
                    seenDoJArticles.add(dojArticle);
                }
            }
        }
    
        return {
            [FOREIGN_TALENT_PROGRAMS]: foreignTalentProgramFlags,
            [DENIED_ENTITIES]: deniedEntityFlags,
            [HIGH_RISK_FUNDING_SOURCES]: highRiskFundingSourceFlags.concat(idToFlags[FUNDER_EOC] || []),
            [HIGH_RISK_FOREIGN_INSTITUTIONS]: idToFlags[COAUTHOR_AFFIL_EOC] || [],
            [HIGH_RISK_APPOINTMENTS]: (formalRelations || []).concat(idToFlags[AUTHOR_AFFIL_EOC] || []),
            [UNIVERSITY_FACULTY_APPOINTMENTS]: universityFacultyFlags,
            [DOJ_PRESS_RELEASES]: dojPRFlags
        }
    }, [idToFlags, formalRelations]);

    const sortedAffiliations = Object.keys(affiliations).toSorted((a, b) => affiliations[b][0] - affiliations[a][0]);
    const showAffiliations = Object.fromEntries(sortedAffiliations.map(affil => [affil, affiliations[affil][1]]))

    function toData(key, name, desc, num = undefined) {
        flagByType[key].forEach(item => console.log(item))
        let flags;
        if ((key === UNIVERSITY_FACULTY_APPOINTMENTS) || (key === DOJ_PRESS_RELEASES)) {
            flags = flagByType[key];
        } else {
            flags = (
                // flagByType[key]
                // // Filter by year
                // .filter((flag) => !flag.year || (((startYear || -Infinity) <= flag.year) && (flag.year <= (endYear || Infinity))))
                // // Filter by affiliation
                // .filter((flag) => (flag.affiliations || []).map(affil => showAffiliations[affil]).reduce((prev, curr) => prev || curr, false))
                flagByType[key]
                    .filter((flag) =>
                        !flag.affiliations?.length ||
                        (flag.affiliations || []).map(affil => showAffiliations[affil]).reduce((prev, curr) => prev || curr, false)
                    )
            );
        }

        // flags.forEach(item => console.log(item));
        return {
            "display_name": name,
            "desc": desc,
            "weight": weights[key],
            "details": flags,
            "num": num || flags.length,
        };
    }

    const data = {
        "name": name,
        "total_score": Object.keys(flagByType).map(flagType => weights[flagType] * flagByType[flagType].length).reduce((prev, curr) => prev + curr, 0),
        "works_count": worksCount,
        [FOREIGN_TALENT_PROGRAMS]: toData(
            FOREIGN_TALENT_PROGRAMS,
            "Papers with foreign talent programs",
            "Authors in these papers are recruited by talent programs that have close ties to high-risk foreign governments"),
        [DENIED_ENTITIES]: toData(
            DENIED_ENTITIES,
            "Papers with denied entities",
            "Some of the parties involved in these works are in the denied entity lists of U.S. government agencies"),
        [HIGH_RISK_FUNDING_SOURCES]: toData(
            HIGH_RISK_FUNDING_SOURCES,
            "Papers with high-risk funding sources",
            "These papers is funded by funding sources that have close ties to high-risk foreign governments"),
        [HIGH_RISK_FOREIGN_INSTITUTIONS]: toData(
            HIGH_RISK_FOREIGN_INSTITUTIONS,
            "Papers with high-risk foreign institutions",
            "Coauthors in these papers are affiliated with high-risk foreign institutions"),
        [HIGH_RISK_APPOINTMENTS]: toData(
            HIGH_RISK_APPOINTMENTS,
            "High-risk appointments at foreign institutions",
            "Papers that list the current author as being affiliated with a high-risk foreign institution or web pages that showcase official appointments at high-risk foreign institutions",
            numHighRiskAppointments),
        [UNIVERSITY_FACULTY_APPOINTMENTS]: toData(
            UNIVERSITY_FACULTY_APPOINTMENTS,
            "Potential high-risk appointments at foreign institutions",
            "The author may be affiliated with high-risk foreign institutions"),
        [DOJ_PRESS_RELEASES]: toData(
            DOJ_PRESS_RELEASES,
            "Miscellaneous potential high-risk associations",
            "The author or an associate may be mentioned in a press release")
    }

    function AffiliationChecklist(props) {
        return <div style={{ fontSize: "14px", marginTop: "10px", display: 'flex', width: "300px", flexDirection: 'column', position: "absolute", padding: "10px", right: 0, backgroundColor: "#323232", borderRadius: "10px", zIndex: 100 }}>
            <div style={{ maxHeight: "500px", overflowX: "scroll" }}>
                {sortedAffiliations.map((affiliation, index) => <div style={{ display: 'flex', flexDirection: 'row', justifyContent: 'flex-start', margin: '5px' }}>
                    <input key={index} style={{ height: "12px", width: "12px", marginRight: '3px', marginTop: '5px' }} type="checkbox" checked={affiliations[affiliation][1]} onClick={() => {
                        setAffiliations(prev => ({ ...prev, [affiliation]: [prev[affiliation][0], !prev[affiliation][1]] }));
                    }} />
                    <text style={{ textAlign: "left", marginLeft: "10px" }}>
                        {affiliation}
                    </text>
                </div>)}
            </div>
            <button style={{ marginTop: "15px", backgroundColor: "#444444", outline: "none", border: "none", padding: "5px", borderRadius: "5px", color: "white", fontWeight: 'bold' }} onClick={() => { setAffiliations(prev => Object.fromEntries(Object.keys(affiliations).map(affil => [affil, [affiliations[affil][0], false]]))) }}>Clear selections</button>
            <button style={{ marginTop: "10px", backgroundColor: "#444444", outline: "none", border: "none", padding: "5px", borderRadius: "5px", color: "white", fontWeight: 'bold' }} onClick={() => { setAffiliations(prev => Object.fromEntries(Object.keys(affiliations).map(affil => [affil, [affiliations[affil][0], true]]))) }}>Select all</button>
        </div>
    }

    return {
        data,
        addAffiliations,
        setWeight: (key, newWeight) => setWeights(prev => ({ ...prev, [key]: newWeight })),
        AffiliationChecklist
    }
}