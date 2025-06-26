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

# Copy the build output to nginx html directory
COPY --from=build /app/build /usr/share/nginx/html

# Copy custom nginx configuration
COPY nginx.conf /etc/nginx/nginx.conf

# Expose port 80
EXPOSE 80

# Start nginx
CMD ["nginx", "-g", "daemon off;"]
