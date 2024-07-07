package video

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func ClearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func moveAndWrite(row, col int, char rune) {
	moveCursor(row, col)
	fmt.Printf("%c", char)
}
