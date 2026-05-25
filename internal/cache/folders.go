package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// FolderEntry is a cached mailbox folder.
type FolderEntry struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Delimiter   string `json:"delimiter"`
}

type foldersCacheFile struct {
	Folders    []FolderEntry `json:"folders"`
	LastFolder string        `json:"lastFolder"`
}

func foldersPath(accountID string) (string, error) {
	root, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, "folders", accountID+".json"), nil
}

// SaveFolders persists folder list and last selected folder for an account.
func SaveFolders(accountID string, folders []FolderEntry, lastFolder string) error {
	if accountID == "" {
		return nil
	}
	path, err := foldersPath(accountID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(foldersCacheFile{
		Folders:    folders,
		LastFolder: lastFolder,
	}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

// LoadFolders returns cached folders and last folder name (may be empty if no cache).
func LoadFolders(accountID string) ([]FolderEntry, string, error) {
	path, err := foldersPath(accountID)
	if err != nil {
		return nil, "", err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", err
	}
	var fc foldersCacheFile
	if err := json.Unmarshal(data, &fc); err != nil {
		return nil, "", err
	}
	return fc.Folders, fc.LastFolder, nil
}
