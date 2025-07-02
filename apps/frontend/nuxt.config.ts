import tailwindcss from "@tailwindcss/vite";

// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: "2025-05-15",
  devtools: { enabled: true },
  css: ["~/assets/css/tailwind.css"],

  vite: {
    plugins: [tailwindcss()],
  },
  runtimeConfig: {
    public: {
      appName: process.env.APP_NAME,
      clerkPublishableKey: process.env.CLERK_PUBLISHABLE_KEY,
      backendUrl: process.env.BACKEND_URL || 'http://localhost:8000',
    },
    clerkSecretKey: process.env.CLERK_SECRET_KEY,
    r2AccountId: process.env.R2_ACCOUNT_ID,
    r2AccessKeyId: process.env.R2_ACCESS_KEY_ID,
    r2SecretAccessKey: process.env.R2_SECRET_ACCESS_KEY,
    r2Endpoint: process.env.R2_ENDPOINT,
    r2BucketName: process.env.R2_BUCKET_NAME,
    togetherApiKey: process.env.TOGETHER_API_KEY,
  },

  modules: ["shadcn-nuxt", "@clerk/nuxt", "@nuxtjs/color-mode", "@vueuse/nuxt"],
  shadcn: {
    /**
     * Prefix for all the imported component
     */
    prefix: "",
    /**
     * Directory that the component lives in.
     * @default "./components/ui"
     */
    componentDir: "./components/ui",
  },
  colorMode: {
    classSuffix: "",
    fallback: "light",
    storageKey: "color-mode",
  },
  components: [
    { path: "~/components", pathPrefix: false },
    { path: "~/components/auth", pathPrefix: true, prefix: "Auth" },
    { path: "~/components/layout", pathPrefix: true, prefix: "Layout" },
  ],
});
