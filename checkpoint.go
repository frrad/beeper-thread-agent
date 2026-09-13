package main

type CheckpointResult struct {
	Messages        []BeeperMessage
	NextCheckpoint  string
	CheckpointFound bool
	Bootstrapped    bool
}

func relevantIncoming(m BeeperMessage) bool {
	return !m.IsSender && !m.IsDeleted && m.Type != "REACTION"
}

// MessagesAfterCheckpoint accepts Beeper's newest-first page and returns unseen incoming messages oldest-first.
func MessagesAfterCheckpoint(items []BeeperMessage, checkpoint string) CheckpointResult {
	newestIncoming := ""
	for _, m := range items {
		if relevantIncoming(m) {
			newestIncoming = m.ID
			break
		}
	}
	if checkpoint == "" {
		return CheckpointResult{NextCheckpoint: newestIncoming, Bootstrapped: true}
	}
	index := -1
	for i, m := range items {
		if m.ID == checkpoint {
			index = i
			break
		}
	}
	if index < 0 {
		if newestIncoming == "" {
			newestIncoming = checkpoint
		}
		return CheckpointResult{NextCheckpoint: newestIncoming}
	}
	var messages []BeeperMessage
	for i := index - 1; i >= 0; i-- {
		if relevantIncoming(items[i]) {
			messages = append(messages, items[i])
		}
	}
	if newestIncoming == "" {
		newestIncoming = checkpoint
	}
	return CheckpointResult{Messages: messages, NextCheckpoint: newestIncoming, CheckpointFound: true}
}
