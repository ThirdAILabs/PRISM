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
  const validation = XMLValidator.validate(xml);
  if (validation !== true) {
    return;
  }
  const doc = new DOMParser().parseFromString(xml, 'text/xml');
  return doc.documentElement.textContent;
}
