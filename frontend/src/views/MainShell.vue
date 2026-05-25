<script setup>
import { ref, onMounted, onUnmounted, watch, computed } from 'vue'
import MailDetail from '../components/MailDetail.vue'
import SettingsModal from '../components/SettingsModal.vue'
import AccountEditModal from '../components/AccountEditModal.vue'
import { displayFolderName } from '../utils/folderName.js'
import { providerBadge } from '../utils/providerMeta.js'
import { usePaneResize } from '../composables/usePaneResize.js'

const { sidebarWidth, folderWidth, listWidth, activePane, startResize } = usePaneResize()
import {
  ListAccounts,
  SwitchAccount,
  PeekAccountCache,
  RemoveAccount,
  ListFolders,
  ListMessages,
  ListMessagesCached,
  GetMessage,
  Logout,
  IsLoggedIn,
  GetCurrentEmail,
  GetSettings,
} from '../../wailsjs/go/app/MailApp'

const emit = defineEmits(['add-account', 'disconnected'])

const accounts = ref([])
const accountEmail = ref('')
const loggedIn = ref(false)
const settingsOpen = ref(false)
const accountEditOpen = ref(false)
const editingAccount = ref(null)
const contextMenu = ref({ show: false, x: 0, y: 0, account: null })
const switchingAccount = ref(false)
const switchingId = ref('')
const activeAccountId = ref('')
const sidebarError = ref('')

const folders = ref([])
const selectedFolder = ref('')
const messages = ref([])
const selectedUid = ref(null)
const detail = ref(null)
const loadingFolders = ref(false)
const loadingMessages = ref(false)
const loadingDetail = ref(false)
const mailError = ref('')
const openLinksInBrowser = ref(true)

let refreshTimer = null

const sortedMessages = computed(() =>
  [...messages.value].sort((a, b) => (b.dateUnix || 0) - (a.dateUnix || 0))
)

const activeAccount = computed(() =>
  accounts.value.find((a) => a.email === accountEmail.value)
)

const currentAccountTag = computed(() => {
  const acc = activeAccount.value
  return acc?.providerTag || providerBadge(acc?.provider).label
})

function accountBadge(acc) {
  const tag = acc.providerTag || providerBadge(acc.provider).label
  const cls = providerBadge(acc.provider).badgeClass
  return { tag, cls }
}

onMounted(async () => {
  document.addEventListener('click', closeContextMenu)
  await refreshAccounts()
  loggedIn.value = await IsLoggedIn()
  if (loggedIn.value) {
    accountEmail.value = await GetCurrentEmail()
    const active = accounts.value.find((a) => a.isActive)
    if (active) activeAccountId.value = active.id
    await loadFolders()
  }
  await loadLinkSetting()
  await setupAutoRefresh()
})

async function loadLinkSetting() {
  try {
    const s = await GetSettings()
    openLinksInBrowser.value = s.openLinksInBrowser !== false
  } catch (_) {
    openLinksInBrowser.value = true
  }
}

onUnmounted(() => {
  document.removeEventListener('click', closeContextMenu)
  clearAutoRefresh()
})

function closeContextMenu() {
  contextMenu.value.show = false
}

function openAccountContextMenu(event, acc) {
  contextMenu.value = {
    show: true,
    x: event.clientX,
    y: event.clientY,
    account: acc,
  }
}

function openAccountEdit() {
  const acc = contextMenu.value.account
  closeContextMenu()
  if (!acc) return
  editingAccount.value = { ...acc }
  accountEditOpen.value = true
}

async function deleteAccountFromMenu() {
  const acc = contextMenu.value.account
  closeContextMenu()
  if (!acc) return
  if (!confirm(`确定删除邮箱 ${acc.email}？删除后需重新添加。`)) return
  sidebarError.value = ''
  try {
    const wasCurrent = loggedIn.value && accountEmail.value === acc.email
    await RemoveAccount(acc.id)
    await refreshAccounts()
    if (wasCurrent) {
      loggedIn.value = false
      accountEmail.value = ''
      folders.value = []
      messages.value = []
      detail.value = null
      selectedFolder.value = ''
      emit('disconnected')
    }
  } catch (e) {
    sidebarError.value = e?.message || String(e)
  }
}

async function onAccountSaved(result) {
  const editedEmail = editingAccount.value?.email
  await refreshAccounts()
  if (loggedIn.value && editedEmail && accountEmail.value === editedEmail) {
    if (result?.email) accountEmail.value = result.email
    await loadFolders()
    if (selectedFolder.value) await loadMessages()
  }
}

