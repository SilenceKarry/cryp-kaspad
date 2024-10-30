package types

import (
	"encoding/json"
	"fmt"

	"github.com/tidwall/gjson"
)

type ChainInfo struct {
	ServerVersion             string `json:"server_version"`
	ChainID                   string `json:"chain_id"`
	HeadBlockNum              int64  `json:"head_block_num"`
	LastIrreversibleBlockNum  int64  `json:"last_irreversible_block_num"`
	LastIrreversibleBlockID   string `json:"last_irreversible_block_id"`
	HeadBlockID               string `json:"head_block_id"`
	HeadBlockTime             string `json:"head_block_time"`
	HeadBlockProducer         string `json:"head_block_producer"`
	VirtualBlockCPULimit      int64  `json:"virtual_block_cpu_limit"`
	VirtualBlockNetLimit      int64  `json:"virtual_block_net_limit"`
	BlockCPULimit             int64  `json:"block_cpu_limit"`
	BlockNetLimit             int64  `json:"block_net_limit"`
	ServerVersionString       string `json:"server_version_string"`
	ForkDBHeadBlockNum        int64  `json:"fork_db_head_block_num"`
	ForkDBHeadBlockID         string `json:"fork_db_head_block_id"`
	ServerFullVersionString   string `json:"server_full_version_string"`
	TotalCPUWeight            string `json:"total_cpu_weight"`
	TotalNetWeight            string `json:"total_net_weight"`
	EarliestAvailableBlockNum int64  `json:"earliest_available_block_num"`
	LastIrreversibleBlockTime string `json:"last_irreversible_block_time"`
}

type Block struct {
	Timestamp           string               `json:"timestamp"`
	Producer            string               `json:"producer"`
	Confirmed           int64                `json:"confirmed"`
	Previous            string               `json:"previous"`
	TransactionMroot    string               `json:"transaction_mroot"`
	ActionMroot         string               `json:"action_mroot"`
	ScheduleVersion     int64                `json:"schedule_version"`
	NewProducers        NewProducers         `json:"new_producers"`
	HeaderExtensions    []int64              `json:"header_extensions"`
	NewProtocolFeatures []NewProtocolFeature `json:"new_protocol_features"`
	ProducerSignature   string               `json:"producer_signature"`
	Transactions        []Transaction        `json:"transactions"`
	BlockExtensions     []int64              `json:"block_extensions"`
	ID                  string               `json:"id"`
	BlockNum            int64                `json:"block_num"`
	RefBlockPrefix      int64                `json:"ref_block_prefix"`
}

type NewProducers struct {
	Version   int64      `json:"version"`
	Producers []Producer `json:"producers"`
}

type Producer struct {
	ProducerName    string `json:"producer_name"`
	BlockSigningKey string `json:"block_signing_key"`
}

type NewProtocolFeature struct {
}

type Transaction struct {
	Status        string `json:"status"`
	CPUUsageUs    int64  `json:"cpu_usage_us"`
	NetUsageWords int64  `json:"net_usage_words"`
	Trx           Trx    `json:"trx"`
}

type Trx struct {
	Raw string
}

func (t *Trx) ID() (string, error) {
	jId := gjson.Get(t.Raw, "id")
	if jId.Exists() {
		return jId.String(), nil
	}
	return "", fmt.Errorf("Trx: 無法取得交易 ID")
}

// UnmarshalJSON 自定義解碼邏輯
func (t *Trx) UnmarshalJSON(data []byte) error {
	// 嘗試將資料解碼為字串
	t.Raw = string(data)
	return nil
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		t.Raw = str
		return nil
	}

	// 嘗試將資料解碼為物件
	// var obj map[string]interface{}
	// if err := json.Unmarshal(data, &obj); err == nil {
	// 	t.Raw = obj
	// 	return nil
	// }

	// 如果兩者都失敗，返回錯誤
	return fmt.Errorf("Trx: 無法解碼資料 %s", string(data))
}

type Account struct {
	AccountName            string              `json:"account_name"`
	HeadBlockNum           int64               `json:"head_block_num"`
	HeadBlockTime          string              `json:"head_block_time"`
	Privileged             bool                `json:"privileged"`
	LastCodeUpdate         string              `json:"last_code_update"`
	Created                string              `json:"created"`
	CoreLiquidBalance      string              `json:"core_liquid_balance"`
	RAMQuota               int64               `json:"ram_quota"`
	NetWeight              int64               `json:"net_weight"`
	CPUWeight              int64               `json:"cpu_weight"`
	NetLimit               Limit               `json:"net_limit"`
	CPULimit               Limit               `json:"cpu_limit"`
	RAMUsage               int64               `json:"ram_usage"`
	Permissions            []PermissionElement `json:"permissions"`
	TotalResources         TotalResources      `json:"total_resources"`
	SelfDelegatedBandwidth interface{}         `json:"self_delegated_bandwidth"`
	RefundRequest          interface{}         `json:"refund_request"`
	VoterInfo              VoterInfo           `json:"voter_info"`
	RexInfo                interface{}         `json:"rex_info"`
	SubjectiveCPUBillLimit Limit               `json:"subjective_cpu_bill_limit"`
	EosioAnyLinkedActions  []interface{}       `json:"eosio_any_linked_actions"`
}

type Limit struct {
	Used                int64  `json:"used"`
	Available           int64  `json:"available"`
	Max                 int64  `json:"max"`
	LastUsageUpdateTime string `json:"last_usage_update_time"`
	CurrentUsed         int64  `json:"current_used"`
}

type PermissionElement struct {
	PermName      string        `json:"perm_name"`
	Parent        string        `json:"parent"`
	RequiredAuth  RequiredAuth  `json:"required_auth"`
	LinkedActions []interface{} `json:"linked_actions"`
}

type RequiredAuth struct {
	Threshold int64         `json:"threshold"`
	Keys      []interface{} `json:"keys"`
	Accounts  []struct {
		Permission AccountPermission `json:"permission"`
		Weight     int64             `json:"weight"`
	} `json:"accounts"`
	Waits []interface{} `json:"waits"`
}

type AccountPermission struct {
	Actor      string `json:"actor"`
	Permission string `json:"permission"`
}

type TotalResources struct {
	Owner     string `json:"owner"`
	NetWeight string `json:"net_weight"`
	CPUWeight string `json:"cpu_weight"`
	RAMBytes  int64  `json:"ram_bytes"`
}

type VoterInfo struct {
	Owner             string        `json:"owner"`
	Proxy             string        `json:"proxy"`
	Producers         []interface{} `json:"producers"`
	Staked            int64         `json:"staked"`
	LastVoteWeight    string        `json:"last_vote_weight"`
	ProxiedVoteWeight string        `json:"proxied_vote_weight"`
	IsProxy           int64         `json:"is_proxy"`
	Flags1            int64         `json:"flags1"`
	Reserved2         int64         `json:"reserved2"`
	Reserved3         string        `json:"reserved3"`
}
