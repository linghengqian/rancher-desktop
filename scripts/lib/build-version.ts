import childProcess from 'child_process';

import semver from 'semver';

const DEFAULT_VERSION = '0.0.0';
const FALLBACK_SUFFIX = '-fallback';

export function resolveBuildVersion(
  packageVersion?: string,
  gitDescribe: () => string = () => childProcess.execFileSync('git', ['describe', '--tags']).toString().trim(),
): string {
  const fallbackVersion = packageVersion ?? DEFAULT_VERSION;
  let fullBuildVersion: string;

  try {
    const described = gitDescribe();
    const validatedVersion = semver.valid(described.replace(/^v/, ''));

    if (!validatedVersion) {
      throw new Error(`Invalid git version ${described}`);
    }
    fullBuildVersion = validatedVersion;
  } catch (error) {
    console.warn(`Failed to parse git version (${ error }); using fallback.`);
    fullBuildVersion = `${fallbackVersion}${FALLBACK_SUFFIX}`;
  }

  if (!semver.valid(fullBuildVersion)) {
    const fallbackBase = semver.valid(fallbackVersion) ? fallbackVersion : DEFAULT_VERSION;

    console.warn(`Invalid build version ${fullBuildVersion}; falling back to ${fallbackBase}${FALLBACK_SUFFIX}`);
    fullBuildVersion = `${fallbackBase}${FALLBACK_SUFFIX}`;
  }

  return fullBuildVersion;
}
