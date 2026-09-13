export type ModelContent = string | Array<
  | { type: "text"; text: string }
  | { type: "image_url"; image_url: { url: string } }
>;

type ChatMessage = { role: "system" | "user" | "assistant"; content: ModelContent };

export class OpenRouterClient {
  constructor(private apiKey: string, private model: string) {}

  async complete(messages: ChatMessage[]): Promise<string> {
    const response = await fetch("https://openrouter.ai/api/v1/chat/completions", {
      method: "POST",
      headers: {
        Authorization: `Bearer ${this.apiKey}`,
        "Content-Type": "application/json",
        "X-Title": "Beeper Thread Agent",
      },
      body: JSON.stringify({ model: this.model, messages }),
    });
    if (!response.ok) throw new Error(`OpenRouter ${response.status}: ${await response.text()}`);
    const data = await response.json() as { choices?: Array<{ message?: { content?: string } }> };
    const text = data.choices?.[0]?.message?.content?.trim();
    if (!text) throw new Error("OpenRouter returned an empty response");
    return text;
  }
}
