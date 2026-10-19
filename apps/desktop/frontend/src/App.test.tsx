import { act, cleanup, fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import App from "./App";

const { handlers, startOperation, cancelOperation } = vi.hoisted(() => ({
  handlers: new Map<string, (event: { data: unknown }) => void>(),
  startOperation: vi.fn(async () => undefined),
  cancelOperation: vi.fn(async () => undefined),
}));

vi.mock("@tanstack/react-virtual", () => ({
  useVirtualizer: ({ count, estimateSize, getItemKey, getScrollElement }: { count: number; estimateSize: () => number; getItemKey: (index: number) => string; getScrollElement: () => Element | null }) => {
    getScrollElement();
    return {
      getVirtualItems: () => Array.from({ length: count }, (_, index) => ({
        key: getItemKey(index), index, size: estimateSize(), start: index * estimateSize(),
      })),
      getTotalSize: () => count * 62,
      measureElement: vi.fn(),
    };
  },
}));

vi.mock("@wailsio/runtime", () => ({
  Events: {
    On: vi.fn((_name: string, handler: (event: { data: unknown }) => void) => {
      handlers.set("buzz:cli-event", handler);
      return () => handlers.delete("buzz:cli-event");
    }),
  },
}));

vi.mock("../bindings/github.com/commandotj/ax-buzz/apps/desktop", () => ({
  CLIService: { StartOperation: startOperation, CancelOperation: cancelOperation },
}));

const formula = {
  name: "alpha", description: "Alpha package", tap: "", version: "2.0", installedVersion: "1.0",
  isCask: false, isInstalled: true, isOutdated: true, isPinned: false, installSupported: true,
};
const cask = {
  name: "beta", description: "Beta app", tap: "homebrew/cask", version: "1.0", installedVersion: "1.0",
  isCask: true, isInstalled: true, isOutdated: false, isPinned: false, installSupported: true,
};
const available = {
  name: "gamma", description: "Gamma tool", tap: "", version: "3.0", installedVersion: "",
  isCask: false, isInstalled: false, isOutdated: false, isPinned: false, installSupported: true,
};
const secondUpdate = {
  name: "delta", description: "Delta package", tap: "homebrew/core", version: "2.0", installedVersion: "1.0",
  isCask: false, isInstalled: true, isOutdated: true, isPinned: false, installSupported: true,
};
const unsupported = {
  name: "omega", description: "", tap: "custom/tap", version: "", installedVersion: "",
  isCask: false, isInstalled: true, isOutdated: true, isPinned: true, installSupported: false,
};

async function send(payload: Record<string, unknown>) {
  await act(async () => {
    handlers.get("buzz:cli-event")?.({ data: { protocolVersion: 1, ...payload } });
  });
}

async function loadPackages(packages = [formula, cask, available, unsupported, secondUpdate]) {
  await send({ event: "data", data: { packages } });
  await send({ event: "result", data: { success: true, count: packages.length, cacheAgeSeconds: 30 } });
}

function clickPackage(name: string) {
  const list = screen.getByRole("list", { name: "Packages" });
  const label = within(list).getByText(name, { selector: ".package-summary strong" });
  fireEvent.click(label.closest("button")!);
}

beforeEach(() => {
  handlers.clear();
  startOperation.mockClear();
  cancelOperation.mockClear();
  localStorage.clear();
  Object.defineProperty(window, "matchMedia", {
    configurable: true,
    value: vi.fn().mockImplementation(() => ({
      matches: false,
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
    })),
  });
  Object.defineProperty(HTMLElement.prototype, "scrollTo", { configurable: true, value: vi.fn() });
  HTMLDialogElement.prototype.showModal = vi.fn(function (this: HTMLDialogElement) { this.open = true; });
  HTMLDialogElement.prototype.close = vi.fn(function (this: HTMLDialogElement) { this.open = false; });
});

afterEach(cleanup);

describe("Buzz desktop app", () => {
  it("loads cached packages, filters sections, and applies appearance preference", async () => {
    render(<App />);
    expect(startOperation).toHaveBeenCalledWith("list", "", false);
    await send({ event: "data", data: { packages: [formula] } });
    await send({ event: "data", data: {} });
    await send({ event: "result", data: { success: true } });
    expect(screen.getAllByText("Catalog not loaded").length).toBe(2);
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await loadPackages();
    await send({ event: "result", data: { success: true } });
    expect(await screen.findByText(/Loaded 5 packages/)).toBeTruthy();
    expect(screen.getAllByText("Cached · just refreshed").length).toBe(2);
    expect(within(screen.getByRole("list", { name: "Packages" })).getByText("alpha")).toBeTruthy();

    fireEvent.click(within(screen.getByRole("navigation")).getByRole("button", { name: /Discover/ }));
    expect(screen.getByText("Browse the Homebrew formula and cask catalog.")).toBeTruthy();
    expect(screen.getAllByRole("button", { name: "All" }).length).toBe(1);
    fireEvent.click(screen.getByRole("button", { name: "Casks" }));
    expect(within(screen.getByRole("list", { name: "Packages" })).getByText("beta", { selector: ".package-summary strong" })).toBeTruthy();
    expect(within(screen.getByRole("list", { name: "Packages" })).queryByText("gamma", { selector: ".package-summary strong" })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "All" }));
    fireEvent.change(screen.getByRole("textbox", { name: "Search packages" }), { target: { value: "GAMMA" } });
    expect(within(screen.getByRole("list", { name: "Packages" })).getByText("gamma", { selector: ".package-summary strong" })).toBeTruthy();
    expect(within(screen.getByRole("list", { name: "Packages" })).queryByText("alpha", { selector: ".package-summary strong" })).toBeNull();

    fireEvent.change(screen.getByRole("combobox", { name: "Appearance" }), { target: { value: "dark" } });
    expect(localStorage.getItem("buzz.appearance")).toBe("dark");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    fireEvent.change(screen.getByRole("textbox", { name: "Search packages" }), { target: { value: "not-found" } });
    expect(screen.getByText("No matching packages")).toBeTruthy();
  });

  it("shows cached empty, stale, and failure states with retry and cancel", async () => {
    render(<App />);
    await send({ event: "progress", data: { message: "Loading catalog" } });
    expect(screen.getByText("Loading catalog")).toBeTruthy();
    await send({ event: "progress" });
    expect(screen.getByText("Working…")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(cancelOperation).toHaveBeenCalledOnce();

    await loadPackages([]);
    expect(screen.getByText("No catalog data yet")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Load catalog" }));
    expect(startOperation).toHaveBeenLastCalledWith("refresh", "", false);
    await send({ event: "result", data: { success: true, count: 0, cacheAgeSeconds: 90000, cacheStale: true } });
    expect(screen.getAllByText("Stale · 1d old").length).toBe(2);
    await send({ event: "error", error: "refresh failed" });
    expect(screen.getByRole("alert").textContent).toContain("refresh failed");
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(startOperation).toHaveBeenLastCalledWith("list", "", false);
    await send({ event: "result", data: { success: false, error: "failed" } });
    expect(screen.getByRole("alert").textContent).toContain("failed");
    await send({ event: "error" });
    expect(screen.getByRole("alert").textContent).toContain("Buzz CLI operation failed.");
  });

  it("confirms package removal and refreshes installed data", async () => {
    render(<App />);
    await loadPackages();
    clickPackage("alpha");
    fireEvent.click(screen.getByRole("button", { name: "Remove package" }));
    expect(screen.getByRole("heading", { name: "Remove alpha?" })).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    fireEvent.click(screen.getByRole("button", { name: "Remove package" }));
    fireEvent.click(within(screen.getByRole("list", { name: "Packages" })).getByRole("button", { name: "Remove alpha" }));
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Cancel" }));
    fireEvent.click(screen.getByRole("button", { name: "Remove package" }));
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Remove alpha" }));
    expect(startOperation).toHaveBeenLastCalledWith("uninstall", "alpha", false);
    await send({ event: "result", data: { success: true, package: "alpha" } });
    expect(screen.getByText("Removed alpha")).toBeTruthy();
    expect(startOperation).toHaveBeenLastCalledWith("list", "", false);
    await loadPackages();
  });

  it("browses package details and requests install", async () => {
    render(<App />);
    await loadPackages();
    fireEvent.click(within(screen.getByRole("navigation")).getByRole("button", { name: /Discover/ }));
    clickPackage("gamma");
    expect(screen.getByRole("heading", { name: "gamma" })).toBeTruthy();
    expect(screen.getByText("Formula")).toBeTruthy();
    fireEvent.click(within(screen.getByRole("list", { name: "Packages" })).getByRole("button", { name: "Install gamma" }));
    expect(screen.getByRole("heading", { name: "Install gamma?" })).toBeTruthy();
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Install gamma" }));
    expect(startOperation).toHaveBeenLastCalledWith("install", "gamma", false);
    await send({ event: "result", data: { success: true, package: "gamma" } });
    expect(screen.getByText("Installed gamma")).toBeTruthy();
    await loadPackages();
    fireEvent.click(screen.getByRole("button", { name: "Install package" }));
    expect(screen.getByText("This will install the Homebrew formula on this Mac.")).toBeTruthy();
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Cancel" }));
  });

  it("upgrades eligible packages in order and handles invalid protocol versions", async () => {
    render(<App />);
    await loadPackages();
    fireEvent.click(within(screen.getByRole("navigation")).getByRole("button", { name: /Updates/ }));
    expect(screen.getByText("Installed packages with newer versions available.")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Upgrade all" }));
    expect(startOperation).toHaveBeenLastCalledWith("upgrade", "alpha", false);
    await send({ event: "result", data: { success: true } });
    expect(screen.getByText("Upgraded package")).toBeTruthy();
    expect(startOperation).toHaveBeenLastCalledWith("upgrade", "delta", false);
    await send({ event: "result", data: { success: true, package: "delta" } });
    expect(startOperation).toHaveBeenLastCalledWith("list", "", false);
    await send({ event: "version", protocolVersion: 7 });
    expect(screen.getByRole("alert").textContent).toContain("Unsupported CLI protocol version 7");
  });

  it("does nothing when update candidates are pinned or unsupported", async () => {
    render(<App />);
    await loadPackages([unsupported]);
    fireEvent.click(within(screen.getByRole("navigation")).getByRole("button", { name: /Updates/ }));
    expect(screen.getByRole("button", { name: "Upgrade all" })).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Upgrade all" }));
    expect(startOperation).toHaveBeenCalledTimes(1);
  });

  it("confirms a single upgrade with matching description", async () => {
    render(<App />);
    await loadPackages();
    fireEvent.click(within(screen.getByRole("navigation")).getByRole("button", { name: /Updates/ }));
    clickPackage("alpha");
    fireEvent.click(screen.getByRole("button", { name: "Upgrade package" }));
    expect(screen.getByRole("heading", { name: "Upgrade alpha?" })).toBeTruthy();
    expect(screen.getByText("This will upgrade the Homebrew formula on this Mac.")).toBeTruthy();
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Upgrade alpha" }));
    expect(startOperation).toHaveBeenLastCalledWith("upgrade", "alpha", false);
    await send({ event: "result", data: { success: true, package: "alpha" } });
    await loadPackages();
  });

  it("routes row upgrade action through confirmation", async () => {
    render(<App />);
    await loadPackages();
    fireEvent.click(within(screen.getByRole("navigation")).getByRole("button", { name: /Updates/ }));
    fireEvent.click(within(screen.getByRole("list", { name: "Packages" })).getByRole("button", { name: "Upgrade alpha" }));
    expect(screen.getByRole("heading", { name: "Upgrade alpha?" })).toBeTruthy();
    fireEvent(screen.getByRole("dialog"), new Event("cancel", { cancelable: true }));
    expect(screen.queryByRole("heading", { name: "Upgrade alpha?" })).toBeNull();
  });

  it("handles persisted appearance, macOS chrome, and cache-age units", async () => {
    let onMediaChange: (() => void) | undefined;
    localStorage.setItem("buzz.appearance", "invalid");
    Object.defineProperty(navigator, "userAgent", { configurable: true, value: "Mozilla Macintosh Mac OS X" });
    Object.defineProperty(window, "matchMedia", {
      configurable: true,
      value: vi.fn().mockImplementation(() => ({
        matches: true,
        addEventListener: (_event: string, listener: () => void) => { onMediaChange = listener; },
        removeEventListener: vi.fn(),
      })),
    });
    render(<App />);
    expect(screen.getByRole("banner", { name: "Buzz window" })).toBeTruthy();
    expect(localStorage.getItem("buzz.appearance")).toBe("system");
    expect(document.documentElement.classList.contains("dark")).toBe(true);
    onMediaChange?.();
    await send({ event: "result", data: { success: true, cacheAgeSeconds: 60, count: 0 } });
    expect(screen.getAllByText("Cached · 1m old").length).toBe(2);
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await send({ event: "result", data: { success: true, cacheAgeSeconds: 3600, count: 0 } });
    expect(screen.getAllByText("Cached · 1h old").length).toBe(2);
  });

  it("restores explicit light appearance", () => {
    localStorage.setItem("buzz.appearance", "light");
    render(<App />);
    expect((screen.getByRole("combobox", { name: "Appearance" }) as HTMLSelectElement).value).toBe("light");
    expect(document.documentElement.classList.contains("dark")).toBe(false);
  });

  it("surfaces launch errors, uninstall details, and system appearance", async () => {
    startOperation.mockRejectedValueOnce(new Error("bridge unavailable"));
    render(<App />);
    await waitFor(() => expect(screen.getByRole("alert").textContent).toContain("bridge unavailable"));
    expect(screen.getAllByText("Catalog not loaded").length).toBe(2);
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    await loadPackages();
    clickPackage("beta");
    expect(screen.getByText("homebrew/cask")).toBeTruthy();
    fireEvent.click(screen.getByRole("button", { name: "Remove package" }));
    expect(screen.getByText("This will uninstall the Homebrew cask from this Mac.")).toBeTruthy();
    fireEvent.click(within(screen.getByRole("dialog")).getByRole("button", { name: "Cancel" }));
    fireEvent.click(screen.getByRole("button", { name: "Close" }));
    fireEvent.change(screen.getByRole("combobox", { name: "Appearance" }), { target: { value: "system" } });
    expect(localStorage.getItem("buzz.appearance")).toBe("system");
    clickPackage("omega");
    expect(screen.getByText("No description provided.")).toBeTruthy();
    expect(screen.getByText("Buzz cannot manage this package.")).toBeTruthy();
    expect(within(screen.getByRole("complementary", { name: "omega details" }).querySelector("dd.mono")!).getByText("Installed")).toBeTruthy();
    expect(screen.getByText("—")).toBeTruthy();
    startOperation.mockRejectedValueOnce("plain failure");
    fireEvent.click(screen.getByRole("button", { name: "Refresh" }));
    await waitFor(() => expect(screen.getByRole("alert").textContent).toContain("plain failure"));
  });
});
