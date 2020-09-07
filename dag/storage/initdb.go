/*
   This file is part of go-palletone.
   go-palletone is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.
   go-palletone is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.
   You should have received a copy of the GNU General Public License
   along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package storage

import (
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/dag/dagconfig"
)

var DBPath = dagconfig.DefaultDataDir()

func Init(path string, cache int, handles int) (ptndb.Database, error) {
	var err error
	if path == "" {
		path = DBPath
	}

	Dbconn, err := ptndb.NewLDBDatabase(path, cache, handles)
	return Dbconn, err
}

//func ReNewDbConn(path string) ptndb.LDBDatabase {
//	if dbconn, err := ptndb.NewLDBDatabase(path, 0, 0); err != nil {
//		return ptndb.LDBDatabase{}
//	} else {
//		return dbconn
//	}
//}
