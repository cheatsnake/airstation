import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// https://vite.dev/config/
export default defineConfig({
    base: "/studio/",
    plugins: [react()],
    server: {
        proxy: {
            "/api": { target: "http://localhost:7331", changeOrigin: true },
            "/static": { target: "http://localhost:7331", changeOrigin: true },
            "/stream": { target: "http://localhost:7331", changeOrigin: true },
        },
    },
});
