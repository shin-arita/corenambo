// @ts-check
const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './tests',
  timeout: 60000,
  retries: 1,
  reporter: 'list',
  use: {
    baseURL: 'http://frontend:5173',
    actionTimeout: 10000,
  },
});