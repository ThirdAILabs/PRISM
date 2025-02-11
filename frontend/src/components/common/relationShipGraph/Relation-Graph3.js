import React, { useEffect, useRef, useState } from 'react';
import RelationGraph from 'relation-graph-react';
import {
    Dialog, DialogTitle, DialogContent, Divider,
    Card, CardContent, Typography, Collapse, IconButton, Link
} from '@mui/material';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';
import ExpandLessIcon from '@mui/icons-material/ExpandLess';
import { onFlagsUpdate } from '../../../services/streamStore';

function generateGraphData(data, parentId = null, level = 0) {
    console.log(`${data}`)
    const nodes = [];
    const lines = [];
    let currentNodeId;

    if (level === 0) {
        // Root node
        currentNodeId = 'a';
        nodes.push({
            id: currentNodeId,
            html: `
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
                ${data.name}
                </div>
            </div>
            `,
            width: 200,
            height: 200,
            data: { url: data?.url, text: data.name },
        });
    } else if (level === 1 && data.count) {
        // Second level node
        currentNodeId = `g${Math.random().toString(36).substring(7)}`;
        const fontSize = 20;
        const padding = 5;
        const textLength = Math.min(data.title.length, 50);
        let dotString = "";
        if (textLength < data.title.length)
            dotString += "...";
        const calculatedWidth = Math.max(textLength * (fontSize * 0.3), 50) + padding;
        nodes.push({
            id: currentNodeId,
            text: data.title.slice(0, textLength) + dotString + "(" + data.count.toString() + ")",
            data: { url: data?.url, text: data.title },
            width: calculatedWidth,
            height: calculatedWidth,
            borderWidth: 3,
            color: 'transparent',
            borderColor: '#ff8c00',
            fontColor: '#ff8c00',
            fontSize: fontSize,
        });
    } else if (level === 2) {
        // Third level node
        currentNodeId = `e${Math.random().toString(36).substring(7)}`;
        const textLength = Math.min(data.title.length, 50);
        let dotString = "";
        if (textLength < data.title.length)
            dotString += "...";
        nodes.push({
            id: currentNodeId,
            text: data.title.slice(0, textLength) + dotString,
            data: { url: data?.url, text: data.title },
            borderWidth: 1, nodeShape: 1, width: 300, height: 60, fontSize: 20
        });

    } else if (level === 3) {
        // Fourth level node

        currentNodeId = `b${Math.random().toString(36).substring(7)}`;
        const textLength = Math.min(data.title.length, 30);
        let dotString = "";
        if (textLength < data.title.length)
            dotString += "...";
        const fontSize = 20;
        const calculatedWidth = Math.max(textLength * (fontSize * 0.25), 50);
        nodes.push({
            id: currentNodeId,
            text: data.title.slice(0, textLength) + dotString,
            data: { url: data?.url, text: data.title },
            width: calculatedWidth,
            height: calculatedWidth,
            fontSize: fontSize,
            borderColor: 'white',
            borderWidth: 1,
            color: "#6F3096"
        });

    } else if (level === 4) {
        const textLength = Math.min(data.title.length, 30);
        const fontSize = 16;
        const calculatedWidth = Math.max(textLength * (fontSize * 0.26), 50);
        currentNodeId = `h${Math.random().toString(36).substring(7)}`;
        let dotString = "";
        if (textLength < data.title.length)
            dotString += "...";
        nodes.push({
            id: currentNodeId,
            text: data.title.slice(0, textLength) + dotString,
            data: { url: data?.url, text: data.title },
            borderColor: '#FD5E53',
            borderWidth: 2,
            color: "#ff69b4",
            nodeShape: 1,
            width: 200, height: 50
        });
    }

    if (parentId) {
        lines.push({ from: parentId, to: currentNodeId });
    }

    if (data.connections && data.connections.length > 0) {
        let limit = data.connections.length;
        if (level)
            limit = Math.min(3, limit);
        for (let connectionIndex = 0; connectionIndex < limit; connectionIndex++) {
            const child = data.connections[connectionIndex];
            const childGraph = generateGraphData(child, currentNodeId, level + 1);
            nodes.push(...childGraph.nodes);
            lines.push(...childGraph.lines);
        }
    }

    return { nodes, lines, rootId: 'a' };
}

