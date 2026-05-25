<script setup>
import { ref, onMounted } from 'vue'
import Login from './views/Login.vue'
import MainShell from './views/MainShell.vue'
import {
  TryAutoLogin,
  IsLoggedIn,
  GetCurrentEmail,
  HasAccounts,
} from '../wailsjs/go/app/MailApp'

const view = ref('loading')
const shellRef = ref(null)

onMounted(async () => {
  try {
    if (await IsLoggedIn()) {
      view.value = 'main'
      return
    }
    const result = await TryAutoLogin()
    if (result.success) {
      view.value = 'main'
      return
    }
    if (await HasAccounts()) {
      view.value = 'main'
    } else {
      view.value = 'add'
    }
  } catch (e) {
    console.error(e)
    view.value = 'add'
  }
})

function onLoginSuccess() {
  view.value = 'main'
}

async function goBackFromAdd() {
  try {
    view.value = (await HasAccounts()) ? 'main' : 'add'
  } catch {
    view.value = 'main'
  }
}

function onAddAccount() {
  view.value = 'add'
}

async function onDisconnected() {
  try {
    if (!(await HasAccounts())) {
      view.value = 'add'
    }
  } catch {
    view.value = 'add'
  }
}
</script>

<template>
  <div class="app-root">
    <div v-if="view === 'loading'" class="loading-screen">正在加载…</div>
    <MainShell
      v-else-if="view === 'main'"
      ref="shellRef"
      @add-account="onAddAccount"
      @disconnected="onDisconnected"
    />
    <Login
      v-else
      @login-success="onLoginSuccess"
      @back="goBackFromAdd"
    />
  </div>
</template>

<style>
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

html, body, #app {
  height: 100%;
  font-family: "Segoe UI", "Microsoft YaHei", sans-serif;
  background: #f5f7fa;
  color: #1a1a2e;
}

.app-root {
  height: 100%;
}

.loading-screen {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  color: #666;
}
</style>