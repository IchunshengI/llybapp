<script setup>
import { ref } from "vue";
import LoginView from "./LoginView.vue";
import HomeLayout from "./home/HomeLayout.vue";

// No backend yet: keep routing purely client-side.
const authed = ref(false);

const handleLoggedIn = () => {
  authed.value = true;
};

const handleLogout = () => {
  authed.value = false;
};
</script>

<template>
  <LoginView v-if="!authed" @logged-in="handleLoggedIn" />
  <HomeLayout v-else @logout="handleLogout" />
</template>

