FROM nginx:alpine

# Copy the static website
COPY apps/mobile/web /usr/share/nginx/html

# Copy custom nginx configuration
COPY infrastructure/docker/nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"] 