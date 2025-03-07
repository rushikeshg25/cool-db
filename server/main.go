package server

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/chzyer/readline"
	"github.com/rushikeshg25/coolDb/internal"
	"github.com/rushikeshg25/coolDb/internal/errors"
)

func Start(dbFile string, dbFilePath string) {
	printBanner()
	f, err := os.Open(dbFilePath)
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
		os.Exit(1)
	}
	defer f.Close()
	c, err := internal.ParseFileConfig(dbFile, dbFilePath, f)
	if err != nil {
		errors.ErrFileParse(dbFilePath)
		os.Exit(1)
	}
	fmt.Println(c)
	l, err := readline.NewEx(&readline.Config{
		Prompt:          "\033[31m»\033[0m ",
		HistoryFile:     "/tmp/readline.tmp",
		AutoComplete:    completer,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",

		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		fmt.Println("Error creating readline")
	}
	defer l.Close()
	l.CaptureExitSignal()

	for {
		input, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(input) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}
		input = strings.TrimSpace(input)
		if strings.HasPrefix(input, ".") {
			fmt.Println(input)
		} else {
			internal.ProcessQuery(input)
		}
	}
}

func printBanner() {
	fmt.Print(`
 ██████╗ ██████╗  ██████╗ ██╗     ██████╗ ██████╗ 
██╔════╝██╔═══██╗██╔═══██╗██║     ██╔══██╗██╔══██╗
██║     ██║   ██║██║   ██║██║     ██║  ██║██████╔╝
██║     ██║   ██║██║   ██║██║     ██║  ██║██╔══██╗
╚██████╗╚██████╔╝╚██████╔╝███████╗██████╔╝██████╔╝
 ╚═════╝ ╚═════╝  ╚═════╝ ╚══════╝╚═════╝ ╚═════╝ 
`)
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem(".",
		readline.PcItem("exit"),
		readline.PcItem("quit"),
		readline.PcItem("dbinfo"),
		readline.PcItem("backup"),
		readline.PcItem("tables"),
		readline.PcItem("help"),
		readline.PcItem("clone"),
	),
)

// func listFiles(path string) func(string) []string {
// 	return func(line string) []string {
// 		names := make([]string, 0)
// 		files, _ := os.ReadDir(path)
// 		for _, f := range files {
// 			names = append(names, f.Name())
// 		}
// 		return names
// 	}
// }

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}
