import { defineConfig } from "vite";
import solid from "vite-plugin-solid";

export default defineConfig({
    plugins: [solid()],
    server: {
        proxy: {
            "/api": { target: "http://localhost:7331", changeOrigin: true },
            "/static": { target: "http://localhost:7331", changeOrigin: true },
        },
    },
});
