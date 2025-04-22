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
  else if (path.includes('admin-page')) {
    showMenuIcon = false;
  }
  return showMenuIcon;
}

export function getTrailingWhiteSpace(count) {
  const stringTrailingWhiteSpace = '\u00A0';
  return Array(count).fill(stringTrailingWhiteSpace).join('');
}

export function getRawTextFromXML(xml) {
  let returnText = '';
  if (!xml) return returnText;

  const length = xml.length;
  const flagStart = '<',
    flagEnd = '</';
  //Need to find the first occurence of < and last occurence of </ to get the correct xml

  let start = xml.indexOf(flagStart);
  let end = xml.lastIndexOf(flagEnd);

  if (start === -1 || end === -1) {
    return xml;
  }

  while (end < length && xml[end] !== '>') {
    end++;
  }
  if (end < length) end++;

  returnText += xml.substring(0, start);
  let newXml = xml.substring(start, end);

  const validation = XMLValidator.validate(newXml);
  if (validation !== true) {
    //In case if xml is not valid, then simply skip the value between < and >

    for (let index = start; index < end; index++) {
      const charValue = xml[index];
      if (charValue === '<') {
        while (index < length && xml[index] !== '>') {
          index++;
        }
      } else returnText += charValue;
    }
    return returnText;
  }

  const doc = new DOMParser().parseFromString(newXml, 'text/xml');
  returnText += doc.documentElement.textContent;

  if (end < length) returnText += xml.substring(end, length);

  return returnText;
}
