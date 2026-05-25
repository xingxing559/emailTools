<script setup>
import { ref, computed, watch, nextTick } from 'vue'
import { openLink } from '../utils/openLink.js'

const props = defineProps({
  detail: { type: Object, default: null },
  loading: { type: Boolean, default: false },
  openLinksInBrowser: { type: Boolean, default: true },
})

const htmlFrameRef = ref(null)
const openLinksInBrowser = ref(props.openLinksInBrowser)

watch(
  () => props.openLinksInBrowser,
  (v) => {
    openLinksInBrowser.value = v
  }
)

const iframeSandbox = computed(() =>
  openLinksInBrowser.value
    ? 'allow-same-origin'
    : 'allow-same-origin allow-popups allow-popups-to-escape-sandbox'
)

// Minimal chrome only — do not override mail background/colors (QQ Mail keeps original HTML look).
const READING_STYLES = `
<style>
  html, body {
    margin: 0;
    padding: 0;
    min-height: 100%;
  }
  body {
    word-wrap: break-word;
    overflow-wrap: break-word;
  }
  img { max-width: 100%; height: auto; }
  table { max-width: 100%; }
</style>
`

const htmlSrcdoc = computed(() => {
  if (!props.detail?.textHtml) return ''
  return READING_STYLES + props.detail.textHtml
})

const hasHtml = computed(() => !!props.detail?.textHtml)
const hasPlain = computed(() => !!props.detail?.textPlain)

const displaySubject = computed(() => {
  const s = (props.detail?.subject || '').trim()
  return s || '（无主题）'
})

function formatSize(bytes) {
  if (!bytes) return '0 字节'
  if (bytes < 1024) return bytes + ' 字节'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
  return (bytes / 1024 / 1024).toFixed(1) + ' MB'
}

