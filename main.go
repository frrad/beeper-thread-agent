package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	loadDotEnv(".env")
	chatID := flag.String("chat", "", "Beeper chatID to watch")
	poll := flag.Int("poll", envInt("POLL_INTERVAL_MS", 2000), "poll interval in milliseconds")
	modelName := flag.String("model", envString("OPENROUTER_MODEL", "anthropic/claude-sonnet-4.5"), "OpenRouter model")
	since := flag.String("since", "", "start after this Beeper message ID")
	send := flag.Bool("send", false, "send generated replies (default is dry-run)")
	dryRun := flag.Bool("dry-run", false, "print replies without sending")
	smoke := flag.Bool("smoke", false, "validate startup without network calls")
	flag.Parse()
	if *dryRun {
		*send = false
	}
	if *poll <= 0 {
		return fmt.Errorf("--poll must be greater than zero")
	}
	if *smoke {
		fmt.Printf("smoke ok: %s; no network calls made\n", map[bool]string{true: "send", false: "dry-run"}[*send])
		return nil
	}
	apiKey := os.Getenv("OPENROUTER_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("OPENROUTER_API_KEY is required (copy .env.example to .env)")
	}

	store := NewStateStore(filepath.Join(".beeper-thread-agent"))
	if err := store.Load(); err != nil {
		return fmt.Errorf("load state: %w", err)
	}
	beeper := &BeeperClient{Endpoint: envString("BEEPER_MCP_URL", "http://localhost:23373/v0/mcp"), Token: os.Getenv("BEEPER_TOKEN"), Store: store}
	ctx := context.Background()
	fmt.Println("Connecting to Beeper Desktop…")
	if err := beeper.Connect(ctx); err != nil {
		return err
	}
	defer beeper.Close()

	reader := bufio.NewReader(os.Stdin)
	if *chatID == "" {
		query, err := prompt(reader, "Search for which Beeper chat? ")
		if err != nil {
			return err
		}
		matches, err := beeper.SearchChats(ctx, query)
		if err != nil {
			return err
		}
		fmt.Println(matches)
		value, err := prompt(reader, "Paste the chatID to watch: ")
		if err != nil {
			return err
		}
		*chatID = strings.TrimSpace(value)
	}
	if *chatID == "" {
		return fmt.Errorf("a chatID is required")
	}
	if *since != "" {
		chat := store.Chat(*chatID)
		chat.Checkpoint = *since
		if err := store.PutChat(*chatID, chat); err != nil {
			return err
		}
	}

	agent := NewThreadAgent(beeper, &OpenRouterClient{APIKey: apiKey, Model: *modelName}, store, AgentOptions{
		ChatID: *chatID, PollInterval: time.Duration(*poll) * time.Millisecond,
		MaxContextMessages: envInt("MAX_CONTEXT_MESSAGES", 30), MaxImageBytes: int64(envInt("MAX_IMAGE_BYTES", 4_000_000)), Send: *send,
	})
	if *send {
		fmt.Println("SEND MODE — model replies will be sent.")
	} else {
		fmt.Println("DRY RUN — replies will only be printed.")
	}
	fmt.Println("Type instructions, then /start. While running, plain text steers the next reply. /help lists commands.")

	lines := make(chan string)
	go func() {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			lines <- scanner.Text()
		}
		close(lines)
	}()
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(signals)
	for {
		select {
		case <-signals:
			fmt.Println("\nStopping…")
			agent.Stop()
			return nil
		case line, ok := <-lines:
			if !ok {
				agent.Stop()
				return nil
			}
			command, ok := ParseCommand(line)
			if !ok {
				fmt.Println("Unknown command. Type /help.")
				continue
			}
			switch command.Name {
			case "steer":
				agent.Steer(command.Text)
			case "start":
				if err := agent.Start(command.Since); err != nil {
					fmt.Println("Start failed:", err)
				} else {
					fmt.Println(agent.Summary())
				}
			case "pause":
				agent.Pause()
				fmt.Println(agent.Summary())
			case "resume":
				agent.Resume()
				fmt.Println(agent.Summary())
			case "status":
				fmt.Println(agent.Summary())
			case "help":
				fmt.Println("/start [messageID]  /pause  /resume  /status  /stop\nAny other line is steering for the next model reply.")
			case "stop":
				agent.Stop()
				return nil
			}
		}
	}
}

func prompt(reader *bufio.Reader, text string) (string, error) {
	fmt.Print(text)
	value, err := reader.ReadString('\n')
	return strings.TrimSpace(value), err
}

func loadDotEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.Trim(strings.TrimSpace(value), "\"'")
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
}

func envString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
func envInt(key string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(key))
	if err == nil {
		return value
	}
	return fallback
}
