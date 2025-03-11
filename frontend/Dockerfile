# Build the react app
FROM node:20-alpine AS builder
WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:stable-alpine
COPY --from=builder /app/build /usr/share/nginx/html

COPY public/config.template.js /usr/share/nginx/html/

COPY default.conf /etc/nginx/conf.d/default.conf

COPY entrypoint.sh /docker-entrypoint.d/entrypoint.sh
RUN chmod +x /docker-entrypoint.d/entrypoint.sh

EXPOSE 80
CMD ["/docker-entrypoint.d/entrypoint.sh", "nginx", "-g", "daemon off;"]
