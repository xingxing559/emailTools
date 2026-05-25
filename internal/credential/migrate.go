package credential

// MigrateLegacyIfNeeded imports old config.json into accounts.json.
func MigrateLegacyIfNeeded() error {
	accountsMu.Lock()
	defer accountsMu.Unlock()
	return ensureMigratedLocked()
}
