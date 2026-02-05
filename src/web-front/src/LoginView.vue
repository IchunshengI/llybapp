<script setup>
import { ref } from "vue";

const emit = defineEmits(["logged-in"]);

const showRegister = ref(false);
const username = ref("");
const password = ref("");
const toastMessage = ref("");
const showToast = ref(false);
const loading = ref(false);
const backendDown = ref(false);

const regUsername = ref("");
const regPassword = ref("");
const regConfirmPassword = ref("");
const regLoading = ref(false);
const regRedirecting = ref(false);
let toastTimer;

const showToastFor = (msg) => {
  toastMessage.value = msg;
  showToast.value = true;
  if (toastTimer) clearTimeout(toastTimer);
  toastTimer = setTimeout(() => {
    showToast.value = false;
  }, 1200);
};

const apiBase = (() => {
  // Provided by Vite define (vite.config.js). Example:
  // BACKEND_HOST=http://47.112.172.204:18081
  const v = typeof __BACKEND_HOST__ === "string" ? __BACKEND_HOST__.trim() : "";
  return v ? v.replace(/\/+$/, "") : "";
})();

const handleLogin = async () => {
  if (loading.value) return;
  if (!apiBase) return showToastFor("未配置 BACKEND_HOST，无法请求后端");
  if (!username.value.trim()) return showToastFor("名称不能为空");
  if (!password.value) return showToastFor("密码不能为空");
  loading.value = true;
  backendDown.value = false;

  // Call backend first; if it fails, still enter Home for now (no auth yet).
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 1800);

  try {
    // Helps debugging in DevTools even if the request is blocked by CORS/mixed content.
    console.log("login ->", `${apiBase}/admin/login`);
    const res = await fetch(`${apiBase}/admin/login`, {
      method: "POST",
      headers: { "Content-Type": "text/plain" },
      body: JSON.stringify({ username: username.value, password: password.value }),
      signal: controller.signal,
    });
    // fetch() doesn't throw on 4xx/5xx; treat non-2xx as failure for now.
    if (!res.ok) throw new Error(`login failed: ${res.status}`);

    const data = await res.json().catch(() => ({}));
    if (data.ok) {
      showToastFor(data.message || "登录成功");
      await new Promise((r) => setTimeout(r, 400));
      emit("logged-in", { username: username.value.trim() });
      return;
    }
    showToastFor(data.message || "登录失败");
  } catch {
    backendDown.value = true;
    showToastFor("后端未正常启动");
  } finally {
    clearTimeout(timer);
    loading.value = false;
  }
};

const handleRegister = async () => {
  if (regLoading.value) return;
  if (!apiBase) return showToastFor("未配置 BACKEND_HOST，无法请求后端");

  const name = regUsername.value.trim();
  if (!name) return showToastFor("名称不能为空");
  if (!regPassword.value) return showToastFor("密码不能为空");
  if (!regConfirmPassword.value) return showToastFor("确认密码不能为空");
  if (regPassword.value !== regConfirmPassword.value) return showToastFor("两次密码不一致");

  regLoading.value = true;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), 3000);

  try {
    console.log("register ->", `${apiBase}/admin/register`);
    const res = await fetch(`${apiBase}/admin/register`, {
      method: "POST",
      headers: { "Content-Type": "text/plain" },
      body: JSON.stringify({ username: name, password: regPassword.value }),
      signal: controller.signal,
    });
    if (!res.ok) throw new Error(`register failed: ${res.status}`);

    const data = await res.json().catch(() => ({}));
    // Backend returns: { code, account_id, message }
    if (data.code === 0) {
      // Keep the modal open (background stays blurred) and show a short
      // "redirecting" state, so the login page doesn't flash back in.
      regRedirecting.value = true;
      showToastFor(data.message || "注册成功");
      await new Promise((r) => setTimeout(r, 700));
      emit("logged-in", { username: regUsername.value.trim() });
      return;
    }
    showToastFor(data.message || `注册失败（${data.code ?? "unknown"}）`);
  } catch {
    showToastFor("后端未正常启动");
  } finally {
    clearTimeout(timer);
    regLoading.value = false;
  }
};

const openRegister = () => {
  regRedirecting.value = false;
  showRegister.value = true;
};

const closeRegister = () => {
  if (regRedirecting.value) return;
  showRegister.value = false;
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
            <span>名称</span>
            <input
              v-model="username"
              :disabled="loading"
              type="text"
              placeholder="请输入名称"
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
          <button class="link-button" type="button" @click="openRegister">
            申请注册
          </button>
        </div>
      </section>
    </main>

    <div v-if="showToast" class="toast">{{ toastMessage }}</div>

    <div v-if="showRegister" class="modal-backdrop" @click="closeRegister"></div>
    <div v-if="showRegister" class="modal" role="dialog" aria-modal="true">
      <div class="modal-header">
        <h3>注册申请</h3>
        <button
          class="icon-button"
          type="button"
          :disabled="regRedirecting"
          @click="closeRegister"
        >
          ×
        </button>
      </div>
      <p class="modal-desc">填写申请信息，我们会尽快为你开通账号。</p>
      <div v-if="regRedirecting" class="form" aria-live="polite">
        <div style="display:flex;align-items:center;justify-content:center;gap:10px;padding:16px 0;">
          <span class="spinner" aria-hidden="true"></span>
          <span>正在进入 Home...</span>
        </div>
      </div>
      <form v-else class="form" @submit.prevent="handleRegister">
        <label class="field">
          <span>名称</span>
          <input v-model="regUsername" :disabled="regLoading" type="text" placeholder="请输入名称" />
        </label>
        <label class="field">
          <span>密码</span>
          <input
            v-model="regPassword"
            :disabled="regLoading"
            type="password"
            placeholder="••••••••"
          />
        </label>
        <label class="field">
          <span>确认密码</span>
          <input
            v-model="regConfirmPassword"
            :disabled="regLoading"
            type="password"
            placeholder="••••••••"
          />
        </label>
        <button class="primary" type="submit" :disabled="regLoading">
          <span v-if="regLoading" class="spinner" aria-hidden="true"></span>
          <span>{{ regLoading ? "提交中..." : "提交申请" }}</span>
        </button>
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
