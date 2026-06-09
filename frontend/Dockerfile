FROM node:20-alpine AS builder

ARG VITE_ONLYOFFICE_DS_URL=http://localhost:9980
ENV VITE_ONLYOFFICE_DS_URL=${VITE_ONLYOFFICE_DS_URL}

WORKDIR /build
COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:1.27-alpine
COPY nginx.conf /etc/nginx/nginx.conf
COPY --from=builder /build/dist /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
