import semver from 'semver';

import { jest } from '@jest/globals';

describe('package version fallback', () => {
  beforeEach(() => {
    jest.restoreAllMocks();
  });

  it('should fall back when git describe yields invalid semver', async() => {
    const { Builder } = await import('../package');
    const resolved = Builder.resolveBuildVersion(undefined, () => 'invalid-tag');

    expect(resolved).toMatch(/-fallback$/);
    expect(semver.valid(resolved.replace(/-fallback$/, ''))).not.toBeNull();
  });
});
