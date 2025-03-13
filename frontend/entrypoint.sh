#!/bin/sh
# Replace placeholders in config.template.js with env variables at runtime
envsubst < /usr/share/nginx/html/config.template.js > /usr/share/nginx/html/config.js
# Execute the CMD provided to the container
exec "$@"
