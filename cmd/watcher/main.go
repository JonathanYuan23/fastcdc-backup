package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"time"

	"fastcdc-backup/pkg/node"

	"github.com/radovskyb/watcher"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func walk(rootOld, rootNew *node.Node, timeOld int64, newFiles, modifiedFiles, deletedFiles []*node.Node) {
	pathToNodeMap := node.GetPathToNodeMappings(rootOld.Children)

	sameDirs := []struct{ oldDir, newDir *node.Node }{}

	files := map[string]bool{}

	for _, oldFile := range rootOld.Children {
		files[oldFile.Path] = false
	}
	for _, newFile := range rootNew.Children {
		if _, ok := files[newFile.Path]; ok {
			if newFile.IsDir { // directory path persists between versions - continue walking
				sameDirs = append(sameDirs, struct{ oldDir, newDir *node.Node }{oldDir: pathToNodeMap[newFile.Path], newDir: newFile})
			} else {
				fileInfo, _ := os.Stat(newFile.Path)
				modifiedTime := fileInfo.ModTime().Unix()

				if modifiedTime > timeOld { // file has been modified since timeOld - mark for chunking
					modifiedFiles = append(modifiedFiles, newFile)
				}
			}

			files[newFile.Path] = true
		} else { // introduction of a new file - mark for chunking
			newFiles = append(newFiles, newFile)
		}
	}

	for path, inNew := range files {
		if !inNew { // file in previous version was not encountered - mark for deletion
			deletedFiles = append(deletedFiles, pathToNodeMap[path])
		}
	}

	for _, dirs := range sameDirs {
		walk(dirs.oldDir, dirs.newDir, timeOld, newFiles, modifiedFiles, deletedFiles)
	}
}

func initHierarchy() error {
	_, err := os.Stat("./hierarchy.json")

	if os.IsNotExist(err) {
		rootNode, err := node.LoadHierarchy("./data")
		if err != nil {
			return err
		}

		tree := node.Tree{
			Root:         rootNode,
			TimeAccessed: time.Now().Unix(),
		}

		hierarchyJSON, err := json.Marshal(tree)
		if err != nil {
			return err
		}

		fo, err := os.Create("./hiearchy.json")
		if err != nil {
			return err
		}
		defer fo.Close()

		_, err = fo.Write(hierarchyJSON)
		if err != nil {
			return err
		}
	} else { // compare tracked hierarchy with previous version
		fi, err := os.Open("./hierarchy.json")
		if err != nil {
			return err
		}
		defer fi.Close()

		bytes, err := io.ReadAll(fi)
		if err != nil {
			return err
		}

		treeOld := node.Tree{}
		json.Unmarshal(bytes, &treeOld)

		rootOld, timeOld := treeOld.Root, treeOld.TimeAccessed

		rootNew, err := node.LoadHierarchy("./data")
		if err != nil {
			return err
		}

		newFiles, modifiedFiles, deletedFiles := []*node.Node{}, []*node.Node{}, []*node.Node{}
		walk(rootOld, rootNew, timeOld, newFiles, modifiedFiles, deletedFiles)
	}

	return nil
}

func main() {
	_ = initHierarchy()
	w := watcher.New()

	rules := []string{}
	directories := []string{"./data"}

	// set regular expression rules
	for _, rule := range rules {
		r := regexp.MustCompile(rule)
		w.AddFilterHook(watcher.RegexFilterHook(r, false))
	}

	// set watched directories
	for _, dir := range directories {
		err := w.AddRecursive(dir)
		check(err)
	}

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event)

			case err := <-w.Error:
				fmt.Println(err)

			case <-w.Closed:
				fmt.Println("Watcher closed")
				return
			}
		}
	}()

	err := w.Start(time.Second)
	check(err)
}
