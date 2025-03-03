import React, { useEffect, useRef, useState } from 'react';
import RelationGraph from 'relation-graph-react';
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
import { onFlagsUpdate } from '../../../services/streamStore.js';
import { AUTHOR_AFFILIATIONS } from '../../../constants/constants.js';

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

function levelZeroNodeHtml(name, riskScore) {
  return `
        <div style="
            display: flex; 
            flex-direction: column; 
            align-items: center; 
            justify-content: center; 
            width: 200px;
            height: 200px;
            background-color: #ff8c00;
            border: 5px solid #ffd700;
            border-radius: 50%;
            padding: 10px;
            box-sizing: border-box;
        ">
            <svg width="50" height="50" viewBox="0 0 24 24" fill="none">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm0 3c1.66 0 3 1.34 3 3s-1.34 3-3 3-3-1.34-3-3 1.34-3 3-3zm0 14.2c-2.5 0-4.71-1.28-6-3.22.03-1.99 4-3.08 6-3.08 1.99 0 5.97 1.09 6 3.08-1.29 1.94-3.5 3.22-6 3.22z" 
                    fill="#ffffff"/>
            </svg>
            <div style="
            color: #ffffff; 
            margin-top: 8px; 
            text-align: center; 
            font-size: 20px;
            font-weight: bold;
            line-height: 1.2;
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            ">
                ${name}
            </div>
            <div style="
            color: #ffffff; 
            margin-top: 4px; 
            text-align: center; 
            font-size: 16px;
            font-weight: normal;
            line-height: 1.2;
            max-width: 100%;
            overflow: hidden;
            text-overflow: ellipsis;
            ">
                [Risk Score: ${riskScore}]
            </div>
            </div>
        </div>
    `;
}

function levelZeroNode(nodeId, name, riskScore, url) {
  return {
    id: nodeId,
    html: levelZeroNodeHtml(name, riskScore),
    width: 200,
    height: 200,
    data: {
      url: url,
      text: name,
    },
  };
}

function levelOneNode(nodeId, title, url, count, connections) {
  return {
    id: nodeId,
    text: title + ' (' + count.toString() + ')',
    data: {
      url: url,
      text: title,
      allConnections: connections,
    },
    styleClass: { width: 'fit-content' },
    width: 180,
    height: 180,
    borderWidth: 3,
    color: 'transparent',
    borderColor: '#ff8c00',
    fontColor: '#ff8c00',
    fontSize: 20,
  };
}

function levelTwoNode(nodeId, title, url, connections) {
  const textLength = Math.min(title.length, 50);
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
    borderWidth: 1,
    nodeShape: 1,
    width: 300,
    height: 60,
    fontSize: 20,
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
    width: calculatedWidth,
    height: calculatedWidth,
    fontSize: fontSize,
    borderColor: 'white',
    borderWidth: 1,
    color: '#6F3096',
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
    borderColor: '#FD5E53',
    borderWidth: 2,
    color: '#ff69b4',
    nodeShape: 1,
    width: 200,
    height: 50,
  };
}

function generateVisibleGraphData(data, parentId = null, level = 0) {
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
    lines.push({ from: parentId, to: currentNodeId });
  }

  let levelNodePairs = [[level, currentNodeId]];

  if (data.connections && data.connections.length > 0) {
    let limit = data.connections.length;
    if (level > 0) {
      limit = Math.min(limit, 3);
    }
    for (let connectionIndex = 0; connectionIndex < limit; connectionIndex++) {
      const child = data.connections[connectionIndex];
      const childGraph = generateVisibleGraphData(child, currentNodeId, level + 1);
      nodes.push(...childGraph.nodes);
      lines.push(...childGraph.lines);
      levelNodePairs = [...levelNodePairs, ...childGraph.levelNodePairs];
    }
  }

  return { nodes, lines, rootId: 'a', levelNodePairs };
}

function getLevelToNodes(levelNodePairs) {
  const levelToNodes = {};
  for (const [level, nodeId] of levelNodePairs) {
    if (!levelToNodes[level]) {
      levelToNodes[level] = [];
    }
    levelToNodes[level].push(nodeId);
  }
  return levelToNodes;
}

