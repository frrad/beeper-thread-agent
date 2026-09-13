export type Command =
  | { name: "start"; since?: string }
  | { name: "pause" | "resume" | "status" | "stop" | "help" }
  | { name: "steer"; text: string };

export function parseCommand(input: string): Command | undefined {
  const text = input.trim();
  if (!text) return undefined;
  if (!text.startsWith("/")) return { name: "steer", text };

  const [rawName, ...args] = text.slice(1).split(/\s+/);
  const name = rawName.toLowerCase();
  if (name === "start") return { name, since: args[0] };
  if (["pause", "resume", "status", "stop", "help"].includes(name)) {
    return { name: name as "pause" | "resume" | "status" | "stop" | "help" };
  }
  return undefined;
}
