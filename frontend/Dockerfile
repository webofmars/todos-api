# Stage 1: Build the React application (force AMD64)
FROM --platform=linux/amd64 node:18-alpine AS build

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm install --only=production

# Copy source code
COPY . .

# Build the application
RUN npm run build

# Stage 2: Serve the application with nginx (force AMD64)
FROM --platform=linux/amd64 nginx:alpine

# Install gettext for envsubst
RUN apk add --no-cache gettext

# Copy the build output to nginx html directory
COPY --from=build /app/build /usr/share/nginx/html

# Copy nginx configuration template and startup script
COPY nginx.conf.template /etc/nginx/nginx.conf.template
COPY docker-entrypoint.sh /docker-entrypoint.sh

# Make the startup script executable
RUN chmod +x /docker-entrypoint.sh

# Expose port 80
EXPOSE 80

# Use our custom entrypoint
ENTRYPOINT ["/docker-entrypoint.sh"]
