import { defineConfig } from 'wxt';

export default defineConfig({
  modules: ['@wxt-dev/module-react'],
  manifest: {
    name: 'Cookie Monitor',
    description: 'Monitor website cookies and sync to WebDAV server',
    permissions: ['cookies', 'storage'],
    host_permissions: ['<all_urls>'],
  },
});
