import { describe, expect, it } from "vitest";
import { messagesAfterCheckpoint } from "../src/checkpoint.js";
import type { BeeperMessage } from "../src/types.js";

const message = (id: string, isSender = false): BeeperMessage => ({ id, chatID: "chat", isSender, type: "TEXT" });

describe("messagesAfterCheckpoint", () => {
  it("bootstraps at the latest incoming message without replaying history", () => {
    expect(messagesAfterCheckpoint([message("3"), message("2", true), message("1")])).toMatchObject({
      messages: [], nextCheckpoint: "3", bootstrapped: true,
    });
  });

  it("returns only new incoming messages, oldest first, and ignores our messages", () => {
    const result = messagesAfterCheckpoint([message("4"), message("3", true), message("2"), message("1")], "1");
    expect(result.messages.map((item) => item.id)).toEqual(["2", "4"]);
    expect(result.nextCheckpoint).toBe("4");
  });

  it("does not replay a page when a stale checkpoint is missing", () => {
    expect(messagesAfterCheckpoint([message("3"), message("2")], "old")).toMatchObject({
      messages: [], nextCheckpoint: "3", checkpointFound: false, bootstrapped: false,
    });
  });
});
