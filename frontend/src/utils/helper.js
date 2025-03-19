import { useLocation } from 'react-router-dom';

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
  let returnText = '';
  if (!xml) return returnText;

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
