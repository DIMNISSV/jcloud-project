<template>
  <UCard class="max-w-sm w-full">
    <template #header>
      <h1 class="text-2xl font-bold text-center">Login to JCloud</h1>
    </template>

    <UForm :state="state" @submit="handleLogin">
      <UFormGroup label="Email" name="email" class="mb-4">
        <UInput v-model="state.email" type="email" placeholder="you@example.com" required />
      </UFormGroup>

      <UFormGroup label="Password" name="password" class="mb-6">
        <UInput v-model="state.password" type="password" required />
      </UFormGroup>

      <UButton type="submit" block size="lg" :loading="loading">
        Login
      </UButton>
    </UForm>

    <div v-if="errorMsg" class="mt-4 text-center text-red-500">
      {{ errorMsg }}
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/store/auth';

const state = reactive({
  email: '',
  password: '',
});

const loading = ref(false);
const errorMsg = ref<string | null>(null);

const authStore = useAuthStore();
const router = useRouter();

async function handleLogin() {
  loading.value = true;
  errorMsg.value = null;

  try {
    const response = await $fetch<{ token: string }>('/api/v1/users/login', {
      method: 'POST',
      body: {
        email: state.email,
        password: state.password,
      },
    });

    authStore.setToken(response.token);

    // In a real app, you would also fetch user details here
    // const user = await $fetch('/api/v1/users/me', ...);
    // authStore.setUser(user);

    await router.push('/');

  } catch (error: any) {
    errorMsg.value = 'Failed to login. Please check your credentials.';
    console.error(error);
  } finally {
    loading.value = false;
  }
}
</script>