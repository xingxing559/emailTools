<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { AddAccount, HasAccounts, GetSettings, ListProviders } from '../../wailsjs/go/app/MailApp'
import { openLink } from '../utils/openLink.js'
import { authCodeHint, addAccountTitle } from '../utils/providerMeta.js'

const emit = defineEmits(['login-success', 'back'])

const providers = ref([])
const selectedProvider = ref('qq')
const email = ref('')
const authCode = ref('')
const label = ref('')
const loading = ref(false)
const error = ref('')
const showBack = ref(false)
const openLinksInBrowser = ref(true)

const currentProvider = computed(() =>
  providers.value.find((p) => p.id === selectedProvider.value)
)
const pageTitle = computed(() => addAccountTitle(selectedProvider.value))
const emailPlaceholder = computed(
  () => currentProvider.value?.emailPlaceholder || 'name@example.com'
)
const authHint = computed(() => authCodeHint(selectedProvider.value))

HasAccounts().then((v) => { showBack.value = v }).catch(() => {})

onMounted(async () => {
  try {
    const s = await GetSettings()
    openLinksInBrowser.value = s.openLinksInBrowser !== false
  } catch (_) {}
  try {
    providers.value = await ListProviders()
    if (providers.value.length && !providers.value.find((p) => p.id === selectedProvider.value)) {
      selectedProvider.value = providers.value[0].id
    }
  } catch (_) {
    providers.value = [
      { id: 'qq', displayName: 'QQ', authType: 'app_password', helpUrl: 'https://help.mail.qq.com/detail/0/985', emailPlaceholder: '123456789@qq.com' },
      { id: 'netease', displayName: '163', authType: 'app_password', helpUrl: 'https://help.mail.163.com/', emailPlaceholder: 'name@163.com' },
      { id: 'microsoft', displayName: 'Outlook', authType: 'app_password', helpUrl: 'https://support.microsoft.com/account-billing/using-app-passwords-with-apps-that-don-t-support-two-step-verification-5896ed9b-4263-6812-ef18-520f4cf8f57c', emailPlaceholder: 'name@outlook.com' },
    ]
  }
})

watch(email, (v) => {
  if (!v || providers.value.length === 0) return
  const lower = v.trim().toLowerCase()
  if (!lower.includes('@')) return
  if (lower.endsWith('@qq.com') || lower.endsWith('@foxmail.com')) selectedProvider.value = 'qq'
  else if (lower.match(/@(163|126|yeah|188)\.com$/)) selectedProvider.value = 'netease'
  else if (lower.match(/@(outlook|hotmail|live|msn)\.com$/)) selectedProvider.value = 'microsoft'
})

watch(selectedProvider, () => {
  error.value = ''
})

function openHelpLink(event) {
  event.preventDefault()
  const url = currentProvider.value?.helpUrl || 'https://help.mail.qq.com/detail/0/985'
  openLink(url, openLinksInBrowser.value)
}

