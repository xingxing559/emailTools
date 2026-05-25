<script setup>
import { ref, watch } from 'vue'
import { GetSettings, SaveSettings, ClearCache } from '../../wailsjs/go/app/MailApp'

const props = defineProps({ open: { type: Boolean, default: false } })
const emit = defineEmits(['close', 'saved'])

const fetchDays = ref(7)
const refreshIntervalMinutes = ref(0)
const openLinksInBrowser = ref(true)
const saving = ref(false)
const clearing = ref(false)
const message = ref('')
const error = ref('')

const refreshOptions = [
  { value: 0, label: '关闭' },
  { value: 1, label: '1 分钟' },
  { value: 2, label: '2 分钟' },
  { value: 5, label: '5 分钟' },
  { value: 10, label: '10 分钟' },
  { value: 15, label: '15 分钟' },
  { value: 30, label: '30 分钟' },
]

watch(
  () => props.open,
  async (v) => {
    if (!v) return
    message.value = ''
    error.value = ''
    try {
      const s = await GetSettings()
      fetchDays.value = s.fetchDays || 7
      refreshIntervalMinutes.value = s.refreshIntervalMinutes ?? 0
      openLinksInBrowser.value = s.openLinksInBrowser !== false
    } catch (e) {
      error.value = e?.message || String(e)
    }
  }
)

async function handleSave() {
  saving.value = true
  error.value = ''
  message.value = ''
  try {
    await SaveSettings({
      fetchDays: Number(fetchDays.value),
      refreshIntervalMinutes: Number(refreshIntervalMinutes.value),
      openLinksInBrowser: !!openLinksInBrowser.value,
    })
    message.value = '设置已保存'
    emit('saved')
    emit('close')
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    saving.value = false
  }
}

async function handleClearCache() {
  if (!confirm('确定清空软件缓存？不会删除已保存的邮箱账号。')) return
  clearing.value = true
  error.value = ''
  message.value = ''
  try {
    await ClearCache()
    message.value = '缓存已清空'
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    clearing.value = false
  }
}
</script>

<template>
  <div v-if="open" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-card">
      <header class="modal-header">
        <h2>设置</h2>
        <button type="button" class="icon-btn" @click="emit('close')">×</button>
      </header>

      <label class="field">
        <span>邮件拉取天数（1–30 天）</span>
        <input v-model.number="fetchDays" type="number" min="1" max="30" />
        <p class="hint">列表将拉取该天数内收到的邮件，最多显示 500 封。</p>
      </label>

      <label class="field">
        <span>自动刷新间隔</span>
        <select v-model.number="refreshIntervalMinutes">
          <option v-for="opt in refreshOptions" :key="opt.value" :value="opt.value">
            {{ opt.label }}
          </option>
        </select>
        <p class="hint">登录后按间隔自动刷新当前文件夹邮件列表；0 表示关闭。</p>
      </label>

      <label class="field">
        <span>邮件链接打开方式</span>
        <select v-model="openLinksInBrowser">
          <option :value="true">系统默认浏览器（推荐）</option>
          <option :value="false">应用内打开</option>
        </select>
        <p class="hint">点击邮件正文中的链接时的打开方式。</p>
      </label>

      <p class="hint cache-hint">邮件列表会按日期缓存到本地，超过 7 天的缓存会自动清理。</p>

      <p v-if="error" class="error">{{ error }}</p>
      <p v-if="message" class="ok">{{ message }}</p>

      <div class="actions">
        <button type="button" class="btn-secondary" :disabled="clearing" @click="handleClearCache">
          {{ clearing ? '清空中…' : '清空缓存' }}
        </button>
        <button type="button" class="btn-primary" :disabled="saving" @click="handleSave">
          {{ saving ? '保存中…' : '保存' }}
        </button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.35);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.modal-card {
  width: 420px;
  max-width: calc(100% - 32px);
  background: #fff;
  border-radius: 12px;
  padding: 20px 24px 24px;
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.12);
}

.modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.modal-header h2 {
  font-size: 1.2rem;
}

.icon-btn {
  border: none;
  background: none;
  font-size: 1.5rem;
  cursor: pointer;
  color: #666;
}

.field {
  display: block;
  margin-bottom: 16px;
}

.field span {
  display: block;
  font-size: 0.9rem;
  margin-bottom: 8px;
}

.field input,
.field select {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.hint {
  margin-top: 8px;
  font-size: 0.8rem;
  color: #888;
}

.cache-hint {
  margin-top: 0;
  margin-bottom: 8px;
}

.error {
  color: #c0392b;
  font-size: 0.9rem;
  margin-top: 12px;
}

.ok {
  color: #2e7d32;
  font-size: 0.9rem;
  margin-top: 12px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 20px;
}

.btn-primary,
.btn-secondary {
  padding: 8px 16px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.9rem;
}

.btn-primary {
  background: #4a90d9;
  color: #fff;
  border: none;
}

.btn-secondary {
  background: #fff;
  border: 1px solid #ddd;
}

.btn-primary:disabled,
.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
