package prompt

import "fmt"

var (
	// Info - info color
	Info = Teal
	// Warn - warning color
	Warn = Yellow
	// Error - error color
	Error = Red
	// Fatal - fatal color
	Fatal = Red
)

var (
	// Red - red color
	Red = Color("\033[1;31m%s\033[0m")
	// Green - green color
	Green = Color("\033[1;32m%s\033[0m")
	// Yellow - yellow color
	Yellow = Color("\033[1;33m%s\033[0m")
	// Purple - purple color
	Purple = Color("\033[1;34m%s\033[0m")
	// Magenta - magenta color
	Magenta = Color("\033[1;35m%s\033[0m")
	// Teal - teal color
	Teal = Color("\033[1;36m%s\033[0m")
	// White - wtite color
	White = Color("\033[1;37m%s\033[0m")
	// Grey - grey color
	Grey = Color("\033[1;30m%s\033[0m")
	// GreenGrey - green foreground, grey background
	GreenGrey = Color("\033[100;32m%s\033[0m")
)

// Color - return color
func Color(colorString string) func(...interface{}) string {
	sprint := func(args ...interface{}) string {
		return fmt.Sprintf(colorString,
			fmt.Sprint(args...))
	}
	return sprint
}
