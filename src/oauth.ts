import { createServer } from "node:http";
import { randomBytes } from "node:crypto";
import open from "open";
import type {
  OAuthClientMetadata,
  OAuthClientProvider,
  OAuthDiscoveryState,
  StoredOAuthClientInformation,
  StoredOAuthTokens,
} from "@modelcontextprotocol/client";
import type { StateStore } from "./state.js";

type OAuthData = {
  client?: StoredOAuthClientInformation;
  tokens?: StoredOAuthTokens;
  verifier?: string;
  discovery?: OAuthDiscoveryState;
};

export class LocalOAuthProvider implements OAuthClientProvider {
  readonly redirectUrl: URL;
  readonly clientMetadata: OAuthClientMetadata;
  private pending?: Promise<URLSearchParams>;
  private resolvePending?: (params: URLSearchParams) => void;
  private expectedState = randomBytes(24).toString("hex");

  constructor(private store: StateStore, port = 33418) {
    this.redirectUrl = new URL(`http://127.0.0.1:${port}/callback`);
    this.clientMetadata = {
      client_name: "Beeper Thread Agent",
      redirect_uris: [this.redirectUrl.toString()],
      grant_types: ["authorization_code", "refresh_token"],
      response_types: ["code"],
      token_endpoint_auth_method: "none",
    };
  }

  private data(): OAuthData {
    return (this.store.data().oauth ??= {}) as OAuthData;
  }

  state() { return this.expectedState; }
  clientInformation() { return this.data().client; }
  async saveClientInformation(value: StoredOAuthClientInformation) { this.data().client = value; await this.store.save(); }
  tokens() { return this.data().tokens; }
  async saveTokens(value: StoredOAuthTokens) { this.data().tokens = value; await this.store.save(); }
  async saveCodeVerifier(value: string) { this.data().verifier = value; await this.store.save(); }
  codeVerifier() {
    const value = this.data().verifier;
    if (!value) throw new Error("OAuth code verifier is missing");
    return value;
  }
  async saveDiscoveryState(value: OAuthDiscoveryState) { this.data().discovery = value; await this.store.save(); }
  discoveryState() { return this.data().discovery; }

  async invalidateCredentials(scope: "all" | "client" | "tokens" | "verifier" | "discovery") {
    const data = this.data();
    if (scope === "all" || scope === "client") delete data.client;
    if (scope === "all" || scope === "tokens") delete data.tokens;
    if (scope === "all" || scope === "verifier") delete data.verifier;
    if (scope === "all" || scope === "discovery") delete data.discovery;
    await this.store.save();
  }

  async redirectToAuthorization(url: URL) {
    this.pending = new Promise((resolve) => { this.resolvePending = resolve; });
    const server = createServer((request, response) => {
      const callback = new URL(request.url ?? "/", this.redirectUrl);
      if (callback.pathname !== this.redirectUrl.pathname) {
        response.writeHead(404).end();
        return;
      }
      response.writeHead(200, { "content-type": "text/plain; charset=utf-8" });
      response.end("Beeper authorized. You can close this tab and return to the terminal.");
      server.close();
      this.resolvePending?.(callback.searchParams);
    });
    await new Promise<void>((resolve, reject) => {
      server.once("error", reject);
      server.listen(Number(this.redirectUrl.port), this.redirectUrl.hostname, resolve);
    });
    console.log(`Authorize Beeper in your browser: ${url}`);
    await open(url.toString());
  }

  async waitForCallback(): Promise<URLSearchParams> {
    if (!this.pending) throw new Error("OAuth authorization was not started");
    const params = await this.pending;
    const returnedState = params.get("state");
    if (!returnedState || returnedState !== this.expectedState) throw new Error("OAuth state mismatch");
    return params;
  }
}
