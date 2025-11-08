<template>
  <UCard>
    <template #header>
      <h1 class="text-xl font-bold">Welcome to your Dashboard</h1>
    </template>

    <p>You are successfully logged in.</p>

    <template #footer>
      <UButton color="error" @click="handleLogout">Logout</UButton>
    </template>
  </UCard>
</template>

<script setup lang="ts">
import { useAuthStore } from '~/store/auth';

const authStore = useAuthStore();
const router = useRouter();

function handleLogout() {
  authStore.logout();
  router.push('/login');
}
</script>