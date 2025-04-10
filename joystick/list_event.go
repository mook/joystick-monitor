/*
 *    Copyright (c) 2023 Unrud <unrud@outlook.com>
 *
 *    This file is part of joystick-monitor.
 *
 *    joystick-monitor is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *
 *    joystick-monitor is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *
 *    You should have received a copy of the GNU General Public License
 *    along with joystick-monitor.  If not, see <http://www.gnu.org/licenses/>.
 */

/*
 * The joystick detection code is derived from libinput:
 *
 * Copyright © 2006-2009 Simon Thum
 * Copyright © 2008-2012 Kristian Høgsberg
 * Copyright © 2010-2012 Intel Corporation
 * Copyright © 2010-2011 Benjamin Franzke
 * Copyright © 2011-2012 Collabora, Ltd.
 * Copyright © 2013-2014 Jonas Ådahl
 * Copyright © 2013-2015 Red Hat, Inc.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a
 * copy of this software and associated documentation files (the "Software"),
 * to deal in the Software without restriction, including without limitation
 * the rights to use, copy, modify, merge, publish, distribute, sublicense,
 * and/or sell copies of the Software, and to permit persons to whom the
 * Software is furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice (including the next
 * paragraph) shall be included in all copies or substantial portions of the
 * Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL
 * THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
 * FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
 * DEALINGS IN THE SOFTWARE.
 */

package joystick

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"strconv"
	"strings"
)

func ListEventJoysticks() (map[string]struct{}, error) {
	joysticks := make(map[string]struct{})
	inputDir, err := os.Open("/dev/input")
	if errors.Is(err, os.ErrNotExist) {
		return joysticks, nil
	}
	if err != nil {
		return nil, err
	}
	defer inputDir.Close()
	inputEntries, err := inputDir.ReadDir(0)
	if err != nil {
		return nil, err
	}
	for _, inputEntry := range inputEntries {
		if !strings.HasPrefix(inputEntry.Name(), "event") {
			continue
		}
		ok, err := isCompatibleDevice(inputEntry.Name())
		if err != nil {
			log.Printf("Failed to check device %s compatibility, ignoring: %s", inputEntry.Name(), err)
			continue
		}
		if !ok {
			continue
		}
		joystick := path.Clean(path.Join(inputDir.Name(), inputEntry.Name()))
		joysticks[joystick] = struct{}{}
	}
	return joysticks, nil
}

var wellKnownKeyboardKeys = map[uint]struct{}{
	29:  {}, /* KEY_LEFTCTRL */
	58:  {}, /* KEY_CAPSLOCK */
	69:  {}, /* KEY_NUMLOCK */
	110: {}, /* KEY_INSERT */
	113: {}, /* KEY_MUTE */
	140: {}, /* KEY_CALC */
	144: {}, /* KEY_FILE */
	155: {}, /* KEY_MAIL */
	164: {}, /* KEY_PLAYPAUSE */
	224: {}, /* KEY_BRIGHTNESSDOWN */
}

// Check if a given device (e.g. "event0") is compatible.
func isCompatibleDevice(deviceName string) (bool, error) {
	deviceDir := path.Join("/sys/class/input/", deviceName)

	// There doesn't seem to be an EV_* type that _can't_ be attached to a
	// joystick / gamepad.

	// We check to make sure that the device has at least one "key", and that
	// the available keys smell like a joystick.
	keyBytes, err := os.ReadFile(path.Join(deviceDir, "device/capabilities/key"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			// Normally the file should exist with a "0", but whatever.
			return false, nil
		}
		return false, fmt.Errorf("failed to parse device %s keys: %w", deviceName, err)
	}
	caps, err := parseCapabitities(string(keyBytes))
	if err != nil {
		return false, err
	}
	if len(caps) < 1 {
		return false, nil
	}

	// We do some libinput-derived key availablity scanning to guess if this is
	// a valid joystick or gamepad.
	keyboardKeysCount := 0
	wellKnownKeyboardKeysCount := 0
	joystickButtonsCount := 0

	for _, cap := range caps {
		if _, ok := wellKnownKeyboardKeys[cap]; ok {
			wellKnownKeyboardKeysCount++
		}
		switch {
		case cap > 0 && cap < 0x100:
			// This is the generic keyboard keys range.
			keyboardKeysCount++
		case cap == 0x11F /* BTN_JOYSTICK - 1 */ :
			// The mout buttons range is just below BTN_JOYSTICK range; if this
			// is found, it's probably a mouse with too many buttons
			return false, nil
		case cap >= 0x120 /* BTN_JOYSTICK */ && cap < 0x140 /* BTN_DIGI */ :
			// This has a joystick or gamepad button
			joystickButtonsCount++
		case cap >= 0x160 /* KEY_OK */ && cap < 0x220 /* BTN_DPAD_UP */ :
			// This is a different range of keyboard keys
			keyboardKeysCount++
		case cap >= 0x220 /* BTN_DPAD_UP */ && cap <= 0x223 /* BTN_DPAD_RIGHT */ :
			joystickButtonsCount++
		case cap >= 0x230 /* KEY_ALS_TOGGLE */ && cap < 0x2C0 /* BTN_TRIGGER_HAPPY */ :
			keyboardKeysCount++
		case cap >= 0x2C0 /* BTN_TRIGGER_HAPPY */ && cap <= 0x2E7 /* BTN_TRIGGER_HAPPY40 */ :
			// This has an extended joystick button.
			joystickButtonsCount++
		}
	}

	if wellKnownKeyboardKeysCount >= 4 {
		// If there are at least 4 keys in the well-known list (per libinput),
		// reject this as being probably a keyboard.
		return false, nil
	}
	if joystickButtonsCount < 2 {
		// If we don't at least have two buttons that look like a joystick,
		// ths is probably some other device.
		return false, nil
	}
	if keyboardKeysCount >= 10 {
		// If there are ten keys that look like keyboard keys, this is probably
		// not a joystick.
		return false, nil
	}

	return true, nil
}

// Parse the capabilities of a device, given the contents of a capabilities file.
func parseCapabitities(capabilitiesData string) ([]uint, error) {
	// The algorithm is based on the description from
	// https://gist.github.com/TriceHelix/de47ed38dcb4f7216b26291c47445d99
	// Each file contains multiple chunks, each of which is a uint64 in hex.
	// The first chunk is the most significant (i.e. largest bit).
	chunks := strings.Split(capabilitiesData, " ")
	var result []uint
	for chunkNum, chunk := range chunks {
		// Each chunk is 64 bits; the nth chunk from the right is offset by
		// n * 64.
		offset := (len(chunks) - chunkNum - 1) * 64
		chunk = strings.ReplaceAll(chunk, "\n", "")
		numeric, err := strconv.ParseUint(chunk, 16, 64)
		if err != nil {
			return nil, err
		}
		for i := range 64 {
			mask := uint64(1) << i
			if numeric&mask != 0 {
				result = append(result, uint(offset+i))
			}
		}
	}
	return result, nil
}
