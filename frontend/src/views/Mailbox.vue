<script setup>
import { ref, onMounted, watch, computed } from 'vue'
import MailDetail from '../components/MailDetail.vue'
import { displayFolderName } from '../utils/folderName.js'
import {
  GetCurrentEmail,
  ListFolders,
  ListMessages,
  GetMessage,
  Logout,
} from '../../wailsjs/go/app/MailApp'

const props = defineProps({
  email: { type: String, default: '' },
})
const emit = defineEmits(['disconnect', 'switch-account'])

const accountEmail = ref(props.email)
const folders = ref([])
const selectedFolder = ref('')
const messages = ref([])
const selectedUid = ref(null)
const detail = ref(null)
const loadingFolders = ref(false)
const loadingMessages = ref(false)
const loadingDetail = ref(false)
const error = ref('')

const sortedMessages = computed(() => {
  return [...messages.value].sort((a, b) => (b.dateUnix || 0) - (a.dateUnix || 0))
})

onMounted(async () => {
  if (!accountEmail.value) {
    try {
      accountEmail.value = await GetCurrentEmail()
    } catch (_) {}
  }
  await loadFolders()
})

watch(selectedFolder, async (name) => {
  if (name) await loadMessages()
})

async function loadFolders() {
  loadingFolders.value = true
  error.value = ''
  try {
    folders.value = await ListFolders()
    if (folders.value.length && !selectedFolder.value) {
      const inbox = folders.value.find((f) =>
        /INBOX|收件箱/i.test(f.name) || /收件箱/.test(f.displayName || '')
      )
      selectedFolder.value = inbox?.name || folders.value[0].name
    }
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    loadingFolders.value = false
  }
}

async function loadMessages() {
  loadingMessages.value = true
  selectedUid.value = null
  detail.value = null
  error.value = ''
  try {
    messages.value = await ListMessages(selectedFolder.value, 0, 50)
    const list = sortedMessages.value
    if (list.length) {
      await selectMessage(list[0].uid)
    }
  } catch (e) {
    error.value = e?.message || String(e)
    messages.value = []
  } finally {
    loadingMessages.value = false
  }
}

async function selectMessage(uid) {
  selectedUid.value = uid
  loadingDetail.value = true
  try {
    detail.value = await GetMessage(selectedFolder.value, uid)
  } catch (e) {
    error.value = e?.message || String(e)
    detail.value = null
  } finally {
    loadingDetail.value = false
  }
}

async function handleRefresh() {
  await loadMessages()
}

async function handleDisconnect() {
  try {
    await Logout()
  } catch (_) {}
  emit('disconnect')
}
</script>

<template>
  <div class="mailbox">
    <header class="toolbar">
      <span class="account">{{ accountEmail }}</span>
      <div class="actions">
        <button class="btn-secondary" :disabled="loadingMessages" @click="handleRefresh">
          刷新
        </button>
        <button class="btn-secondary" @click="emit('switch-account')">切换账号</button>
        <button class="btn-secondary" @click="handleDisconnect">断开连接</button>
      </div>
    </header>

    <p v-if="error" class="banner-error">{{ error }}</p>

    <div class="main-panels">
      <aside class="folders">
        <h3>文件夹</h3>
        <div v-if="loadingFolders" class="hint">加载中…</div>
        <ul v-else>
          <li
            v-for="f in folders"
            :key="f.name"
            :class="{ active: f.name === selectedFolder }"
            @click="selectedFolder = f.name"
          >
            {{ displayFolderName(f) }}
          </li>
        </ul>
      </aside>

      <section class="message-list">
        <h3>邮件列表</h3>
        <div v-if="loadingMessages" class="hint">加载中…</div>
        <ul v-else-if="sortedMessages.length">
          <li
            v-for="m in sortedMessages"
            :key="m.uid"
            :class="{ active: m.uid === selectedUid, unread: !m.seen }"
            @click="selectMessage(m.uid)"
          >
            <div class="subject">{{ m.subject }}</div>
            <div class="meta">
              <span class="from">{{ m.from }}</span>
              <span class="date">{{ m.date }}</span>
            </div>
          </li>
        </ul>
        <div v-else class="hint">暂无邮件</div>
      </section>

      <MailDetail :detail="detail" :loading="loadingDetail" />
    </div>
  </div>
</template>

<style scoped>
.mailbox {
  display: flex;
  flex-direction: column;
  height: 100%;
}

.toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: #fff;
  border-bottom: 1px solid #e8eaed;
}

.account {
  font-weight: 600;
  color: #333;
}

.actions {
  display: flex;
  gap: 8px;
}

.btn-secondary {
  padding: 6px 14px;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: #fff;
  cursor: pointer;
  font-size: 0.9rem;
}

.btn-secondary:hover:not(:disabled) {
  background: #f0f4f8;
}

.btn-secondary:disabled {
  opacity: 0.5;
}

.banner-error {
  padding: 8px 16px;
  background: #fdecea;
  color: #c0392b;
  font-size: 0.9rem;
}

.main-panels {
  display: flex;
  flex: 1;
  min-height: 0;
}

.folders {
  width: 180px;
  background: #fff;
  border-right: 1px solid #e8eaed;
  overflow-y: auto;
}

.folders h3,
.message-list h3 {
  padding: 12px;
  font-size: 0.85rem;
  color: #666;
  border-bottom: 1px solid #eee;
}

.folders ul,
.message-list ul {
  list-style: none;
}

.folders li {
  padding: 10px 12px;
  cursor: pointer;
  font-size: 0.9rem;
  border-bottom: 1px solid #f5f5f5;
}

.folders li:hover,
.folders li.active {
  background: #e8f0fe;
  color: #1a5fb4;
}

.message-list {
  width: 320px;
  background: #fff;
  border-right: 1px solid #e8eaed;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
}

.message-list ul {
  flex: 1;
  overflow-y: auto;
}

.message-list li {
  padding: 12px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
}

.message-list li:hover,
.message-list li.active {
  background: #f0f7ff;
}

.message-list li.unread .subject {
  font-weight: 600;
}

.subject {
  font-size: 0.9rem;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.75rem;
  color: #888;
}

.meta .from {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  max-width: 60%;
}

.hint {
  padding: 16px;
  color: #999;
  font-size: 0.9rem;
}
</style>
