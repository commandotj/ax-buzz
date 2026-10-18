import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Events } from "@wailsio/runtime";
import { useVirtualizer } from "@tanstack/react-virtual";
import {
  Check, ChevronRight, CircleAlert, Download, LoaderCircle,
  Package, RefreshCw, Search, Settings2, Trash2, Upload,
} from "lucide-react";
import buzzIcon from "../../../../logos/export/icon.svg";
import { CLIService } from "../bindings/github.com/commandotj/ax-buzz/apps/desktop";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Separator } from "@/components/ui/separator";

type PackageInfo = {
  name: string; description: string; tap: string; version: string;
  installedVersion: string; isCask: boolean; isInstalled: boolean;
  isOutdated: boolean; isPinned: boolean; installSupported: boolean;
};
type StreamEvent = {
  protocolVersion: number; event: "progress" | "data" | "result" | "error";
  data?: { message?: string; operation?: string; success?: boolean; count?: number;
    cacheAgeSeconds?: number; cacheStale?: boolean; packages?: PackageInfo[]; error?: string; package?: string };
  error?: string;
};
type Action = { operation: "install" | "uninstall" | "upgrade"; pkg: PackageInfo };
type View = "Installed" | "Discover" | "Updates";
type Appearance = "system" | "light" | "dark";
type WailsEvent<T> = { data: T };

const packageKey = (pkg: PackageInfo) => `${pkg.isCask ? "cask" : "formula"}:${pkg.name}`;
const displaySource = (pkg: PackageInfo) => `${pkg.isCask ? "Cask" : "Formula"} · ${pkg.tap || "homebrew/core"}`;

