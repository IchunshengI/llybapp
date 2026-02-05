import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";

export default defineConfig(({ mode }) => {
  // Allow using BACKEND_HOST without the VITE_ prefix by exposing it as a global
  // compile-time constant. This is loaded from .env* files and/or container env.
  const env = loadEnv(mode, process.cwd(), "");

  return {
    plugins: [vue()],
    define: {
      __BACKEND_HOST__: JSON.stringify(env.BACKEND_HOST || ""),
    },
  };
});
