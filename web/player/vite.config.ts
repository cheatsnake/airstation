import { defineConfig } from "vite";
import solid from "vite-plugin-solid";
import { VitePWA } from "vite-plugin-pwa";

export default defineConfig({
    plugins: [
        solid(),
        VitePWA({
            scope: "/",
            registerType: "autoUpdate",
            workbox: { cleanupOutdatedCaches: true },
            devOptions: {
                enabled: true,
            },
            manifest: {
                scope: "/",
                start_url: "/",
                lang: "en",
                name: "Radio",
                short_name: "Radio",
                icons: [
                    {
                        src: "icon48.png",
                        sizes: "48x48",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                    {
                        src: "icon72.png",
                        sizes: "72x72",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                    {
                        src: "icon96.png",
                        sizes: "96x96",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                    {
                        src: "icon128.png",
                        sizes: "128x128",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                    {
                        src: "icon144.png",
                        sizes: "144x144",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                    {
                        src: "icon152.png",
                        sizes: "152x152",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                    {
                        src: "icon192.png",
                        sizes: "192x192",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                    {
                        src: "icon256.png",
                        sizes: "256x256",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                    {
                        src: "icon512.png",
                        sizes: "512x512",
                        type: "image/png",
                        purpose: "maskable any",
                    },
                ],
            },
        }),
    ],
    server: {
        proxy: {
            "/api": { target: "http://localhost:7331", changeOrigin: true },
            "/stream": { target: "http://localhost:7331", changeOrigin: true },
            "/static": { target: "http://localhost:7331", changeOrigin: true },
        },
    },
});
