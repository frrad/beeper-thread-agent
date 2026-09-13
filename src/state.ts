import { mkdir, readFile, writeFile } from "node:fs/promises";
import path from "node:path";
import type { PersistedState } from "./types.js";

export class StateStore {
  readonly directory: string;
  readonly file: string;
  private state: PersistedState = { chats: {} };

  constructor(directory = path.resolve(".beeper-thread-agent")) {
    this.directory = directory;
    this.file = path.join(directory, "state.json");
  }

  async load(): Promise<void> {
    try {
      this.state = JSON.parse(await readFile(this.file, "utf8")) as PersistedState;
      this.state.chats ??= {};
    } catch (error) {
      if ((error as NodeJS.ErrnoException).code !== "ENOENT") throw error;
    }
  }

  data(): PersistedState {
    return this.state;
  }

  chat(chatID: string) {
    return (this.state.chats[chatID] ??= { history: [] });
  }

  async save(): Promise<void> {
    await mkdir(this.directory, { recursive: true, mode: 0o700 });
    await writeFile(this.file, `${JSON.stringify(this.state, null, 2)}\n`, { mode: 0o600 });
  }
}
