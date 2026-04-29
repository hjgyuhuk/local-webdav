import { Browser } from 'wxt/browser';

export function cookiesToNetscape(cookies: Browser.cookies.Cookie[]): string {
  const lines = ['# Netscape HTTP Cookie File', '# https://curl.se/docs/http-cookies.html', ''];

  for (const cookie of cookies) {
    const domain = cookie.domain.startsWith('.') ? cookie.domain : `.${cookie.domain}`;
    const includeSubdomains = cookie.domain.startsWith('.') ? 'TRUE' : 'FALSE';
    const path = cookie.path || '/';
    const secure = cookie.secure ? 'TRUE' : 'FALSE';
    const expiration = cookie.expirationDate ? Math.floor(cookie.expirationDate).toString() : '0';
    const line = [domain, includeSubdomains, path, secure, expiration, cookie.name, cookie.value].join('\t');
    lines.push(line);
  }

  return lines.join('\n');
}

export function getDomainFromUrl(url: string): string {
  try {
    const hostname = new URL(url).hostname;
    return hostname.startsWith('www.') ? hostname.slice(4) : hostname;
  } catch {
    return url;
  }
}
