import React, { useEffect, useRef, useState } from 'react';
import RelationGraph from 'relation-graph-react';
import '../../../styles/components/_graph2.scss';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  Divider,
  Card,
  CardContent,
  Typography,
  Collapse,
  IconButton,
  Link,
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import { AUTHOR_AFFILIATIONS } from '../../../constants/constants.js';
import { getRawTextFromXML } from '../../../utils/helper.js';

function getNodeTitle(flagType, flag) {
  if (flagType == AUTHOR_AFFILIATIONS) {
    return flag.Affiliations[0];
  } else if (flag.Work) {
    return flag.Work.DisplayName;
  } else if (flag.University) {
    return flag.University;
  } else if (flag.DocTitle) {
    return flag.DocTitle;
  }
  return '';
}

function getNodeUrl(flag) {
  if (flag.Work) {
    return flag.Work.WorkUrl;
  } else if (flag.UniversityUrl) {
    return flag.UniversityUrl;
  } else if (flag.DocUrl) {
    return flag.DocUrl;
  }
  return '';
}

function convertDataToGraphFormat(authorName, reportContent) {
  const flagTypeToTitle = {
    AssociationsWithDeniedEntities: 'Funding from Denied Entities',
    AuthorAffiliations: 'Author Affiliations',
    CoauthorAffiliations: "Co-authors' affiliations with High Risk Foreign Institutes",
    HighRiskFunders: 'Affiliations with High Risk Foreign Institutes',
    MiscHighRiskAssociations: 'Miscellaneous High Risk Connections',
    PotentialAuthorAffiliations: 'Appointments at High Risk Foreign Institutes',
    TalentContracts: 'Talent Contracts',
  };

  let graphData = {
    name: authorName,
    connections: [],
  };

  let riskScore = 0;
  Object.keys(reportContent).forEach((flagType, index) => {
    riskScore += reportContent[flagType].length;
    let works = [];

    Object.keys(reportContent[flagType]).forEach((flag, index2) => {
      if (reportContent[flagType][index2].hasOwnProperty('Connections')) {
        // This is a connection with potentially more connections out of it
        if (reportContent[flagType][index2].Connections !== null) {
          // secondary or tertiary connection
          let work = {
            title: reportContent[flagType][index2].Connections[0].DocTitle,
            url: reportContent[flagType][index2].Connections[0].DocUrl,
            connections: [],
          };

          if (
            reportContent[flagType][index2].hasOwnProperty('EntityMentioned') &&
            !reportContent[flagType][index2].FrequentCoauthor
          ) {
            work.title += ' [' + reportContent[flagType][index2].EntityMentioned + ']';
          }

          if (reportContent[flagType][index2].Connections.length > 1) {
            // tertiary connection
            work.connections.push({
              title: reportContent[flagType][index2].Connections[1].DocTitle,
              url: reportContent[flagType][index2].Connections[1].DocUrl,
              connections: [
                {
                  title: reportContent[flagType][index2].DocTitle,
                  url: reportContent[flagType][index2].DocUrl,
                  connections: [],
                },
              ],
            });
          } else {
            work.connections.push({
              title: reportContent[flagType][index2].DocTitle,
              url: reportContent[flagType][index2].DocUrl,
              connections: [],
            });
          }
          works.push(work);
        } else {
          // primary connection
          works.push({
            title: reportContent[flagType][index2]?.DocTitle || '',
            url: reportContent[flagType][index2]?.DocUrl || '',
            connections: [],
          });
        }
      } else {
        works.push({
          title: getNodeTitle(flagType, reportContent[flagType][index2]),
          url: getNodeUrl(reportContent[flagType][index2]),
          connections: [],
        });
      }
    });

    graphData.connections.push({
      title: flagTypeToTitle[flagType],
      count: works.length,
      connections: works,
    });
  });

  graphData['risk_score'] = riskScore;

  return graphData;
}

function levelZeroNode(nodeId, name, riskScore, url) {
  return {
    id: nodeId,
    data: {
      url: url,
      text: name,
    },
    text: name,
  };
}

function levelOneNode(nodeId, title, url, count, connections) {
  const textLength = Math.min(title.length, 40);
  let dotString = '';
  if (textLength < title.length) dotString += '...';
  return {
    id: nodeId,
    text: title.slice(0, textLength) + dotString + ' (' + count.toString() + ')',
    data: {
      url: url,
      text: title,
      allConnections: connections,
    },
  };
}

