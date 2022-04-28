package main

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFilterOut(t *testing.T) {
	test_cases := []struct {
		name     string
		file     string
		ext      string
		size     int64
		expected bool
	}{
		{name: "test_case_1", file: "test.txt", ext: "", size: 0, expected: false},
		{name: "test_case_2", file: "test.txt", ext: ".txt", size: 5, expected: false},
		{name: "test_case_3", file: "test.txt", ext: ".go", size: 20, expected: true},
	}
	for _, tc := range test_cases {
		t.Run(tc.name, func(t *testing.T) {
			info, err := os.Stat(tc.file)
			if err != nil {
				t.Fatal(err)
			}
			res := filter_out(tc.file, tc.size, tc.ext, info)
			if res != tc.expected {
				t.Errorf("Expected %t, got %t", tc.expected, res)
			}
		})
	}
}
func TestRun(t *testing.T) {
	test_cases := []struct {
		name     string
		root     string
		ext      string
		list     bool
		size     int64
		expected string
	}{
		{name: "test_case_1", root: "tests", ext: "", list: true, size: 0,
			expected: "tests/ten.txt\ntests/zero.txt\n"},
		{name: "test_case_2", root: "tests", ext: ".txt", list: true, size: 5, expected: "tests/ten.txt\n"},
		{name: "test_case_3", root: "tests", ext: ".go", list: true, size: 20, expected: ""},
	}
	for _, tc := range test_cases {
		t.Run(tc.name, func(t *testing.T) {
			config := Config{
				root_dir: tc.root,
				ext:      tc.ext,
				list:     tc.list,
				size:     tc.size,
				del:      false,
			}
			var buf bytes.Buffer
			// run
			if err := run(&buf, config); err != nil {
				t.Fatal(err)
			}
			res := buf.String()
			if res != tc.expected {
				t.Errorf("Expected '%q', got '%q'", tc.expected, res)
			}
		})
	}
}
func TestDel(t *testing.T) {
	test_cases := []struct {
		name          string
		config        Config
		ext_no_delete string
		num_del       int
		num_no_del    int
		expected      string
	}{
		{name: "test_case_1", config: Config{ext: ".log", del: true}, ext_no_delete: "", expected: ""},
		{name: "test_case_2", config: Config{ext: ".log", del: true}, ext_no_delete: ".txt", expected: ""},
		{name: "test_case_3", config: Config{ext: ".log", del: true}, ext_no_delete: ".log", expected: ""},
	}

	for _, tc := range test_cases {
		if tc.config.ext != tc.ext_no_delete {
			rand.Seed(time.Now().UnixNano())
			tc.num_del = rand.Intn(10)
			tc.num_no_del = rand.Intn(10)
		} else {
			tc.num_del = rand.Intn(10)
			tc.num_no_del = 0
		}

		t.Run(tc.name, func(t *testing.T) {
			temp_dir, cleanup := create_temp_dir(t,
				map[string]int{
					tc.config.ext:    tc.num_del,
					tc.ext_no_delete: tc.num_no_del,
				})
			defer cleanup()
			var buf bytes.Buffer
			var loger_buf bytes.Buffer
			tc.config.root_dir = temp_dir
			tc.config.work_log = &loger_buf
			if err := run(&buf, tc.config); err != nil {
				t.Fatal(err)
			}
			res := buf.String()
			if res != tc.expected {
				t.Errorf("Expected '%q', got '%q'", tc.expected, res)
			}
			left_files, err := os.ReadDir(temp_dir)
			if err != nil {
				t.Fatal(err)
			}
			if len(left_files) != tc.num_no_del {
				t.Errorf("Expected %d files, got %d", tc.num_no_del, len(left_files))
			}
			fmt.Println(tc.num_del, tc.num_no_del)
			fmt.Println(loger_buf.String())
			var expected_num_log_lines int
			if tc.config.ext != tc.ext_no_delete {
				expected_num_log_lines = tc.num_del + 1
			} else {
				expected_num_log_lines = tc.num_no_del + 1
			}
			log_lines := bytes.Split(loger_buf.Bytes(), []byte("\n"))
			if len(log_lines) != expected_num_log_lines {
				t.Errorf("Expected %d log lines, got %d", expected_num_log_lines, len(log_lines))
			}

		})
	}
}

// test helper function
func create_temp_dir(t *testing.T,
	files map[string]int) (dir_name string, cleanup func()) {
	// Mark this function as a test helper by calling the t.Helper() method
	t.Helper()
	// create temp dir
	temp_dir, err := os.MkdirTemp("", "walk-del-tests")
	if err != nil {
		t.Fatal(err)
	}
	for k, v := range files {
		for j := 0; j < v; j++ {
			file_name := fmt.Sprintf("file-%d%s", j, k)
			file_path := filepath.Join(temp_dir, file_name)
			if err := os.WriteFile(file_path, []byte("hello"), 0644); err != nil {
				t.Fatal(err)
			}
		}
	}
	return temp_dir, func() {
		os.RemoveAll(temp_dir)
	}
}
