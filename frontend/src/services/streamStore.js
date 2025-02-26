let streamFlags = null;
const listeners = new Set();

export const setStreamFlags = (flags) => {
  streamFlags = flags;
  listeners.forEach((callback) => callback(flags));
};

export const getStreamFlags = () => streamFlags;

export const onFlagsUpdate = (callback) => {
  listeners.add(callback);
  if (streamFlags !== null) {
    callback(streamFlags);
  }
  return () => listeners.delete(callback);
};

