import { useState, useEffect } from 'react';
import {
  Button,
  Text,
  Badge,
  Switch,
  Toasty,
  useKumoToastManager,
} from '@cloudflare/kumo';
import {
  Gear,
  Globe,
  ArrowsClockwise,
} from '@phosphor-icons/react';
import { getConfig, updateConfig, type AppConfig } from '../../utils/storage';

const TOAST_TIMEOUT = 3000;

function PopupContent() {
  const [config, setConfigState] = useState<AppConfig | null>(null);
  const [syncing, setSyncing] = useState(false);
  const toast = useKumoToastManager();

  useEffect(() => {
    getConfig().then(setConfigState);
  }, []);

  const handleToggle = async (checked: boolean) => {
    await updateConfig({ enabled: checked });
    setConfigState((c) => c ? { ...c, enabled: checked } : c);
    toast.add({
      variant: 'success',
      title: checked ? 'Monitoring enabled' : 'Monitoring paused',
      timeout: TOAST_TIMEOUT,
    });
  };

  const handleSyncAll = async () => {
    setSyncing(true);
    try {
      const response = await browser.runtime.sendMessage({ type: 'syncAll' });
      if (response.success) {
        const entries = Object.entries(response.results) as [string, number | string][];
        const successCount = entries.filter(([, v]) => typeof v === 'number').length;
        const totalCookies = entries.reduce((sum, [, v]) => sum + (typeof v === 'number' ? v : 0), 0);
        toast.add({
          variant: 'success',
          title: `Synced ${totalCookies} cookies from ${successCount} domains`,
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
      setSyncing(false);
    }
  };

  const openOptions = () => {
    browser.runtime.openOptionsPage();
  };

  if (!config) {
    return (
      <div className="popup-container">
        <Text variant="secondary">Loading...</Text>
      </div>
    );
  }

  return (
    <div className="popup-container">
      <div className="popup-header">
        <div style={{ display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <Globe size={20} />
          <Text variant="heading3">Cookie Monitor</Text>
        </div>
        <Badge variant={config.enabled ? 'success' : 'secondary'}>
          {config.enabled ? 'Active' : 'Paused'}
        </Badge>
      </div>

      <div className="status-list">
        <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', marginBottom: '0.5rem' }}>
          <Text variant="body">Monitoring</Text>
          <Switch
            checked={config.enabled}
            onCheckedChange={handleToggle}
          />
        </div>

        {config.websites.length === 0 ? (
          <Text variant="secondary">No websites configured</Text>
        ) : (
          <>
            <Text variant="secondary">
              {config.websites.length} website{config.websites.length > 1 ? 's' : ''}
            </Text>
            <div className="status-list">
              {config.websites.map((domain) => (
                <div key={domain} className="status-item">
                  <div className={`status-dot ${config.enabled ? 'active' : 'inactive'}`} />
                  <Text variant="mono" truncate>{domain}</Text>
                </div>
              ))}
            </div>
          </>
        )}
      </div>

      <div className="popup-footer">
        <Button
          variant="primary"
          size="sm"
          icon={<ArrowsClockwise />}
          loading={syncing}
          onClick={handleSyncAll}
          disabled={config.websites.length === 0}
          style={{ flex: 1 }}
        >
          Sync All
        </Button>
        <Button
          variant="secondary"
          size="sm"
          icon={<Gear />}
          onClick={openOptions}
        >
          Settings
        </Button>
      </div>
    </div>
  );
}

function App() {
  return (
    <Toasty>
      <PopupContent />
    </Toasty>
  );
}

export default App;
