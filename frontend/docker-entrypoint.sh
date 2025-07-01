#!/bin/sh

# Set default values for environment variables
export API_HOST=${API_HOST:-api}
export API_PORT=${API_PORT:-8080}

echo "🔧 Configuring nginx with API_HOST=${API_HOST} and API_PORT=${API_PORT}"

# Generate nginx.conf from template by substituting environment variables
envsubst '${API_HOST} ${API_PORT}' < /etc/nginx/nginx.conf.template > /etc/nginx/nginx.conf

echo "✅ Nginx configuration generated from template:"
echo "   Template: /etc/nginx/nginx.conf.template"
echo "   Generated: /etc/nginx/nginx.conf"
echo "   API upstream: http://${API_HOST}:${API_PORT}"

# Start nginx
exec nginx -g "daemon off;"
