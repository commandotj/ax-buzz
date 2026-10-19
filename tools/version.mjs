import { execFileSync } from 'node:child_process';
import { readFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const versionFiles = {
  cli: 'apps/cli/.version',
  wails: 'apps/desktop/build/config.yml',
  plist: 'apps/desktop/build/darwin/Info.plist',
  manifest: '.release-please-manifest.json',
};

export function parseSemver(value, name) {
  if (typeof value !== 'string') throw new Error(`${name} must be a string`);
  const match = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/.exec(value);
  if (!match) throw new Error(`${name} must be a numeric MAJOR.MINOR.PATCH version; got ${value}`);
  return match.slice(1).map(Number);
}

function singleMatch(source, expression, name) {
  const matches = [...source.matchAll(expression)];
  if (matches.length !== 1) throw new Error(`Expected one ${name}; found ${matches.length}`);
  return matches[0][1];
}

export function parseVersionSnapshot(files) {
  const cliVersion = files.cli.trim().split(/\s+/, 1)[0];
  if (!cliVersion.startsWith('v')) throw new Error('CLI version must keep leading v prefix');
  const cli = cliVersion.slice(1);
  const wails = singleMatch(
    files.wails,
    /^\s+version:\s*["']([^"']+)["']\s*(?:#.*)?$/gm,
    'Wails app version marker',
  );
  const shortVersion = singleMatch(
    files.plist,
    /<key>CFBundleShortVersionString<\/key>\s*<string>([^<]+)<\/string>/g,
    'CFBundleShortVersionString',
  );
  const bundleVersion = singleMatch(
    files.plist,
    /<key>CFBundleVersion<\/key>\s*<string>([^<]+)<\/string>/g,
    'CFBundleVersion',
  );
  const manifest = JSON.parse(files.manifest)['.'];
  if (typeof manifest !== 'string') throw new Error('Release Please manifest must define root package version');

  const values = { cli, wails, shortVersion, bundleVersion, manifest };
  for (const [name, value] of Object.entries(values)) parseSemver(value, name);
  if (new Set(Object.values(values)).size !== 1) {
    throw new Error(`Version sources disagree: ${JSON.stringify(values)}`);
  }
  return cli;
}

export function compareVersions(left, right) {
  const a = parseSemver(left, 'left version');
  const b = parseSemver(right, 'right version');
  for (let index = 0; index < a.length; index += 1) {
    if (a[index] !== b[index]) return Math.sign(a[index] - b[index]);
  }
  return 0;
}

function readSnapshot(ref) {
  const read = (path) => ref
    ? execFileSync('git', ['show', `${ref}:${path}`], { encoding: 'utf8' })
    : readFileSync(path, 'utf8');
  return Object.fromEntries(Object.entries(versionFiles).map(([key, path]) => [key, read(path)]));
}

function main(args) {
  const command = args[0];
  const current = parseVersionSnapshot(readSnapshot());
  if (command === 'check') {
    process.stdout.write(`${current}\n`);
    return;
  }
  if (command === 'read' && args[1]) {
    process.stdout.write(`${parseVersionSnapshot(readSnapshot(args[1]))}\n`);
    return;
  }
  if (command === 'compare' && args[1]) {
    const other = parseVersionSnapshot(readSnapshot(args[1]));
    if (compareVersions(current, other) <= 0) {
      throw new Error(`Current version ${current} is not newer than ${other}`);
    }
    process.stdout.write(`${current}\n`);
    return;
  }
  throw new Error('Usage: node tools/version.mjs check | read <git-ref> | compare <git-ref>');
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    main(process.argv.slice(2));
  } catch (error) {
    process.stderr.write(`::error::${error.message}\n`);
    process.exitCode = 1;
  }
}
