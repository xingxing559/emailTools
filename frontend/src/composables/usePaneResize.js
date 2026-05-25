import { ref, onUnmounted } from 'vue'

const STORAGE_KEY = 'emailtools-pane-widths'

const DEFAULTS = {
  sidebar: 240,
  folder: 180,
  list: 320,
}

const LIMITS = {
  sidebar: { min: 160, max: 420 },
  folder: { min: 120, max: 360 },
  list: { min: 220, max: 640 },
}

function loadWidths() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (!raw) return { ...DEFAULTS }
    const parsed = JSON.parse(raw)
    return {
      sidebar: clamp(parsed.sidebar, LIMITS.sidebar) ?? DEFAULTS.sidebar,
      folder: clamp(parsed.folder, LIMITS.folder) ?? DEFAULTS.folder,
      list: clamp(parsed.list, LIMITS.list) ?? DEFAULTS.list,
    }
  } catch {
    return { ...DEFAULTS }
  }
}

function clamp(value, { min, max }) {
  const n = Number(value)
  if (!Number.isFinite(n)) return null
  return Math.min(max, Math.max(min, Math.round(n)))
}

function saveWidths(widths) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(widths))
  } catch (_) {}
}

/**
 * @param {'sidebar'|'folder'|'list'} pane
 */
export function usePaneResize() {
  const initial = loadWidths()
  const sidebarWidth = ref(initial.sidebar)
  const folderWidth = ref(initial.folder)
  const listWidth = ref(initial.list)
  const activePane = ref(null)

  let startX = 0
  let startWidth = 0

  function persist() {
    saveWidths({
      sidebar: sidebarWidth.value,
      folder: folderWidth.value,
      list: listWidth.value,
    })
  }

  function onMouseMove(e) {
    if (!activePane.value) return
    const dx = e.clientX - startX
    const limits = LIMITS[activePane.value]
    if (activePane.value === 'sidebar') {
      sidebarWidth.value = clamp(startWidth + dx, limits)
    } else if (activePane.value === 'folder') {
      folderWidth.value = clamp(startWidth + dx, limits)
    } else if (activePane.value === 'list') {
      listWidth.value = clamp(startWidth + dx, limits)
    }
  }

  function stopResize() {
    if (!activePane.value) return
    activePane.value = null
    document.removeEventListener('mousemove', onMouseMove)
    document.removeEventListener('mouseup', stopResize)
    document.body.style.cursor = ''
    document.body.style.userSelect = ''
    persist()
  }

  function startResize(pane, event) {
    activePane.value = pane
    startX = event.clientX
    if (pane === 'sidebar') startWidth = sidebarWidth.value
    else if (pane === 'folder') startWidth = folderWidth.value
    else startWidth = listWidth.value

    document.addEventListener('mousemove', onMouseMove)
    document.addEventListener('mouseup', stopResize)
    document.body.style.cursor = 'col-resize'
    document.body.style.userSelect = 'none'
  }

  onUnmounted(stopResize)

  return {
    sidebarWidth,
    folderWidth,
    listWidth,
    activePane,
    startResize,
  }
}
