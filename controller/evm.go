package controller

func (c *Controller) ShareMoney(amount string) interface{} {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.evm.ShareMoney(amount)
}

func (c *Controller) PressEVM(count, tps int) interface{} {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.evm.PressEVMV2(count, tps)
}
