import childProcess from 'child_process';

import semver from 'semver';

import { jest } from '@jest/globals';

describe('package version fallback', () => {
  beforeEach(() => {
    jest.restoreAllMocks();
  });

  it('should fall back when git describe yields invalid semver', async() => {
    jest.spyOn(childProcess, 'execFileSync').mockReturnValueOnce(Buffer.from('invalid-tag'));
    const { default: PackageBuilder } = await import('../package');
    const builder = new PackageBuilder();
    const resolved = (builder as any).resolveBuildVersion();

    expect(resolved).toMatch(/-fallback$/);
    expect(semver.valid(resolved.replace(/-fallback$/, ''))).not.toBeNull();
  });
});
