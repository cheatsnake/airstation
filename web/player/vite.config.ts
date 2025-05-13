import { defineConfig, loadEnv } from "vite";
import solid from "vite-plugin-solid";
import { VitePWA } from "vite-plugin-pwa";
import path from "path";

export default defineConfig(({ mode }) => {
    const globalEnv = loadEnv(mode, path.join(process.cwd(), "..", ".."), "");
    const localEnv = loadEnv(mode, process.cwd(), "");
    const appTitle = globalEnv.AIRSTATION_PLAYER_TITLE || localEnv.AIRSTATION_PLAYER_TITLE || "Radio";
    return {
        plugins: [
            solid(),
            VitePWA({
                scope: "/",
                registerType: "autoUpdate",
                workbox: {
                    cleanupOutdatedCaches: true,
                    navigateFallback: "/index.html",
                    navigateFallbackDenylist: [/^\/studio\//],
                },
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
        envPrefix: "AIRSTATION_PLAYER_",
        define: {
            "import.meta.env.AIRSTATION_PLAYER_TITLE": JSON.stringify(appTitle),
        },
    };
});
