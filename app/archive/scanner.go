package archive

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/iyear/tdl/core/util/fsutil"
	"github.com/iyear/tdl/pkg/consts"
	"github.com/iyear/tdl/pkg/filterMap"
)

type FileInfo struct {
	Path     string // Full path
	Folder   string // Folder path (relative to strip prefix)
	Thumb    string // Thumbnail path if exists
	Size     int64
}

type FolderTree struct {
	Folders    []string   // List of all folder relative paths
	Files      []FileInfo // List of all files with their folder info
	TotalFiles int
}

func scanFolderTree(rootPath string, includes, excludes []string) (*FolderTree, error) {
	tree := &FolderTree{
		Folders: make([]string, 0),
		Files:   make([]FileInfo, 0),
	}

	includesMap := filterMap.New(includes, fsutil.AddPrefixDot)
	excludesMap := filterMap.New(excludes, fsutil.AddPrefixDot)
	excludesMap[consts.UploadThumbExt] = struct{}{} // ignore thumbnail files

	foldersSet := make(map[string]struct{})

	err := filepath.WalkDir(rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Record folders
		if d.IsDir() {
			if path != rootPath {
				foldersSet[path] = struct{}{}
			}
			return nil
		}

		// Process files
		ext := filepath.Ext(path)

		// Apply filters
		if _, ok := includesMap[ext]; len(includesMap) > 0 && !ok {
			return nil
		}
		if _, ok := excludesMap[ext]; len(excludesMap) > 0 && ok {
			return nil
		}

		stat, err := os.Stat(path)
		if err != nil {
			return err
		}

		file := FileInfo{
			Path:   path,
			Folder: filepath.Dir(path),
			Size:   stat.Size(),
		}

		// Check for thumbnail
		thumbPath := path[:len(path)-len(ext)] + consts.UploadThumbExt
		if fsutil.PathExists(thumbPath) {
			file.Thumb = thumbPath
		}

		// Ensure parent folder is recorded
		foldersSet[filepath.Dir(path)] = struct{}{}

		tree.Files = append(tree.Files, file)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert folder set to slice
	for folder := range foldersSet {
		tree.Folders = append(tree.Folders, folder)
	}

	tree.TotalFiles = len(tree.Files)
	return tree, nil
}

type UploadItem struct {
	File     FileInfo
	TopicID  int
	AsPhoto  bool
}

func prepareUploadItems(tree *FolderTree, topicMgr *topicManager, asPhoto bool) ([]UploadItem, error) {
	items := make([]UploadItem, 0, len(tree.Files))

	for _, file := range tree.Files {
		topicID, err := topicMgr.getTopicID(file.Folder)
		if err != nil {
			return nil, err
		}

		items = append(items, UploadItem{
			File:    file,
			TopicID: topicID,
			AsPhoto: asPhoto,
		})
	}

	return items, nil
}