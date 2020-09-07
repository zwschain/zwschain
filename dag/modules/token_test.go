/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package modules

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestToken2Json(t *testing.T) {
	token := &FungibleToken{Name: "PalletOne BTC", Symbol: "PBTC", Decimals: 8, TotalSupply: 9000000000000}
	txt, _ := json.Marshal(token)
	fmt.Println(string(txt))

	tkdef := &TokenDefine{TokenDefineJson: txt}
	txtdefJson, _ := json.Marshal(tkdef)
	fmt.Println(string(txtdefJson))

	fmt.Println("=================================")

	tkdefNew := TokenDefine{}
	json.Unmarshal(txtdefJson, &tkdefNew)
	fmt.Println(string(tkdefNew.TokenDefineJson))

	tokenNew := FungibleToken{}
	json.Unmarshal(tkdefNew.TokenDefineJson, &tokenNew)
	fmt.Println(tokenNew)
}
