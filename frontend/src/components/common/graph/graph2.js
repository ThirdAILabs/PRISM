import React, { useEffect, useRef, useState } from 'react';
import RelationGraph from 'relation-graph-react';
import '../../../styles/components/_graph2.scss';
import { AUTHOR_AFFILIATIONS } from '../../../constants/constants.js';
import MyToolbarToolbar from './toolbar.tsx';
import { Tooltip } from 'react-tooltip';

function getNodeTitle(flagType, flag) {
  if (flagType === AUTHOR_AFFILIATIONS) {
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
      flagType: flagType,
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

function levelOneNode(nodeId, title, url, count, connections, flagType) {
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
      flagType: flagType,
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
    nodes.push(
      levelOneNode(currentNodeId, data.title, data.url, data.count, data.connections, data.flagType)
    );
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

const MyComponent = ({ authorName, reportContent, review, setReview, setGraphNodeInfo }) => {
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
    const selectedFlag = nodeObject?.data?.flagType;
    if (selectedFlag) setReview(nodeObject?.data?.flagType);
    else {
      setReview('NodeData');
      setGraphNodeInfo({
        title: 'Detailed Information',
        text: nodeObject?.data?.text,
        url: nodeObject?.data?.url,
        allConnections: nodeObject?.data?.allConnections,
      });
    }
    console.log('nodedataclick', nodeObject?.data.flagType);
    setSelectedNode({
      ...nodeObject,
      data: {
        ...nodeObject.data,
        allConnections: nodeObject.data.allConnections,
      },
    });
    // setDialogOpen(true);
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
            <div className="my-card-header" style={{ color: 'rgba(0, 0, 0, 0)' }}>
              .
            </div>
            <div className="my-card-body">{node.text}</div>
          </div>
        )}
        {!node.lot ||
          (node.lot.level >= 3 && (
            <div className="my-industy-node my-industy-node-level-3">
              <div className="my-card-header" style={{ color: 'rgba(0, 0, 0, 0)' }}>
                .
              </div>
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
  const [width, setWidht] = useState('96%');

  const onSizeOptionChanged = () => {
    setTimeout(() => {
      const graphInstance = graphRef.current?.getInstance();
      graphInstance?.onGraphResize();
    });
  };
  useEffect(() => {
    async function handleWidth() {
      if (review && review !== '') {
        setWidht('56%');
      } else {
        setWidht('96%');
      }
    }
    handleWidth();
    onSizeOptionChanged();
  }, [review]);
  return (
    <div>
      <div
        style={{
          height: '716px',
          width: width,
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
          toolBarSlot={() => <MyToolbarToolbar />}
          graphPlugSlot={
            <div>
              <Tooltip id="my-tooltip" />
            </div>
          }
        ></RelationGraph>
      </div>
    </div>
  );
};

export default MyComponent;