function levelTwoNode(nodeId, title, url, connections) {
  const textLength = Math.min(title.length, 40);
  let dotString = '';
  if (textLength < title.length) dotString += '...';
  return {
    id: nodeId,
    text: title.slice(0, textLength) + dotString,
    data: {
      url: url,
      text: title,
      allConnections: connections,
    },
  };
}

function levelThreeNode(nodeId, title, url, connections) {
  const textLength = Math.min(title.length, 30);
  let dotString = '';
  if (textLength < title.length) dotString += '...';
  const fontSize = 20;
  const calculatedWidth = Math.max(textLength * (fontSize * 0.25), 50);
  return {
    id: nodeId,
    text: title.slice(0, textLength) + dotString,
    data: {
      url: url,
      text: title,
      allConnections: connections,
    },
  };
}

function levelFourNode(nodeId, title, url, connections) {
  const textLength = Math.min(title.length, 30);
  let dotString = '';
  if (textLength < title.length) dotString += '...';
  return {
    id: nodeId,
    text: title.slice(0, textLength) + dotString,
    data: {
      url: url,
      text: title,
      allConnections: connections,
    },
  };
}

function generateGraphData(data, parentId = null, level = 0) {
  if (level === 1 && data.count === 0) {
    return { nodes: [], lines: [], rootId: 'a', levelNodePairs: [] };
  }

  const nodes = [];
  const lines = [];
  let currentNodeId;

  if (level === 0) {
    // Root node
    currentNodeId = 'a';
    nodes.push(levelZeroNode(currentNodeId, data.name, data.risk_score, data.url));
  } else if (level === 1) {
    // Second level node
    currentNodeId = `g${Math.random().toString(36).substring(7)}`;
    nodes.push(levelOneNode(currentNodeId, data.title, data.url, data.count, data.connections));
  } else if (level === 2) {
    // Third level node
    currentNodeId = `e${Math.random().toString(36).substring(7)}`;
    nodes.push(levelTwoNode(currentNodeId, data.title, data.url, data.connections));
  } else if (level === 3) {
    // Fourth level node
    currentNodeId = `b${Math.random().toString(36).substring(7)}`;
    nodes.push(levelThreeNode(currentNodeId, data.title, data.url, data.connections));
  } else if (level === 4) {
    currentNodeId = `h${Math.random().toString(36).substring(7)}`;
    nodes.push(levelFourNode(currentNodeId, data.title, data.url, data.connections));
  }

  if (parentId) {
    lines.push({ id: `${parentId}-${currentNodeId}`, from: parentId, to: currentNodeId });
  }

  let levelNodePairs = [[level, currentNodeId]];

  if (data.connections && data.connections.length > 0) {
    let limit = data.connections.length;
    if (level > 0) {
      limit = Math.min(limit, 5);
    }
    for (let connectionIndex = 0; connectionIndex < limit; connectionIndex++) {
      const child = data.connections[connectionIndex];
      const childGraph = generateGraphData(child, currentNodeId, level + 1);
      nodes.push(...childGraph.nodes);
      lines.push(...childGraph.lines);
      levelNodePairs = [...levelNodePairs, ...childGraph.levelNodePairs];
    }
  }

  return { nodes, lines, rootId: 'a', levelNodePairs };
}

