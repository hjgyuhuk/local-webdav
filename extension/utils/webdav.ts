import type { WebDAVConfig } from './storage';

export class WebDAVClient {
  private config: WebDAVConfig;

  constructor(config: WebDAVConfig) {
    this.config = config;
  }

  private getAuthHeaders(): Record<string, string> {
    const headers: Record<string, string> = {
      'Content-Type': 'application/octet-stream',
    };
    if (this.config.username && this.config.password) {
      const credentials = btoa(`${this.config.username}:${this.config.password}`);
      headers['Authorization'] = `Basic ${credentials}`;
    }
    return headers;
  }

  async uploadFile(path: string, content: string): Promise<void> {
    const url = `${this.config.url.replace(/\/$/, '')}${path}`;
    const response = await fetch(url, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: content,
    });
    if (!response.ok) {
      throw new Error(`Upload failed: ${response.status} ${response.statusText}`);
    }
  }

  async testConnection(): Promise<boolean> {
    try {
      const url = `${this.config.url.replace(/\/$/, '')}${this.config.folder}`;
      const response = await fetch(url, {
        method: 'PROPFIND',
        headers: {
          ...this.getAuthHeaders(),
          Depth: '0',
        },
      });
      return response.ok || response.status === 207;
    } catch {
      return false;
    }
  }
}

export async function uploadCookie(domain: string, content: string): Promise<void> {
  const { getConfig } = await import('./storage');
  const config = await getConfig();
  if (!config.enabled) return;
  if (!config.webdav.url) throw new Error('WebDAV URL not configured');

  const client = new WebDAVClient(config.webdav);
  const folder = config.webdav.folder.replace(/\/$/, '');
  const path = `${folder}/${domain}-cookie.txt`;
  await client.uploadFile(path, content);
}

export async function syncDomain(domain: string): Promise<number> {
  const cookies = await browser.cookies.getAll({ domain });
  if (cookies.length === 0) return 0;

  const { cookiesToNetscape } = await import('./netscape');
  const content = cookiesToNetscape(cookies);
  await uploadCookie(domain, content);
  return cookies.length;
}

export async function syncAllDomains(): Promise<Record<string, number | string>> {
  const { getConfig } = await import('./storage');
  const config = await getConfig();
  if (!config.enabled) throw new Error('Monitoring is disabled');

  const results: Record<string, number | string> = {};
  for (const domain of config.websites) {
    try {
      results[domain] = await syncDomain(domain);
    } catch (err) {
      results[domain] = err instanceof Error ? err.message : 'Unknown error';
    }
  }
  return results;
}
