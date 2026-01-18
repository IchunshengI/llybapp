<script setup>
import { ref } from "vue";

const showRegister = ref(false);

</script>

<template>
  <div class="page">
    <div class="orb orb-1"></div>
    <div class="orb orb-2"></div>
    <div class="grid"></div>

    <main class="shell" :class="{ blurred: showRegister }" aria-hidden="showRegister">
      <section class="hero">
        <span class="pill">LLYB 控制台</span>
        <h1>欢迎回来</h1>
        <p>命 / 势，由我们负责解答</p>
      </section>

      <section class="panel">
        <div class="panel-header">
          <h2>账号登录</h2>
          <p>使用你的工作区账号信息登录。</p>
        </div>

        <form class="form" @submit.prevent>
          <label class="field">
            <span>邮箱</span>
            <input type="email" placeholder="you@company.com" />
          </label>
          <label class="field">
            <span>密码</span>
            <input type="password" placeholder="••••••••" />
          </label>
          <div class="row">
            <label class="check">
              <input type="checkbox" />
              <span>保持登录状态</span>
            </label>
            <a class="link" href="#">忘记密码？</a>
          </div>
          <button class="primary" type="submit">登录</button>
        </form>

        <div class="panel-footer">
          <span>没有账号？</span>
          <button class="link-button" type="button" @click="showRegister = true">
            申请注册
          </button>
        </div>
      </section>
    </main>

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
