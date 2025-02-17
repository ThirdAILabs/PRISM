import React, { useEffect, useRef } from "react";
import "./loadedPapers.css"

function LoadedPapers({titles}) {
    console.log(titles);
    // const scrollableAreaEndRef = useScrollToBottom();
    return <div className="LoadedPapers-card">
        <div className="LoadedPapers-message-container">
            Loaded {titles.length} papers
        </div>
        <div className="LoadedPapers-titles-container">
            <div className="LoadedPapers-titles-scrollable">
                {
                    titles.reverse().map(title => <div className="LoadedPapers-title">{title}</div>)
                }
            </div>
        </div>
    </div>
}

export default LoadedPapers