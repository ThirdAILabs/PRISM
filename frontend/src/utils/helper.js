import { useLocation } from 'react-router-dom';
import { XMLValidator } from 'fast-xml-parser';
export function GetShowMenuIcon() {
  const location = useLocation();
  let showMenuIcon = true;
  const path = location.pathname;
  if (path.includes('report')) {
    showMenuIcon = false;
  } else if (path.includes('error')) {
    showMenuIcon = false;
  }
  return showMenuIcon;
}

export function getTrailingWhiteSpace(count) {
  const stringTrailingWhiteSpace = '\u00A0';
  return Array(count).fill(stringTrailingWhiteSpace).join('');
}

export function getRawTextFromXML(xml) {
  console.log('consoling the xml', xml);
  let returnText = '';
  if (!xml) return returnText;
  // let xmlStart = false;
  // let newXML = "";
  // for (let index = 0; index < xml.length; index++) {
  //   const charValue = xml[index];
  //   if (charValue === '<') {
  //     xmlStart = true;
  //   }
  //   if (xmlStart && charValue === '>') {
  //     newXML += charValue;
  //   }
  //   else {
  //     returnText += charValue;
  //   }
  // }
  // const validation = XMLValidator.validate(newXML);
  // if (validation !== true) {
  //   return xml;
  // }

  // const doc = new DOMParser().parseFromString(newXML, 'text/xml');
  // returnText += doc.documentElement.textContent;
  for (let index = 0; index < xml.length; index++) {
    const charValue = xml[index];
    if (charValue === '<') {
      while (index < xml.length && xml[index] !== '>') {
        index++;
      }
    } else returnText += charValue;
  }
  return returnText;
}
