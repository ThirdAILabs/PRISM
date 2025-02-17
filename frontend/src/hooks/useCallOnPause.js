import { useEffect, useRef } from 'react';

function useCallOnPause(minSeconds) {
    const timeoutIds = useRef([]);
    const lastCallId = useRef();

    useEffect(() => {
        // Clear timeouts on component unmount to avoid acting on stale calls.
        return () => {
            for (const timeoutId in timeoutIds.current) {
                clearTimeout(timeoutId);
            }
        }
    }, []);

    const call = callable => {
        const callId = Math.random().toString() + Date.now().toString();
        lastCallId.current = callId;
        // Call after minSeconds pause.
        const timeoutId = setTimeout(() => {
            // Only call the last item.
            if (lastCallId.current === callId) {
                callable();
            }
        }, minSeconds / 1000);
        timeoutIds.current = [...timeoutIds.current, timeoutId];
    };

    return call;
}

export default useCallOnPause;