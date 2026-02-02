<script setup>
import { ref } from "vue";

const emit = defineEmits(["logged-in"]);

const showRegister = ref(false);
const email = ref("");
const password = ref("");
const toastMessage = ref("");
const showToast = ref(false);
const loading = ref(false);
const backendDown = ref(false);
let toastTimer;

const showToastFor = (msg) => {
  toastMessage.value = msg;
  showToast.value = true;
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    showToast.value = false;
  }, 1200);
};

const handleLogin = async () => {
  if (loading.value) return;
  loading.value = true;
  backendDown.value = false;

  // Call backend first; if it fails, still enter Home for now (no auth yet).
  const apiBase = "http://llyb-backend:8081";
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 1800);

  try {
    const res = await fetch(`${apiBase}/admin/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ username: email.value, password: password.value }),
      signal: controller.signal,
    });
    // fetch() doesn't throw on 4xx/5xx; treat non-2xx as failure for now.
    if (!res.ok) throw new Error(`login failed: ${res.status}`);

    showToastFor("登录成功");
  } catch {
    backendDown.value = true;
    showToastFor("后端未正常启动，已进入 Home（模拟）");
  } finally {
    clearTimeout(timer);
    loading.value = false;
    // Give the user a moment to see the warning/toast before leaving the page.
    if (backendDown.value) {
      await new Promise((r) => setTimeout(r, 800));
    }
    emit("logged-in");
  }
};
</script>

<template>
  <div class="page">
    <div class="orb orb-1"></div>
    <div class="orb orb-2"></div>
    <div class="grid"></div>

    <main class="shell" :class="{ blurred: showRegister }" :aria-hidden="showRegister">
      <section class="hero">
        <span class="pill">LLYB 控制台</span>
        <h1>欢迎回来</h1>
        <p>命 / 势，由我们负责解答</p>
      </section>

      <section class="panel">
        <div class="panel-header">
          <h2>账号登录</h2>
          <p>后端未接入完成前：登录接口失败也会进入 Home。</p>
        </div>

        <form class="form" @submit.prevent="handleLogin">
          <label class="field">
            <span>邮箱</span>
            <input
              v-model="email"
              :disabled="loading"
              type="email"
              placeholder="you@company.com"
            />
          </label>
          <label class="field">
            <span>密码</span>
            <input
              v-model="password"
              :disabled="loading"
              type="password"
              placeholder="••••••••"
            />
          </label>
          <div class="row">
            <label class="check">
              <input :disabled="loading" type="checkbox" />
              <span>保持登录状态</span>
            </label>
            <a class="link" href="#">忘记密码？</a>
          </div>
          <button class="primary" type="submit" :disabled="loading">
            <span v-if="loading" class="spinner" aria-hidden="true"></span>
            <span>{{ loading ? "登录中..." : "登录" }}</span>
          </button>
          <p v-if="backendDown" class="backend-down">后端未正常启动</p>
        </form>

        <div class="panel-footer">
          <span>没有账号？</span>
          <button class="link-button" type="button" @click="showRegister = true">
            申请注册
          </button>
        </div>
      </section>
    </main>

    <div v-if="showToast" class="toast">{{ toastMessage }}</div>

    <div v-if="showRegister" class="modal-backdrop" @click="showRegister = false"></div>
    <div v-if="showRegister" class="modal" role="dialog" aria-modal="true">
      <div class="modal-header">
        <h3>注册申请</h3>
        <button class="icon-button" type="button" @click="showRegister = false">
          ×
        </button>
      </div>
      <p class="modal-desc">填写申请信息，我们会尽快为你开通账号。</p>
      <form class="form" @submit.prevent>
        <label class="field">
          <span>昵称</span>
          <input type="text" placeholder="小李" />
        </label>
        <label class="field">
          <span>邮箱</span>
          <input type="email" placeholder="you@company.com" />
        </label>
        <label class="field">
          <span>密码</span>
          <input type="password" placeholder="••••••••" />
        </label>
        <label class="field">
          <span>确认密码</span>
          <input type="password" placeholder="••••••••" />
        </label>
        <button class="primary" type="submit">提交申请</button>
      </form>
    </div>
  </div>
</template>

<style scoped>
.backend-down {
  margin-top: 10px;
  font-size: 0.9rem;
  color: rgba(255, 122, 89, 0.95);
  font-weight: 600;
}

.spinner {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid rgba(255, 255, 255, 0.35);
  border-top-color: rgba(255, 255, 255, 0.95);
  display: inline-block;
  margin-right: 10px;
  animation: spin 0.9s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
