import { writeFileSync } from 'node:fs';
import { resolve } from 'node:path';
import { pathToFileURL } from 'node:url';

const semver = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;
const sha256 = /^[a-f0-9]{64}$/;

export function renderBuzzCask(version, armSha, intelSha) {
  if (!semver.test(version)) throw new Error(`Invalid Buzz version: ${version}`);
  if (!sha256.test(armSha) || !sha256.test(intelSha)) throw new Error('Both DMG checksums must be lowercase SHA-256 values');

  return `cask "buzz" do
  arch arm: "arm64", intel: "amd64"

  version "${version}"
  sha256 arm:   "${armSha}",
         intel: "${intelSha}"

  url "https://github.com/commandotj/ax-buzz/releases/download/v#{version}/Buzz-v#{version}-macos-#{arch}.dmg"
  name "Buzz"
  desc "Manage Homebrew software from a native macOS app"
  homepage "https://github.com/commandotj/ax-buzz"
  depends_on macos: ">= :monterey"

  app "Buzz.app"
end
`;
}

if (process.argv[1] && import.meta.url === pathToFileURL(resolve(process.argv[1])).href) {
  try {
    const [version, armSha, intelSha, outputPath] = process.argv.slice(2);
    if (!outputPath) throw new Error('Usage: node tools/render-buzz-cask.mjs <version> <arm-sha256> <intel-sha256> <output-path>');
    writeFileSync(outputPath, renderBuzzCask(version, armSha, intelSha));
  } catch (error) {
    process.stderr.write(`::error::${error.message}\n`);
    process.exitCode = 1;
  }
}
