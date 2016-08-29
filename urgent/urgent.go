/*
Package main implements the urgent command.

The urgent command takes no input and prints out a comma separated
list of windows with the ICCCM urgency hint set:

    The UrgencyHint flag, if set in the flags field, indicates that
    the client deems the window contents to be urgent, requiring the
    timely response of the user.

*/
package main

import (
	"fmt"
	"log"

	"strings"

	"github.com/BurntSushi/xgb/xproto"
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/icccm"
)

const urgencyHint = (1 << 8)

var urgent []string

// windowName returns the X11 property WM_NAME if available, otherwise
// it tries to get the ICCCM WM_NAME
func windowName(X *xgbutil.XUtil, win xproto.Window) (string, error) {
	reply, err := xproto.GetProperty(X.Conn(), false, win, xproto.AtomWmName, xproto.AtomString, 0, 128).Reply()

	if err == nil && len(reply.Value) > 0 {
		return string(reply.Value), nil
	}

	return icccm.WmNameGet(X, win)
}

// findUrgent recursively checks all children of parent and appends to
// the global array urgent any window with a name and the urgency hint
// set
func findUrgent(X *xgbutil.XUtil, parent xproto.Window) error {
	tree, err := xproto.QueryTree(X.Conn(), parent).Reply()
	if err != nil {
		return fmt.Errorf("queryTree failed: %s", err)
	}

	for _, child := range tree.Children {
		findUrgent(X, child)

		if hints, err := icccm.WmHintsGet(X, child); err != nil || hints.Flags&urgencyHint == 0 {
			continue
		}

		if name, err := windowName(X, child); err == nil && name != "" {
			urgent = append(urgent, name)
		}
	}
	return nil
}

func main() {
	X, err := xgbutil.NewConn()
	if err != nil {
		log.Fatalf("Error connecting to X, quitting... %v", err)
	}

	defer X.Conn().Close()

	if err := findUrgent(X, X.RootWin()); err != nil {
		log.Println(err)
	}

	if len(urgent) != 0 {
		fmt.Println(strings.Join(urgent, ", "))
	}
}
