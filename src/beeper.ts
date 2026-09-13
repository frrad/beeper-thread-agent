import {
  Client,
  StreamableHTTPClientTransport,
  UnauthorizedError,
  type AuthProvider,
} from "@modelcontextprotocol/client";
import type { LocalOAuthProvider } from "./oauth.js";
import type { MessagePage } from "./types.js";

export class BeeperClient {
  private client = new Client({ name: "beeper-thread-agent", version: "0.1.0" });
  private transport?: StreamableHTTPClientTransport;

  constructor(
    private url: string,
    private oauth: LocalOAuthProvider,
    private token?: string,
  ) {}

  async connect(): Promise<void> {
    const authProvider: AuthProvider | LocalOAuthProvider = this.token
      ? { token: async () => this.token }
      : this.oauth;
    this.transport = new StreamableHTTPClientTransport(new URL(this.url), { authProvider });
    try {
      await this.client.connect(this.transport);
    } catch (error) {
      if (!(error instanceof UnauthorizedError) || this.token) throw error;
      const callback = await this.oauth.waitForCallback();
      await this.transport.finishAuth(callback);
      this.client = new Client({ name: "beeper-thread-agent", version: "0.1.0" });
      this.transport = new StreamableHTTPClientTransport(new URL(this.url), { authProvider });
      await this.client.connect(this.transport);
    }
  }

  async close() { await this.client.close(); }

  private async call(name: string, args: Record<string, unknown>): Promise<string> {
    const result = await this.client.callTool({ name, arguments: args });
    const text = result.content.find((item) => item.type === "text");
    if (!text || text.type !== "text") throw new Error(`${name} returned no text result`);
    if (result.isError) throw new Error(text.text);
    return text.text;
  }

  async listMessages(chatID: string): Promise<MessagePage> {
    return JSON.parse(await this.call("list_messages", { chatID })) as MessagePage;
  }

  async searchChats(query: string): Promise<string> {
    return this.call("search", { query });
  }

  async sendMessage(chatID: string, text: string): Promise<string> {
    return this.call("send_message", { chatID, text });
  }
}
