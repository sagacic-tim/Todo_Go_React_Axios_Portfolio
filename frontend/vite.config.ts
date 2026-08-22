import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

export default defineConfig({
  root: ".",            // project root
  base: "/",            // publicPath
  plugins: [react()],
  css: {
    modules: {
      // turn kebab-case class-names into camelCase keys
      localsConvention: "camelCaseOnly"
    }
  },
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:8080",
        changeOrigin: true,
        secure: false
      }
    }
  }
})
