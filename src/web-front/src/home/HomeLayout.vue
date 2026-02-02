<script setup>
import { computed, ref } from "vue";
import TestFeature from "./pages/TestFeature.vue";

const emit = defineEmits(["logout"]);

const activeKey = ref("test");
const menus = [
  { key: "test", label: "基础推理" },
];

const activeLabel = computed(() => menus.find((m) => m.key === activeKey.value)?.label ?? "");
</script>

<template>
  <div class="home">
    <aside class="sidebar">
      <div class="brand">
        <div class="logo">LLYB</div>
        <div class="title">
          <div class="name">控制台</div>
        </div>
      </div>

      <nav class="nav">
        <button
          v-for="m in menus"
          :key="m.key"
          class="nav-item"
          :class="{ active: activeKey === m.key }"
          type="button"
          @click="activeKey = m.key"
        >
          {{ m.label }}
        </button>
      </nav>

      <div class="sidebar-footer">
        <button class="logout" type="button" @click="emit('logout')">退出登录</button>
      </div>
    </aside>

    <section class="content">
      <header class="topbar">
        <div class="crumb">Home / {{ activeLabel }}</div>
      </header>

      <main class="main">
        <TestFeature v-if="activeKey === 'test'" />
      </main>
    </section>
  </div>
</template>

<style scoped>
.home {
  min-height: 100vh;
  display: flex;
  background: #0b1012;
  color: rgba(255, 255, 255, 0.92);
}

.sidebar {
  width: 240px;
  background: rgba(255, 255, 255, 0.04);
  border-right: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  flex-direction: column;
  padding: 16px;
}

.brand {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 10px 10px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.logo {
  width: 40px;
  height: 40px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-weight: 800;
  letter-spacing: 0.04em;
  color: #0b1012;
  background: linear-gradient(135deg, #4dd7c8, #ff7a59);
}

.name {
  font-weight: 700;
  line-height: 1.1;
}

.sub {
  font-size: 12px;
  opacity: 0.7;
  margin-top: 2px;
}

.nav {
  display: grid;
  gap: 8px;
  padding: 14px 6px;
}

.nav-item {
  text-align: left;
  border: 1px solid rgba(255, 255, 255, 0.10);
  background: rgba(255, 255, 255, 0.03);
  color: rgba(255, 255, 255, 0.9);
  border-radius: 12px;
  padding: 10px 12px;
  cursor: pointer;
}

.nav-item:hover {
  background: rgba(255, 255, 255, 0.06);
}

.nav-item.active {
  border-color: rgba(77, 215, 200, 0.65);
  background: rgba(77, 215, 200, 0.12);
}

.sidebar-footer {
  margin-top: auto;
  padding: 12px 6px 6px;
}

.logout {
  width: 100%;
  border: 1px solid rgba(255, 255, 255, 0.12);
  background: transparent;
  color: rgba(255, 255, 255, 0.85);
  border-radius: 12px;
  padding: 10px 12px;
  cursor: pointer;
}

.logout:hover {
  background: rgba(255, 255, 255, 0.06);
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.topbar {
  height: 56px;
  display: flex;
  align-items: center;
  padding: 0 18px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.03);
}

.crumb {
  font-size: 13px;
  opacity: 0.75;
}

.main {
  padding: 18px;
}
</style>
