package main

import "testing"

func msg(id string, sender bool) BeeperMessage {
	return BeeperMessage{ID: id, ChatID: "chat", IsSender: sender, Type: "TEXT"}
}

func TestCheckpointBootstrapsWithoutHistory(t *testing.T) {
	result := MessagesAfterCheckpoint([]BeeperMessage{msg("3", false), msg("2", true), msg("1", false)}, "")
	if !result.Bootstrapped || result.NextCheckpoint != "3" || len(result.Messages) != 0 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestCheckpointReturnsIncomingOldestFirst(t *testing.T) {
	result := MessagesAfterCheckpoint([]BeeperMessage{msg("4", false), msg("3", true), msg("2", false), msg("1", false)}, "1")
	if len(result.Messages) != 2 || result.Messages[0].ID != "2" || result.Messages[1].ID != "4" || result.NextCheckpoint != "4" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestCheckpointDoesNotReplayWhenMissing(t *testing.T) {
	result := MessagesAfterCheckpoint([]BeeperMessage{msg("3", false), msg("2", false)}, "old")
	if result.CheckpointFound || result.Bootstrapped || len(result.Messages) != 0 || result.NextCheckpoint != "3" {
		t.Fatalf("unexpected result: %#v", result)
	}
}
