#!/usr/bin/env node
import "dotenv/config";
import { createInterface } from "node:readline";
import { stdin, stdout } from "node:process";
import { BeeperClient } from "./beeper.js";
import { parseCommand } from "./commands.js";
import { ThreadAgent } from "./agent.js";
import { LocalOAuthProvider } from "./oauth.js";
import { OpenRouterClient } from "./openrouter.js";
import { StateStore } from "./state.js";

type Args = { chat?: string; poll?: number; model?: string; send: boolean; since?: string; smoke: boolean };

function argsFrom(argv: string[]): Args {
  const args: Args = { send: false, smoke: false };
  for (let i = 0; i < argv.length; i += 1) {
    const value = argv[i];
    if (value === "--") continue;
    else if (value === "--chat") args.chat = argv[++i];
    else if (value === "--poll") args.poll = Number(argv[++i]);
    else if (value === "--model") args.model = argv[++i];
    else if (value === "--since") args.since = argv[++i];
    else if (value === "--send") args.send = true;
    else if (value === "--dry-run") args.send = false;
    else if (value === "--smoke") args.smoke = true;
    else if (value === "--help" || value === "-h") {
      console.log("Usage: beeper-thread-agent [--chat ID] [--model MODEL] [--poll MS] [--since MESSAGE_ID] [--send|--dry-run] [--smoke]");
      process.exit(0);
    } else throw new Error(`Unknown option: ${value}`);
  }
  return args;
}

function question(rl: ReturnType<typeof createInterface>, prompt: string): Promise<string> {
  return new Promise((resolve) => rl.question(prompt, resolve));
}

async function main() {
  const args = argsFrom(process.argv.slice(2));
  if (args.smoke) {
    console.log(`smoke ok: ${args.send ? "send" : "dry-run"}; no network calls made`);
    return;
  }
  const apiKey = process.env.OPENROUTER_API_KEY;
  if (!apiKey) throw new Error("OPENROUTER_API_KEY is required (copy .env.example to .env)");

  const store = new StateStore();
  await store.load();
  const oauth = new LocalOAuthProvider(store);
  const beeper = new BeeperClient(
    process.env.BEEPER_MCP_URL ?? "http://localhost:23373/v0/mcp",
    oauth,
    process.env.BEEPER_TOKEN,
  );
  console.log("Connecting to Beeper Desktop…");
  await beeper.connect();
  const rl = createInterface({ input: stdin, output: stdout });

  let chatID = args.chat;
  if (!chatID) {
    const query = await question(rl, "Search for which Beeper chat? ");
    console.log(await beeper.searchChats(query));
    chatID = (await question(rl, "Paste the chatID to watch: ")).trim();
  }
  if (!chatID) throw new Error("A chatID is required");

  const pollIntervalMs = args.poll ?? Number(process.env.POLL_INTERVAL_MS ?? 2000);
  const modelName = args.model ?? process.env.OPENROUTER_MODEL ?? "anthropic/claude-sonnet-4.5";
  const agent = new ThreadAgent(beeper, new OpenRouterClient(apiKey, modelName), store, {
    chatID,
    pollIntervalMs,
    maxContextMessages: Number(process.env.MAX_CONTEXT_MESSAGES ?? 30),
    maxImageBytes: Number(process.env.MAX_IMAGE_BYTES ?? 4_000_000),
    send: args.send,
  });
  if (args.since) store.chat(chatID).checkpoint = args.since;

  console.log(`${args.send ? "SEND MODE — model replies will be sent." : "DRY RUN — replies will only be printed."}`);
  console.log("Type instructions, then /start. While running, plain text steers the next reply. /help lists commands.");

  let closing = false;
  const close = async () => {
    if (closing) return;
    closing = true;
    await agent.stop();
    await beeper.close();
    rl.close();
  };
  rl.on("line", async (line) => {
    const command = parseCommand(line);
    if (!command) { console.log("Unknown command. Type /help."); return; }
    if (command.name === "steer") agent.steer(command.text);
    else if (command.name === "start") { agent.start(command.since); console.log(agent.summary()); }
    else if (command.name === "pause") { agent.pause(); console.log(agent.summary()); }
    else if (command.name === "resume") { agent.resume(); console.log(agent.summary()); }
    else if (command.name === "status") console.log(agent.summary());
    else if (command.name === "help") console.log("/start [messageID]  /pause  /resume  /status  /stop\nAny other line is steering for the next model reply.");
    else if (command.name === "stop") await close();
  });
  rl.on("SIGINT", () => { console.log("\nStopping…"); void close(); });
}

main().catch((error) => {
  console.error((error as Error).message);
  process.exitCode = 1;
});
