import { BrowserOpenURL } from '../../wailsjs/runtime/runtime'

/**
 * @param {string} url
 * @param {boolean} useBrowser - true: system default browser; false: in-app navigation
 */
export function openLink(url, useBrowser) {
  if (!url || url.startsWith('javascript:')) return
  if (useBrowser) {
    BrowserOpenURL(url)
  } else {
    window.open(url, '_blank')
  }
}
