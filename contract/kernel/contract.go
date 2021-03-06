package kernel

import (
	"errors"
	"fmt"

	"github.com/xuperchain/xupercore/common"
	"github.com/xuperchain/xupercore/contract"
	"github.com/xuperchain/xupercore/contract/bridge"
	"github.com/xuperchain/xupercore/permission/acl/utils"
)

// contractMethods manage methods about contract
type contractMethods struct {
	xbridge *bridge.XBridge
}

// Deploy deploys contract
func (c *contractMethods) Deploy(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	// check if account exist
	accountName := args["account_name"]
	contractName := args["contract_name"]
	if accountName == nil || contractName == nil {
		return nil, errors.New("invoke DeployMethod error, account name or contract name is nil")
	}
	// check if contractName is ok
	if contractErr := common.ValidContractName(string(contractName)); contractErr != nil {
		return nil, fmt.Errorf("deploy failed, contract `%s` contains illegal character, error: %s", contractName, contractErr)
	}
	_, err := ctx.ModelCache.Get(utils.GetAccountBucket(), accountName)
	if err != nil {
		return nil, fmt.Errorf("get account `%s` error: %s", accountName, err)
	}

	out, resourceUsed, err := c.xbridge.DeployContract(ctx.ContextConfig, args)
	if err != nil {
		return nil, err
	}
	ctx.AddResourceUsed(resourceUsed)

	// key: contract, value: account
	err = ctx.ModelCache.Put(utils.GetContract2AccountBucket(), contractName, accountName)
	if err != nil {
		return nil, err
	}
	key := utils.MakeAccountContractKey(string(accountName), string(contractName))
	err = ctx.ModelCache.Put(utils.GetAccount2ContractBucket(), []byte(key), []byte(utils.GetAccountContractValue()))
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Upgrade upgrades contract
func (c *contractMethods) Upgrade(ctx *KContext, args map[string][]byte) (*contract.Response, error) {
	contractName := args["contract_name"]
	if contractName == nil {
		return nil, errors.New("invoke Upgrade error, contract name is nil")
	}
	err := ctx.ContextConfig.Core.VerifyContractOwnerPermission(string(contractName), ctx.AuthRequire)
	if err != nil {
		return nil, err
	}
	resp, resourceUsed, err := c.xbridge.UpgradeContract(ctx.ContextConfig, args)
	if err != nil {
		return nil, err
	}
	ctx.AddResourceUsed(resourceUsed)
	return resp, nil
}
