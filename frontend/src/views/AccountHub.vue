<script setup>
import { ref, onMounted } from 'vue'
import {
  ListAccounts,
  SwitchAccount,
  RemoveAccount,
} from '../../wailsjs/go/app/MailApp'

const emit = defineEmits(['enter-mailbox', 'add-account'])

const accounts = ref([])
const loading = ref(false)
const switchingId = ref('')
const error = ref('')

onMounted(loadAccounts)

async function loadAccounts() {
  loading.value = true
  error.value = ''
  try {
    accounts.value = await ListAccounts()
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    loading.value = false
  }
}

async function enterAccount(id) {
  switchingId.value = id
  error.value = ''
  try {
    const result = await SwitchAccount(id)
    if (result.success) {
      emit('enter-mailbox', result.email)
    } else {
      error.value = result.error || '连接失败'
    }
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    switchingId.value = ''
  }
}

async function deleteAccount(id) {
  if (!confirm('确定删除该邮箱账号？删除后需重新添加。')) return
  error.value = ''
  try {
    await RemoveAccount(id)
    await loadAccounts()
  } catch (e) {
    error.value = e?.message || String(e)
  }
}
</script>

<template>
  <div class="hub-page">
    <div class="hub-card">
      <h1>邮箱账号</h1>
      <p class="subtitle">选择要查看的 QQ 邮箱，或添加新账号</p>

      <p v-if="error" class="error">{{ error }}</p>
      <div v-if="loading" class="hint">加载中…</div>

      <ul v-else-if="accounts.length" class="account-list">
        <li v-for="acc in accounts" :key="acc.id" class="account-item">
          <div class="info">
            <div class="label">{{ acc.label || acc.email }}</div>
            <div class="email">{{ acc.email }}</div>
          </div>
          <div class="actions">
            <button
              class="btn-primary"
              :disabled="!!switchingId"
              @click="enterAccount(acc.id)"
            >
              {{ switchingId === acc.id ? '连接中…' : '进入邮箱' }}
            </button>
            <button class="btn-danger" @click="deleteAccount(acc.id)">删除</button>
          </div>
        </li>
      </ul>

      <div v-else class="hint">暂无已保存的邮箱账号</div>

      <button class="btn-add" @click="emit('add-account')">+ 添加邮箱</button>
    </div>
  </div>
</template>

<style scoped>
.hub-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100%;
  padding: 24px;
}

.hub-card {
  width: 100%;
  max-width: 520px;
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}

h1 {
  font-size: 1.5rem;
  margin-bottom: 8px;
}

.subtitle {
  color: #666;
  font-size: 0.9rem;
  margin-bottom: 20px;
}

.error {
  color: #c0392b;
  font-size: 0.9rem;
  margin-bottom: 12px;
}

.hint {
  color: #999;
  padding: 16px 0;
}

.account-list {
  list-style: none;
  margin-bottom: 20px;
}

.account-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 0;
  border-bottom: 1px solid #eee;
}

.label {
  font-weight: 600;
  font-size: 0.95rem;
}

.email {
  font-size: 0.85rem;
  color: #666;
  margin-top: 4px;
}

.actions {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex-shrink: 0;
}

.btn-primary {
  padding: 8px 14px;
  background: #4a90d9;
  color: #fff;
  border: none;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.85rem;
}

.btn-primary:disabled {
  opacity: 0.6;
}

.btn-danger {
  padding: 6px 14px;
  background: #fff;
  color: #c0392b;
  border: 1px solid #e8c4c4;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.8rem;
}

.btn-add {
  width: 100%;
  padding: 12px;
  background: #f0f7ff;
  color: #1a5fb4;
  border: 1px dashed #4a90d9;
  border-radius: 8px;
  cursor: pointer;
  font-size: 1rem;
}

.btn-add:hover {
  background: #e8f0fe;
}
</style>
