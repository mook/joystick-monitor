package joystick

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sys/unix"
)

type praseCababilitiesTestCase struct {
	Input    string
	Expected []uint
}

func TestParseCapabitities(t *testing.T) {
	testCases := []praseCababilitiesTestCase{
		{"0", []uint{}},
		// "AT Translated Set 2 keyboard" /ev query
		{"120013", []uint{unix.EV_SYN, unix.EV_KEY, unix.EV_MSC, unix.EV_LED, unix.EV_REP}},
		// "Lid Switch" /ev query
		{"21", []uint{unix.EV_SYN, unix.EV_SW}},
		// "PC Speaker" /ev query
		{"40001", []uint{unix.EV_SYN, unix.EV_SND}},
		// Mouse /key query
		{"30000 0 0 0 0", []uint{0x110 /*BTN_LEFT*/, 0x111 /*BTN_RIGHT*/}},
		// Xbox Wireless Controller /ev query
		{"20001b", []uint{unix.EV_SYN, unix.EV_KEY, unix.EV_ABS, unix.EV_MSC, unix.EV_FF}},
		// Xbox Wireless Controller /key query
		{"7fff000000000000 0 100040000000 0 0", []uint{
			0x9E /*KEY_BACK*/, 0xAC, /*KEY_HOMEPAGE*/
			0x130 /*BTN_SOUTH*/, 0x131, 0x132, 0x133, 0x134, 0x135, 0x136,
			0x137, 0x138, 0x139, 0x13A, 0x13B, 0x13C, 0x13D, 0x13E, /* BTN_THUMBR */
		}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Input, func(t *testing.T) {
			results, err := parseCapabitities(testCase.Input)
			if assert.NoError(t, err) {
				slices.Sort(results)
				assert.ElementsMatch(t, testCase.Expected, results)
			}
		})
	}
}
