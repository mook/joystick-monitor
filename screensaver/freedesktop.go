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

package screensaver

import (
	"github.com/godbus/dbus/v5"
)

type Screensaver struct {
	bus         *dbus.Conn
	screenSaver dbus.BusObject
}

func NewScreensaver() (*Screensaver, error) {
	bus, err := dbus.SessionBusPrivate()
	if err != nil {
		return nil, err
	}
	err = bus.Auth(nil)
	if err != nil {
		bus.Close()
		return nil, err
	}
	err = bus.Hello()
	if err != nil {
		bus.Close()
		return nil, err
	}
	return &Screensaver{
		bus: bus,
		screenSaver: bus.Object("org.freedesktop.ScreenSaver",
			"/org/freedesktop/ScreenSaver"),
	}, nil
}

func (s *Screensaver) Simulate() error {
	return s.screenSaver.Call("org.freedesktop.ScreenSaver.SimulateUserActivity", 0).Store()
}

func (s *Screensaver) Close() error {
	return s.bus.Close()
}