function senderInitial(from) {
  if (!from) return '?'
  const nameMatch = from.match(/^([^<]+)</)
  const label = nameMatch ? nameMatch[1].trim() : from.split('@')[0]
  const ch = label.replace(/["']/g, '').trim()[0]
  return (ch || '?').toUpperCase()
}

function handleIframeClick(event) {
  const doc = event.currentTarget
  const anchor = event.target?.closest?.('a')
  if (!anchor || !doc.contains(anchor)) return

  const href = anchor.getAttribute('href')
  if (!href || href.startsWith('#')) return

  if (openLinksInBrowser.value) {
    event.preventDefault()
    event.stopPropagation()
    try {
      openLink(anchor.href || href, true)
    } catch (_) {
      openLink(href, true)
    }
  }
}

async function bindIframeLinks() {
  await nextTick()
  const frame = htmlFrameRef.value
  const doc = frame?.contentDocument
  if (!doc?.body) return
  doc.body.removeEventListener('click', handleIframeClick, true)
  doc.body.addEventListener('click', handleIframeClick, true)
}

watch(htmlSrcdoc, () => {
  if (hasHtml.value) bindIframeLinks()
})

watch(openLinksInBrowser, () => {
  if (hasHtml.value) bindIframeLinks()
})
</script>

<template>
  <section class="detail-panel">
    <div v-if="loading" class="state-panel">
      <div class="detail-skeleton">
        <div class="skeleton skeleton-title" />
        <div class="skeleton skeleton-meta" />
        <div class="skeleton skeleton-meta short" />
        <div class="skeleton skeleton-body" />
      </div>
    </div>

    <div v-else-if="!detail" class="state-panel empty">
      <div class="state-icon" aria-hidden="true">
        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
          <path
            d="M4 6h16v12H4V6zm0 0l8 6 8-6"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </div>
      <p class="state-title">选择一封邮件</p>
      <p class="state-desc">从左侧列表点击邮件即可在此阅读</p>
    </div>

    <template v-else>
      <header class="detail-header">
        <h2 class="subject">{{ displaySubject }}</h2>
        <div class="sender-block">
          <span class="sender-avatar" aria-hidden="true">{{ senderInitial(detail.from) }}</span>
          <div class="sender-meta">
            <div class="sender-line">
              <span class="label">发件人</span>
              <span class="value">{{ detail.from }}</span>
            </div>
            <div v-if="detail.to?.length" class="sender-line">
              <span class="label">收件人</span>
              <span class="value">{{ detail.to.join(', ') }}</span>
            </div>
            <div v-if="detail.date" class="sender-line">
              <span class="label">时间</span>
              <span class="value">{{ detail.date }}</span>
            </div>
          </div>
        </div>
        <div v-if="detail.attachments?.length" class="attachments">
          <span class="attachments-label">附件</span>
          <div class="attachment-chips">
            <span v-for="(a, i) in detail.attachments" :key="i" class="att-chip">
              <svg class="att-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path
                  d="M21.44 11.05l-9.19 9.19a6 6 0 01-8.49-8.49l9.19-9.19a4 4 0 015.66 5.66l-9.2 9.19a2 2 0 01-2.83-2.83l8.49-8.48"
                  stroke-linecap="round"
                  stroke-linejoin="round"
                />
              </svg>
              <span class="att-name">{{ a.filename || '未命名' }}</span>
              <span class="att-size">{{ formatSize(a.size) }}</span>
            </span>
          </div>
        </div>
      </header>

      <div class="detail-body">
        <div class="reading-surface">
          <iframe
            v-if="hasHtml"
            ref="htmlFrameRef"
            class="html-frame"
            :sandbox="iframeSandbox"
            :srcdoc="htmlSrcdoc"
            title="邮件正文"
            @load="bindIframeLinks"
          />
          <pre v-else-if="hasPlain" class="plain-body">{{ detail.textPlain }}</pre>
          <div v-else class="body-empty">（无正文内容）</div>
        </div>
      </div>
    </template>
  </section>
</template>

<style scoped>
.detail-panel {
  flex: 1;
  min-width: 0;
  margin: 0 12px 12px 12px;
  display: flex;
  flex-direction: column;
  background: var(--color-surface);
  border-radius: var(--radius-md);
  border: 1px solid var(--color-border);
  overflow: hidden;
  box-shadow: var(--shadow-sm);
}

.state-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 24px;
  color: var(--color-text-muted);
}

.state-panel.empty {
  gap: 8px;
}

.state-icon {
  width: 56px;
  height: 56px;
  margin-bottom: 8px;
  color: var(--color-border);
}

.state-icon svg {
  width: 100%;
  height: 100%;
}

.state-title {
  font-size: 1rem;
  font-weight: 600;
  color: var(--color-text-secondary);
}

.state-desc {
  font-size: 0.875rem;
}

.detail-skeleton {
  width: 100%;
  max-width: 480px;
  padding: 24px;
}

.skeleton-title {
  height: 24px;
  width: 75%;
  margin-bottom: 20px;
}

.skeleton-meta {
  height: 14px;
  width: 55%;
  margin-bottom: 10px;
}

.skeleton-meta.short {
  width: 35%;
}

.skeleton-body {
  height: 200px;
  margin-top: 24px;
  border-radius: var(--radius-sm);
}

.detail-header {
  padding: 20px 24px 16px;
  border-bottom: 1px solid var(--color-border-subtle);
  flex-shrink: 0;
}

.subject {
  font-size: 1.25rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.35;
  margin-bottom: 16px;
  word-break: break-word;
  color: var(--color-text);
}

.sender-block {
  display: flex;
  gap: 14px;
  align-items: flex-start;
}

.sender-avatar {
  flex-shrink: 0;
  width: 44px;
  height: 44px;
  border-radius: var(--radius-full);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 1rem;
  font-weight: 600;
  background: var(--color-avatar-bg);
  color: var(--color-avatar-text);
}

.sender-meta {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.sender-line {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 12px;
  font-size: 0.8125rem;
  line-height: 1.45;
}

.sender-line .label {
  flex-shrink: 0;
  width: 48px;
  color: var(--color-text-muted);
  font-weight: 500;
}

.sender-line .value {
  flex: 1;
  min-width: 0;
  color: var(--color-text-secondary);
  word-break: break-word;
}

.attachments {
  margin-top: 16px;
  padding-top: 14px;
  border-top: 1px solid var(--color-border-subtle);
}

.attachments-label {
  display: block;
  font-size: 0.75rem;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.04em;
  color: var(--color-text-muted);
  margin-bottom: 10px;
}

.attachment-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.att-chip {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--color-surface-muted);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  font-size: 0.8125rem;
  color: var(--color-text-secondary);
}

.att-icon {
  width: 14px;
  height: 14px;
  color: var(--color-primary);
  flex-shrink: 0;
}

.att-name {
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.att-size {
  color: var(--color-text-muted);
  font-size: 0.75rem;
}

.detail-body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  /* QQ 邮箱阅读区：浅灰底，正文块自带背景色（如 Steam 深色模板） */
  background: #eceff4;
}

.reading-surface {
  height: 100%;
  margin: 0;
  background: transparent;
  border: none;
  border-radius: 0;
  overflow: hidden;
  box-shadow: none;
}

.html-frame {
  width: 100%;
  height: 100%;
  border: none;
  background: transparent;
  display: block;
}

.plain-body {
  margin: 0;
  padding: 20px 24px 28px;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: var(--font-sans);
  font-size: 0.9375rem;
  line-height: 1.65;
  color: var(--color-text);
  overflow: auto;
  height: 100%;
  max-width: 720px;
}

.body-empty {
  padding: 32px 24px;
  color: var(--color-text-muted);
  font-size: 0.875rem;
}
</style>