watch(selectedFolder, async (name) => {
  if (name && loggedIn.value && !switchingAccount.value) await loadMessages()
})

async function refreshAccounts() {
  try {
    accounts.value = await ListAccounts()
  } catch (e) {
    sidebarError.value = e?.message || String(e)
  }
}

function clearAutoRefresh() {
  if (refreshTimer) {
    clearInterval(refreshTimer)
    refreshTimer = null
  }
}

async function setupAutoRefresh() {
  clearAutoRefresh()
  try {
    const s = await GetSettings()
    const minutes = s.refreshIntervalMinutes || 0
    if (minutes <= 0) return
    refreshTimer = setInterval(() => {
      if (loggedIn.value && selectedFolder.value && !switchingAccount.value) {
        loadMessages({ silent: true })
      }
    }, minutes * 60 * 1000)
  } catch (_) {}
}

function applyPeekCache(peek) {
  if (peek?.folders?.length) {
    folders.value = peek.folders
  }
  if (peek?.lastFolder) {
    selectedFolder.value = peek.lastFolder
  }
  if (peek?.messages?.length) {
    messages.value = peek.messages
    selectedUid.value = null
    detail.value = null
  }
}

async function selectAccount(id) {
  if (switchingAccount.value) return
  if (loggedIn.value && activeAccountId.value === id) return

  switchingAccount.value = true
  switchingId.value = id
  sidebarError.value = ''
  mailError.value = ''
  detail.value = null
  selectedUid.value = null

  try {
    try {
      const peek = await PeekAccountCache(id)
      applyPeekCache(peek)
    } catch (_) {}

    const result = await SwitchAccount(id)
    if (!result.success) {
      sidebarError.value = result.error || '连接失败'
      return
    }
    loggedIn.value = true
    accountEmail.value = result.email
    activeAccountId.value = id
    refreshAccounts().catch(() => {})

    // 连接完成后先结束全屏遮罩，后台刷新列表
    switchingAccount.value = false
    switchingId.value = ''

    await loadFolders()
    if (selectedFolder.value) {
      await loadMessages({ skipAutoSelect: true })
    }
  } catch (e) {
    sidebarError.value = e?.message || String(e)
  } finally {
    switchingId.value = ''
    switchingAccount.value = false
  }
}

async function loadFolders() {
  if (!loggedIn.value) return
  loadingFolders.value = true
  mailError.value = ''
  try {
    folders.value = await ListFolders()
    if (folders.value.length && !selectedFolder.value) {
      const inbox = folders.value.find((f) =>
        /INBOX|收件箱/i.test(f.name) || /收件箱/.test(f.displayName || '')
      )
      selectedFolder.value = inbox?.name || folders.value[0].name
    }
  } catch (e) {
    mailError.value = e?.message || String(e)
  } finally {
    loadingFolders.value = false
  }
}

async function loadMessages(opts = {}) {
  const { silent = false, skipAutoSelect = false } = opts
  if (!selectedFolder.value) return

  if (!silent) {
    loadingMessages.value = true
    if (!skipAutoSelect) {
      selectedUid.value = null
      detail.value = null
    }
  }
  mailError.value = ''

  try {
    try {
      const cached = await ListMessagesCached(selectedFolder.value)
      if (cached?.length) {
        messages.value = cached
      }
    } catch (_) {}

    const fresh = await ListMessages(selectedFolder.value, 0, 50)
    messages.value = fresh

    if (skipAutoSelect) {
      return
    }

    const prevUid = selectedUid.value
    const list = sortedMessages.value
    if (!list.length) {
      selectedUid.value = null
      detail.value = null
      return
    }
    const stillThere = prevUid != null && list.some((m) => m.uid === prevUid)
    if (stillThere) {
      if (!silent) await selectMessage(prevUid, { keepLoading: silent })
    } else {
      await selectMessage(list[0].uid, { keepLoading: silent })
    }
  } catch (e) {
    mailError.value = e?.message || String(e)
    if (!silent) {
      messages.value = []
      selectedUid.value = null
      detail.value = null
    }
  } finally {
    if (!silent) loadingMessages.value = false
  }
}

async function selectMessage(uid, opts = {}) {
  const { keepLoading = false } = opts
  selectedUid.value = uid
  if (!keepLoading) loadingDetail.value = true
  try {
    detail.value = await GetMessage(selectedFolder.value, uid)
    mailError.value = ''
  } catch (e) {
    mailError.value = e?.message || String(e)
    detail.value = null
  } finally {
    if (!keepLoading) loadingDetail.value = false
  }
}

