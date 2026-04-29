import { useState, useEffect } from 'react';
import {
  Button,
  Input,
  Switch,
  Text,
  LayerCard,
  SensitiveInput,
  Empty,
  Toasty,
  useKumoToastManager,
} from '@cloudflare/kumo';
import {
  FloppyDisk,
  Plugs,
  Trash,
  Plus,
  Globe,
  ArrowsClockwise,
} from '@phosphor-icons/react';
import { getConfig, setConfig, type AppConfig } from '../../utils/storage';
import { WebDAVClient } from '../../utils/webdav';

const TOAST_TIMEOUT = 3000;

function AppContent() {
  const [config, setConfigState] = useState<AppConfig>({
    webdav: { url: '', username: '', password: '', folder: '/cookies' },
    websites: [],
    enabled: true,
  });
  const [newWebsite, setNewWebsite] = useState('');
  const [saving, setSaving] = useState(false);
  const [testing, setTesting] = useState(false);
  const [syncing, setSyncing] = useState<string | null>(null);
  const toast = useKumoToastManager();

  useEffect(() => {
    getConfig().then(setConfigState);
  }, []);

  const handleSave = async () => {
    setSaving(true);
    try {
      await setConfig(config);
      toast.add({ variant: 'success', title: 'Settings saved', timeout: TOAST_TIMEOUT });
    } catch {
      toast.add({ variant: 'error', title: 'Failed to save settings', timeout: TOAST_TIMEOUT });
    } finally {
      setSaving(false);
    }
  };

  const handleTestConnection = async () => {
    setTesting(true);
    try {
      const client = new WebDAVClient(config.webdav);
      const ok = await client.testConnection();
      toast.add({
        variant: ok ? 'success' : 'error',
        title: ok ? 'Connection successful' : 'Connection failed',
        timeout: TOAST_TIMEOUT,
      });
    } catch {
      toast.add({ variant: 'error', title: 'Connection failed', timeout: TOAST_TIMEOUT });
    } finally {
      setTesting(false);
    }
  };

  const handleSync = async (domain: string) => {
    setSyncing(domain);
    try {
      const response = await browser.runtime.sendMessage({
        type: 'syncDomain',
        domain,
      });
      if (response.success) {
        toast.add({
          variant: 'success',
          title: `Synced ${response.count} cookies for ${domain}`,
          timeout: TOAST_TIMEOUT,
        });
      } else {
        toast.add({ variant: 'error', title: response.error || 'Sync failed', timeout: TOAST_TIMEOUT });
      }
    } catch (err) {
      toast.add({
        variant: 'error',
        title: err instanceof Error ? err.message : 'Sync failed',
        timeout: TOAST_TIMEOUT,
      });
    } finally {
      setSyncing(null);
    }
  };

  const handleAddWebsite = async () => {
    const domain = newWebsite.trim().toLowerCase();
    if (!domain) return;
    if (config.websites.includes(domain)) {
      toast.add({ variant: 'warning', title: 'Website already added', timeout: TOAST_TIMEOUT });
      return;
    }
    const newConfig = { ...config, websites: [...config.websites, domain] };
    setConfigState(newConfig);
    setNewWebsite('');
    await setConfig(newConfig);
    toast.add({ variant: 'success', title: `Added ${domain}`, timeout: TOAST_TIMEOUT });
  };

  const handleRemoveWebsite = async (domain: string) => {
    const newConfig = { ...config, websites: config.websites.filter((w) => w !== domain) };
    setConfigState(newConfig);
    await setConfig(newConfig);
    toast.add({ variant: 'success', title: `Removed ${domain}`, timeout: TOAST_TIMEOUT });
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      e.preventDefault();
      handleAddWebsite();
    }
  };

  return (
    <div className="options-container">
      <header className="options-header">
        <Text variant="heading1" as="h1">Cookie Monitor</Text>
        <Text variant="secondary">Monitor website cookies and sync to WebDAV server</Text>
      </header>

      <div className="sections">
        <LayerCard>
          <LayerCard.Primary>
            <div className="toggle-row">
              <Text variant="heading3">Monitoring</Text>
              <Switch
                label={config.enabled ? 'Enabled' : 'Disabled'}
                checked={config.enabled}
                onCheckedChange={async (checked: boolean) => {
                  const newConfig = { ...config, enabled: checked };
                  setConfigState(newConfig);
                  await setConfig(newConfig);
                }}
              />
            </div>
            <Text variant="secondary" size="sm">
              When enabled, cookie changes for monitored websites will be automatically synced.
            </Text>
          </LayerCard.Primary>
        </LayerCard>

        <LayerCard>
          <LayerCard.Primary>
            <Text variant="heading3" as="h2" DANGEROUS_className="mb-4">
              WebDAV Configuration
            </Text>
            <div className="form-grid">
              <Input
                label="Server URL"
                placeholder="http://127.0.0.1:43621/cookies"
                value={config.webdav.url}
                onValueChange={(value: string) =>
                  setConfigState((c) => ({
                    ...c,
                    webdav: { ...c.webdav, url: value },
                  }))
                }
              />
              <Input
                label="Folder Path"
                placeholder="/cookies"
                value={config.webdav.folder}
                onValueChange={(value: string) =>
                  setConfigState((c) => ({
                    ...c,
                    webdav: { ...c.webdav, folder: value },
                  }))
                }
              />
              <Input
                label="Username"
                placeholder="username"
                value={config.webdav.username}
                onValueChange={(value: string) =>
                  setConfigState((c) => ({
                    ...c,
                    webdav: { ...c.webdav, username: value },
                  }))
                }
              />
              <SensitiveInput
                label="Password"
                placeholder="password"
                value={config.webdav.password}
                onValueChange={(value: string) =>
                  setConfigState((c) => ({
                    ...c,
                    webdav: { ...c.webdav, password: value },
                  }))
                }
              />
            </div>
            <div className="form-actions">
              <Button
                variant="primary"
                icon={<FloppyDisk />}
                loading={saving}
                onClick={handleSave}
              >
                Save Settings
              </Button>
              <Button
                variant="secondary"
                icon={<Plugs />}
                loading={testing}
                onClick={handleTestConnection}
              >
                Test Connection
              </Button>
            </div>
          </LayerCard.Primary>
        </LayerCard>

        <LayerCard>
          <LayerCard.Primary>
            <Text variant="heading3" as="h2" DANGEROUS_className="mb-4">
              Monitored Websites
            </Text>
            <div className="add-website">
              <Input
                placeholder="youtube.com"
                value={newWebsite}
                onValueChange={setNewWebsite}
                onKeyDown={handleKeyDown}
              />
              <Button
                variant="secondary"
                icon={<Plus />}
                onClick={handleAddWebsite}
              >
                Add
              </Button>
            </div>
            {config.websites.length === 0 ? (
              <Empty
                icon={<Globe />}
                title="No websites added"
                description="Add domains to monitor their cookies"
              />
            ) : (
              <div className="website-list">
                {config.websites.map((domain) => (
                  <div key={domain} className="website-item">
                    <Text variant="mono">{domain}</Text>
                    <div style={{ display: 'flex', gap: '0.25rem' }}>
                      <Button
                        variant="ghost"
                        size="sm"
                        icon={<ArrowsClockwise />}
                        loading={syncing === domain}
                        onClick={() => handleSync(domain)}
                        title="Sync now"
                      />
                      <Button
                        variant="ghost"
                        size="sm"
                        icon={<Trash />}
                        onClick={() => handleRemoveWebsite(domain)}
                        title="Remove"
                      />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </LayerCard.Primary>
        </LayerCard>
      </div>
    </div>
  );
}

function App() {
  return (
    <Toasty>
      <AppContent />
    </Toasty>
  );
}

export default App;
