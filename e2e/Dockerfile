FROM mcr.microsoft.com/playwright:v1.60.0-noble

WORKDIR /e2e

COPY package.json package-lock.json ./
RUN npm ci

COPY playwright.config.js ./
COPY tests ./tests