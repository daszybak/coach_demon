# Stage 1: build
FROM node:22 as builder
WORKDIR /app
COPY . .
RUN npm install && npm run build

# Stage 2: serve static files
FROM nginx:alpine
COPY --from=builder /app/dist /usr/share/nginx/html
EXPOSE 80
