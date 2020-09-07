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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package ptn

import (
	"time"

	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/common/p2p"
	"github.com/palletone/go-palletone/common/p2p/discover"
	"github.com/palletone/go-palletone/common/util"
	mp "github.com/palletone/go-palletone/consensus/mediatorplugin"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
)

func (pm *ProtocolManager) newProducedUnitBroadcastLoop() {
	for {
		select {
		case event := <-pm.newProducedUnitCh:
			log.Debugf("receive NewProducedUnitEvent")
			pm.IsExistInCache(event.Unit.Hash().Bytes())
			go pm.BroadcastUnit(event.Unit, true)
			//self.BroadcastCorsHeader(event.Unit.Header(), self.SubProtocols[0].Name)

		case <-pm.newProducedUnitSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) toGroupSignEventRecvLoop() {
	log.Debugf("toGroupSignEventRecvLoop")
	for {
		select {
		case event := <-pm.toGroupSignCh:
			go pm.toGroupSign(event)

		// Err() channel will be closed when unsubscribing.
		case <-pm.toGroupSignSub.Err():
			return
		}
	}
}

func (pm *ProtocolManager) toGroupSign(event modules.ToGroupSignEvent) {
	log.Debugf("receive toGroupSign event")

	// 判断是否满足群签名的条件
	if !pm.dag.IsSynced(false) {
		log.Debugf(errStr)
		return
	}

	if !pm.producer.LocalHaveActiveMediator() && !pm.producer.LocalHavePrecedingMediator() {
		log.Debugf("the current node has no mediator")
		return
	}

	// 获取最高稳定单元的高度
	gasToken := dagconfig.DagConfig.GetGasToken()
	iun := pm.dag.GetIrreversibleUnitNum(gasToken)

	// 对稳定单元后一个unit进行群签名
	newHash, err := pm.dag.GetUnitHash(&modules.ChainIndex{AssetID: gasToken, Index: iun + 1})
	if err != nil {
		log.Debugf(err.Error())
		return
	}

	if pm.IsExistInCache(util.RlpHash(newHash).Bytes()) {
		return
	}
	go pm.producer.AddToTBLSSignBufs(newHash)
}

// @author Albert·Gou
func (pm *ProtocolManager) sigShareTransmitLoop() {
	for {
		select {
		case event := <-pm.sigShareCh:
			pm.IsExistInCache(event.Hash().Bytes())
			go pm.transmitSigShare(&event)

			// Err() channel will be closed when unsubscribing.
		case <-pm.sigShareSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) transmitSigShare(sigShare *mp.SigShareEvent) {
	unitHash := sigShare.UnitHash
	header, err := pm.dag.GetHeaderByHash(unitHash)
	if err != nil {
		log.Debugf("fail to get header of unit(%v), err: %v", unitHash.TerminalString(), err.Error())
		return
	}

	// 判读该unit是否是本地mediator生产的
	if pm.producer.IsLocalMediator(header.Author()) {
		go pm.producer.AddToTBLSRecoverBuf(sigShare)
	} else {
		go pm.BroadcastSigShare(sigShare)
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) groupSigBroadcastLoop() {
	for {
		select {
		case event := <-pm.groupSigCh:
			pm.IsExistInCache(event.Hash().Bytes())
			go pm.dag.SetUnitGroupSign(event.UnitHash, event.GroupSig, pm.txpool)
			go pm.BroadcastGroupSig(&event)

		// Err() channel will be closed when unsubscribing.
		case <-pm.groupSigSub.Err():
			return
		}
	}
}

// @author Albert·Gou
// BroadcastGroupSig will propagate the group signature of unit to p2p network
func (pm *ProtocolManager) BroadcastGroupSig(groupSig *mp.GroupSigEvent) {
	now := uint64(time.Now().Unix())
	if now > groupSig.Deadline {
		return
	}

	peers := pm.peers.PeersWithoutGroupSig(groupSig.Hash())
	for _, peer := range peers {
		go peer.SendGroupSig(groupSig)
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) vssDealTransmitLoop() {
	for {
		select {
		case event := <-pm.vssDealCh:
			go pm.transmitVSSDeal(&event)

			// Err() channel will be closed when unsubscribing.
		case <-pm.vssDealSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) transmitVSSDeal(deal *mp.VSSDealEvent) {
	// 重复判断
	//// 判断是否同步, 如果没同步完成，发起的vss deal是无效的，浪费带宽
	//if !pm.dag.IsSynced(true) {
	//	log.Debugf(errStr)
	//	return
	//}

	// vss deal 一定是请求其他mediator的消息
	pm.IsExistInCache(deal.Hash().Bytes())
	//go pm.BroadcastVSSDeal(deal)

	// 处理 同一个节点配置多个mediator的情况
	ma := pm.dag.GetActiveMediatorAddr(int(deal.DstIndex))
	if pm.producer.IsLocalMediator(ma) {
		go pm.producer.AddToDealBuf(deal)
	} else {
		go pm.BroadcastVSSDeal(deal)
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) vssResponseBroadcastLoop() {
	for {
		select {
		case event := <-pm.vssResponseCh:
			go pm.broadcastVssResp(&event)

			// Err() channel will be closed when unsubscribing.
		case <-pm.vssResponseSub.Err():
			return
		}
	}
}

// @author Albert·Gou
func (pm *ProtocolManager) broadcastVssResp(resp *mp.VSSResponseEvent) {
	//peers := pm.GetActiveMediatorPeers()
	//for _, peer := range peers {
	//	if peer == nil { // 此时为本节点
	//		go pm.producer.AddToResponseBuf(resp)
	//		continue
	//	}
	//
	//	err := peer.SendVSSResponse(resp)
	//	if err != nil {
	//		log.Debugf(err.Error())
	//	}
	//}

	pm.IsExistInCache(resp.Hash().Bytes())
	go pm.producer.AddToResponseBuf(resp)
	go pm.BroadcastVSSResponse(resp)
}

// GetPeer, retrieve specified peer. If it is the node itself, p is nil and self is true
// @author Albert·Gou
func (pm *ProtocolManager) GetPeer(node *discover.Node) (p *peer, self bool) {
	id := node.ID
	if pm.srvr.Self().ID == id {
		self = true
	}

	p = pm.peers.Peer(id.TerminalString())
	if p == nil && !self {
		log.Debugf("the Peer is not exist: %v", node.String())
	}

	return
}

// GetActiveMediatorPeers retrieves a list of peers that active mediator.
// If the value is nil, it is the node itself
// @author Albert·Gou
func (pm *ProtocolManager) GetActiveMediatorPeers() map[string]*peer {
	nodes := pm.dag.GetActiveMediatorNodes()
	list := make(map[string]*peer, len(nodes))

	for id, node := range nodes {
		peer, self := pm.GetPeer(node)
		if peer != nil || self {
			list[id] = peer
		}
	}

	return list
}

// @author Albert·Gou
func (p *peer) SendVSSDeal(deal *mp.VSSDealEvent) error {
	p.MarkVSSDeal(deal.Hash())
	return p2p.Send(p.rw, VSSDealMsg, deal)
}

// @author Albert·Gou
func (p *peer) SendVSSResponse(resp *mp.VSSResponseEvent) error {
	p.MarkVSSResponse(resp.Hash())
	return p2p.Send(p.rw, VSSResponseMsg, resp)
}

// @author Albert·Gou
func (p *peer) SendSigShare(sigShare *mp.SigShareEvent) error {
	p.MarkSigShare(sigShare.Hash())
	return p2p.Send(p.rw, SigShareMsg, sigShare)
}

//BroadcastGroupSig
func (p *peer) SendGroupSig(groupSig *mp.GroupSigEvent) error {
	p.MarkGroupSig(groupSig.Hash())
	return p2p.Send(p.rw, GroupSigMsg, groupSig)
}