async function handleRefresh() {
  await loadMessages()
}

async function onSettingsSaved() {
  await loadLinkSetting()
  await setupAutoRefresh()
  if (loggedIn.value && selectedFolder.value) {
    await loadMessages()
  }
}

async function handleDisconnect() {
  clearAutoRefresh()
  try {
    await Logout()
  } catch (_) {}
  loggedIn.value = false
  accountEmail.value = ''
  activeAccountId.value = ''
  folders.value = []
  messages.value = []
  detail.value = null
  selectedFolder.value = ''
  await refreshAccounts()
  emit('disconnected')
}

defineExpose({ refreshAccounts })
</script>

<template>
  <div class="shell">
    <aside class="sidebar" :style="{ width: sidebarWidth + 'px' }">
      <div class="sidebar-head">
        <h2>邮箱账号</h2>
        <button type="button" class="btn-icon" title="设置" @click="settingsOpen = true">⚙</button>
      </div>
      <p v-if="sidebarError" class="sidebar-error">{{ sidebarError }}</p>
      <ul class="account-list">
        <li
          v-for="acc in accounts"
          :key="acc.id"
          :class="{ active: loggedIn && accountEmail === acc.email, switching: switchingId === acc.id }"
          @click="selectAccount(acc.id)"
          @contextmenu.prevent="openAccountContextMenu($event, acc)"
        >
          <div class="account-row">
            <span class="provider-badge" :class="accountBadge(acc).cls">{{ accountBadge(acc).tag }}</span>
            <span class="label">{{ acc.label || acc.email }}</span>
          </div>
          <div class="email">{{ acc.email }}</div>
        </li>
      </ul>
      <div class="sidebar-foot">
        <button type="button" class="btn-block" :disabled="switchingAccount" @click="emit('add-account')">
          + 添加邮箱
        </button>
        <button
          v-if="loggedIn"
          type="button"
          class="btn-block secondary"
          :disabled="switchingAccount"
          @click="handleDisconnect"
        >
          断开连接
        </button>
      </div>
    </aside>

    <div
      class="resize-handle"
      :class="{ dragging: activePane === 'sidebar' }"
      title="拖拽调整宽度"
      @mousedown.prevent="startResize('sidebar', $event)"
    />

    <main class="content">
      <header v-if="loggedIn" class="toolbar">
        <span class="account">
          <span
            v-if="currentAccountTag"
            class="provider-badge toolbar-badge"
            :class="providerBadge(activeAccount?.provider).badgeClass"
          >
            {{ currentAccountTag }}
          </span>
          {{ accountEmail }}
        </span>
        <button
          class="btn-secondary"
          :disabled="loadingMessages || switchingAccount"
          @click="handleRefresh"
        >
          刷新
        </button>
      </header>

      <div v-if="!loggedIn" class="empty-mail">
        <p>请从左侧选择一个邮箱账号，或添加新账号。</p>
      </div>

      <template v-else>
        <p v-if="mailError" class="banner-error">{{ mailError }}</p>
        <div class="main-panels">
          <aside class="folders" :style="{ width: folderWidth + 'px' }">
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

          <div
            class="resize-handle"
            :class="{ dragging: activePane === 'folder' }"
            title="拖拽调整宽度"
            @mousedown.prevent="startResize('folder', $event)"
          />

          <section class="message-list" :style="{ width: listWidth + 'px' }">
            <h3>邮件列表</h3>
            <div v-if="loadingMessages && !messages.length" class="hint">加载中…</div>
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
            <div v-else-if="!loadingMessages" class="hint">暂无邮件</div>
          </section>

          <div
            class="resize-handle"
            :class="{ dragging: activePane === 'list' }"
            title="拖拽调整宽度"
            @mousedown.prevent="startResize('list', $event)"
          />

          <MailDetail
            :detail="detail"
            :loading="loadingDetail"
            :open-links-in-browser="openLinksInBrowser"
          />
        </div>
      </template>

      <div v-if="switchingAccount" class="content-loading-overlay">
        <div class="loading-box">
          <div class="spinner" />
          <p>加载中…</p>
        </div>
      </div>
    </main>

    <ul
      v-if="contextMenu.show"
      class="account-context-menu"
      :style="{ left: contextMenu.x + 'px', top: contextMenu.y + 'px' }"
      @click.stop
    >
      <li @click="openAccountEdit">编辑配置</li>
      <li class="danger" @click="deleteAccountFromMenu">删除账号</li>
    </ul>

    <SettingsModal
      :open="settingsOpen"
      @close="settingsOpen = false"
      @saved="onSettingsSaved"
    />

    <AccountEditModal
      :open="accountEditOpen"
      :account="editingAccount"
      @close="accountEditOpen = false"
      @saved="onAccountSaved"
    />
  </div>
