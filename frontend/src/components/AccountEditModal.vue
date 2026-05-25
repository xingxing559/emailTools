<script setup>
import { ref, watch, computed } from 'vue'
import { UpdateAccount } from '../../wailsjs/go/app/MailApp'
import { authCodeHint, providerBadge } from '../utils/providerMeta.js'

const props = defineProps({
  open: { type: Boolean, default: false },
  account: { type: Object, default: null },
})
const emit = defineEmits(['close', 'saved'])

const label = ref('')
const authCode = ref('')
const saving = ref(false)
const error = ref('')

const providerTag = computed(
  () => props.account?.providerTag || providerBadge(props.account?.provider).label
)
const authHint = computed(() => authCodeHint(props.account?.provider))

watch(
  () => props.open,
  (v) => {
    if (!v || !props.account) return
    label.value = props.account.label || ''
    authCode.value = ''
    error.value = ''
  }
)

async function handleSave() {
  if (!props.account?.id) return
  saving.value = true
  error.value = ''
  try {
    const result = await UpdateAccount(
      props.account.id,
      authCode.value.trim(),
      label.value.trim()
    )
    if (!result.success) {
      error.value = result.error || '保存失败'
      return
    }
    emit('saved', result)
    emit('close')
  } catch (e) {
    error.value = e?.message || String(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div v-if="open && account" class="modal-overlay" @click.self="emit('close')">
    <div class="modal-card">
      <header class="modal-header">
        <h2>编辑邮箱</h2>
        <button type="button" class="icon-btn" @click="emit('close')">×</button>
      </header>

      <div class="provider-line">
        <span class="provider-badge" :class="providerBadge(account.provider).badgeClass">
          {{ providerTag }}
        </span>
        <span class="provider-email">{{ account.email }}</span>
      </div>

      <label class="field">
        <span>备注名称</span>
        <input v-model="label" type="text" placeholder="例如：工作邮箱" />
      </label>

      <label class="field">
        <span>授权码 / 应用密码</span>
        <input
          v-model="authCode"
          type="password"
          placeholder="留空则不修改；填写后将更新并验证连接"
          autocomplete="new-password"
        />
        <p class="hint">{{ authHint }}</p>
      </label>

      <p v-if="error" class="error">{{ error }}</p>

      <div class="actions">
        <button type="button" class="btn-secondary" @click="emit('close')">取消</button>
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
  z-index: 110;
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

.provider-line {
  display: flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 16px;
  padding: 10px 12px;
  background: #f5f8fc;
  border-radius: 8px;
}

.provider-badge {
  font-size: 0.7rem;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
}

.badge-qq { background: #e6f7ff; color: #12b7f5; }
.badge-netease { background: #ffeded; color: #d43c33; }
.badge-microsoft { background: #e8f4fc; color: #0078d4; }

.provider-email {
  font-size: 0.9rem;
  color: #444;
  overflow: hidden;
  text-overflow: ellipsis;
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

.field input {
  width: 100%;
  padding: 8px 10px;
  border: 1px solid #ddd;
  border-radius: 8px;
}

.field input:focus {
  outline: none;
  border-color: #4a90d9;
}

.hint {
  margin-top: 8px;
  font-size: 0.8rem;
  color: #888;
}

.error {
  color: #c0392b;
  font-size: 0.9rem;
  margin-bottom: 12px;
}

.actions {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
  margin-top: 8px;
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

.btn-primary:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}
</style>
