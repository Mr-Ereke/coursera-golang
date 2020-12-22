package main

import (
	"fmt"
	"io"
	"os"
	"io/ioutil"
	// "strings"
	"strconv"
)

func getFileName(file os.FileInfo, printFiles bool) string {
	name := ""

	if (printFiles && !file.IsDir()) {
		size := "empty"

		if (file.Size() > 0) {
			size = strconv.FormatInt(file.Size(), 10) + "b"
		}

		size = " (" + size + ")"
		name = file.Name() + size
	} else {
		name = file.Name()
	}

	return "───" + name + "\n"
}

func buildTree(out io.Writer, path string, printFiles bool, level int, levels []bool) error {
	files, error := ioutil.ReadDir(path)

	if !printFiles {
		temp := files[:0]

		for _, f := range files {
			if f.IsDir() {
				temp = append(temp, f)
			}
		}

		files = temp
	}

	if error != nil {
      return error
  }

	for _, f := range files {
		if (!f.IsDir() && !printFiles) {
			continue
		}

		if (level == 0 && f.Name() == path) {
			continue
		}

		fileName := ""

		for i := 0; i < level; i++ {
			if ((files[len(files)-1] != f && !levels[i]) || (!levels[i] && i < level)) {
				fmt.Fprint(out, "│")
			}

			fmt.Fprint(out, "\t")
		}

		if files[len(files)-1] == f  {
			fileName = "└"
		} else {
			fileName = "├"
		}

		if (f.IsDir()) {
				fileName = fileName + getFileName(f, printFiles)
				fmt.Fprint(out, fileName)
				sublevel := level + 1
				sublevels := append(levels, files[len(files)-1] == f)
				error := buildTree(out, path + string(os.PathSeparator) + f.Name(), printFiles, sublevel, sublevels)

				if error != nil {
			      return error
			  }
		}

		if (printFiles && !f.IsDir()) {
			fileName = fileName + getFileName(f, printFiles)
			fmt.Fprint(out, fileName)
		}
  }

	return nil
}

func dirTree(out io.Writer, path string, printFiles bool) error  {
	levels := []bool{}
  error := buildTree(out, path, printFiles, 0, levels)

  if error != nil {
      return error
  }

	return nil
}

func main() {
	out := os.Stdout

	if !(len(os.Args) == 2 || len(os.Args) == 3) {
		panic("usage go run main.go . [-f]")
	}

	path := os.Args[1]
	printFiles := len(os.Args) == 3 && os.Args[2] == "-f"
	err := dirTree(out, path, printFiles)

	if err != nil {
		panic(err.Error())
	}
}
