const folderMap = {
  INBOX: '收件箱',
  'Sent Messages': '已发送',
  Sent: '已发送',
  Drafts: '草稿箱',
  'Deleted Messages': '已删除',
  Trash: '垃圾箱',
  Junk: '垃圾邮件',
  Spam: '垃圾邮件',
}

export function displayFolderName(folder) {
  if (!folder) return ''
  return folder.displayName || folderMap[folder.name] || folder.name
}