function getNodeToNumConnections(lines) {
  const nodeToNumConnections = {};
  for (const line of lines) {
    if (!nodeToNumConnections[line.from]) {
      nodeToNumConnections[line.from] = 0;
    }
    nodeToNumConnections[line.from]++;
  }
  return nodeToNumConnections;
}

function invisibleNode(nodeId) {
  return {
    id: nodeId,
    text: '',
    width: 0,
    height: 0,
    color: 'transparent',
  };
}

function invisibleLine(from, to) {
  return {
    from: from,
    to: to,
    lineWidth: 0,
    color: 'transparent',
  };
}

function generateGraphData(data) {
  const visibleGraphData = generateVisibleGraphData(data);

  /*
        Nodes of a particular level (level is distance from root node)
        are evenly spaced out in a radial pattern.

        Imagine an unbalanced graph where one branch that is much longer
        than other branches. In this case, leaf nodes from the longer branch
        will be distributed all around the graph instead of being localized
        (leaf nodes close to the associated subtree).

        To ensure that leaf nodes are localized we add invisible children
        nodes to nodes that have less children, thus constraining the leaves
        of the longer branches to their respective localities.
    */

  const levelToNodes = getLevelToNodes(visibleGraphData.levelNodePairs);
  const nodeToNumConnections = getNodeToNumConnections(visibleGraphData.lines);

  const invisibleNodes = [];
  const invisibleLines = [];

  const maxLevel = Math.max(...Object.keys(levelToNodes).map(Number));
  for (let level = 0; level <= maxLevel; level++) {
    const maxConnections = Math.max(
      ...levelToNodes[level].map((nodeId) => nodeToNumConnections[nodeId] || 0)
    );
    for (const nodeId of levelToNodes[level]) {
      for (let i = nodeToNumConnections[nodeId] || 0; i < maxConnections; i++) {
        const invisibleNodeId = `invisible${level}${Math.random().toString(36).substring(7)}`;
        invisibleNodes.push(invisibleNode(invisibleNodeId));
        invisibleLines.push(invisibleLine(nodeId, invisibleNodeId));
        levelToNodes[level + 1].push(invisibleNodeId);
      }
    }
  }

  return {
    nodes: [...visibleGraphData.nodes, ...invisibleNodes],
    lines: [...visibleGraphData.lines, ...invisibleLines],
    rootId: visibleGraphData.rootId,
  };
}

const Graph = ({ authorName, reportContent }) => {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [selectedNode, setSelectedNode] = useState(null);
  const [expandedChildren, setExpandedChildren] = useState(false);

  const graphRef = useRef(null);
  const graphOptions = {
    // Here you can refer to the parameters in "Graph Graph" for settings
  };

  const graphData = generateGraphData(convertDataToGraphFormat(authorName, reportContent));
  const { nodes, lines, rootId } = graphData;

  useEffect(() => {
    if (reportContent) {
      const graphInstance = graphRef.current.getInstance();
      graphInstance.setJsonData(graphData).then(() => {
        graphInstance.moveToCenter();
        graphInstance.zoomToFit();
      });
    }
  }, [reportContent]);

  const showGraph = () => {
    const graphInstance = graphRef.current.getInstance();

    // Clear the existing graph data
    graphInstance.setJsonData({ nodes: [], lines: [], rootId: '' }).then(() => {
      // Set the new graph data
      const __graph_json_data = {
        rootId: rootId,
        nodes: nodes,
        lines: lines,
      };

      graphInstance.setJsonData(__graph_json_data).then(() => {
        graphInstance.moveToCenter();
        graphInstance.zoomToFit();
      });
    });
  };

  const onLineClick = (lineObject, linkObject, $event) => {
    console.log('onLineClick:', lineObject);
  };

  const onNodeClick = (nodeObject) => {
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

  useEffect(() => {
    showGraph();
  }, []);

  return (
    <div>
      <div
        style={{
          height: '100vh',
          width: '90%',
          marginLeft: '5%',
          overflow: 'hidden',
          position: 'relative',
          backgroundColor: '#000000',
        }}
      >
        <RelationGraph ref={graphRef} options={graphOptions} onNodeClick={onNodeClick} />
        <Dialog open={dialogOpen} onClose={() => setDialogOpen(false)} fullWidth maxWidth="sm">
          <DialogTitle sx={{ bgcolor: '#2A2A2A', color: 'white' }}>
            {selectedNode?.data.text}
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
                            {child.title}
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

export default Graph;
