// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  modules: [
    "@nuxt/ui",
    "@pinia/nuxt",
    '@nuxtjs/tailwindcss',
  ],

  routeRules: {
    '/api/**': {
      proxy: {
        to: "http://localhost:8080/api/**", // Base path for user-service
      }
    },
  }
})