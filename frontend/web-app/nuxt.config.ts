import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2025-07-15',
  devtools: { enabled: true },

  modules: [
    "@nuxt/ui",
    "@pinia/nuxt",
  ],

  nitro: {
    alias: {
      pinia: 'pinia'
    }
  },

  pinia: {
    storesDirs: ['./app/store/**'],
  },

  css: ['~/assets/css/main.css'],
  colorMode: { classSuffix: '' },

  vite: { plugins: [tailwindcss()] },

  routeRules: {
    '/api/**': {
      proxy: {
        to: "http://localhost:8080/api/**",
      }
    },
  }
})