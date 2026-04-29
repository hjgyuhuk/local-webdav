export interface WebDAVConfig {
  url: string;
  username: string;
  password: string;
  folder: string;
}

export interface AppConfig {
  webdav: WebDAVConfig;
  websites: string[];
  enabled: boolean;
}

const STORAGE_KEY = 'cookie_monitor_config';

const defaultConfig: AppConfig = {
  webdav: {
    url: '',
    username: '',
    password: '',
    folder: '/cookies',
  },
  websites: [],
  enabled: true,
};

export async function getConfig(): Promise<AppConfig> {
  const result = await browser.storage.local.get(STORAGE_KEY);
  return (result[STORAGE_KEY] as AppConfig) ?? defaultConfig;
}

export async function setConfig(config: AppConfig): Promise<void> {
  await browser.storage.local.set({ [STORAGE_KEY]: config });
}

export async function updateConfig(partial: Partial<AppConfig>): Promise<void> {
  const config = await getConfig();
  await setConfig({ ...config, ...partial });
}

export async function addWebsite(domain: string): Promise<void> {
  const config = await getConfig();
  if (!config.websites.includes(domain)) {
    config.websites.push(domain);
    await setConfig(config);
  }
}

export async function removeWebsite(domain: string): Promise<void> {
  const config = await getConfig();
  config.websites = config.websites.filter((w) => w !== domain);
  await setConfig(config);
}
