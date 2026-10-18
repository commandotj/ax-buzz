import { existsSync } from "node:fs";
import { homedir } from "node:os";
import { delimiter, dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { spawn, spawnSync } from "node:child_process";

const desktopDir = resolve(dirname(fileURLToPath(import.meta.url)), "..");
const rootDir = resolve(desktopDir, "../..");
const cliPath = resolve(rootDir, "apps/cli/bin/buzz");

if (!existsSync(cliPath)) {
  const build = spawnSync("pnpm", ["--filter", "@buzz/cli", "build"], {
    cwd: rootDir,
    stdio: "inherit",
  });
  if (build.error) throw build.error;
  if (build.status !== 0) process.exit(build.status ?? 1);
}

const installedWails = resolve(homedir(), "go/bin/wails3");
const wails = process.env.WAILS3_BIN || (existsSync(installedWails) ? installedWails : "wails3");
const wailsPath = wails.includes("/") ? `${dirname(wails)}${delimiter}${process.env.PATH ?? ""}` : process.env.PATH;
const child = spawn(wails, process.argv.slice(2), {
  cwd: desktopDir,
  stdio: "inherit",
  env: { ...process.env, PATH: wailsPath, BUZZ_CLI_PATH: cliPath },
});

child.on("error", (error) => {
  process.stderr.write(`Unable to start Wails CLI: ${error.message}\n`);
  process.exitCode = 1;
});
child.on("exit", (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exitCode = code ?? 1;
});