</template>

<style scoped>
.shell {
  display: flex;
  height: 100%;
}

.resize-handle {
  flex-shrink: 0;
  width: 5px;
  margin: 0 -2px;
  cursor: col-resize;
  background: transparent;
  position: relative;
  z-index: 5;
  transition: background 0.15s ease;
}

.resize-handle::after {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  left: 50%;
  width: 1px;
  transform: translateX(-50%);
  background: #e8eaed;
  transition: background 0.15s ease, width 0.15s ease;
}

.resize-handle:hover::after,
.resize-handle.dragging::after {
  width: 3px;
  background: #4a90d9;
}

.resize-handle:hover,
.resize-handle.dragging {
  background: rgba(74, 144, 217, 0.08);
}

.sidebar {
  flex-shrink: 0;
  background: #fff;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.sidebar-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px;
  border-bottom: 1px solid #eee;
}

.sidebar-head h2 {
  font-size: 1rem;
}

.btn-icon {
  border: none;
  background: #f0f4f8;
  border-radius: 6px;
  width: 32px;
  height: 32px;
  cursor: pointer;
}

.account-list {
  list-style: none;
  flex: 1;
  overflow-y: auto;
}

.account-list li {
  padding: 12px 16px;
  cursor: pointer;
  border-bottom: 1px solid #f5f5f5;
}

.account-list li:hover,
.account-list li.active {
  background: #e8f0fe;
}

.account-list li.switching {
  opacity: 0.7;
}

.account-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;
}

.account-row .label {
  font-weight: 500;
  font-size: 0.9rem;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.account-list .email {
  font-size: 0.8rem;
  color: #666;
  padding-left: 0;
}

.provider-badge {
  flex-shrink: 0;
  font-size: 0.7rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  line-height: 1.2;
}

.badge-qq {
  background: #e6f7ff;
  color: #12b7f5;
}

.badge-netease {
  background: #ffeded;
  color: #d43c33;
}

.badge-microsoft {
  background: #e8f4fc;
  color: #0078d4;
}

.badge-default {
  background: #f0f0f0;
  color: #666;
}

.toolbar-badge {
  margin-right: 8px;
  vertical-align: middle;
}

.account-context-menu {
  position: fixed;
  z-index: 200;
  list-style: none;
  min-width: 120px;
  background: #fff;
  border: 1px solid #e0e0e0;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.12);
  padding: 4px 0;
}

.account-context-menu li {
  padding: 8px 16px;
  font-size: 0.9rem;
  cursor: pointer;
}

.account-context-menu li:hover {
  background: #f0f4f8;
}

.account-context-menu li.danger {
  color: #c0392b;
}

.account-context-menu li.danger:hover {
  background: #fdecea;
}

.label {
  font-weight: 600;
  font-size: 0.9rem;
}

.email {
  font-size: 0.8rem;
  color: #666;
  margin-top: 4px;
}

.sidebar-error {
  color: #c0392b;
  font-size: 0.85rem;
  padding: 8px 16px;
}

.sidebar-foot {
  padding: 12px;
  border-top: 1px solid #eee;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.btn-block {
  width: 100%;
  padding: 10px;
  border-radius: 8px;
  border: 1px dashed #4a90d9;
  background: #f0f7ff;
  color: #1a5fb4;
  cursor: pointer;
}

.btn-block.secondary {
  border: 1px solid #ddd;
  background: #fff;
  color: #444;
}

.btn-block:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.content {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  position: relative;
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
}

.btn-secondary {
  padding: 6px 14px;
  border: 1px solid #ddd;
  border-radius: 6px;
  background: #fff;
  cursor: pointer;
}

.btn-secondary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.empty-mail {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #888;
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
  flex-shrink: 0;
  background: #fff;
  overflow-y: auto;
  min-width: 0;
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
  flex-shrink: 0;
  background: #fff;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  min-width: 0;
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

.content-loading-overlay {
  position: absolute;
  inset: 0;
  background: rgba(255, 255, 255, 0.72);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 50;
}

.loading-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  padding: 24px 32px;
  background: #fff;
  border-radius: 12px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.1);
}

.loading-box p {
  color: #555;
  font-size: 0.95rem;
}

.spinner {
  width: 36px;
  height: 36px;
  border: 3px solid #e8eaed;
  border-top-color: #4a90d9;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to {
    transform: rotate(360deg);
  }
}
</style>
