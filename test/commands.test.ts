import { describe, expect, it } from "vitest";
import { parseCommand } from "../src/commands.js";

describe("parseCommand", () => {
  it("parses lifecycle commands", () => {
    expect(parseCommand("/start abc")).toEqual({ name: "start", since: "abc" });
    expect(parseCommand(" /PAUSE ")).toEqual({ name: "pause" });
    expect(parseCommand("/stop")).toEqual({ name: "stop" });
  });

  it("treats plain text as steering", () => {
    expect(parseCommand("keep it very short")).toEqual({ name: "steer", text: "keep it very short" });
  });

  it("rejects empty and unknown commands", () => {
    expect(parseCommand(" ")).toBeUndefined();
    expect(parseCommand("/wat")).toBeUndefined();
  });
});
