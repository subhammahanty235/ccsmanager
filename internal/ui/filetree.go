package ui

import (
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"ccmanager/internal/session"
)

type FileNode struct {
	Name      string
	Path      string
	IsDir     bool
	Operation string
	Count     int
	Children  []*FileNode
}

func BuildFileTree(changes []session.FileChange) *FileNode {
	root := &FileNode{
		Name:     ".",
		IsDir:    true,
		Children: make([]*FileNode, 0),
	}

	for _, change := range changes {
		addToTree(root, change)
	}

	sortTree(root)
	return root
}

func addToTree(root *FileNode, change session.FileChange) {
	path := change.Path
	if strings.HasPrefix(path, "/") {
		path = makeRelative(path)
	}

	parts := strings.Split(path, "/")
	current := root

	for i, part := range parts {
		if part == "" || part == "." {
			continue
		}

		isLast := i == len(parts)-1

		var child *FileNode
		for _, c := range current.Children {
			if c.Name == part {
				child = c
				break
			}
		}

		if child == nil {
			child = &FileNode{
				Name:     part,
				Path:     strings.Join(parts[:i+1], "/"),
				IsDir:    !isLast,
				Children: make([]*FileNode, 0),
			}
			current.Children = append(current.Children, child)
		}

		if isLast {
			child.Operation = change.Operation
			child.Count = change.Count
			child.IsDir = false
		}

		current = child
	}
}

func makeRelative(path string) string {
	prefixes := []string{"/Users/", "/home/", "/var/"}

	for _, prefix := range prefixes {
		if strings.HasPrefix(path, prefix) {
			remaining := path[len(prefix):]
			parts := strings.SplitN(remaining, "/", 3)
			if len(parts) >= 3 {
				return parts[2]
			}
			return remaining
		}
	}

	return strings.TrimPrefix(path, "/")
}

func sortTree(node *FileNode) {
	sort.Slice(node.Children, func(i, j int) bool {
		if node.Children[i].IsDir != node.Children[j].IsDir {
			return node.Children[i].IsDir
		}
		return node.Children[i].Name < node.Children[j].Name
	})

	for _, child := range node.Children {
		sortTree(child)
	}
}

func RenderFileTree(root *FileNode, maxLines int) []string {
	var lines []string
	renderNode(root, "", true, &lines, maxLines)
	return lines
}

func renderNode(node *FileNode, prefix string, isLast bool, lines *[]string, maxLines int) {
	if maxLines > 0 && len(*lines) >= maxLines {
		return
	}

	if node.Name != "." {
		var line strings.Builder

		if isLast {
			line.WriteString(prefix + "└── ")
		} else {
			line.WriteString(prefix + "├── ")
		}

		line.WriteString(node.Name)

		if !node.IsDir && node.Operation != "" {
			line.WriteString(" ")
			switch node.Operation {
			case "created":
				line.WriteString("[created]")
			case "edited":
				if node.Count > 1 {
					line.WriteString("[edited x")
					line.WriteString(strconv.Itoa(node.Count))
					line.WriteString("]")
				} else {
					line.WriteString("[edited]")
				}
			case "deleted":
				line.WriteString("[deleted]")
			}
		}

		*lines = append(*lines, line.String())
	}

	var newPrefix string
	if node.Name == "." {
		newPrefix = ""
	} else if isLast {
		newPrefix = prefix + "    "
	} else {
		newPrefix = prefix + "|   "
	}

	for i, child := range node.Children {
		if maxLines > 0 && len(*lines) >= maxLines {
			remaining := len(node.Children) - i
			if remaining > 0 {
				*lines = append(*lines, newPrefix+"... and "+strconv.Itoa(remaining)+" more")
			}
			return
		}
		isLastChild := i == len(node.Children)-1
		renderNode(child, newPrefix, isLastChild, lines, maxLines)
	}
}

func CountFiles(root *FileNode) int {
	count := 0
	countFilesRecursive(root, &count)
	return count
}

func countFilesRecursive(node *FileNode, count *int) {
	if !node.IsDir {
		*count++
	}
	for _, child := range node.Children {
		countFilesRecursive(child, count)
	}
}

func GetFileExtension(path string) string {
	return strings.TrimPrefix(filepath.Ext(path), ".")
}
