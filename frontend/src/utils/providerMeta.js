/** Provider badge styles (aligned with backend displayName). */

const meta = {
  qq: { badgeClass: 'badge-qq', label: 'QQ' },
  netease: { badgeClass: 'badge-netease', label: '163' },
  microsoft: { badgeClass: 'badge-microsoft', label: 'Outlook' },
}

export function providerBadge(providerOrTag) {
  const key = normalizeKey(providerOrTag)
  return meta[key] || { badgeClass: 'badge-default', label: providerOrTag || '?' }
}

function normalizeKey(id) {
  if (!id) return 'qq'
  const s = String(id).toLowerCase()
  if (s === '163' || s === 'outlook') {
    if (s === '163') return 'netease'
    if (s === 'outlook') return 'microsoft'
  }
  if (meta[s]) return s
  return 'qq'
}

export function authCodeHint(provider) {
  switch (normalizeKey(provider)) {
    case 'netease':
      return '客户端授权密码（非邮箱登录密码）'
    case 'microsoft':
      return '微软应用密码（非 Outlook 登录密码）'
    case 'qq':
    default:
      return '16 位 IMAP 授权码（非 QQ 登录密码）'
  }
}

export function addAccountTitle(provider) {
  switch (normalizeKey(provider)) {
    case 'netease':
      return '添加网易邮箱'
    case 'microsoft':
      return '添加 Outlook 邮箱'
    default:
      return '添加 QQ 邮箱'
  }
}
