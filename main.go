package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	root_dir string
	ext      string
	list     bool
	size     int64
	del      bool
	work_log io.Writer
}

func main() {
	//parse flags
	root := flag.String("root", ".", "root directory")
	ext := flag.String("ext", "", "extension")
	list := flag.Bool("list", false, "list files")
	// When -size specified, only match files whose size is larger than this value.
	size := flag.Int64("size", 0, "min size")
	del := flag.Bool("del", false, "delete files")
	// default value for this flag is an empty string so if the user doesn’t provide a name, the program will send output to STDOUT.
	log_file := flag.String("log", "", "log file")
	flag.Parse()
	var f io.Writer
	if *log_file != "" {
		f, err := os.OpenFile(*log_file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		defer f.Close()
	}

	//create config
	config := Config{
		root_dir: *root,
		ext:      *ext,
		list:     *list,
		size:     *size,
		del:      *del,
		work_log: f,
	}
	//run
	if err := run(os.Stdout, config); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// walk directory
// retrieve all files and directories in the root directory
func run(out io.Writer, config Config) error {
	// create a logger
	del_logger := log.New(config.work_log, "delete file: ", log.LstdFlags)
	return filepath.Walk(config.root_dir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// skip files
			if filter_out(path, config.size, config.ext, info) {
				return nil
			}
			if config.del {
				return del_file(path, del_logger)
			}
			// default output is list of files and directories
			return list_file(path, out)

		})
}

// define which files or directories to filter out
func filter_out(path string, min_size int64, ext string, info os.FileInfo) bool {
	// when occurring a directory or the size of fileinfo is less than config.size
	if info.IsDir() || info.Size() < min_size {
		return true
	}
	// when config.ext is not empty and the extension of file is equal to config.ext
	if ext != "" && ext != filepath.Ext(path) {
		return true
	}
	return false
}

func list_file(path string, out io.Writer) error {
	_, err := fmt.Fprintln(out, path)
	return err
}

// del_logger to log the file that is deleted without errors
func del_file(path string, del_logger *log.Logger) error {
	if err := os.Remove(path); err != nil {
		return err
	}
	del_logger.Println(path)
	return nil
}