function App() {
  const macOSWindow = /Macintosh|Mac OS X/.test(navigator.userAgent);
  const [view, setView] = useState<View>("Installed");
  const [kindFilter, setKindFilter] = useState<"all" | "formula" | "cask">("all");
  const [packages, setPackages] = useState<PackageInfo[]>([]);
  const [search, setSearch] = useState("");
  const [selectedKey, setSelectedKey] = useState("");
  const [appearance, setAppearance] = useState<Appearance>(() => readAppearance());
  const [status, setStatus] = useState("Loading cached Homebrew data…");
  const [error, setError] = useState("");
  const [progress, setProgress] = useState("");
  const [cacheAge, setCacheAge] = useState<number | null>(null);
  const [cacheStale, setCacheStale] = useState(false);
  const [busy, setBusy] = useState(false);
  const [confirm, setConfirm] = useState<Action | null>(null);
  const [detailOpen, setDetailOpen] = useState(false);
  const [actionNotice, setActionNotice] = useState("");
  const incoming = useRef<PackageInfo[]>([]);
  const operation = useRef("");
  const actions = useRef<Action[]>([]);
  const dialog = useRef<HTMLDialogElement>(null);
  const listRef = useRef<HTMLDivElement>(null);

  const launch = async (name: string, pkg?: PackageInfo) => {
    operation.current = name;
    incoming.current = [];
    setProgress("");
    setBusy(true);
    setError("");
    try {
      await CLIService.StartOperation(name, pkg?.name ?? "", pkg?.isCask ?? false);
    } catch (cause) {
      setBusy(false);
      operation.current = "";
      setError(cause instanceof Error ? cause.message : String(cause));
    }
  };

  useEffect(() => {
    const unsubscribe = Events.On("buzz:cli-event", (event: WailsEvent<StreamEvent>) => {
      const payload = event.data;
      if (payload.protocolVersion !== 1) {
        setError(`Unsupported CLI protocol version ${payload.protocolVersion}.`);
        setBusy(false);
        return;
      }
      if (payload.event === "progress") {
        setProgress(payload.data?.message ?? "Working…");
        return;
      }
      if (payload.event === "data") {
        incoming.current.push(...(payload.data?.packages ?? []));
        return;
      }
      if (payload.event === "error" || payload.data?.success === false) {
        setError(payload.error ?? payload.data?.error ?? "Buzz CLI operation failed.");
        setBusy(false);
        operation.current = "";
        actions.current = [];
        return;
      }
      const completed = operation.current;
      if (completed === "list" || completed === "refresh" || completed === "installed") {
        setPackages(incoming.current);
        setCacheAge(payload.data?.cacheAgeSeconds ?? null);
        setCacheStale(payload.data?.cacheStale ?? false);
        setStatus(`Loaded ${payload.data?.count ?? incoming.current.length} packages`);
        setBusy(false);
        operation.current = "";
        return;
      }
      if (completed === "install" || completed === "uninstall" || completed === "upgrade") {
        setActionNotice(`${completed === "uninstall" ? "Removed" : completed === "install" ? "Installed" : "Upgraded"} ${payload.data?.package ?? "package"}`);
        const next = actions.current.shift();
        if (next) void launch(next.operation, next.pkg);
        else void launch("list");
      }
    });
    void launch("list");
    return unsubscribe;
  }, []);

  useEffect(() => {
    localStorage.setItem("buzz.appearance", appearance);
    const media = window.matchMedia("(prefers-color-scheme: dark)");
    const apply = () => document.documentElement.classList.toggle("dark", appearance === "dark" || (appearance === "system" && media.matches));
    apply();
    media.addEventListener("change", apply);
    return () => media.removeEventListener("change", apply);
  }, [appearance]);

  useEffect(() => {
    if (confirm) dialog.current?.showModal();
    else if (dialog.current?.open) dialog.current.close();
  }, [confirm]);

  const visiblePackages = useMemo(() => {
    const query = search.trim().toLowerCase();
    return packages.filter((pkg) => {
      const inView = view === "Installed" ? pkg.isInstalled : view === "Updates" ? pkg.isInstalled && pkg.isOutdated : true;
      const inKind = view !== "Discover" || kindFilter === "all" || pkg.isCask === (kindFilter === "cask");
      return inView && inKind && (!query || `${pkg.name} ${pkg.description} ${pkg.tap}`.toLowerCase().includes(query));
    }).sort((a, b) => a.name.localeCompare(b.name));
  }, [kindFilter, packages, search, view]);
  const getPackageKey = useCallback((index: number) => packageKey(visiblePackages[index]), [visiblePackages]);
  const virtualizer = useVirtualizer({
    count: visiblePackages.length,
    getScrollElement: () => listRef.current,
    estimateSize: () => 62,
    getItemKey: getPackageKey,
    overscan: 12,
    directDomUpdates: true,
  });
  const virtualRows = virtualizer.getVirtualItems();
  const selected = visiblePackages.find((pkg) => packageKey(pkg) === selectedKey) ?? visiblePackages[0] ?? null;
  const updateCount = packages.filter((pkg) => pkg.isInstalled && pkg.isOutdated).length;
  const installedCount = packages.filter((pkg) => pkg.isInstalled).length;
  const cacheText = cacheAge === null ? "Catalog not loaded" : `${cacheStale ? "Stale · " : "Cached · "}${formatAge(cacheAge)}`;

  useEffect(() => listRef.current?.scrollTo({ top: 0 }), [kindFilter, search, view]);

  const queueAction = (action: Action) => {
    actions.current = [];
    setConfirm(action);
  };

  const confirmAction = () => {
    if (!confirm) return;
    const action = confirm;
    setConfirm(null);
    setActionNotice("");
    actions.current = [];
    void launch(action.operation, action.pkg);
  };

  const upgradeAll = () => {
    const queue = packages.filter((pkg) => pkg.isInstalled && pkg.isOutdated && !pkg.isPinned && pkg.installSupported);
    if (queue.length === 0) return;
    actions.current = queue.slice(1).map((pkg) => ({ operation: "upgrade", pkg }));
    setActionNotice("");
    void launch("upgrade", queue[0]);
  };

  const navItems: { name: View; count: number; icon: typeof Package }[] = [
    { name: "Installed", count: installedCount, icon: Package },
    { name: "Discover", count: packages.length, icon: Search },
    { name: "Updates", count: updateCount, icon: Upload },
  ];

  return (
    <div className={`app-shell ${macOSWindow ? "macos-shell" : ""}`}>
      {macOSWindow && <header className="window-chrome" aria-label="Buzz window"><span className="traffic-light-reserve" aria-hidden="true" /><strong>Buzz</strong><span className="window-chrome-context">HOMEBREW</span></header>}
      <aside className="sidebar" aria-label="Main navigation">
        <div className="brand-lockup"><div className="brand-mark" aria-hidden="true"><img src={buzzIcon} alt="" /></div><div><strong>Buzz</strong><span>Homebrew library</span></div></div>
        <Separator className="sidebar-rule" />
        <p className="nav-heading">LIBRARY</p>
        <nav className="nav-list">
          {navItems.map(({ name, count, icon: Icon }) => <Button key={name} variant={view === name ? "secondary" : "ghost"} className={`nav-item ${view === name ? "nav-current" : ""}`} aria-current={view === name ? "page" : undefined} onClick={() => setView(name)}><Icon size={16} />{name}<span className="nav-count">{count}</span></Button>)}
        </nav>
        <div className="sidebar-spacer" />
        <label className="appearance-control"><span><Settings2 size={15} /> Appearance</span><select aria-label="Appearance" value={appearance} onChange={(event) => setAppearance(event.target.value as Appearance)}><option value="system">System</option><option value="light">Light</option><option value="dark">Dark</option></select></label>
        <div className="sidebar-footer"><span className="status-dot" />On this Mac</div>
      </aside>

      <main className="main-content">
        <header className="page-header">
          <div><p className="breadcrumb">Buzz <ChevronRight size={13} /> Homebrew</p><h1>{view}</h1><p className="page-subtitle">{view === "Installed" ? "Software managed by Homebrew on this Mac." : view === "Updates" ? "Installed packages with newer versions available." : "Browse the Homebrew formula and cask catalog."}</p></div>
          <div className="header-actions"><span className={`cache-label ${cacheStale ? "is-stale" : ""}`} title={cacheAge === null ? "" : new Date(Date.now() - cacheAge * 1000).toLocaleString()}>{cacheStale && <CircleAlert size={14} />}{cacheText}</span><Button variant="outline" onClick={() => void launch("refresh")} disabled={busy}><RefreshCw size={15} className={busy ? "spin" : ""} />Refresh</Button></div>
        </header>

        <div className="toolbar"><label className="search-box"><Search size={16} /><input aria-label="Search packages" placeholder="Search name or description" value={search} onChange={(event) => setSearch(event.target.value)} /></label>{view === "Discover" && <div className="kind-filters" role="group" aria-label="Package type">{(["all", "formula", "cask"] as const).map((kind) => <Button key={kind} size="sm" variant={kindFilter === kind ? "secondary" : "ghost"} aria-pressed={kindFilter === kind} onClick={() => setKindFilter(kind)}>{kind === "all" ? "All" : kind === "formula" ? "Formulae" : "Casks"}</Button>)}</div>}<span className="result-count">{visiblePackages.length} {visiblePackages.length === 1 ? "package" : "packages"}</span>{view === "Updates" && updateCount > 0 && <Button onClick={upgradeAll} disabled={busy}><Upload size={15} />Upgrade all</Button>}</div>

        <div className="workspace">
          <section className="package-pane" aria-label={`${view} packages`}>
            {busy && <div className="progress-line" role="status" aria-live="polite"><LoaderCircle size={15} className="spin" />{progress || "Loading package data…"}<Button size="sm" variant="ghost" onClick={() => void CLIService.CancelOperation()}>Cancel</Button></div>}
            {error && <div className="inline-alert" role="alert"><CircleAlert size={16} /><span>{error}</span><Button size="sm" variant="outline" onClick={() => void launch("list")}>Retry</Button></div>}
            {actionNotice && <div className="success-note" role="status"><Check size={15} />{actionNotice}</div>}
            {!busy && !error && visiblePackages.length === 0 ? <div className="empty-state"><Package size={24} /><strong>{packages.length === 0 ? "No catalog data yet" : "No matching packages"}</strong><span>{packages.length === 0 ? "Refresh Homebrew catalog to get started." : "Try another search or section."}</span>{packages.length === 0 && <Button onClick={() => void launch("refresh")}>Load catalog</Button>}</div> : <div ref={listRef} className="package-list" role="list" aria-label="Packages">
              <div ref={virtualizer.containerRef} className="package-virtual-space">
                {virtualRows.map((row) => {
                  const pkg = visiblePackages[row.index];
                  return <div key={row.key} ref={virtualizer.measureElement} data-index={row.index} className="package-virtual-row" style={{ height: row.size }}>
                    <PackageRow pkg={pkg} view={view} index={row.index + 1} selected={selected ? packageKey(selected) === packageKey(pkg) : false} onSelect={() => { setSelectedKey(packageKey(pkg)); setDetailOpen(true); }} onAction={queueAction} disabled={busy} />
                  </div>;
                })}
              </div>
            </div>}
            <footer className="list-footer"><span>{cacheText}</span><span>Buzz CLI · {status}</span></footer>
          </section>

          {selected && <PackageInspector pkg={selected} view={view} onAction={queueAction} onClose={() => setDetailOpen(false)} open={detailOpen} disabled={busy} />}
        </div>
      </main>

      <dialog ref={dialog} className="confirm-dialog" onCancel={() => setConfirm(null)} aria-labelledby="confirm-title">
        {confirm && <><div className="dialog-icon"><Trash2 size={18} /></div><h2 id="confirm-title">Remove {confirm.pkg.name}?</h2><p>This will uninstall the Homebrew {confirm.pkg.isCask ? "cask" : "formula"} from this Mac.</p><div className="dialog-actions"><Button variant="outline" onClick={() => setConfirm(null)}>Cancel</Button><Button variant="destructive" onClick={confirmAction}>Remove {confirm.pkg.name}</Button></div></>}
      </dialog>
    </div>
  );
}

