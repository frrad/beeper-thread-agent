export type Attachment = {
  type?: string;
  mimeType?: string;
  fileName?: string;
  fileSize?: number;
  srcURL?: string;
  transcription?: { transcription?: string; engine?: string };
};

export type BeeperMessage = {
  id: string;
  chatID: string;
  senderID?: string;
  senderName?: string;
  timestamp?: string;
  sortKey?: string;
  type?: string;
  text?: string;
  isSender?: boolean;
  isDeleted?: boolean;
  attachments?: Attachment[];
};

export type MessagePage = {
  items: BeeperMessage[];
  hasMore?: boolean;
  oldestCursor?: string;
  newestCursor?: string;
};

export type ConversationTurn = {
  role: "user" | "assistant";
  content: string;
};

export type PersistedState = {
  oauth?: Record<string, unknown>;
  chats: Record<string, { checkpoint?: string; history: ConversationTurn[] }>;
};