const Demo = () => {

    const [dialogOpen, setDialogOpen] = useState(false);
    const [selectedNode, setSelectedNode] = useState(null);
    const [expandedChildren, setExpandedChildren] = useState(false);

    const graphRef = useRef(null);
    const graphOptions = {
        // Here you can refer to the parameters in "Graph Graph" for settings
    };

    const [graphData, setGraphData] = useState(null);
    let nodes = [], lines = [], rootId = null;

    useEffect(() => {
        const unsubscribe = onFlagsUpdate((flags) => {
            setGraphData(flags);
        });
        return () => unsubscribe();
    }, []);

    useEffect(() => {
        if (graphData) {
            console.log("GraphData: ", graphData);
            const elements = generateGraphData(graphData);
            nodes = elements.nodes;
            lines = elements.lines;
            rootId = elements.rootId;

            const graphInstance = graphRef.current.getInstance();
            graphInstance.setJsonData(elements).then(() => {
                graphInstance.moveToCenter();
                graphInstance.zoomToFit();
            });
        }
    }, [graphData]);

    const showGraph = () => {
        const __graph_json_data = {
            rootId: rootId,
            nodes: nodes,
            lines: lines
        };
        const graphInstance = graphRef.current.getInstance();
        graphInstance.setJsonData(__graph_json_data).then(() => {
            graphInstance.moveToCenter();
            graphInstance.zoomToFit();
        });
    };

    // const onNodeClick = (nodeObject, $event) => {
    //     console.log('onNodeClick:', nodeObject);
    // };
    const onNodeClick = (nodeObject) => {
        console.log("onNOdeClick", nodeObject);
        setSelectedNode(nodeObject);
        setDialogOpen(true);
        setExpandedChildren(false);
    };

    const onLineClick = (lineObject, linkObject, $event) => {
        console.log('onLineClick:', lineObject);
    };

    useEffect(() => {
        showGraph();
    }, []);

    return (
        <div>
            <div style={{
                height: '100vh', width: "90%", marginLeft: "5%", overflow: 'hidden',
                position: 'relative',
                backgroundColor: "#000000"
            }}>
                <RelationGraph ref={graphRef} options={graphOptions} onNodeClick={onNodeClick} onLineClick={onLineClick} />
                <Dialog
                    open={dialogOpen}
                    onClose={() => setDialogOpen(false)}
                    fullWidth
                    maxWidth="sm"

                >
                    <DialogTitle sx={{ bgcolor: '#2A2A2A', color: 'white' }}>
                        Detailed Information
                    </DialogTitle>
                    <Divider sx={{
                        borderBottomWidth: 2,
                        borderColor: 'text.primary'
                    }} />
                    <DialogContent sx={{ bgcolor: '#2A2A2A' }}>
                        {/* Node Text Card */}
                        <Card variant="outlined" sx={{ mb: 2, bgcolor: "#6e6e6e" }}>
                            <CardContent>
                                <Typography variant="body1" color="text.primary" sx={{ fontWeight: 600 }}>
                                    {selectedNode?.data.text}
                                </Typography>

                                {selectedNode?.data?.url && (
                                    <>
                                        <Divider sx={{
                                            borderBottomWidth: 1,
                                            borderColor: 'white',
                                            marginTop: 1,
                                        }} />
                                        <Link
                                            href={selectedNode.data.url}
                                            target="_blank"
                                            sx={{ mt: 1, display: 'block', color: "orange" }}
                                        >
                                            {selectedNode.data.url}
                                        </Link>
                                    </>
                                )}
                            </CardContent>
                        </Card>

                        {/* Children Section */}
                        {selectedNode?.targetTo.length > 0 && (
                            <>
                                <div
                                    onClick={() => setExpandedChildren(!expandedChildren)}
                                    style={{ cursor: 'pointer', display: 'flex', alignItems: 'center' }}

                                >
                                    <Typography variant="subtitle1" sx={{ color: 'white' }}>
                                        Child Nodes
                                    </Typography>
                                    <IconButton size="small" sx={{ color: 'white' }}>
                                        {expandedChildren ?
                                            <ExpandLessIcon color="inherit" /> :
                                            <ExpandMoreIcon color="inherit" />}
                                    </IconButton>
                                </div>

                                <Collapse in={expandedChildren}>
                                    <Card variant="outlined" sx={{ mt: 1, bgcolor: "#6e6e6e" }}>
                                        <CardContent>
                                            {selectedNode.targetTo.map((child, index) => (
                                                <Typography
                                                    key={index}
                                                    variant="body1"
                                                    sx={{ mb: 1, color: 'black' }}
                                                >
                                                    {child.data.text}
                                                    <Divider sx={{
                                                        borderBottomWidth: 1,
                                                        borderColor: 'white',
                                                        marginTop: 1,
                                                    }} />
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

export default Demo;