function PackageRow({ pkg, view, index, selected, onSelect, onAction, disabled }: {
  pkg: PackageInfo; view: View; index: number; selected: boolean; onSelect: () => void;
  onAction: (action: Action) => void; disabled: boolean;
}) {
  return <div className={`package-row ${selected ? "selected" : ""}`} role="listitem">
    <button className="row-main" onClick={onSelect} aria-pressed={selected}>
      <span className="row-number">{String(index).padStart(2, "0")}</span><span className="package-symbol"><Package size={16} strokeWidth={1.7} /></span>
      <span className="package-summary"><strong>{pkg.name}</strong><span>{displaySource(pkg)}</span></span>
      <span className="row-version">{view === "Discover" ? pkg.version : pkg.isInstalled ? pkg.installedVersion || "Installed" : pkg.version}</span>
      {view === "Discover" && pkg.isInstalled ? <Badge variant="outline" className="installed-badge">{pkg.isOutdated ? "Installed · Update" : "Installed"}</Badge> : pkg.isOutdated && <Badge variant="outline" className="update-badge">Update</Badge>}
    </button>
    <div className="row-actions">{view !== "Discover" && pkg.isInstalled && pkg.isOutdated && pkg.installSupported && !pkg.isPinned && <Button size="sm" variant="outline" aria-label={`Upgrade ${pkg.name}`} disabled={disabled} onClick={() => onAction({ operation: "upgrade", pkg })}><Upload size={14} />Upgrade</Button>}{view === "Installed" && pkg.isInstalled && pkg.installSupported ? <Button size="icon-sm" variant="ghost" className="row-remove" aria-label={`Remove ${pkg.name}`} disabled={disabled} onClick={() => onAction({ operation: "uninstall", pkg })}><Trash2 size={15} /></Button> : view === "Discover" && !pkg.isInstalled && pkg.installSupported ? <Button size="sm" aria-label={`Install ${pkg.name}`} disabled={disabled} onClick={() => onAction({ operation: "install", pkg })}><Download size={14} />Install</Button> : null}</div>
  </div>;
}

