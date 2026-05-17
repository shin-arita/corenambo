import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";

const validHostname = /^[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?(\.[a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$/;

const extraAllowedHosts = process.env.VITE_SERVER_ALLOWED_HOSTS
  ? process.env.VITE_SERVER_ALLOWED_HOSTS
      .split(",")
      .map((h) => h.trim())
      .filter(Boolean)
      .filter((h) => validHostname.test(h))
  : undefined;

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    allowedHosts: extraAllowedHosts,
  },
});
