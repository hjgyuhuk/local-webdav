import { getConfig } from '../utils/storage';
import { cookiesToNetscape } from '../utils/netscape';
import { uploadCookie } from '../utils/webdav';

export default defineBackground(() => {
  console.log('Cookie Monitor background started');

  // Listen for cookie changes
  browser.cookies.onChanged.addListener(async (changeInfo) => {
    const { cookie, removed } = changeInfo;
    if (removed) return;

    const config = await getConfig();
    if (!config.enabled) return;

    const domain = cookie.domain.startsWith('.') ? cookie.domain.slice(1) : cookie.domain;
    const isMonitored = config.websites.some(
      (w) => domain === w || domain.endsWith(`.${w}`)
    );

    if (!isMonitored) return;

    console.log(`Cookie changed for ${domain}, syncing...`);

    try {
      const cookies = await browser.cookies.getAll({ domain });
      const content = cookiesToNetscape(cookies);
      await uploadCookie(domain, content);
      console.log(`Synced ${cookies.length} cookies for ${domain}`);
    } catch (err) {
      console.error(`Failed to sync cookies for ${domain}:`, err);
    }
  });

  // Handle messages from popup/options
  browser.runtime.onMessage.addListener((message, sender, sendResponse) => {
    if (message.type === 'syncDomain') {
      (async () => {
        try {
          const cookies = await browser.cookies.getAll({ domain: message.domain });
          if (cookies.length === 0) {
            sendResponse({ success: false, error: 'No cookies found for this domain' });
            return;
          }
          const content = cookiesToNetscape(cookies);
          await uploadCookie(message.domain, content);
          sendResponse({ success: true, count: cookies.length });
        } catch (err) {
          sendResponse({
            success: false,
            error: err instanceof Error ? err.message : 'Sync failed',
          });
        }
      })();
      return true;
    }

    if (message.type === 'syncAll') {
      (async () => {
        try {
          const config = await getConfig();
          if (!config.enabled) {
            sendResponse({ success: false, error: 'Monitoring is disabled' });
            return;
          }

          const results: Record<string, number | string> = {};
          for (const domain of config.websites) {
            try {
              const cookies = await browser.cookies.getAll({ domain });
              if (cookies.length === 0) {
                results[domain] = 'No cookies found';
                continue;
              }
              const content = cookiesToNetscape(cookies);
              await uploadCookie(domain, content);
              results[domain] = cookies.length;
            } catch (err) {
              results[domain] = err instanceof Error ? err.message : 'Sync failed';
            }
          }
          sendResponse({ success: true, results });
        } catch (err) {
          sendResponse({
            success: false,
            error: err instanceof Error ? err.message : 'Sync failed',
          });
        }
      })();
      return true;
    }

    return false;
  });
});
