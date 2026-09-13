import type { BeeperMessage } from "./types.js";

export type CheckpointResult = {
  messages: BeeperMessage[];
  nextCheckpoint?: string;
  checkpointFound: boolean;
  bootstrapped: boolean;
};

function isRelevantIncoming(message: BeeperMessage): boolean {
  return !message.isSender && !message.isDeleted && message.type !== "REACTION";
}

/** Accepts Beeper's newest-first page and returns unseen incoming messages oldest-first. */
export function messagesAfterCheckpoint(
  newestFirst: BeeperMessage[],
  checkpoint?: string,
): CheckpointResult {
  const incoming = newestFirst.filter(isRelevantIncoming);
  const newestIncoming = incoming[0]?.id;

  if (!checkpoint) {
    return {
      messages: [],
      nextCheckpoint: newestIncoming,
      checkpointFound: false,
      bootstrapped: true,
    };
  }

  const checkpointIndex = newestFirst.findIndex((message) => message.id === checkpoint);
  if (checkpointIndex === -1) {
    // A stale checkpoint must never cause a reply storm over an entire page of history.
    return {
      messages: [],
      nextCheckpoint: newestIncoming ?? checkpoint,
      checkpointFound: false,
      bootstrapped: false,
    };
  }

  const messages = newestFirst.slice(0, checkpointIndex).filter(isRelevantIncoming).reverse();
  return {
    messages,
    nextCheckpoint: newestIncoming ?? checkpoint,
    checkpointFound: true,
    bootstrapped: false,
  };
}
