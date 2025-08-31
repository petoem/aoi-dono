package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const defaultEditor = "vi"

// editorEnvs are the environment variable used to detect users editor,
// lower index has higher precedence
var editorEnvs = []string{
	"AOI_EDITOR",
	"VISUAL",
	"EDITOR",
}

func getEditor() string {
	var editor string
	for _, e := range editorEnvs {
		editor = os.Getenv(e)
		if editor != "" {
			return editor
		}
	}
	return defaultEditor
}

func openEditor(editor string, post string) error {
	fmt.Fprintln(os.Stderr, "hint: Waiting for your editor to close the file...")
	args := strings.Split(editor, " ")
	args = append(args, post)
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("unable to start editor: %w", err)
	}
	return nil
}