async function handleSubmit() {
  error.value = ''
  if (!email.value.trim() || !authCode.value.trim()) {
    error.value = '请填写邮箱地址和授权码'
    return
  }
  loading.value = true
  try {
    const result = await AddAccount(
      selectedProvider.value,
      email.value.trim(),
      authCode.value.trim(),
      label.value.trim()
    )
    if (result.success) {
      emit('login-success', result.email)
    } else {
      error.value = result.error || '添加失败'
    }
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="login-page">
    <div class="login-card">
      <button v-if="showBack" type="button" class="back-link" @click="emit('back')">
        ← 返回账号列表
      </button>
      <h1>{{ pageTitle }}</h1>
      <p class="subtitle">通过 IMAP 连接，仅用于查看邮件</p>

      <div v-if="providers.length" class="provider-picker">
        <button
          v-for="p in providers"
          :key="p.id"
          type="button"
          class="provider-card"
          :class="{ active: selectedProvider === p.id }"
          @click="selectedProvider = p.id"
        >
          <span class="provider-badge" :class="'badge-' + p.id">{{ p.displayName }}</span>
        </button>
      </div>

      <form @submit.prevent="handleSubmit">
        <label>
          <span>备注名称（可选）</span>
          <input v-model="label" type="text" placeholder="例如：工作邮箱" />
        </label>
        <label>
          <span>邮箱地址</span>
          <input
            v-model="email"
            type="email"
            :placeholder="emailPlaceholder"
            autocomplete="username"
          />
        </label>
        <label>
          <span>授权码 / 应用密码</span>
          <input
            v-model="authCode"
            type="password"
            :placeholder="authHint"
            autocomplete="current-password"
          />
        </label>

        <p v-if="error" class="error">{{ error }}</p>

        <button type="submit" :disabled="loading">
          {{ loading ? '验证连接中…' : '添加并进入' }}
        </button>
      </form>

      <div class="help">
        <template v-if="selectedProvider === 'qq'">
          <p>首次使用请在 QQ 邮箱网页端：</p>
          <ol>
            <li>设置 → 账号与安全 → 安全管理</li>
            <li>开启 IMAP/SMTP 服务</li>
            <li>生成 16 位授权码</li>
          </ol>
        </template>
        <template v-else-if="selectedProvider === 'netease'">
          <p>首次使用请在网易邮箱网页端：</p>
          <ol>
            <li>设置 → POP3/SMTP/IMAP</li>
            <li>开启 IMAP/SMTP 服务</li>
            <li>设置客户端授权密码</li>
          </ol>
        </template>
        <template v-else>
          <p>Outlook 使用应用密码（与 QQ/163 相同流程）：</p>
          <ol>
            <li>登录 <a href="https://outlook.live.com" @click.prevent="openLink('https://outlook.live.com', openLinksInBrowser)">outlook.live.com</a>，开启 IMAP</li>
            <li>微软账户安全中心开启双重验证后创建「应用密码」</li>
            <li>将应用密码填入上方（非登录密码）</li>
          </ol>
        </template>
        <a v-if="currentProvider?.helpUrl" :href="currentProvider.helpUrl" @click="openHelpLink">
          查看官方说明
        </a>
      </div>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100%;
  padding: 24px;
}

.login-card {
  width: 100%;
  max-width: 460px;
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.08);
}

.back-link {
  background: none;
  border: none;
  color: #4a90d9;
  cursor: pointer;
  font-size: 0.9rem;
  margin-bottom: 12px;
  padding: 0;
}

h1 {
  font-size: 1.5rem;
  margin-bottom: 8px;
}

.subtitle {
  color: #666;
  font-size: 0.9rem;
  margin-bottom: 16px;
}

.provider-picker {
  display: flex;
  gap: 8px;
  margin-bottom: 20px;
}

.provider-card {
  flex: 1;
  padding: 10px 8px;
  border: 1px solid #ddd;
  border-radius: 8px;
  background: #fafafa;
  cursor: pointer;
  text-align: center;
}

.provider-card.active {
  border-color: #4a90d9;
  background: #e8f0fe;
}

.provider-badge {
  font-size: 0.85rem;
  font-weight: 600;
}

.badge-qq { color: #12b7f5; }
.badge-netease { color: #d43c33; }
.badge-microsoft { color: #0078d4; }

form label {
  display: block;
  margin-bottom: 16px;
}

form label span {
  display: block;
  font-size: 0.85rem;
  color: #444;
  margin-bottom: 6px;
}

input[type="email"],
input[type="password"],
input[type="text"] {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid #ddd;
  border-radius: 8px;
  font-size: 1rem;
}

input:focus {
  outline: none;
  border-color: #4a90d9;
}

.error {
  color: #c0392b;
  font-size: 0.9rem;
  margin-bottom: 12px;
}

button[type="submit"] {
  width: 100%;
  padding: 12px;
  background: #4a90d9;
  color: #fff;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  cursor: pointer;
}

button[type="submit"]:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.help {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid #eee;
  font-size: 0.8rem;
  color: #666;
}

.help ol {
  margin: 8px 0 8px 20px;
}

.help a {
  color: #4a90d9;
}
</style>
