import os from 'os';

import { jest } from '@jest/globals';

import mockModules from '@pkg/utils/testUtils/mockModules';

describe('setupUpdate', () => {
  const modules = mockModules({
    '@pkg/utils/logging': undefined,
    electron:             undefined,
  });
  const updaterMocks = {
    AppImageUpdater: jest.fn(),
    MacUpdater:      jest.fn(),
    MsiUpdater:      jest.fn(),
  };
  jest.unstable_mockModule('electron-updater', () => ({
    __esModule: true,
    ...updaterMocks,
    Provider: class {},
    AppUpdater: class {},
    NsisUpdater: class {},
    ProgressInfo: class {},
    UpdateInfo: class {},
  }));
  jest.unstable_mockModule('electron-updater/out/ElectronAppAdapter', () => ({
    __esModule: true,
    ElectronAppAdapter: jest.fn(() => ({ appUpdateConfigPath: '/tmp/app-update.yml' })),
  }));
  jest.unstable_mockModule('fs', () => {
    const actual = jest.requireActual<typeof import('fs')>('fs');
    return {
      __esModule: true,
      default: {
        ...actual,
        promises: {
          readFile: jest.fn(() => Promise.resolve('provider: custom')),
        },
      },
      ...actual,
    };
  });
  jest.unstable_mockModule('yaml', () => ({
    __esModule: true,
    default: {
      parse: jest.fn(() => ({ upgradeServer: 'https://example.com' })),
    },
  }));

  beforeEach(() => {
    const app = modules.electron.app as typeof modules.electron.app & { getVersion: jest.Mock };
    const ipcMain = modules.electron.ipcMain as typeof modules.electron.ipcMain & { on: jest.Mock };
    app.getVersion = jest.fn();
    ipcMain.on = jest.fn();
    app.getVersion.mockReturnValue('unknown');
    delete process.env.APPIMAGE;
  });

  afterEach(() => {
    updaterMocks.AppImageUpdater.mockReset();
    updaterMocks.MacUpdater.mockReset();
    updaterMocks.MsiUpdater.mockReset();
    jest.resetModules();
  });

  it('skips updater when app version is invalid', async() => {
    const { default: setupUpdate } = await import('../index');
    await expect(setupUpdate(true)).resolves.toBe(false);
    expect(updaterMocks.AppImageUpdater).not.toHaveBeenCalled();
  });

  if (os.platform() === 'linux') {
    it('skips updater on non-AppImage Linux builds', async() => {
      const app = modules.electron.app as typeof modules.electron.app & { getVersion: jest.Mock };
      app.getVersion.mockReturnValue('1.22.0');
      const { default: setupUpdate } = await import('../index');
      await expect(setupUpdate(true)).resolves.toBe(false);
      expect(updaterMocks.AppImageUpdater).not.toHaveBeenCalled();
    });
  }
});
