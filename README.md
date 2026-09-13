# Beeper Thread Agent

A small Go program that watches one Beeper thread, asks an OpenRouter model what to say, and optionally sends the reply through Beeper Desktop's MCP server.

It is deliberately **dry-run by default**. Add `--send` only when you want model replies to go out automatically.

## What it does

- Connects to the local [Beeper Desktop MCP server](https://developers.beeper.com/desktop-api/mcp/) using browser OAuth or an optional bearer token.
- Lets you search for and choose a chat, or accept a known `chatID` on the command line.
- Starts at the newest message by default, so it never answers an old backlog accidentally.
- Polls every two seconds, ignores your own outbound messages, deduplicates incoming messages, and checkpoints progress on disk.
- Includes voice-note transcriptions and attachment metadata. Small locally cached images are passed to vision-capable models when available.
- Keeps stdin open: type a sentence at any time to steer the next response.
- Supports `/pause`, `/resume`, `/status`, and `/stop`, and backs off after transient errors.

This is a local hobby tool, not a hosted service. Keep Beeper Desktop and the terminal session running.

## Requirements

- Go 1.24 or later
- Beeper Desktop running, with its local Desktop API enabled
- An [OpenRouter](https://openrouter.ai/docs/quickstart) API key

## Install

```sh
git clone <repository-url>
cd beeper-thread-agent
cp .env.example .env
```

Put your OpenRouter key in `.env`:

```dotenv
OPENROUTER_API_KEY=your_key_here
```

The first live run opens Beeper's authorization page in your browser. OAuth tokens and per-chat checkpoints are kept in `.beeper-thread-agent/state.json`, which is gitignored and created with user-only permissions. If your Beeper setup gives you a manual token instead, set `BEEPER_TOKEN` in `.env`.

## Run it

Start safely in dry-run mode:

```sh
go run .
```

Search for a chat at the prompt, paste the displayed `chatID`, type instructions, and start:

```text
Help troubleshoot this issue. Give one simple test at a time and wait for the result.
/start
```

New messages are polled and draft replies are printed, but nothing is sent. While it runs, type any plain line to steer its **next** reply:

```text
Ask for a screenshot this time, and keep it to one sentence.
```

When the drafts look right, restart in send mode:

```sh
go run . --send
```

You can skip chat search when you already know the ID:

```sh
go run . --chat 'your-chat-id' --send
```

To deliberately begin after a particular message rather than at the latest one:

```sh
go run . --chat 'your-chat-id' --since 'message-id'
```

Or pass the ID to `/start message-id` during the session.

For a reusable binary:

```sh
go build -o beeper-thread-agent .
./beeper-thread-agent --chat 'your-chat-id'
```

## Commands

| Command | Effect |
| --- | --- |
| `/start [messageID]` | Begin polling, optionally after one known message |
| `/pause` | Keep the process open but stop polling |
| `/resume` | Resume polling |
| `/status` | Show mode, chat, checkpoint, and queued steering |
| `/stop` | Save state, disconnect, and exit |
| `/help` | Show commands |
| any other line | Add steering for the next generated reply |

`Ctrl-C` also shuts down cleanly.

## Configuration

See `.env.example`. Command-line `--model` and `--poll` override their environment counterparts.

The default model is `anthropic/claude-sonnet-4.5`; choose any OpenRouter model that supports the message content you expect. Images are omitted when the cached file is unavailable, too large, or the configured model cannot use them. The prompt still receives their filename and type metadata.

## Safety and limitations

- `--send` is intentionally explicit. Once enabled, replies are sent without per-message approval until you pause or stop it.
- The first run checkpoints the newest incoming message and waits for the next one. This prevents accidental replies to history.
- A checkpoint that has fallen outside Beeper's returned page is advanced without replying. This trades a possible missed message for avoiding a history-wide reply storm.
- Steering typed while a model request is already in flight applies to the following reply.
- No prompt can make autonomous messaging risk-free. Supervise it, keep instructions narrow, and use dry-run first.
- Local state can contain message excerpts and OAuth tokens. It is permissions-restricted and gitignored, but you should still treat it as sensitive.

## Development

```sh
go test ./...
go test -race ./...
go vet ./...
go build ./...
go run . --smoke
```

The smoke command makes no network calls and sends nothing.

## License

MIT