const MyComponent = ({ authorName, reportContent }) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);
  const [expandedChildren, setExpandedChildren] = useState(false);

  const graphRef = useRef(null);

  useEffect(() => {
    showGraph();
  }, []);
  const graphData = generateGraphData(convertDataToGraphFormat(authorName, reportContent));
  const { nodes, lines, rootId } = graphData;
  const showGraph = async () => {
    const __graph_json_data = {
      rootId: rootId,
      nodes: nodes,
      lines: lines,
    };
    // const __graph_json_data = {
    //   rootId: '0',
    //   nodes: [
    //     { id: '0', text: 'Semiconductor' },
    //     { id: '1', text: 'Final Test' },
    //     { id: '2', text: 'Traditional Packaging' },
    //     { id: '3', text: 'COG (Glass Encapsulation)' },
    //     { id: '5', text: 'TO (Coaxial Encapsulation)' },
    //     { id: '8', text: 'BGA (Ball Grid Array Encapsulation)' },
    //     { id: '10', text: 'QFN (Quad Flat No-Lead Encapsulation)' },
    //     { id: '12', text: 'Other Traditional Packaging' },
    //     { id: '30', text: 'Testing' },
    //     { id: '31', text: 'Finished Product Testing' },
    //     { id: '44', text: 'Wafer Testing' },
    //     { id: '49', text: 'Other Professional Testing' },
    //     { id: '51', text: 'Advanced Packaging' },
    //     { id: '52', text: 'MEMS (Micro-Electro-Mechanical Systems Packaging)' },
    //     { id: '54', text: 'WLCSP (Wafer-Level Chip-Scale Packaging)' },
    //     { id: '56', text: 'TSV (Through-Silicon Via Technology)' },
    //     { id: '58', text: 'FOWLP (Fan-Out Wafer-Level Packaging)' },
    //     { id: '60', text: 'BUMP (Bump Packaging Technology)' },
    //     { id: '70', text: 'Raw Materials' },
    //     { id: '71', text: 'Other Raw Materials' },
    //     { id: '72', text: 'Sapphire Crystal' },
    //     { id: '74', text: 'Plating Additives for Semiconductor Testing' },
    //     { id: '75', text: 'Probe Card' },
    //     { id: '77', text: 'Aluminum Electrolytic Capacitor Electrode Foil' },
    //     { id: '78', text: 'Carbon Materials, New Polymer Composite Materials' },
    //     { id: '95', text: 'Wafer Manufacturing Materials' },
    //     { id: '96', text: 'Photoresist Materials' },
    //     { id: '101', text: 'Chemical Mechanical Polishing Materials' },
    //     { id: '105', text: 'Electronic Gases' },
    //     { id: '116', text: 'Second and Third Generation Semiconductors' },
    //     { id: '149', text: 'Deposition Materials' },
    //     { id: '178', text: 'Packaging Materials' },
    //     { id: '179', text: 'Ceramic Substrate' },
    //     { id: '185', text: 'Lead Frame' },
    //     { id: '197', text: 'Bonding Wire' },
    //     { id: '217', text: 'Packaging Substrate' },
    //     { id: '252', text: 'Chip Bonding Materials' },
    //     { id: '268', text: 'Design' },
    //     { id: '269', text: 'Analog Circuit' },
    //     { id: '270', text: 'RF Front-End and Receiver' },
    //     { id: '272', text: 'Analog/Mixed-Signal' },
    //     { id: '273', text: 'Driver Chip' },
    //     { id: '309', text: 'Data Conversion' },
    //     { id: '325', text: 'High-Performance Analog Chip' },
    //     { id: '465', text: 'Discrete Devices' },
    //     { id: '466', text: 'IGBT' },
    //     { id: '467', text: 'Power Devices' },
    //     { id: '470', text: 'Diode/Transistor' },
    //     { id: '502', text: 'MOSFET' },
    //     { id: '518', text: 'FRD' },
    //     { id: '552', text: 'Other Chips' },
    //     { id: '553', text: 'Clock Chip' },
    //     { id: '555', text: 'Biochip' },
    //     { id: '556', text: '25G and 56G High-Speed IO Connector' },
    //     { id: '557', text: 'Aerospace and Military Special Chips' },
    //     { id: '558', text: 'Security Chip' },
    //     { id: '560', text: 'Sensor' },
    //     { id: '561', text: 'Gas Sensor' },
    //     { id: '563', text: 'Photoelectric Sensor' },
    //     { id: '574', text: 'Temperature Sensor' },
    //     { id: '580', text: 'Optical Sensor' },
    //     { id: '582', text: 'Humidity Sensor' },
    //     { id: '669', text: 'Optoelectronics' },
    //     { id: '670', text: 'LED/LD' },
    //     { id: '681', text: 'Optical Chip' },
    //     { id: '1167', text: 'Manufacturing' },
    //     { id: '1168', text: 'Production Mode' },
    //     { id: '1169', text: 'Foundry' },
    //     { id: '1200', text: 'IDM' },
    //     { id: '1218', text: 'FOUNDRY' },
    //     { id: '1238', text: 'Equipment' },
    //     { id: '1239', text: 'Front-End Test Equipment' },
    //     { id: '1240', text: 'Measurement' },
    //     { id: '1245', text: 'Defect Detection' },
    //     { id: '1253', text: 'Other Equipment' },
    //     { id: '1254', text: 'Other Detection Equipment' },
    //     { id: '1259', text: 'Industrial Robot' },
    //     { id: '1260', text: 'Wafer Bonding System' },
    //     { id: '1262', text: 'Integrated Circuit Equipment - Gas Purifier' },
    //     { id: '1263', text: 'Die Bonder' },
    //     { id: '1269', text: 'Front-End Process Equipment' },
    //     { id: '1270', text: 'Ion Implantation Equipment' },
    //     { id: '1272', text: 'Debonding Equipment' },
    //     { id: '1274', text: 'Photolithography' },
    //     { id: '1279', text: 'Single Crystal Furnace' },
    //     { id: '1280', text: 'Plating Equipment' },
    //     { id: '1302', text: 'Back-End Test Equipment' },
    //     { id: '1303', text: 'Tester' },
    //     { id: '1316', text: 'Probe Station' },
    //     { id: '1319', text: 'Sorter' },
    //     { id: '1324', text: 'Back-End Packaging Equipment' },
    //     { id: '1325', text: 'Wafer Alignment/Bonding Equipment' },
    //     { id: '1328', text: 'Dicing Machine' },
    //     { id: '1335', text: 'Other Packaging Equipment' },
    //     { id: '1337', text: 'Mounter' },
    //     { id: '1346', text: 'Die Bonder' },
    //     { id: '123456', text: 'Random test' }
    //   ],
    //   lines: [
    //     { text: '', from: '0', to: '1' },
    //     { text: '', from: '1', to: '2' },
    //     { text: '', from: '2', to: '3' },
    //     { text: '', from: '2', to: '5' },
    //     { text: '', from: '2', to: '8' },
    //     { text: '', from: '2', to: '10' },
    //     { text: '', from: '2', to: '12' },
    //     { text: '', from: '1', to: '30' },
    //     { text: '', from: '30', to: '31' },
    //     { text: '', from: '30', to: '44' },
    //     { text: '', from: '30', to: '49' },
    //     { text: '', from: '1', to: '51' },
    //     { text: '', from: '51', to: '52' },
    //     { text: '', from: '51', to: '54' },
    //     { text: '', from: '51', to: '56' },
    //     { text: '', from: '51', to: '58' },
    //     { text: '', from: '51', to: '60' },
    //     { text: '', from: '0', to: '70' },
    //     { text: '', from: '70', to: '71' },
    //     { text: '', from: '71', to: '72' },
    //     { text: '', from: '71', to: '74' },
    //     { text: '', from: '71', to: '75' },
    //     { text: '', from: '71', to: '77' },
    //     { text: '', from: '71', to: '78' },
    //     { text: '', from: '70', to: '95' },
    //     { text: '', from: '95', to: '96' },
    //     { text: '', from: '95', to: '101' },
    //     { text: '', from: '95', to: '105' },
    //     { text: '', from: '95', to: '116' },
    //     { text: '', from: '95', to: '149' },
    //     { text: '', from: '70', to: '178' },
    //     { text: '', from: '178', to: '179' },
    //     { text: '', from: '178', to: '185' },
    //     { text: '', from: '178', to: '197' },
    //     { text: '', from: '178', to: '217' },
    //     { text: '', from: '178', to: '252' },
    //     { text: '', from: '0', to: '268' },
    //     { text: '', from: '268', to: '269' },
    //     { text: '', from: '269', to: '270' },
    //     { text: '', from: '269', to: '272' },
    //     { text: '', from: '269', to: '273' },
    //     { text: '', from: '269', to: '309' },
    //     { text: '', from: '269', to: '325' },
    //     { text: '', from: '268', to: '465' },
    //     { text: '', from: '465', to: '466' },
    //     { text: '', from: '465', to: '467' },
    //     { text: '', from: '465', to: '470' },
    //     { text: '', from: '465', to: '502' },
    //     { text: '', from: '465', to: '518' },
    //     { text: '', from: '268', to: '552' },
    //     { text: '', from: '552', to: '553' },
    //     { text: '', from: '552', to: '555' },
    //     { text: '', from: '552', to: '556' },
    //     { text: '', from: '552', to: '557' },
    //     { text: '', from: '552', to: '558' },
    //     { text: '', from: '268', to: '560' },
    //     { text: '', from: '560', to: '561' },
    //     { text: '', from: '560', to: '563' },
    //     { text: '', from: '560', to: '574' },
    //     { text: '', from: '560', to: '580' },
    //     { text: '', from: '560', to: '582' },
    //     { text: '', from: '268', to: '669' },
    //     { text: '', from: '669', to: '670' },
    //     { text: '', from: '669', to: '681' },
    //     { text: '', from: '0', to: '1167' },
    //     { text: '', from: '1167', to: '1168' },
    //     { text: '', from: '1168', to: '1169' },
    //     { text: '', from: '1168', to: '1200' },
    //     { text: '', from: '1168', to: '1218' },
    //     { text: '', from: '0', to: '1238' },
    //     { text: '', from: '1238', to: '1239' },
    //     { text: '', from: '1239', to: '1240' },
    //     { text: '', from: '1239', to: '1245' },
    //     { text: '', from: '1238', to: '1253' },
    //     { text: '', from: '1253', to: '1254' },
    //     { text: '', from: '1253', to: '1259' },
    //     { text: '', from: '1253', to: '1260' },
    //     { text: '', from: '1253', to: '1262' },
    //     { text: '', from: '1253', to: '1263' },
    //     { text: '', from: '1238', to: '1269' },
    //     { text: '', from: '1269', to: '1270' },
    //     { text: '', from: '1269', to: '1272' },
    //     { text: '', from: '1269', to: '1274' },
    //     { text: '', from: '1269', to: '1279' },
    //     { text: '', from: '1269', to: '1280' },
    //     { text: '', from: '1238', to: '1302' },
    //     { text: '', from: '1302', to: '1303' },
    //     { text: '', from: '1302', to: '1316' },
    //     { text: '', from: '1302', to: '1319' },
    //     { text: '', from: '1238', to: '1324' },
    //     { text: '', from: '1324', to: '1325' },
    //     { text: '', from: '1324', to: '1328' },
    //     { text: '', from: '1324', to: '1335' },
    //     { text: '', from: '1324', to: '1337' },
    //     { text: '', from: '1324', to: '1346' },
    //     { text: '', from: '533', to: '123456' }
    //   ]
    // };
    const graphInstance = graphRef.current?.getInstance();
    if (graphInstance) {
      graphInstance.loading();
      await graphInstance.setJsonData(__graph_json_data);
      await graphInstance.doLayout();
      await openByLevel(1);
      await graphInstance.zoomToFit();
      graphInstance.clearLoading();
    }
  };

  const onNodeClick = (nodeObject, $event) => {
    setSelectedNode({
      ...nodeObject,
      data: {
        ...nodeObject.data,
        allConnections: nodeObject.data.allConnections,
      },
    });
    setDialogOpen(true);
    setExpandedChildren(false);
  };

  const onLineClick = (lineObject, linkObject, $event) => {
    console.log('onLineClick:', lineObject);
  };

  const openByLevel = async (level) => {
    const graphInstance = graphRef.current?.getInstance();
    if (graphInstance) {
      // Reset data

      graphInstance.getNodes().forEach((node) => {
        node.expanded = true;
        node.alignItems = 'top';
      });

      // Reset data

      graphInstance.getNodes().forEach((node) => {
        node.className = 'my-industy-node-level-' + Math.abs(node.lot?.level || 0);
      });

      graphInstance.getNodes().forEach((node) => {
        if (Math.abs(node.lot?.level || 0) >= level) {
          node.expanded = false;
        }
      });

      await graphInstance.doLayout();
    }
  };
  const MyNodeSlot = ({ node }) => {
    return (
      <div slot="node" slot-scope="{node}">
        {node.lot && node.lot.level === 0 && (
          <div className="my-industy-node my-industy-node-level-0">
            {/* root node */}
            <div className="my-card-header">Author Name</div>
            <div className="my-card-body">{node.text}</div>
          </div>
        )}
        {node.lot && node.lot.level === 1 && (
          <div className="my-industy-node my-industy-node-level-1">
            {/* level 1 nodes */}
            <div className="my-card-header">Flag Type</div>
            <div className="my-card-body">{node.text}</div>
          </div>
        )}
        {node.lot && node.lot.level === 2 && (
          <div className="my-industy-node my-industy-node-level-2">
            {/* level 2 nodes */}
            <div className="my-card-header">Subsection Link</div>
            <div className="my-card-body">{node.text}</div>
          </div>
        )}
        {
          (node.lot.level >= 3 && (
            <div className="my-industy-node my-industy-node-level-3">
              <div className="my-card-header">Product Type</div>
              <div className="my-card-body">{node.text}</div>
            </div>
          ))}
      </div>
    );
  };

  const graphOptions = {
    debug: true,
    backgrounImageNoRepeat: true,
    moveToCenterWhenRefresh: true,
    zoomToFitWhenRefresh: true,
    useAnimationWhenRefresh: false,
    useAnimationWhenExpanded: true,
    reLayoutWhenExpandedOrCollapsed: true,
    defaultExpandHolderPosition: 'bottom',
    defaultNodeShape: 1,
    defaultNodeBorderWidth: 0,
    defaultLineShape: 2,
    defaultJunctionPoint: 'tb',
    lineUseTextPath: true,
    defaultLineWidth: 1,
    defaultLineColor: '#09abff',
    layouts: [
      {
        layoutName: 'tree',
        from: 'top',
        levelDistance: '250,250,250,250',
        min_per_width: 205,
      },
    ],
  };

  return (
    <div>
      <div
        style={{
          height: '700px',
          width: '96%',
          marginLeft: '2%',
          overflow: 'hidden',
          position: 'relative',
          border: '1px solid rgb(230,230,230)',
          borderRadius: '8px',
          marginTop: '20px',
        }}
      >
        <RelationGraph
          ref={graphRef}
          options={graphOptions}
          nodeSlot={MyNodeSlot}
          onNodeClick={onNodeClick}
          onLineClick={onLineClick}
        ></RelationGraph>
        <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} fullWidth maxWidth="sm">
          <DialogTitle sx={{ bgcolor: '#2A2A2A', color: 'white' }}>
            {'Detailed Information'}
          </DialogTitle>
          <Divider
            sx={{
              borderBottomWidth: 2,
              borderColor: 'text.primary',
            }}
          />
          <DialogContent sx={{ bgcolor: '#2A2A2A' }}>
            <Card variant="outlined" sx={{ mb: 2, bgcolor: '#6e6e6e' }}>
              <CardContent>
                <Typography variant="body1" color="text.primary" sx={{ fontWeight: 600 }}>
                  {selectedNode?.data.text}
                </Typography>
                {selectedNode?.data?.url && (
                  <>
                    <Divider
                      sx={{
                        borderBottomWidth: 1,
                        borderColor: 'white',
                        marginTop: 1,
                      }}
                    />
                    <Link
                      href={selectedNode.data.url}
                      target="_blank"
                      sx={{ mt: 1, display: 'block', color: 'orange' }}
                    >
                      {selectedNode.data.url}
                    </Link>
                  </>
                )}
              </CardContent>
            </Card>

            {selectedNode?.data?.allConnections?.length > 0 && (
              <>
                <div
                  onClick={() => setExpandedChildren(!expandedChildren)}
                  style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}
                >
                  <Typography variant="subtitle1" sx={{ color: 'white' }}>
                    Relevant Documents / Webpages ({selectedNode.data.allConnections.length})
                  </Typography>
                  <IconButton size="small" sx={{ color: 'white' }}>
                    {expandedChildren ? (
                      <ExpandLessIcon color="inherit" />
                    ) : (
                      <ExpandMoreIcon color="inherit" />
                    )}
                  </IconButton>
                </div>

                <Collapse in={expandedChildren}>
                  <Card variant="outlined" sx={{ mt: 1, bgcolor: '#6e6e6e' }}>
                    <CardContent>
                      {selectedNode.data.allConnections.map((child, index) => (
                        <Typography key={index} variant="body1" sx={{ mb: 1, color: 'black' }}>
                          <Link
                            href={child.url}
                            target="_blank"
                            rel="noopener noreferrer"
                            style={{ textDecoration: 'none', color: 'inherit' }}
                            onMouseEnter={(e) => (e.target.style.textDecoration = 'underline')}
                            onMouseLeave={(e) => (e.target.style.textDecoration = 'none')}
                          >
                            {getRawTextFromXML(child.title)}
                          </Link>
                          <Divider
                            sx={{
                              borderBottomWidth: 1,
                              borderColor: 'white',
                              marginTop: 1,
                            }}
                          />
                        </Typography>
                      ))}
                    </CardContent>
                  </Card>
                </Collapse>
              </>
            )}
          </DialogContent>
        </Dialog>
      </div>
    </div>
  );
};

export default MyComponent;
