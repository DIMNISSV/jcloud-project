// app/store/auth.ts

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: null as string | null,
    user: null as object | null,
  }),

  getters: {
    isLoggedIn: (state) => !!state.token,
  },

  actions: {
    setToken(newToken: string) {
      this.token = newToken
    },
    setUser(newUser: object) {
      this.user = newUser
    },
    logout() {
      this.token = null
      this.user = null
    },
  },
})