package main

import (
	"bufio"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"fastcdc-backup/pkg/fastcdc"
	"fastcdc-backup/pkg/node"
	"fastcdc-backup/pkg/sqlite-chunks"

	"github.com/radovskyb/watcher"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func getFileChunks(path string) ([]fastcdc.Chunk, error) {
	fi, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fi.Close()

	// set options
	opt := fastcdc.Options{}
	opt.SetDefaults()

	// buffered reader is more efficient for many small reads
	br := bufio.NewReaderSize(fi, 4096)
	chunker := fastcdc.NewChunker(br, opt)

	chunks := []fastcdc.Chunk{}

	for {
		chunk, err := chunker.NextChunk()

		if err == io.EOF {
			fmt.Printf("Finished chunking %s\n", path)
			break
		} else if err != nil {
			return nil, err
		}

		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

func writeChunk(chunk *fastcdc.Chunk) error {
	filename := fmt.Sprintf("%x", chunk.Checksum())
	path := filepath.Join("./chunks", filename)

	fo, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fo.Close()

	n, err := fo.Write(chunk.Data)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote %d bytes to %s\n", n, filename)

	return nil
}

func writeChunklist(fnode *node.FNode) error {
	chunklistJSON, err := json.Marshal(fnode)
	if err != nil {
		return err
	}

	path := strings.Replace(fnode.Path, "data", "chunklists", 1)

	fo, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fo.Close()

	_, err = fo.Write(chunklistJSON)
	if err != nil {
		return err
	}
	fmt.Printf("Wrote chunklist to %s\n", path)

	return nil
}

func processNewChunk(db *sql.DB, checksum string, newChunk *fastcdc.Chunk) error {
	if sqlitechunks.Exists(db, checksum) { // chunk entry exists in internal db
		// increase instance count of chunk entry
		sqlitechunks.IncreaseCount(db, checksum)

	} else {
		// write to chunk directory and send it to server for addition to r2 bucket
		writeChunk(newChunk)

		// TODO: send chunk to server for addition to R2
	}

	return nil
}

func processNewFile(db *sql.DB, file *node.Node) error {
	if file.IsDir {
		for _, child := range file.Children {
			err := processNewFile(db, child)
			if err != nil {
				return err
			}
		}
	} else {
		chunks, err := getFileChunks(file.Path)
		if err != nil {
			return err
		}

		size := int64(0)
		checksums := []string{}

		for _, chunk := range chunks {
			checksum := fmt.Sprintf("%x", chunk.Checksum())
			checksums = append(checksums, checksum)

			size += int64(chunk.Size)

			err := processNewChunk(db, checksum, &chunk)
			if err != nil {
				return err
			}
		}

		// write chunklist to node.Path, replacing "data" with "chunklists" as the path base
		path := file.Path

		fnode := &node.FNode{
			Path:   path,
			Size:   size,
			Chunks: checksums,
		}
		writeChunklist(fnode)
	}

	return nil
}

func processModifiedFile(db *sql.DB, file *node.Node) error {
	if file.IsDir {
		for _, child := range file.Children {
			err := processModifiedFile(db, child)
			if err != nil {
				return err
			}
		}
	} else {
		path := strings.Replace(file.Path, "data", "chunklists", 1)

		fi, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fi.Close()

		bytes, err := io.ReadAll(fi)
		if err != nil {
			return err
		}

		oldChunkList := node.FNode{}
		json.Unmarshal(bytes, &oldChunkList)

		newChunkList, err := getFileChunks(file.Path)
		if err != nil {
			return err
		}

		size := int64(0)
		checksums := []string{}

		newChunks := []fastcdc.Chunk{}

		chunks := map[string]bool{}

		for _, chunk := range oldChunkList.Chunks {
			chunks[chunk] = false
		}

		for _, chunk := range newChunkList {
			checksum := fmt.Sprintf("%x", chunk.Checksum())
			checksums = append(checksums, checksum)

			size += int64(chunk.Size)

			if _, ok := chunks[checksum]; ok { // same chunk persists between both list versions
				chunks[checksum] = true
			} else { // introduction of a new chunk
				newChunks = append(newChunks, chunk)
			}
		}

		for checksum, inNew := range chunks {
			if !inNew { // chunk in previous list was not encountered
				err := processDeletedChunk(db, checksum)
				if err != nil {
					return err
				}
			}
		}

		// add new chunks
		for _, newChunk := range newChunks {
			checksum := fmt.Sprintf("%x", newChunk.Checksum())

			err := processNewChunk(db, checksum, &newChunk)
			if err != nil {
				return err
			}
		}

		fnode := &node.FNode{
			Path:   file.Path,
			Size:   size,
			Chunks: checksums,
		}
		// replace old chunklist
		writeChunklist(fnode)
	}

	return nil
}

func processDeletedChunk(db *sql.DB, checksum string) error {
	// decrease chunk instance_count in db
	sqlitechunks.DecreaseCount(db, checksum)
	instances := sqlitechunks.GetCount(db, checksum)

	if instances == 0 { // instance_count reaches 0 after decrease
		path := filepath.Join("./chunks", checksum)
		err := os.Remove(path)
		if err != nil {
			return err
		}

		// TODO: remove chunk from db, request server for R2 removal of chunk
		// SERVER: if chunk does not exist in snapshot cache, proceed with removal
	}

	return nil
}

func processDeletedFile(db *sql.DB, file *node.Node) error {
	if file.IsDir {
		for _, child := range file.Children {
			err := processDeletedFile(db, child)
			if err != nil {
				return err
			}
		}
	} else {
		path := strings.Replace(file.Path, "data", "chunklists", 1)

		fi, err := os.Open(path)
		if err != nil {
			return err
		}
		defer fi.Close()

		bytes, err := io.ReadAll(fi)
		if err != nil {
			return err
		}

		oldChunklist := node.FNode{}
		json.Unmarshal(bytes, &oldChunklist)

		for _, checksum := range oldChunklist.Chunks {
			err := processDeletedChunk(db, checksum)
			if err != nil {
				return err
			}
		}

		// delete chunklist at node.Path, replacing "data" with "chunklists" as the path base
		err = os.Remove(path)
		if err != nil {
			return err
		}
	}

	return nil
}

func walk(rootOld, rootNew *node.Node, timeOld int64) ([]*node.Node, []*node.Node, []*node.Node) {
	pathToNodeMap := node.GetPathToNodeMappings(rootOld.Children)

	sameDirs := []struct{ oldDir, newDir *node.Node }{}

	newFiles, modifiedFiles, deletedFiles := []*node.Node{}, []*node.Node{}, []*node.Node{}

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
		tmpN, tmpM, tmpD := walk(dirs.oldDir, dirs.newDir, timeOld)

		newFiles = append(newFiles, tmpN...)
		modifiedFiles = append(modifiedFiles, tmpM...)
		deletedFiles = append(deletedFiles, tmpD...)
	}

	return newFiles, modifiedFiles, deletedFiles
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

		fo, err := os.Create("./hierarchy.json")
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

		newFiles, modifiedFiles, deletedFiles := walk(rootOld, rootNew, timeOld)

		db, err := sqlitechunks.OpenDB("./db/chunks.sqlite")
		if err != nil {
			return err
		}
		defer db.Close()

		fmt.Println("new files")
		for _, newFile := range newFiles {
			processNewFile(db, newFile)
		}

		fmt.Println("modified files")
		for _, modifiedFile := range modifiedFiles {
			processModifiedFile(db, modifiedFile)
		}

		fmt.Println("deleted files")
		for _, deletedFile := range deletedFiles {
			processDeletedFile(db, deletedFile)
		}
	}

	return nil
}

func main() {
	defer func() {
		if str := recover(); str != nil {
			fmt.Println(str)
		}
	}()

	err := initHierarchy()
	check(err)
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

	err = w.Start(time.Second)
	check(err)
}
