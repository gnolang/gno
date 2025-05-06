package evm

type SmartContracts []*SmartContract

type SmartContract struct {
	ContractId int
	Code       []byte
	Creator    *Caller
}

func (s *SmartContracts) Push(c *SmartContract) int {
	c.ContractId = len(*s)
	*s = append(*s, c)
	return c.ContractId
}
