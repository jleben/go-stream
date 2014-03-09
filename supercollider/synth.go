/*
Interface to SuperCollider audio processing server

Copyright (C) 2014 Jakob Leben <jakob.leben@gmail.com>

This program is free software; you can redistribute it and/or
modify it under the terms of the GNU General Public License
as published by the Free Software Foundation; either version 2
of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program; if not, write to the Free Software
Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
*/

package supercollider

type Synth struct {
  valid bool
  server *Server
  id int32
}

func NewSynth ( server *Server, name string, params ... interface {} ) Synth {
  id, err := server.NewSynth(name, params...)
  valid := err == nil
  synth := Synth { valid, server, id }
  return synth
}

func (synth Synth) Set ( params ... interface {} ) {
  if !synth.valid { return }
  synth.server.SetNodeControls( synth.id, params... )
}

func (synth Synth) Free () {
  if !synth.valid { return }
  synth.server.FreeNode( synth.id )
  synth.valid = false
}
