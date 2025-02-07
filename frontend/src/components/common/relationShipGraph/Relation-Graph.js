import React, { useEffect, useRef } from 'react';
import RelationGraph from 'relation-graph-react';

const data =
{
  name: "Dr. Ram Kumar",
  connections: [
    {
      relation: "Doj press release",
      entity: "P-Body",
      children: [
        {
          name: "Dr. QingyunSun", leaf: [{
            name: "Left-Nodde",
            role: "Leaft-Role"
          }]
        },
        {
          name: "Department of Physics", leaf: [
            {
              name: "Leaf-Node",
              role: "Leaft-Role"
            }
          ]
        },
        {
          name: "Graduate Program", leaf: [
            {
              name: "Leaf-Node",
              role: "Leaft-Role"
            }
          ]
        },
        {
          name: "Research Center X", role: "Senior Researcher", leaf: [
            {
              name: "Leaf-Node",
              role: "Leaft-Role"
            }
          ]
        },
        {
          name: "Innovation Hub", role: "Advisor", leaf: [
            {
              name: "Leaf-Node",
              role: "Leaft-Role"
            }
          ]
        },
        {
          name: "Tech Transfer Office", role: "Committee Member", leaf: [
            {
              name: "Leaf-Node",
              role: "Leaft-Role"
            }
          ]
        },
        {
          name: "Student Mentorship", role: "Lead Mentor", leaf: [
            {
              name: "Leaf-Node",
              role: "Leaft-Role"
            }
          ]
        },
        {
          name: "Grant Committee", role: "Member", leaf: [
            {
              name: "Leaf-Node",
              role: "Leaft-Role"
            }
          ]
        }
      ],
    },
    {
      relation: "appointmentWithShivaji",
      entity: "Institute Anand",
      years: "2012-2032",
      children: [
        { name: "Project Alpha", role: "Consultant" },
        { name: "Research Initiative B", role: "Advisor" }
      ],
    },
    {
      relation: "collaboration",
      entity: "Dr. C",
      years: "2020",
      children: [],
    },
    {
      relation: "secondary_affiliation",
      entity: "Entity ABC LLC",
      years: "2019-2021",
      children: [
        { name: "Project X", role: "Lead" },
        { name: "Initiative Y", role: "Advisor" },
        { name: "Program Z", role: "Member" }
      ],
    },
  ],
  risk_score: 91,
};

const ResearcherNetwork = () => {
  const graphRef = useRef(null);

  const graphOptions = {
    layout: {
      layoutName: 'force',
      layoutClassName: 'seeks-layout-force',
    },
    defaultNodeBorderWidth: 0,
    defaultNodeShape: 1,
    defaultLineShape: 3,
    defaultJunctionPoint: 'border',
    defaultNodeColor: '#1E88E5',
    defaultLineColor: '#666666',
    defaultNodeWidth: 'auto', // Auto-adjust node width
    defaultNodeHeight: 'auto', // Auto-adjust node height
    background: '#000000'
  };

  useEffect(() => {
    const setGraphData = async () => {
      const graphData = {
        rootId: 'root',
        nodes: [],
        lines: []
      };

      // Add central node
      graphData.nodes.push({
        id: 'root',
        text: data.name,
        html: `
          <div style="display: flex; flex-direction: column; align-items: center; padding: 8px; border: 1px solid #ccc; border-radius: 8px;">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="blue" stroke="black" stroke-width="1">
            <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"></path>
            <circle cx="12" cy="7" r="4"></circle>
          </svg>
          <div style="margin-top: 4px; color: black;">${data.name}</div>
        </div>
        `,
      });

      // Add connections
      data.connections.forEach((conn, index) => {
        const nodeId = `conn_${index}`;
        graphData.nodes.push({
          id: nodeId,
          text: `${conn.entity}\n(${conn.relation})`,
          color: conn.type === 'direct' ? '#4CAF50' : '#FFA726',
        });
        graphData.lines.push({
          from: 'root',
          to: nodeId,
          text: conn.relation,
        });

        // Add children nodes
        if (conn.children && conn.children.length > 0) {
          const children = conn.children;
          const maxChildrenToShow = 3;
          const childrenToShow = children.slice(0, maxChildrenToShow);
          const remainingChildrenCount = children.length - maxChildrenToShow;

          childrenToShow.forEach((child, childIndex) => {
            const childNodeId = `child_${index}_${childIndex}`;
            graphData.nodes.push({
              id: childNodeId,
              text: `${child.name}\n(${child.role})`,
              color: '#9C27B0',
            });
            graphData.lines.push({
              from: nodeId,
              to: childNodeId,
              text: child.role,
            });
            if (child.leaf && child.leaf.length > 0) {
              const leaf = child.leaf;
              leaf.forEach((thisLeaf, leafIndex) => {
                const leafNodeId = `leaf_${index}_${childIndex}_${leafIndex}`;
                graphData.nodes.push({
                  id: leafNodeId,
                  text: `${thisLeaf.name}\n(${thisLeaf.role})`,
                  color: '#47f029',
                });
                graphData.lines.push({
                  from: childNodeId,
                  to: leafNodeId,
                  text: "Random it is"
                });
              })
            }
          });

          // if (remainingChildrenCount > 0) {
          //   const moreNodeId = `more_${index}`;
          //   graphData.nodes.push({
          //     id: moreNodeId,
          //     text: `${remainingChildrenCount} more`,
          //     color: '#9C27B0',
          //   });
          //   graphData.lines.push({
          //     from: nodeId,
          //     to: moreNodeId,
          //     text: 'More',
          //   });
          // }
        }
      });

      const instance = graphRef.current?.getInstance();
      if (instance) {
        await instance.setJsonData(graphData);
        await instance.moveToCenter();
        await instance.zoomToFit();
      }
    };

    setGraphData();
  }, []);

  return (
    <div style={{ backgroundColor: "#000000", marginLeft: "5%", height: '1000px', width: "90%" }}>
      <RelationGraph
        ref={graphRef}
        options={graphOptions}
      />
    </div>

  );
};

export default ResearcherNetwork;