import { useEffect, useState } from "react";
import { Events } from "@wailsio/runtime";
import {
  ArrowUpRight,
  Check,
  CircleAlert,
  Command,
  LoaderCircle,
  Package,
  RefreshCw,
  Sparkles,
  SquareTerminal,
} from "lucide-react";
import { CLIService } from "../bindings/github.com/commandotj/ax-buzz/apps/desktop";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";

type ConnectionState = "idle" | "checking" | "connected" | "error";

type CLIStreamEvent = {
  protocolVersion: number;
  event: "progress" | "result" | "error";
  data?: {
    message?: string;
    product?: string;
    version?: string;
  };
  error?: string;
};

type WailsEvent<T> = { data: T };

function App() {
  const [connection, setConnection] = useState<ConnectionState>("idle");
  const [cliVersion, setCliVersion] = useState("");
  const [message, setMessage] = useState("Check Buzz CLI to connect this workspace.");

  useEffect(() => {
    return Events.On("buzz:cli-event", (event: WailsEvent<CLIStreamEvent>) => {
      const payload = event.data;
      if (payload.protocolVersion !== 1) {
        setConnection("error");
        setMessage(`Unsupported CLI protocol version ${payload.protocolVersion}.`);
        return;
      }
      if (payload.event === "progress") {
        setConnection("checking");
        setMessage(payload.data?.message ?? "Checking Buzz CLI…");
        return;
      }
      if (payload.event === "result") {
        setConnection("connected");
        setCliVersion(payload.data?.version ?? "");
        setMessage("Buzz CLI is ready on this Mac.");
        return;
      }
      setConnection("error");
      setMessage(payload.error ?? "Buzz CLI check failed.");
    });
  }, []);

  const checkConnection = async () => {
    setConnection("checking");
    setMessage("Starting Buzz CLI…");
    try {
      await CLIService.StartHandshake();
    } catch (error) {
      setConnection("error");
      setMessage(error instanceof Error ? error.message : String(error));
    }
  };

  const cancelCheck = async () => {
    try {
      await CLIService.CancelHandshake();
    } catch (error) {
      setConnection("error");
      setMessage(error instanceof Error ? error.message : String(error));
    }
  };

  const statusIcon =
    connection === "checking" ? (
      <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
    ) : connection === "connected" ? (
      <Check className="size-4" aria-hidden="true" />
    ) : connection === "error" ? (
      <CircleAlert className="size-4" aria-hidden="true" />
    ) : (
      <SquareTerminal className="size-4" aria-hidden="true" />
    );

  return (
    <div className="app-shell">
      <aside className="sidebar">
        <div className="brand-lockup">
          <div className="brand-mark" aria-hidden="true">
            <Package className="size-5" />
            <span />
          </div>
          <div>
            <p className="brand-name">Buzz</p>
            <p className="brand-caption">Software, in order.</p>
          </div>
        </div>

        <Separator className="sidebar-rule" />

        <div className="nav-section">
          <p className="nav-heading">Workspace</p>
          <Button className="nav-item nav-current" variant="secondary" aria-current="page">
            <Command className="size-4" aria-hidden="true" />
            Overview
          </Button>
        </div>

        <div className="sidebar-note">
          <div className="note-icon" aria-hidden="true">
            <Sparkles className="size-4" />
          </div>
          <p>Made for the software you actually use.</p>
        </div>

        <div className="sidebar-footer">
          <span className="status-dot" />
          <span>Runs on this Mac</span>
        </div>
      </aside>

      <main className="main-content">
        <header className="topbar">
          <div>
            <p className="breadcrumb">Buzz / Overview</p>
            <h1>Overview</h1>
          </div>
          <Badge variant="outline" className="platform-badge">macOS</Badge>
        </header>

        <section className="welcome-panel" aria-labelledby="welcome-title">
          <div className="welcome-copy">
            <p className="welcome-kicker">Your Mac software shelf</p>
            <h2 id="welcome-title">A clearer view of what’s on your Mac.</h2>
            <p className="welcome-description">
              Find packages, see what’s installed, and keep software up to date from one calm workspace.
            </p>
            <div className="welcome-actions">
              <Button onClick={checkConnection} disabled={connection === "checking"}>
                {connection === "checking" ? (
                  <LoaderCircle className="size-4 animate-spin" aria-hidden="true" />
                ) : (
                  <SquareTerminal className="size-4" aria-hidden="true" />
                )}
                {connection === "connected" ? "Check again" : "Connect Buzz CLI"}
              </Button>
              {connection === "checking" && (
                <Button variant="ghost" onClick={cancelCheck}>Cancel</Button>
              )}
            </div>
          </div>
          <div className="shelf-art" aria-hidden="true">
            <div className="shelf-orbit orbit-one" />
            <div className="shelf-orbit orbit-two" />
            <div className="package-tile tile-one"><Package /></div>
            <div className="package-tile tile-two"><Command /></div>
            <div className="package-tile tile-three"><Sparkles /></div>
            <div className="shelf-base" />
          </div>
        </section>

        <section className="connection-section" aria-labelledby="connection-title">
          <div className="section-heading">
            <div>
              <p className="section-kicker">Local connection</p>
              <h2 id="connection-title">Buzz CLI</h2>
            </div>
            <Badge className={`connection-badge state-${connection}`} variant="outline">
              {statusIcon}
              {connection === "connected" ? "Connected" : connection === "checking" ? "Checking" : connection === "error" ? "Needs attention" : "Not checked"}
            </Badge>
          </div>

          <Card className="connection-card">
            <CardHeader className="connection-card-header">
              <div className="cli-glyph" aria-hidden="true"><SquareTerminal className="size-5" /></div>
              <div className="connection-card-copy">
                <CardTitle>{cliVersion ? `Buzz CLI ${cliVersion}` : "Connect your local CLI"}</CardTitle>
                <CardDescription>{message}</CardDescription>
              </div>
              <Button variant="outline" onClick={connection === "checking" ? cancelCheck : checkConnection}>
                {connection === "checking" ? "Cancel check" : connection === "connected" ? "Refresh" : "Check connection"}
                {connection !== "checking" && <RefreshCw className="size-4" aria-hidden="true" />}
              </Button>
            </CardHeader>
            <CardContent className="connection-details">
              <span>{connection === "connected" ? "Protocol ready" : "Local process"}</span>
              <span className="detail-value">{connection === "connected" ? "Versioned JSON stream" : "Uses bundled Buzz CLI"}</span>
            </CardContent>
          </Card>
        </section>

        <section className="empty-workspace" aria-label="Software workspace">
          <div className="empty-icon"><Package className="size-5" aria-hidden="true" /></div>
          <div>
            <h2>Your software workspace starts here.</h2>
            <p>Connect Buzz CLI to bring package information into this window.</p>
          </div>
          <ArrowUpRight className="empty-arrow" aria-hidden="true" />
        </section>
      </main>
    </div>
  );
}

export default App;
