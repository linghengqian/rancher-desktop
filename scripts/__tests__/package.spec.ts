import semver from 'semver';

import { jest } from '@jest/globals';

describe('package version fallback', () => {
  beforeEach(() => {
    jest.restoreAllMocks();
  });

  it('should fall back when git describe yields invalid semver', async() => {
    const { resolveBuildVersion } = await import('../lib/build-version');
    const resolved = resolveBuildVersion('1.2.3', () => 'invalid-tag');

    expect(resolved).toMatch(/-fallback$/);
    expect(semver.valid(resolved.replace(/-fallback$/, ''))).not.toBeNull();
  });
});