function PackageInspector({ pkg, view, onAction, onClose, open, disabled }: {
  pkg: PackageInfo; view: View; onAction: (action: Action) => void; onClose: () => void; open: boolean; disabled: boolean;
}) {
  return <aside className={`inspector ${open ? "inspector-open" : ""}`} aria-label={`${pkg.name} details`}>
    <div className="inspector-top"><span>PACKAGE DETAILS</span><Button variant="ghost" size="sm" className="inspector-close" onClick={onClose}>Close</Button><span className="inspector-mark"><Package size={19} /></span><h2>{pkg.name}</h2><p>{pkg.description || "No description provided."}</p></div>
    <dl className="package-facts"><div><dt>Type</dt><dd>{pkg.isCask ? "Cask" : "Formula"}</dd></div><div><dt>Source</dt><dd>{pkg.tap || "homebrew/core"}</dd></div><div><dt>Installed</dt><dd className="mono">{pkg.isInstalled ? pkg.installedVersion || "Installed" : "Not installed"}</dd></div><div><dt>Available</dt><dd className="mono">{pkg.version || "—"}</dd></div></dl>
    <div className="inspector-actions">{view !== "Discover" && pkg.isInstalled && pkg.isOutdated && pkg.installSupported && !pkg.isPinned && <Button className="action-full" onClick={() => onAction({ operation: "upgrade", pkg })} disabled={disabled}><Upload size={15} />Upgrade package</Button>}{view === "Installed" && pkg.isInstalled && pkg.installSupported && <Button className="action-full destructive-action" variant="outline" onClick={() => onAction({ operation: "uninstall", pkg })} disabled={disabled}><Trash2 size={15} />Remove package</Button>}{view === "Discover" && !pkg.isInstalled && pkg.installSupported && <Button className="action-full" onClick={() => onAction({ operation: "install", pkg })} disabled={disabled}><Download size={15} />Install package</Button>}{view === "Discover" && pkg.isInstalled && <p className="unsupported-note">Already installed on this Mac.</p>}{!pkg.installSupported && <p className="unsupported-note">Buzz cannot manage this package.</p>}</div>
  </aside>;
}

function readAppearance(): Appearance {
  const stored = localStorage.getItem("buzz.appearance");
  return stored === "light" || stored === "dark" ? stored : "system";
}

function formatAge(seconds: number): string {
  if (seconds < 60) return "just refreshed";
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m old`;
  if (seconds < 86400) return `${Math.floor(seconds / 3600)}h old`;
  return `${Math.floor(seconds / 86400)}d old`;
}

export default App;
