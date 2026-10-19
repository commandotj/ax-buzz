import assert from 'node:assert/strict';
import test from 'node:test';
import { renderBuzzCask } from './render-buzz-cask.mjs';

const armSha = 'a'.repeat(64);
const intelSha = 'b'.repeat(64);

test('renders a macOS Buzz.app Cask with architecture URLs and verified hashes', () => {
  const cask = renderBuzzCask('0.7.0', armSha, intelSha);
  assert.match(cask, /cask "buzz" do/);
  assert.match(cask, /arch arm: "arm64", intel: "amd64"/);
  assert.match(cask, /sha256 arm: +"a{64}",\n +intel: "b{64}"/);
  assert.match(cask, /Buzz-v#\{version\}-macos-#\{arch\}\.dmg/);
  assert.match(cask, /app "Buzz\.app"/);
  assert.match(cask, /depends_on macos: ">= :monterey"/);
  assert.doesNotMatch(cask, /Formula|PLACEHOLDER/);
});

test('rejects malformed release versions and checksums', () => {
  assert.throws(() => renderBuzzCask('v0.7.0', armSha, intelSha), /Invalid Buzz version/);
  assert.throws(() => renderBuzzCask('0.7.0', 'bad', intelSha), /lowercase SHA-256/);
});
