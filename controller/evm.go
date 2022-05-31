package controller

func (c *Controller) ShareMoney(amount string) interface{} {
	return c.evm.ShareMoney(amount)
}

func (c *Controller) PressEVM(count, tps int) interface{} {
	return c.evm.PressEVMV2(count, tps)
}
