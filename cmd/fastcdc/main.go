package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"fastcdc-backup/pkg/fastcdc"
)

// func check(e error) {
// 	if e != nil {
// 		panic(e)
// 	}
// }

func createChunkDir() error {
	if _, err := os.Stat("./chunks"); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir("chunks", 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		err = os.RemoveAll("./chunks")
		if err != nil {
			return err
		}

		err = os.Mkdir("chunks", 0755)
		if err != nil {
			return err
		}
	}

	return nil
}

func writeChunk(chunk fastcdc.Chunk) error {
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

func writeFile(filePath string) error {
	defer func() {
		if str := recover(); str != nil {
			fmt.Println(str)
		}
	}()

	fi, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer fi.Close()

	// set options
	opt := fastcdc.Options{}
	opt.SetDefaults()

	// buffered reader is more efficient for many small reads
	br := bufio.NewReaderSize(fi, 4096)
	chunker := fastcdc.NewChunker(br, opt)

	err = createChunkDir()
	if err != nil {
		return err
	}

	for {
		chunk, err := chunker.NextChunk()

		if err == io.EOF {
			fmt.Println("Finished chunking")
			break
		} else if err != nil {
			return err
		}

		err = writeChunk(chunk)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	_ = writeFile("./shakespeare.txt")
}
