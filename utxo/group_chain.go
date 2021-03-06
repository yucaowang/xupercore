package utxo

import (
	"strings"

	"github.com/xuperchain/xupercore/contract"
	"github.com/xuperchain/xupercore/xmodel"
)

func (uv *UtxoVM) QueryChainInList() map[string]bool {
	return uv.getChainInList()
}

func (uv *UtxoVM) QueryPeerIDsInList(bcname string) map[string]bool {
	return uv.getPeerIDsInList(bcname)
}

func (uv *UtxoVM) getPeerIDsInList(bcname string) map[string]bool {
	peerIDMap := map[string]bool{}
	args := map[string][]byte{
		"bcname": []byte(bcname),
	}

	groupChainContract := uv.GetGroupChainContract()
	if groupChainContract == nil {
		return peerIDMap
	}
	moduleName := groupChainContract.ModuleName
	contractName := groupChainContract.ContractName
	methodName := groupChainContract.MethodName + "Node"

	uv.xlog.Trace("check IP list of group", "moduleName:", moduleName, "contractName:", contractName, "methodName:", methodName, "bcname", bcname)

	if moduleName == "" && contractName == "" && methodName == "" {
		return peerIDMap
	}

	status, target, err := uv.queryGroupChain(moduleName, contractName, methodName, args)
	if status >= 400 || err != nil || string(target) == "" {
		return peerIDMap
	}
	res := strings.Split(string(target), "\x01")
	for _, v := range res {
		arr := strings.Split(v, "/")
		if len(arr) <= 0 {
			continue
		}
		peerID := arr[len(arr)-1]
		if peerID == "" {
			continue
		}
		peerIDMap[peerID] = true
	}
	return peerIDMap
}

func (uv *UtxoVM) getChainInList() map[string]bool {
	chainMap := map[string]bool{}
	args := map[string][]byte{}

	groupChainContract := uv.GetGroupChainContract()
	if groupChainContract == nil {
		return chainMap
	}
	moduleName := groupChainContract.ModuleName
	contractName := groupChainContract.ContractName
	methodName := groupChainContract.MethodName + "Chain"

	if moduleName == "" && contractName == "" && methodName == "" {
		return chainMap
	}

	status, target, err := uv.queryGroupChain(moduleName, contractName, methodName, args)
	if status >= 400 || err != nil || string(target) == "" {
		return chainMap
	}
	res := strings.Split(string(target), "\x01")
	for _, v := range res {
		chainMap[v] = true
	}

	return chainMap
}

func (uv *UtxoVM) queryGroupChain(moduleName, contractName, methodName string, args map[string][]byte) (int, []byte, error) {
	modelCache, err := xmodel.NewXModelCache(uv.GetXModel(), uv)
	if err != nil {
		return 400, nil, err
	}
	contextConfig := &contract.ContextConfig{
		XMCache:        modelCache,
		ResourceLimits: contract.MaxLimits,
		ContractName:   contractName,
	}
	vm, err := uv.vmMgr3.GetVM(moduleName)
	if err != nil {
		return 400, nil, err
	}
	ctx, err := vm.NewContext(contextConfig)
	if err != nil {
		return 400, nil, err
	}
	invokeRes, invokeErr := ctx.Invoke(methodName, args)
	if invokeErr != nil {
		ctx.Release()
		return 400, nil, invokeErr
	}
	ctx.Release()

	return invokeRes.Status, invokeRes.Body, nil
}
