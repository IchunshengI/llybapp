<script setup>
import { onBeforeUnmount, onMounted, ref } from "vue";
import LoginView from "./LoginView.vue";
import HomeLayout from "./home/HomeLayout.vue";

// Tiny client-side routing without vue-router:
// - /        -> Login
// - /home    -> Home
const routePath = ref(window.location.pathname || "/");
const currentUsername = ref("");

const setPath = (path) => {
  const next = path && path.startsWith("/") ? path : "/";
  window.history.pushState({}, "", next);
  routePath.value = window.location.pathname || "/";
};

const handleLoggedIn = (payload) => {
  const name = typeof payload === "string" ? payload : payload?.username;
  currentUsername.value = (name || "").trim();
  setPath("/home");
};
const handleLogout = () => setPath("/");

const handlePopState = () => {
  routePath.value = window.location.pathname || "/";
};

onMounted(() => {
  window.addEventListener("popstate", handlePopState);
});

onBeforeUnmount(() => {
  window.removeEventListener("popstate", handlePopState);
});
</script>

<template>
  <LoginView v-if="!routePath.startsWith('/home')" @logged-in="handleLoggedIn" />
  <HomeLayout v-else :username="currentUsername" @logout="handleLogout" />
</template>
