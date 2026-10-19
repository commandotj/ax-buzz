import assert from 'node:assert/strict';
import test from 'node:test';
import { compareVersions, parseSemver, parseVersionSnapshot } from './version.mjs';

const files = {
  cli: 'v0.6.2 # x-release-please-version\n',
  wails: 'version: \'3\'\ninfo:\n  version: "0.6.2" # app\n',
  plist: '<key>CFBundleShortVersionString</key>\n<string>0.6.2</string>\n<key>CFBundleVersion</key>\n<string>0.6.2</string>',
  manifest: '{".":"0.6.2"}',
};

test('reads one synchronized version across app and bundled CLI metadata', () => {
  assert.equal(parseVersionSnapshot(files), '0.6.2');
});

test('rejects missing and duplicated version fields', () => {
  assert.throws(() => parseVersionSnapshot({ ...files, wails: "version: '3'\ninfo: {}\n" }), /Expected one Wails app version/);
  assert.throws(() => parseVersionSnapshot({ ...files, plist: `${files.plist}\n<key>CFBundleVersion</key><string>0.6.2</string>` }), /Expected one CFBundleVersion/);
});

test('rejects mismatched or malformed version sources', () => {
  assert.throws(() => parseVersionSnapshot({ ...files, manifest: '{".":"0.6.3"}' }), /Version sources disagree/);
  assert.throws(() => parseVersionSnapshot({ ...files, cli: '0.6.2\n' }), /leading v prefix/);
  assert.throws(() => parseSemver('v0.6.2', 'test'), /numeric MAJOR.MINOR.PATCH/);
});

test('compares semantic versions numerically', () => {
  assert.equal(compareVersions('0.10.0', '0.9.9'), 1);
  assert.equal(compareVersions('1.0.0', '1.0.0'), 0);
  assert.equal(compareVersions('0.5.9', '0.6.0'), -1);
});
